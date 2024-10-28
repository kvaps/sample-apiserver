package registry

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
)

// REST реализует RESTStorage для API-сервисов с использованием etcd
type REST struct {
	*genericregistry.Store
	GVK schema.GroupVersionKind
}

// Реализация интерфейса GroupVersionKindProvider
func (r *REST) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	return r.GVK
}

// RESTInPeace — это простая функция, которая паникает при ошибке
// В противном случае возвращает данное хранилище. Предназначена
// для оборачивания wardle хранилищ.
func RESTInPeace(storage *REST, err error) *REST {
	if err != nil {
		err = fmt.Errorf("unable to create REST storage for a resource due to %v, will die", err)
		panic(err)
	}
	return storage
}
