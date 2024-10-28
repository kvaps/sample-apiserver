/*
Copyright 2016 The Kubernetes Authors.

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

package apiserver

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/dynamic"

	"k8s.io/sample-apiserver/pkg/apis/apps"
	"k8s.io/sample-apiserver/pkg/apis/apps/install"
	appsregistry "k8s.io/sample-apiserver/pkg/registry"
	applicationstorage "k8s.io/sample-apiserver/pkg/registry/apps/application"
)

var (
	// Scheme defines methods for serializing and deserializing API objects.
	Scheme = runtime.NewScheme()
	// Codecs provides methods for retrieving codecs and serializers for specific
	// versions and content types.
	Codecs            = serializer.NewCodecFactory(Scheme)
	AppsComponentName = "apps"
)

func init() {
	install.Install(Scheme)

	// we need to add the options to empty v1
	// TODO fix the server code to avoid this
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	// TODO: keep the generic API server from wanting this
	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}

// ExtraConfig holds custom apiserver config
type ExtraConfig struct {
	// Place you custom config here.
}

// Config defines the config for the apiserver
type Config struct {
	GenericConfig *genericapiserver.RecommendedConfig
	ExtraConfig   ExtraConfig
}

// AppsServer contains state for a Kubernetes cluster master/api server.
type AppsServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *ExtraConfig
}

// CompletedConfig embeds a private pointer that cannot be instantiated outside of this package.
type CompletedConfig struct {
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (cfg *Config) Complete() CompletedConfig {
	c := completedConfig{
		cfg.GenericConfig.Complete(),
		&cfg.ExtraConfig,
	}

	return CompletedConfig{&c}
}

// New returns a new instance of AppsServer from the given config.
func (c completedConfig) New() (*AppsServer, error) {
	genericServer, err := c.GenericConfig.New("sample-apiserver", genericapiserver.NewEmptyDelegate())
	if err != nil {
		return nil, err
	}

	s := &AppsServer{
		GenericAPIServer: genericServer,
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(apps.GroupName, Scheme, metav1.ParameterCodec, Codecs)

	// Создание динамического клиента для HelmRelease
	dynamicClient, err := dynamic.NewForConfig(c.GenericConfig.LoopbackClientConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create dynamic client: %v", err)
	}

	v1alpha1storage := map[string]rest.Storage{}
	v1alpha1storage["kuberneteses"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "Kubernetes"}))
	v1alpha1storage["postgreses"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "Postgres"}))
	v1alpha1storage["redises"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "Redis"}))
	v1alpha1storage["kafkas"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "Kafka"}))
	v1alpha1storage["rabbitmqs"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "RabbitMQ"}))
	v1alpha1storage["ferretdbs"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "FerretDB"}))
	v1alpha1storage["vmdisks"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "VMDisk"}))
	v1alpha1storage["vminstances"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "VMInstance"}))

	apiGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = v1alpha1storage

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	return s, nil
}
