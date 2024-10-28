/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package application

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/sample-apiserver/pkg/apis/apps/validation"

	"k8s.io/sample-apiserver/pkg/apis/apps"
)

// NewStrategy creates and returns an applicationStrategy instance
func NewStrategy(typer runtime.ObjectTyper, resourceName string) applicationStrategy {
	return applicationStrategy{typer, names.SimpleNameGenerator, resourceName}
}

// GetAttrs returns labels.Set, fields.Set, and error in case the given runtime.Object is not a Application
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	apiserver, ok := obj.(*apps.Application)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a Application")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), SelectableFields(apiserver), nil
}

// MatchApplication is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchApplication(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// SelectableFields returns a field set that represents the object.
func SelectableFields(obj *apps.Application) fields.Set {
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
	application := obj.(*apps.Application)
	return validation.ValidateApplication(application)
}

// WarningsOnCreate returns warnings for the creation of the given object.
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

// WarningsOnUpdate returns warnings for the given update.
func (applicationStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}
