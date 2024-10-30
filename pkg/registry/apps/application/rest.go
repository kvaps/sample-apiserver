package application

import (
	"context"
	"fmt"
	"log"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/client-go/dynamic"
	"github.com/aenix.io/cozystack/cozystack-api/pkg/apis/apps"
	appsv1alpha1 "github.com/aenix.io/cozystack/cozystack-api/pkg/apis/apps/v1alpha1"
	"github.com/aenix.io/cozystack/cozystack-api/pkg/conversion"
)

var helmReleaseGVR = schema.GroupVersionResource{
	Group:    "helm.toolkit.fluxcd.io",
	Version:  "v2",
	Resource: "helmreleases",
}

type REST struct {
	dynamicClient dynamic.Interface
	gvr           schema.GroupVersionResource
	kindName      string
	GVK           schema.GroupVersionKind
}

func NewREST(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, kindName string) *REST {
	return &REST{
		dynamicClient: dynamicClient,
		gvr:           gvr,
		kindName:      kindName,
		GVK:           schema.GroupVersion{Group: apps.GroupName, Version: "v1alpha1"}.WithKind(kindName),
	}
}

func (r *REST) NamespaceScoped() bool {
	return true
}

// GetSingularName реализует SingularNameProvider
func (r *REST) GetSingularName() string {
	return r.gvr.Resource
}

// Create создает новый Application, транслируя его в HelmRelease
func (r *REST) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *v1.CreateOptions) (runtime.Object, error) {
	app, ok := obj.(*appsv1alpha1.Application)
	if !ok {
		return nil, fmt.Errorf("expected Application object, got %T", obj)
	}

	// Конвертация Application в HelmRelease
	helmRelease, err := r.ConvertApplicationToHelmRelease(app)
	if err != nil {
		return nil, fmt.Errorf("conversion error: %v", err)
	}

	// Преобразование HelmRelease в Unstructured
	unstructuredHR, err := runtime.DefaultUnstructuredConverter.ToUnstructured(helmRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HelmRelease to unstructured: %v", err)
	}

	log.Printf("Creating HelmRelease %s in namespace %s", helmRelease.Name, app.Namespace)

	// Создание HelmRelease через динамический клиент с использованием правильного GVR
	createdHR, err := r.dynamicClient.Resource(helmReleaseGVR).Namespace(app.Namespace).Create(ctx, &unstructured.Unstructured{Object: unstructuredHR}, *options)
	if err != nil {
		log.Printf("Failed to create HelmRelease %s: %v", helmRelease.Name, err)
		return nil, fmt.Errorf("failed to create HelmRelease: %v", err)
	}

	// Конвертация обратно в Application для ответа
	convertedApp := &appsv1alpha1.Application{}
	err = r.ConvertHelmReleaseToApplication(createdHR, convertedApp)
	if err != nil {
		log.Printf("Conversion error from HelmRelease to Application for resource %s: %v", createdHR.GetName(), err)
		return nil, fmt.Errorf("conversion error: %v", err)
	}

	log.Printf("Successfully created and converted HelmRelease %s to Application", createdHR.GetName())
	return convertedApp, nil
}

// Get получает Application, транслируя его из HelmRelease, и возвращает как unstructured объект
func (r *REST) Get(ctx context.Context, name string, options *v1.GetOptions) (runtime.Object, error) {
	log.Printf("Attempting to retrieve resource %s of kind %s in namespace tenant-kvaps", name, r.gvr.Resource)

	// Получаем HelmRelease как unstructured
	hr, err := r.dynamicClient.Resource(helmReleaseGVR).Namespace("tenant-kvaps").Get(ctx, name, *options)
	if err != nil {
		log.Printf("Error retrieving HelmRelease for resource %s: %v", name, err)
		return nil, err
	}

	// Конвертация HelmRelease в Application
	var app appsv1alpha1.Application
	err = r.ConvertHelmReleaseToApplication(hr, &app)
	if err != nil {
		log.Printf("Conversion error from HelmRelease to Application for resource %s: %v", name, err)
		return nil, fmt.Errorf("conversion error: %v", err)
	}

	// Конвертация Application в unstructured.Unstructured
	unstructuredApp, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&app)
	if err != nil {
		log.Printf("Failed to convert Application to unstructured for resource %s: %v", name, err)
		return nil, fmt.Errorf("failed to convert Application to unstructured: %v", err)
	}

	log.Printf("Successfully retrieved and converted resource %s of kind %s to unstructured", name, r.gvr.Resource)
	return &unstructured.Unstructured{Object: unstructuredApp}, nil
}

