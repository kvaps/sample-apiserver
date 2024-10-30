package application

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/client-go/dynamic"

	appsv1alpha1 "github.com/aenix.io/cozystack/cozystack-api/pkg/apis/apps/v1alpha1"
	"github.com/aenix.io/cozystack/cozystack-api/pkg/config"
)

var helmReleaseGVR = schema.GroupVersionResource{
	Group:    "helm.toolkit.fluxcd.io",
	Version:  "v2",
	Resource: "helmreleases",
}

// REST implements the RESTStorage interface for Application.
type REST struct {
	dynamicClient dynamic.Interface
	gvr           schema.GroupVersionResource
	gvk           schema.GroupVersionKind
	kindName      string
	releaseConfig config.ReleaseConfig
}

// NewREST creates a new REST storage for Application with specific configuration.
func NewREST(dynamicClient dynamic.Interface, config *config.Resource) *REST {
	return &REST{
		dynamicClient: dynamicClient,
		gvr:           schema.GroupVersionResource{Group: appsv1alpha1.GroupName, Version: "v1alpha1", Resource: config.Application.Plural},
		gvk:           schema.GroupVersion{Group: appsv1alpha1.GroupName, Version: "v1alpha1"}.WithKind(config.Application.Kind),
		kindName:      config.Application.Kind,
		releaseConfig: config.Release,
	}
}

func (r *REST) NamespaceScoped() bool {
	return true
}

// GetSingularName implements SingularNameProvider.
func (r *REST) GetSingularName() string {
	return r.gvr.Resource
}

// Create creates a new Application by translating it into a HelmRelease.
func (r *REST) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	app, ok := obj.(*appsv1alpha1.Application)
	if !ok {
		return nil, fmt.Errorf("expected Application object, got %T", obj)
	}

	// Convert Application to HelmRelease using the configuration.
	helmRelease, err := r.ConvertApplicationToHelmRelease(app)
	if err != nil {
		return nil, fmt.Errorf("conversion error: %v", err)
	}

	// Convert HelmRelease to Unstructured.
	unstructuredHR, err := runtime.DefaultUnstructuredConverter.ToUnstructured(helmRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HelmRelease to unstructured: %v", err)
	}

	log.Printf("Creating HelmRelease %s in namespace %s", helmRelease.Name, app.Namespace)

	// Create HelmRelease using the dynamic client with the correct GVR.
	createdHR, err := r.dynamicClient.Resource(helmReleaseGVR).Namespace(app.Namespace).Create(ctx, &unstructured.Unstructured{Object: unstructuredHR}, *options)
	if err != nil {
		log.Printf("Failed to create HelmRelease %s: %v", helmRelease.Name, err)
		return nil, fmt.Errorf("failed to create HelmRelease: %v", err)
	}

	// Convert back to Application for the response.
	convertedApp, err := r.ConvertHelmReleaseToApplication(createdHR)
	if err != nil {
		log.Printf("Conversion error from HelmRelease to Application for resource %s: %v", createdHR.GetName(), err)
		return nil, fmt.Errorf("conversion error: %v", err)
	}

	log.Printf("Successfully created and converted HelmRelease %s to Application", createdHR.GetName())

	// Convert Application to unstructured.Unstructured.
	unstructuredApp, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&convertedApp)
	if err != nil {
		log.Printf("Failed to convert Application to unstructured for resource %s: %v", convertedApp.GetName(), err)
		return nil, fmt.Errorf("failed to convert Application to unstructured: %v", err)
	}

	log.Printf("Successfully retrieved and converted resource %s of type %s to unstructured", convertedApp.GetName(), r.gvr.Resource)
	return &unstructured.Unstructured{Object: unstructuredApp}, nil
}

