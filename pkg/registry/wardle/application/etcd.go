package application

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/sample-apiserver/pkg/apis/wardle"
	"k8s.io/sample-apiserver/pkg/registry"
)

type resourceRESTOptionsGetter struct {
	delegate       generic.RESTOptionsGetter
	resourcePrefix string
}

func (r *resourceRESTOptionsGetter) GetRESTOptions(
	resource schema.GroupResource,
	obj runtime.Object,
) (generic.RESTOptions, error) {
	restOptions, err := r.delegate.GetRESTOptions(resource, obj)
	if err != nil {
		return restOptions, err
	}
	restOptions.ResourcePrefix = r.resourcePrefix
	return restOptions, nil
}

func NewREST(
	scheme *runtime.Scheme,
	optsGetter generic.RESTOptionsGetter,
	resourceName string,
	singularResourceName string,
	kindName string,
) (*registry.REST, error) {
	strategy := NewStrategy(scheme, resourceName)

	customOptsGetter := &resourceRESTOptionsGetter{
		delegate:       optsGetter,
		resourcePrefix: resourceName,
	}

	// Use the internal group version
	groupVersion := wardle.SchemeGroupVersion

	expectedGVK := groupVersion.WithKind(kindName)
	expectedListGVK := groupVersion.WithKind(kindName + "List")

	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			obj, err := scheme.New(expectedGVK)
			if err != nil {
				panic(err)
			}
			return obj
		},
		NewListFunc: func() runtime.Object {
			obj, err := scheme.New(expectedListGVK)
			if err != nil {
				panic(err)
			}
			return obj
		},
		PredicateFunc: MatchApplication,
		DefaultQualifiedResource: schema.GroupResource{
			Group:    wardle.GroupName,
			Resource: resourceName,
		},
		SingularQualifiedResource: schema.GroupResource{
			Group:    wardle.GroupName,
			Resource: singularResourceName,
		},
		CreateStrategy: strategy,
		UpdateStrategy: strategy,
		DeleteStrategy: strategy,
		TableConvertor: rest.NewDefaultTableConvertor(schema.GroupResource{
			Group:    wardle.GroupName,
			Resource: resourceName,
		}),
	}

	options := &generic.StoreOptions{
		RESTOptions: customOptsGetter,
		AttrFunc:    GetAttrs,
	}

	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}

	// Set the GVK for the discovery endpoint to the versioned GVK
	versionedGVK := schema.GroupVersion{Group: wardle.GroupName, Version: "v1alpha1"}.WithKind(kindName)

	return &registry.REST{
		Store: store,
		GVK:   versionedGVK,
	}, nil
}
