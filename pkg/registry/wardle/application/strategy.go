package application

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/sample-apiserver/pkg/apis/wardle/validation"

	"k8s.io/sample-apiserver/pkg/apis/wardle"
)

// NewStrategy создает и возвращает экземпляр applicationStrategy
func NewStrategy(typer runtime.ObjectTyper, resourceName string) applicationStrategy {
	return applicationStrategy{typer, names.SimpleNameGenerator, resourceName}
}

// GetAttrs возвращает labels.Set, fields.Set и ошибку, если переданный объект не является Application
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	apiserver, ok := obj.(*wardle.Application)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a Application")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), SelectableFields(apiserver), nil
}

// MatchApplication фильтр, используемый backend etcd для просмотра событий
func MatchApplication(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// SelectableFields возвращает набор полей, представляющих объект.
func SelectableFields(obj *wardle.Application) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

type applicationStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
	resourceName string
}

func (s applicationStrategy) NamespaceScoped() bool {
	return true
}

func (applicationStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (applicationStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (applicationStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	application := obj.(*wardle.Application)
	return validation.ValidateApplication(application)
}

// Реализуем метод ObjectKinds для соответствия интерфейсу RESTCreateStrategy, RESTUpdateStrategy и RESTDeleteStrategy
func (s applicationStrategy) ObjectKinds(obj runtime.Object) ([]schema.GroupVersionKind, bool, error) {
	return s.ObjectTyper.ObjectKinds(obj)
}

// WarningsOnCreate возвращает предупреждения для создания данного объекта.
func (applicationStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (applicationStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (applicationStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (applicationStrategy) Canonicalize(obj runtime.Object) {
}

func (applicationStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

// WarningsOnUpdate возвращает предупреждения для данного обновления.
func (applicationStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
