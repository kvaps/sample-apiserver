package application

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/client-go/dynamic"
	"k8s.io/sample-apiserver/pkg/apis/apps/v1alpha1"
	"k8s.io/sample-apiserver/pkg/conversion"
)

var helmReleaseGVR = schema.GroupVersionResource{
	Group:    "helm.toolkit.fluxcd.io",
	Version:  "v2",
	Resource: "helmreleases",
}

type REST struct {
	*registry.Store
	dynamicClient dynamic.Interface
	gvr           schema.GroupVersionResource
	kindName      string
	GVK           schema.GroupVersionKind
}

func NewREST(dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, kindName string, scheme *runtime.Scheme) *REST {
	store := &registry.Store{
		NewFunc:     func() runtime.Object { return &v1alpha1.Application{} },
		NewListFunc: func() runtime.Object { return &v1alpha1.ApplicationList{} },
		KeyFunc: func(ctx context.Context, name string) (string, error) {
			return name, nil
		},
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			meta, ok := obj.(metav1.Object)
			if !ok {
				return "", fmt.Errorf("object is not of type metav1.Object")
			}
			return meta.GetName(), nil
		},
		PredicateFunc: func(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
			return storage.SelectionPredicate{Label: label, Field: field}
		},
		StorageVersioner: schema.GroupVersion{Group: gvr.Group, Version: gvr.Version},
	}

	return &REST{
		Store:         store,
		dynamicClient: dynamicClient,
		gvr:           gvr,
		kindName:      kindName,
		GVK: schema.GroupVersionKind{
			Group:   gvr.Group,
			Version: gvr.Version,
			Kind:    kindName,
		},
	}
}

// NamespaceScoped указывает, что ресурс является пространственно ограниченным
func (r *REST) NamespaceScoped() bool {
	return true
}

// New создает новый экземпляр REST для хранения.
func (r *REST) New() runtime.Object {
	return &v1alpha1.Application{}
}

// Create создает новый Application, транслируя его в HelmRelease
func (r *REST) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, bool, error) {
	app, ok := obj.(*v1alpha1.Application)
	if !ok {
		return nil, false, fmt.Errorf("expected Application object, got %T", obj)
	}

	helmRelease, err := conversion.ConvertApplicationToHelmRelease(app)
	if err != nil {
		return nil, false, fmt.Errorf("conversion error: %v", err)
	}

	unstructuredHR, err := runtime.DefaultUnstructuredConverter.ToUnstructured(helmRelease)
	if err != nil {
		return nil, false, fmt.Errorf("failed to convert HelmRelease to unstructured: %v", err)
	}

	createdHR, err := r.dynamicClient.Resource(helmReleaseGVR).Namespace(app.Namespace).Create(ctx, &unstructured.Unstructured{Object: unstructuredHR}, *options)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create HelmRelease: %v", err)
	}

	createdHelmRelease, err := conversion.ConvertUnstructuredToHelmRelease(createdHR)
	if err != nil {
		return nil, false, fmt.Errorf("conversion error: %v", err)
	}

	convertedApp, err := conversion.ConvertHelmReleaseToApplication(createdHelmRelease)
	if err != nil {
		return nil, false, fmt.Errorf("conversion error: %v", err)
	}

	return convertedApp, true, nil
}

// Update обновляет объект Application
func (r *REST) Update(ctx context.Context, name string, obj runtime.Object, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	app, ok := obj.(*v1alpha1.Application)
	if !ok {
		return nil, false, fmt.Errorf("expected Application object, got %T", obj)
	}

	helmRelease, err := conversion.ConvertApplicationToHelmRelease(app)
	if err != nil {
		return nil, false, fmt.Errorf("conversion error: %v", err)
	}

	unstructuredHR, err := runtime.DefaultUnstructuredConverter.ToUnstructured(helmRelease)
	if err != nil {
		return nil, false, fmt.Errorf("failed to convert HelmRelease to unstructured: %v", err)
	}

	updatedHR, err := r.dynamicClient.Resource(helmReleaseGVR).Namespace(app.Namespace).Update(ctx, &unstructured.Unstructured{Object: unstructuredHR}, *options)
	if err != nil {
		return nil, false, fmt.Errorf("failed to update HelmRelease: %v", err)
	}

	updatedHelmRelease, err := conversion.ConvertUnstructuredToHelmRelease(updatedHR)
	if err != nil {
		return nil, false, fmt.Errorf("conversion error: %v", err)
	}

	convertedApp, err := conversion.ConvertHelmReleaseToApplication(updatedHelmRelease)
	if err != nil {
		return nil, false, fmt.Errorf("conversion error: %v", err)
	}

	return convertedApp, true, nil
}

// Delete удаляет объект Application по имени
func (r *REST) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	err := r.dynamicClient.Resource(helmReleaseGVR).Namespace(r.gvr.Resource).Delete(ctx, name, *options)
	if err != nil {
		return nil, false, fmt.Errorf("failed to delete HelmRelease: %v", err)
	}
	return &metav1.Status{Status: metav1.StatusSuccess}, true, nil
}