// Get retrieves an Application by translating it from a HelmRelease and returns it as an unstructured object.
func (r *REST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	// Извлекаем неймспейс из контекста
	namespace, ok := request.NamespaceFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("namespace not found in context")
	}

	log.Printf("Attempting to retrieve resource %s of type %s in namespace %s", name, r.gvr.Resource, namespace)

	// Retrieve HelmRelease as unstructured.
	hr, err := r.dynamicClient.Resource(helmReleaseGVR).Namespace(namespace).Get(ctx, r.releaseConfig.Prefix+name, *options)
	if err != nil {
		log.Printf("Error retrieving HelmRelease for resource %s: %v", name, err)
		return nil, err
	}

	// Convert HelmRelease to Application.
	convertedApp, err := r.ConvertHelmReleaseToApplication(hr)
	if err != nil {
		log.Printf("Conversion error from HelmRelease to Application for resource %s: %v", name, err)
		return nil, fmt.Errorf("conversion error: %v", err)
	}

	// Convert Application to unstructured.Unstructured.
	unstructuredApp, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&convertedApp)
	if err != nil {
		log.Printf("Failed to convert Application to unstructured for resource %s: %v", name, err)
		return nil, fmt.Errorf("failed to convert Application to unstructured: %v", err)
	}

	log.Printf("Successfully retrieved and converted resource %s of kind %s to unstructured", name, r.gvr.Resource)
	return &unstructured.Unstructured{Object: unstructuredApp}, nil
}

// List retrieves a list of Application resources by translating them from HelmReleases.
// It filters HelmReleases based on releaseConfig.Chart.Name and releaseConfig.Chart.SourceRef,
// removes the releaseConfig.Prefix from their names, and returns them as unstructured objects.
// List реализует метод интерфейса Lister.
func (r *REST) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	// Извлекаем неймспейс из контекста
	namespace, ok := request.NamespaceFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("namespace not found in context")
	}

	log.Printf("Attempting to list all HelmReleases in namespace %s", namespace)

	// Преобразуем internalversion.ListOptions в metav1.ListOptions
	metaOptions := metav1.ListOptions{
		LabelSelector: options.LabelSelector.String(),
		FieldSelector: options.FieldSelector.String(),
		// Добавьте другие поля при необходимости
	}

	// Получаем список HelmReleases
	hrList, err := r.dynamicClient.Resource(helmReleaseGVR).Namespace(namespace).List(ctx, metaOptions)
	if err != nil {
		log.Printf("Error listing HelmReleases: %v", err)
		return nil, err
	}

	// Создаем список для хранения отфильтрованных и преобразованных Application объектов
	appList := &appsv1alpha1.ApplicationList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.cozystack.io/v1alpha1",
			Kind:       "ApplicationList",
		},
		ListMeta: metav1.ListMeta{
			ResourceVersion: hrList.GetResourceVersion(),
		},
		Items: []appsv1alpha1.Application{},
	}

	// Фильтруем и преобразуем HelmReleases в Applications
	for _, hr := range hrList.Items {
		// Фильтрация по Chart Name
		chartName, found, err := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "chart")
		if err != nil || !found {
			log.Printf("HelmRelease %s missing spec.chart.spec.chart field: %v", hr.GetName(), err)
			continue
		}
		if chartName != r.releaseConfig.Chart.Name {
			log.Printf("HelmRelease %s chart name %s does not match expected %s", hr.GetName(), chartName, r.releaseConfig.Chart.Name)
			continue
		}

		// Фильтрация по SourceRefConfig
		sourceRefKind, found, err := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "sourceRef", "kind")
		if err != nil || !found {
			log.Printf("HelmRelease %s missing spec.chart.spec.sourceRef.kind field: %v", hr.GetName(), err)
			continue
		}
		sourceRefName, found, err := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "sourceRef", "name")
		if err != nil || !found {
			log.Printf("HelmRelease %s missing spec.chart.spec.sourceRef.name field: %v", hr.GetName(), err)
			continue
		}
		sourceRefNamespace, found, err := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "sourceRef", "namespace")
		if err != nil || !found {
			log.Printf("HelmRelease %s missing spec.chart.spec.sourceRef.namespace field: %v", hr.GetName(), err)
			continue
		}
		if sourceRefKind != r.releaseConfig.Chart.SourceRef.Kind ||
			sourceRefName != r.releaseConfig.Chart.SourceRef.Name ||
			sourceRefNamespace != r.releaseConfig.Chart.SourceRef.Namespace {
			log.Printf("HelmRelease %s sourceRef does not match expected values", hr.GetName())
			continue
		}

		// Преобразуем HelmRelease в Application
		app, err := r.ConvertHelmReleaseToApplication(&hr)
		if err != nil {
			log.Printf("Error converting HelmRelease %s to Application: %v", hr.GetName(), err)
			continue
		}

		// Удаляем префикс из имени
		app.Name = strings.TrimPrefix(app.Name, r.releaseConfig.Prefix)

		// Добавляем Application в список
		appList.Items = append(appList.Items, app)
	}

	log.Printf("Successfully listed %d Application resources in namespace %s", len(appList.Items), namespace)
	return appList, nil
}

