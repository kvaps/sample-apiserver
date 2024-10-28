package application

import (
	"context"
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/client-go/dynamic"
	appsv1alpha1 "k8s.io/sample-apiserver/pkg/apis/apps/v1alpha1"
	"k8s.io/sample-apiserver/pkg/conversion"
)

// REST реализует rest.Storage для Application ресурсов
type REST struct {
	dynamicClient dynamic.Interface
	gvr           schema.GroupVersionResource
}

// NewREST создает новое REST хранилище для Application
func NewREST(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource) *REST {
	return &REST{
		dynamicClient: dynamicClient,
		gvr:           gvr,
	}
}

func (r *REST) NamespaceScoped() bool {
	// Верните true, если ресурсы должны быть связаны с namespace
	return true
}

// SingularName возвращает единственное число для ресурса.
// TODO: automate this
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
	helmRelease, err := ConvertApplicationToHelmRelease(app)
	if err != nil {
		return nil, fmt.Errorf("conversion error: %v", err)
	}

	// Преобразование HelmRelease в Unstructured
	unstructuredHR, err := runtime.DefaultUnstructuredConverter.ToUnstructured(helmRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HelmRelease to unstructured: %v", err)
	}

	// Создание HelmRelease через динамический клиент
	createdHR, err := r.dynamicClient.Resource(r.gvr).Namespace(app.Namespace).Create(ctx, &unstructured.Unstructured{Object: unstructuredHR}, *options)
	if err != nil {
		return nil, fmt.Errorf("failed to create HelmRelease: %v", err)
	}

	// Конвертация обратно в Application для ответа
	convertedApp := &appsv1alpha1.Application{}
	err = ConvertHelmReleaseToApplication(createdHR, convertedApp)
	if err != nil {
		return nil, fmt.Errorf("conversion error: %v", err)
	}

	return convertedApp, nil
}

// Get получает Application, транслируя его из HelmRelease
func (r *REST) Get(ctx context.Context, name string, options *v1.GetOptions) (runtime.Object, error) {
	hr, err := r.dynamicClient.Resource(r.gvr).Namespace("tenant-kvaps").Get(ctx, name, *options)
	if err != nil {
		return nil, err
	}

	var app appsv1alpha1.Application
	err = ConvertHelmReleaseToApplication(hr, &app)
	if err != nil {
		return nil, fmt.Errorf("conversion error: %v", err)
	}

	return &app, nil
}

// List получает список Application ресурсов, транслируя их из HelmRelease
func (r *REST) List(ctx context.Context, options *v1.ListOptions) (runtime.Object, error) {
	hrList, err := r.dynamicClient.Resource(r.gvr).Namespace("tenant-kvaps").List(ctx, *options)
	if err != nil {
		return nil, err
	}

	var appList appsv1alpha1.ApplicationList
	appList.Items = make([]appsv1alpha1.Application, 0, len(hrList.Items))
	for _, hr := range hrList.Items {
		var app appsv1alpha1.Application
		err := ConvertHelmReleaseToApplication(&hr, &app)
		if err != nil {
			continue
		}
		appList.Items = append(appList.Items, app)
	}

	return &appList, nil
}

// Update обновляет существующий Application, транслируя его в HelmRelease
func (r *REST) Update(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, options *v1.UpdateOptions) (runtime.Object, error) {
	app, ok := obj.(*appsv1alpha1.Application)
	if !ok {
		return nil, fmt.Errorf("expected Application object, got %T", obj)
	}

	helmRelease, err := ConvertApplicationToHelmRelease(app)
	if err != nil {
		return nil, fmt.Errorf("conversion error: %v", err)
	}

	unstructuredHR, err := runtime.DefaultUnstructuredConverter.ToUnstructured(helmRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HelmRelease to unstructured: %v", err)
	}

	updatedHR, err := r.dynamicClient.Resource(r.gvr).Namespace(app.Namespace).Update(ctx, &unstructured.Unstructured{Object: unstructuredHR}, *options)
	if err != nil {
		return nil, fmt.Errorf("failed to update HelmRelease: %v", err)
	}

	convertedApp := &appsv1alpha1.Application{}
	err = ConvertHelmReleaseToApplication(updatedHR, convertedApp)
	if err != nil {
		return nil, fmt.Errorf("conversion error: %v", err)
	}

	return convertedApp, nil
}

// Delete удаляет Application, транслируя его удаление в HelmRelease
func (r *REST) Delete(ctx context.Context, name string, options *v1.DeleteOptions) error {
	return r.dynamicClient.Resource(r.gvr).Namespace("tenant-kvaps").Delete(ctx, name, *options)
}

// Destroy освобождает ресурсы, связанные с REST.
func (r *REST) Destroy() {
	// Пустая реализация метода, так как нет необходимости в дополнительных действиях для освобождения ресурсов
}

// New создает новый экземпляр REST для хранения.
func (r *REST) New() runtime.Object {
	return &appsv1alpha1.Application{}
}

// ConvertHelmReleaseToApplication конвертирует HelmRelease в Application
func ConvertHelmReleaseToApplication(hr *unstructured.Unstructured, app *appsv1alpha1.Application) error {
	var helmRelease helmv2.HelmRelease
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(hr.Object, &helmRelease)
	if err != nil {
		return err
	}

	convertedApp, err := conversion.ConvertHelmReleaseToApplication(&helmRelease)
	if err != nil {
		return err
	}

	*app = *convertedApp
	return nil
}

// ConvertApplicationToHelmRelease конвертирует Application в HelmRelease
func ConvertApplicationToHelmRelease(app *appsv1alpha1.Application) (*helmv2.HelmRelease, error) {
	helmRelease, err := conversion.ConvertApplicationToHelmRelease(app)
	if err != nil {
		return nil, err
	}
	return helmRelease, nil
}
