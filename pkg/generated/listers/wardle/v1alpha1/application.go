/*
Copyright The Kubernetes Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
	wardlev1alpha1 "k8s.io/sample-apiserver/pkg/apis/wardle/v1alpha1"
)

// ApplicationLister helps list Applications.
// All objects returned here must be treated as read-only.
type ApplicationLister interface {
	// List lists all Applications in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*wardlev1alpha1.Application, err error)
	// Applications returns an object that can list and get Applications.
	Applications(namespace string) ApplicationNamespaceLister
	ApplicationListerExpansion
}

// applicationLister implements the ApplicationLister interface.
type applicationLister struct {
	listers.ResourceIndexer[*wardlev1alpha1.Application]
}

// NewApplicationLister returns a new ApplicationLister.
func NewApplicationLister(indexer cache.Indexer) ApplicationLister {
	return &applicationLister{listers.New[*wardlev1alpha1.Application](indexer, wardlev1alpha1.Resource("application"))}
}

// Applications returns an object that can list and get Applications.
func (s *applicationLister) Applications(namespace string) ApplicationNamespaceLister {
	return applicationNamespaceLister{listers.NewNamespaced[*wardlev1alpha1.Application](s.ResourceIndexer, namespace)}
}

// ApplicationNamespaceLister helps list and get Applications.
// All objects returned here must be treated as read-only.
type ApplicationNamespaceLister interface {
	// List lists all Applications in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*wardlev1alpha1.Application, err error)
	// Get retrieves the Application from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*wardlev1alpha1.Application, error)
	ApplicationNamespaceListerExpansion
}

// applicationNamespaceLister implements the ApplicationNamespaceLister
// interface.
type applicationNamespaceLister struct {
	listers.ResourceIndexer[*wardlev1alpha1.Application]
}
