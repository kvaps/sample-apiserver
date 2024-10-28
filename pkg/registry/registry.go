package registry

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/sample-apiserver/pkg/registry/apps/application"
)

// REST реализует RESTStorage для API ресурсов.
type REST struct {
	*registry.Store
	GVK schema.GroupVersionKind
}

// RESTInPeace возвращает rest.Storage для Application ресурса.
func RESTInPeace(restStorage *application.REST) rest.Storage {
	return restStorage
}
