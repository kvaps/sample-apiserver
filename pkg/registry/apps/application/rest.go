package application

import (
	"context"
	"fmt"
	"log"
	"strings"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	unstructuredApp, err := runtime.DefaultUnstructuredConverter.ToUnstructured(convertedApp)
	if err != nil {
		log.Printf("Failed to convert Application to unstructured for resource %s: %v", convertedApp.GetName(), err)
		return nil, fmt.Errorf("failed to convert Application to unstructured: %v", err)
	}

	log.Printf("Successfully retrieved and converted resource %s of type %s to unstructured", convertedApp.GetName(), r.gvr.Resource)
	return &unstructured.Unstructured{Object: unstructuredApp}, nil
}

// Get retrieves an Application by translating it from a HelmRelease and returns it as an unstructured object.
func (r *REST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	log.Printf("Attempting to retrieve resource %s of type %s in namespace %s", name, r.gvr.Resource, "tenant-kvaps")

	// Retrieve HelmRelease as unstructured.
	hr, err := r.dynamicClient.Resource(helmReleaseGVR).Namespace("tenant-kvaps").Get(ctx, r.releaseConfig.Prefix+name, *options)
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
func (r *REST) List(ctx context.Context, options *metav1.ListOptions) (runtime.Object, error) {
	log.Printf("Attempting to list all resources of type %s in namespace %s", r.gvr.Resource, "tenant-kvaps")

	hrList, err := r.dynamicClient.Resource(helmReleaseGVR).Namespace("tenant-kvaps").List(ctx, *options)
	if err != nil {
		log.Printf("Error listing HelmReleases for resource type %s: %v", r.gvr.Resource, err)
		return nil, err
	}

	var appList appsv1alpha1.ApplicationList
	appList.Items = make([]appsv1alpha1.Application, 0, len(hrList.Items))
	for _, hr := range hrList.Items {
		app, err := r.ConvertHelmReleaseToApplication(&hr)
		if err != nil {
			log.Printf("Error converting HelmRelease to Application for resource %s: %v", hr.GetName(), err)
			continue
		}
		appList.Items = append(appList.Items, app)
	}

	log.Printf("Successfully listed all resources of type %s in namespace %s", r.gvr.Resource, "tenant-kvaps")
	return &appList, nil
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
	log.Printf("Deleting HelmRelease %s in namespace %s", name, "tenant-kvaps")

	err := r.dynamicClient.Resource(helmReleaseGVR).Namespace("tenant-kvaps").Delete(ctx, name, *options)
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

// Kind returns the resource kind used for API discovery.
func (r *REST) Kind() string {
	return r.gvk.Kind
}

// GroupVersionKind returns the GroupVersionKind for REST.
func (r *REST) GroupVersionKind(schema.GroupVersion) schema.GroupVersionKind {
	return r.gvk
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
