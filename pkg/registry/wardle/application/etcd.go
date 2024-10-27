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

// Custom RESTOptionsGetter that sets the ResourcePrefix
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

// NewREST returns a RESTStorage object that will work against API services.
func NewREST(
	scheme *runtime.Scheme,
	optsGetter generic.RESTOptionsGetter,
	resourceName string,
	singularResourceName string,
) (*registry.REST, error) {
	strategy := NewStrategy(scheme, resourceName)

	customOptsGetter := &resourceRESTOptionsGetter{
		delegate:       optsGetter,
		resourcePrefix: resourceName,
	}

	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			return &wardle.Application{}
		},
		NewListFunc: func() runtime.Object {
			return &wardle.ApplicationList{}
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

	return &registry.REST{Store: store}, nil
}
