package registry

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/sample-apiserver/pkg/registry/apps/application"
)

// REST implements a RESTStorage for API services against etcd
type REST struct {
	*genericregistry.Store
	GVK schema.GroupVersionKind
}

// Implement the GroupVersionKindProvider interface
func (r *REST) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	return r.GVK
}

// RESTInPeace создает REST для Application
func RESTInPeace(r *application.REST) rest.Storage {
	return r
}
