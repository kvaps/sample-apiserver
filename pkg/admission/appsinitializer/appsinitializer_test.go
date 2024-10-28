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

package appsinitializer_test

import (
	"context"
	"testing"
	"time"

	"k8s.io/apiserver/pkg/admission"
	"k8s.io/sample-apiserver/pkg/admission/appsinitializer"
	"k8s.io/sample-apiserver/pkg/generated/clientset/versioned/fake"
	informers "k8s.io/sample-apiserver/pkg/generated/informers/externalversions"
)

// TestWantsInternalAppsInformerFactory ensures that the informer factory is injected
// when the WantsInternalAppsInformerFactory interface is implemented by a plugin.
func TestWantsInternalAppsInformerFactory(t *testing.T) {
	cs := &fake.Clientset{}
	sf := informers.NewSharedInformerFactory(cs, time.Duration(1)*time.Second)
	target := appsinitializer.New(sf)

	wantAppsInformerFactory := &wantInternalAppsInformerFactory{}
	target.Initialize(wantAppsInformerFactory)
	if wantAppsInformerFactory.sf != sf {
		t.Errorf("expected informer factory to be initialized")
	}
}

// wantInternalAppsInformerFactory is a test stub that fulfills the WantsInternalAppsInformerFactory interface
type wantInternalAppsInformerFactory struct {
	sf informers.SharedInformerFactory
}

func (f *wantInternalAppsInformerFactory) SetInternalAppsInformerFactory(sf informers.SharedInformerFactory) {
	f.sf = sf
}
func (f *wantInternalAppsInformerFactory) Admit(ctx context.Context, a admission.Attributes, o admission.ObjectInterfaces) error {
	return nil
}
func (f *wantInternalAppsInformerFactory) Handles(o admission.Operation) bool { return false }
func (f *wantInternalAppsInformerFactory) ValidateInitialization() error      { return nil }

var _ admission.Interface = &wantInternalAppsInformerFactory{}
var _ appsinitializer.WantsInternalAppsInformerFactory = &wantInternalAppsInformerFactory{}