// Update updates an existing Application by translating it into a HelmRelease.
func (r *REST) Update(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, options *metav1.UpdateOptions) (runtime.Object, error) {
	app, ok := obj.(*appsv1alpha1.Application)
	if !ok {
		return nil, fmt.Errorf("expected Application object, got %T", obj)
	}

	// Convert Application to HelmRelease using the configuration.
	helmRelease, err := r.ConvertApplicationToHelmRelease(app)
	if err != nil {
		return nil, fmt.Errorf("conversion error: %v", err)
	}

	// Convert HelmRelease to Unstructured.
	unstructuredHR, err := runtime.DefaultUnstructuredConverter.ToUnstructured(helmRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HelmRelease to unstructured: %v", err)
	}

	log.Printf("Updating HelmRelease %s in namespace %s", helmRelease.Name, app.Namespace)

	// Update HelmRelease using the dynamic client with the correct GVR.
	updatedHR, err := r.dynamicClient.Resource(helmReleaseGVR).Namespace(app.Namespace).Update(ctx, &unstructured.Unstructured{Object: unstructuredHR}, *options)
	if err != nil {
		log.Printf("Failed to update HelmRelease %s: %v", helmRelease.Name, err)
		return nil, fmt.Errorf("failed to update HelmRelease: %v", err)
	}

	// Convert back to Application for the response.
	convertedApp, err := r.ConvertHelmReleaseToApplication(updatedHR)
	if err != nil {
		log.Printf("Conversion error from HelmRelease to Application for resource %s: %v", updatedHR.GetName(), err)
		return nil, fmt.Errorf("conversion error: %v", err)
	}

	log.Printf("Successfully updated and converted HelmRelease %s to Application", updatedHR.GetName())

	// Convert Application to unstructured.Unstructured.
	unstructuredApp, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&convertedApp)
	if err != nil {
		log.Printf("Failed to convert Application to unstructured for resource %s: %v", convertedApp.GetName(), err)
		return nil, fmt.Errorf("failed to convert Application to unstructured: %v", err)
	}

	log.Printf("Successfully retrieved and converted resource %s of type %s to unstructured", convertedApp.GetName(), r.gvr.Resource)
	return &unstructured.Unstructured{Object: unstructuredApp}, nil
}

// Delete deletes an Application by translating its deletion into a HelmRelease deletion.
func (r *REST) Delete(ctx context.Context, name string, options *metav1.DeleteOptions) error {
	// Извлекаем неймспейс из контекста
	namespace, ok := request.NamespaceFrom(ctx)
	if !ok {
		return fmt.Errorf("namespace not found in context")
	}

	log.Printf("Deleting HelmRelease %s in namespace %s", name, namespace)

	err := r.dynamicClient.Resource(helmReleaseGVR).Namespace(namespace).Delete(ctx, r.releaseConfig.Prefix+name, *options)
	if err != nil {
		log.Printf("Failed to delete HelmRelease %s: %v", name, err)
		return err
	}

	log.Printf("Successfully deleted HelmRelease %s", name)
	return nil
}

// Destroy releases resources associated with REST.
func (r *REST) Destroy() {
	// Empty implementation as no additional actions are needed to release resources.
}

// New creates a new instance of REST for storage.
func (r *REST) New() runtime.Object {
	return &appsv1alpha1.Application{}
}

