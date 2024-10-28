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

// resourceRESTOptionsGetter используется для получения REST опций
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

// NewREST создает и возвращает REST-хранилище для ресурса
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

	// Используем внутреннюю группу версии
	groupVersion := wardle.SchemeGroupVersion

	// Создаем GVK для ресурса и списка
	versionedGVK := groupVersion.WithKind(kindName)
	versionedListGVK := groupVersion.WithKind(kindName + "List")

	store := &genericregistry.Store{
		NewFunc: func() runtime.Object {
			obj, err := scheme.New(versionedGVK)
			if err != nil {
				panic(err)
			}
			return obj
		},
		NewListFunc: func() runtime.Object {
			obj, err := scheme.New(versionedListGVK)
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

	// Устанавливаем Decorator для установки TypeMeta
	store.Decorator = func(obj runtime.Object) {
		// Устанавливаем GVK для объекта
		obj.GetObjectKind().SetGroupVersionKind(versionedGVK)

		// Если объект является списком, устанавливаем GVK для списка и для каждого элемента
		if list, ok := obj.(*wardle.ApplicationList); ok {
			list.TypeMeta.Kind = kindName + "List"
			list.TypeMeta.APIVersion = versionedGVK.GroupVersion().String()
			for i := range list.Items {
				list.Items[i].TypeMeta.Kind = kindName
				list.Items[i].TypeMeta.APIVersion = versionedGVK.GroupVersion().String()
			}
		}
	}

	return &registry.REST{
		Store: store,
		GVK:   versionedGVK,
	}, nil
}