// List получает список Application ресурсов, транслируя их из HelmRelease
func (r *REST) List(ctx context.Context, options *v1.ListOptions) (runtime.Object, error) {
	log.Printf("Attempting to list all resources of kind %s in namespace tenant-kvaps", r.gvr.Resource)

	hrList, err := r.dynamicClient.Resource(helmReleaseGVR).Namespace("tenant-kvaps").List(ctx, *options)
	if err != nil {
		log.Printf("Error listing HelmReleases for resource kind %s: %v", r.gvr.Resource, err)
		return nil, err
	}

	var appList appsv1alpha1.ApplicationList
	appList.Items = make([]appsv1alpha1.Application, 0, len(hrList.Items))
	for _, hr := range hrList.Items {
		var app appsv1alpha1.Application
		err := r.ConvertHelmReleaseToApplication(&hr, &app)
		if err != nil {
			log.Printf("Error converting HelmRelease to Application for resource %s: %v", hr.GetName(), err)
			continue
		}
		appList.Items = append(appList.Items, app)
	}

	log.Printf("Successfully listed all resources of kind %s in namespace tenant-kvaps", r.gvr.Resource)
	return &appList, nil
}

// Update обновляет существующий Application, транслируя его в HelmRelease
func (r *REST) Update(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, options *v1.UpdateOptions) (runtime.Object, error) {
	app, ok := obj.(*appsv1alpha1.Application)
	if !ok {
		return nil, fmt.Errorf("expected Application object, got %T", obj)
	}

	helmRelease, err := r.ConvertApplicationToHelmRelease(app)
	if err != nil {
		return nil, fmt.Errorf("conversion error: %v", err)
	}

	unstructuredHR, err := runtime.DefaultUnstructuredConverter.ToUnstructured(helmRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HelmRelease to unstructured: %v", err)
	}

	log.Printf("Updating HelmRelease %s in namespace %s", helmRelease.Name, app.Namespace)

	// Обновление HelmRelease через динамический клиент с использованием правильного GVR
	updatedHR, err := r.dynamicClient.Resource(helmReleaseGVR).Namespace(app.Namespace).Update(ctx, &unstructured.Unstructured{Object: unstructuredHR}, *options)
	if err != nil {
		log.Printf("Failed to update HelmRelease %s: %v", helmRelease.Name, err)
		return nil, fmt.Errorf("failed to update HelmRelease: %v", err)
	}

	// Конвертация обратно в Application для ответа
	convertedApp := &appsv1alpha1.Application{}
	err = r.ConvertHelmReleaseToApplication(updatedHR, convertedApp)
	if err != nil {
		log.Printf("Conversion error from HelmRelease to Application for resource %s: %v", updatedHR.GetName(), err)
		return nil, fmt.Errorf("conversion error: %v", err)
	}

	log.Printf("Successfully updated and converted HelmRelease %s to Application", updatedHR.GetName())
	return convertedApp, nil
}

// Delete удаляет Application, транслируя его удаление в HelmRelease
func (r *REST) Delete(ctx context.Context, name string, options *v1.DeleteOptions) error {
	log.Printf("Deleting HelmRelease %s in namespace tenant-kvaps", name)

	err := r.dynamicClient.Resource(helmReleaseGVR).Namespace("tenant-kvaps").Delete(ctx, name, *options)
	if err != nil {
		log.Printf("Failed to delete HelmRelease %s: %v", name, err)
		return err
	}

	log.Printf("Successfully deleted HelmRelease %s", name)
	return nil
}

// Destroy освобождает ресурсы, связанные с REST.
func (r *REST) Destroy() {
	// Пустая реализация метода, так как нет необходимости в дополнительных действиях для освобождения ресурсов
}

// New создает новый экземпляр REST для хранения.
func (r *REST) New() runtime.Object {
	return &appsv1alpha1.Application{}
}

// Kind возвращает вид ресурса, который используется для обнаружения API
func (r *REST) Kind() string {
	return r.GVK.Kind
}

func (r *REST) GroupVersionKind(schema.GroupVersion) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   r.GVK.Group,
		Version: r.GVK.Version,
		Kind:    r.kindName,
	}
}

// ConvertHelmReleaseToApplication конвертирует HelmRelease в Application
func (r *REST) ConvertHelmReleaseToApplication(hr *unstructured.Unstructured, app *appsv1alpha1.Application) error {
	log.Printf("Converting HelmRelease to Application for resource %s", hr.GetName())

	var helmRelease helmv2.HelmRelease
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(hr.Object, &helmRelease)
	if err != nil {
		log.Printf("Error in conversion from unstructured to HelmRelease: %v", err)
		return err
	}

	convertedApp, err := conversion.ConvertHelmReleaseToApplication(&helmRelease)
	if err != nil {
		log.Printf("Error in conversion from HelmRelease to Application struct: %v", err)
		return err
	}

	*app = *convertedApp
	log.Printf("Successfully converted HelmRelease %s to Application", hr.GetName())
	return nil
}

// ConvertApplicationToHelmRelease конвертирует Application в HelmRelease
func (r *REST) ConvertApplicationToHelmRelease(app *appsv1alpha1.Application) (*helmv2.HelmRelease, error) {
	helmRelease, err := conversion.ConvertApplicationToHelmRelease(app)
	if err != nil {
		return nil, err
	}
	return helmRelease, nil
}