// NewList возвращает пустой список Application объектов.
func (r *REST) NewList() runtime.Object {
	return &appsv1alpha1.ApplicationList{}
}

// Kind returns the resource kind used for API discovery.
func (r *REST) Kind() string {
	return r.gvk.Kind
}

// GroupVersionKind returns the GroupVersionKind for REST.
func (r *REST) GroupVersionKind(schema.GroupVersion) schema.GroupVersionKind {
	return r.gvk
}

// ConvertToTable реализует интерфейс TableConvertor для вашей структуры REST.
// Он создает metav1.Table с колонками NAME, VERSION, READY, AGE и заполняет их соответствующими данными.
func (r *REST) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	log.Printf("ConvertToTable: received object of type %T", object)

	var table metav1.Table

	switch obj := object.(type) {
	case *appsv1alpha1.ApplicationList:
		// Define table columns
		table.ColumnDefinitions = []metav1.TableColumnDefinition{
			{Name: "NAME", Type: "string", Description: "Name of the Application", Priority: 0},
			{Name: "VERSION", Type: "string", Description: "Version of the Application", Priority: 0},
			{Name: "READY", Type: "boolean", Description: "Ready status of the Application", Priority: 0},
			{Name: "AGE", Type: "string", Description: "Age of the Application", Priority: 0},
		}
		table.Rows = make([]metav1.TableRow, 0, len(obj.Items))
		now := time.Now()

		for _, app := range obj.Items {
			name := app.GetName()
			version := app.Status.Version
			if version == "" {
				version = "<unknown>"
			}

			ready := false
			for _, condition := range app.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == metav1.ConditionTrue {
					ready = true
					break
				}
			}

			age := computeAge(app.GetCreationTimestamp().Time, now)

			row := metav1.TableRow{
				Cells:  []interface{}{name, version, ready, age},
				Object: runtime.RawExtension{Object: &app},
			}
			table.Rows = append(table.Rows, row)
		}

		table.ListMeta = metav1.ListMeta{
			ResourceVersion: obj.GetResourceVersion(),
		}

	case *appsv1alpha1.Application:
		// Define table columns
		table.ColumnDefinitions = []metav1.TableColumnDefinition{
			{Name: "NAME", Type: "string", Description: "Name of the Application", Priority: 0},
			{Name: "VERSION", Type: "string", Description: "Version of the Application", Priority: 0},
			{Name: "READY", Type: "boolean", Description: "Ready status of the Application", Priority: 0},
			{Name: "AGE", Type: "string", Description: "Age of the Application", Priority: 0},
		}
		table.Rows = []metav1.TableRow{}
		now := time.Now()

		name := obj.GetName()
		version := obj.Status.Version
		if version == "" {
			version = "<unknown>"
		}

		ready := false
		for _, condition := range obj.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == metav1.ConditionTrue {
				ready = true
				break
			}
		}

		age := computeAge(obj.GetCreationTimestamp().Time, now)

		row := metav1.TableRow{
			Cells:  []interface{}{name, version, ready, age},
			Object: runtime.RawExtension{Object: obj},
		}
		table.Rows = append(table.Rows, row)

		table.ListMeta = metav1.ListMeta{
			ResourceVersion: obj.GetResourceVersion(),
		}

	default:
		// Return an error if object type is not supported
		resource := schema.GroupResource{}
		if info, ok := request.RequestInfoFrom(ctx); ok {
			resource = schema.GroupResource{Group: info.APIGroup, Resource: info.Resource}
		}
		return nil, errNotAcceptable{
			resource: resource,
			message:  "object does not implement the Object interfaces",
		}
	}

	// Handle table options
	if opt, ok := tableOptions.(*metav1.TableOptions); ok && opt != nil && opt.NoHeaders {
		table.ColumnDefinitions = nil
	}

	// Set TypeMeta
	table.TypeMeta = metav1.TypeMeta{
		APIVersion: "meta.k8s.io/v1",
		Kind:       "Table",
	}

	log.Printf("ConvertToTable: returning table with %d rows", len(table.Rows))

	return &table, nil
}

// computeAge вычисляет возраст объекта на основе CreationTimestamp и текущего времени.
func computeAge(creationTime, currentTime time.Time) string {
	duration := currentTime.Sub(creationTime)
	return duration.Round(time.Minute).String()
}

// ConvertHelmReleaseToApplication converts a HelmRelease to an Application using the configuration.
func (r *REST) ConvertHelmReleaseToApplication(hr *unstructured.Unstructured) (appsv1alpha1.Application, error) {
	log.Printf("Converting HelmRelease to Application for resource %s", hr.GetName())

	var helmRelease helmv2.HelmRelease
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(hr.Object, &helmRelease)
	if err != nil {
		log.Printf("Error converting from unstructured to HelmRelease: %v", err)
		return appsv1alpha1.Application{}, err
	}

	app, err := r.convertHelmReleaseToApplication(&helmRelease)
	if err != nil {
		log.Printf("Error converting from HelmRelease to Application: %v", err)
		return app, err
	}

	log.Printf("Successfully converted HelmRelease %s to Application", hr.GetName())
	return app, nil
}

// ConvertApplicationToHelmRelease converts an Application to a HelmRelease using the configuration.
func (r *REST) ConvertApplicationToHelmRelease(app *appsv1alpha1.Application) (*helmv2.HelmRelease, error) {
	helmRelease, err := r.convertApplicationToHelmRelease(app)
	if err != nil {
		return nil, err
	}
	return helmRelease, nil
}

// convertHelmReleaseToApplication implements the actual conversion logic using the configuration.
func (r *REST) convertHelmReleaseToApplication(hr *helmv2.HelmRelease) (appsv1alpha1.Application, error) {
	app := appsv1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.cozystack.io/v1alpha1",
			Kind:       r.kindName,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              strings.TrimPrefix(hr.Name, r.releaseConfig.Prefix),
			Namespace:         hr.Namespace,
			CreationTimestamp: hr.CreationTimestamp,
			DeletionTimestamp: hr.DeletionTimestamp,
		},
		Spec: hr.Spec.Values,
		Status: appsv1alpha1.ApplicationStatus{
			Version: hr.Status.LastAttemptedRevision,
		},
	}

	conditions := []metav1.Condition{}
	for _, hrCondition := range hr.GetConditions() {
		if hrCondition.Type == "Ready" || hrCondition.Type == "Released" {
			conditions = append(conditions, metav1.Condition{
				LastTransitionTime: hrCondition.LastTransitionTime,
				Reason:             hrCondition.Reason,
				Message:            hrCondition.Message,
				Status:             hrCondition.Status,
				Type:               hrCondition.Type,
			})
		}
	}
	app.SetConditions(conditions)
	return app, nil
}

// convertApplicationToHelmRelease implements the actual conversion logic using the configuration.
func (r *REST) convertApplicationToHelmRelease(app *appsv1alpha1.Application) (*helmv2.HelmRelease, error) {
	helmRelease := &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "helm.toolkit.fluxcd.io/v2",
			Kind:       "HelmRelease",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.releaseConfig.Prefix + app.Name,
			Namespace: app.Namespace,
			Labels:    r.releaseConfig.Labels,
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: &helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:             r.releaseConfig.Chart.Name,
					Version:           *&app.AppVersion,
					ReconcileStrategy: "Revision",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      r.releaseConfig.Chart.SourceRef.Kind,
						Name:      r.releaseConfig.Chart.SourceRef.Name,
						Namespace: r.releaseConfig.Chart.SourceRef.Namespace,
					},
				},
			},
			Values: app.Spec,
		},
	}

	return helmRelease, nil
}

// errNotAcceptable указывает, что ресурс не поддерживает конвертацию в Table
type errNotAcceptable struct {
	resource schema.GroupResource
	message  string
}

func (e errNotAcceptable) Error() string {
	return e.message
}

func (e errNotAcceptable) Status() metav1.Status {
	return metav1.Status{
		Status:  metav1.StatusFailure,
		Code:    http.StatusNotAcceptable,
		Reason:  metav1.StatusReason("NotAcceptable"),
		Message: e.Error(),
	}
}
