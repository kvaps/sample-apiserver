package apiserver

import (
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/dynamic"
	restclient "k8s.io/client-go/rest"

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

	// Регистрация типов HelmRelease
	if err := helmv2.AddToScheme(Scheme); err != nil {
		panic(fmt.Sprintf("Failed to add HelmRelease types to scheme: %v", err))
	}

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

// AppsServer содержит состояние для Kubernetes master/api server.
type AppsServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *ExtraConfig
}

// CompletedConfig внедряет приватный указатель, который нельзя создать за пределами этого пакета.
type CompletedConfig struct {
	*completedConfig
}

// Complete заполняет любые поля, которые не установлены, но необходимы для корректной работы.
func (cfg *Config) Complete() CompletedConfig {
	c := completedConfig{
		cfg.GenericConfig.Complete(),
		&cfg.ExtraConfig,
	}

	return CompletedConfig{&c}
}

// New возвращает новый экземпляр AppsServer из данной конфигурации.
func (c completedConfig) New() (*AppsServer, error) {
	genericServer, err := c.GenericConfig.New("sample-apiserver", genericapiserver.NewEmptyDelegate())
	if err != nil {
		return nil, err
	}

	s := &AppsServer{
		GenericAPIServer: genericServer,
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(apps.GroupName, Scheme, metav1.ParameterCodec, Codecs)

	// Создание динамического клиента для HelmRelease с использованием InClusterConfig
	inClusterConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get in-cluster config: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(inClusterConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create dynamic client: %v", err)
	}

	v1alpha1storage := map[string]rest.Storage{}
	v1alpha1storage["kuberneteses"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "kuberneteses"}))
	v1alpha1storage["postgreses"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "postgreses"}))
	v1alpha1storage["redises"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "redises"}))
	v1alpha1storage["kafkas"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "kafkas"}))
	v1alpha1storage["rabbitmqs"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "rabbitmqs"}))
	v1alpha1storage["ferretdbs"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "ferretdbs"}))
	v1alpha1storage["vmdisks"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "vmdisks"}))
	v1alpha1storage["vminstances"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(dynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1alpha1", Resource: "vminstances"}))

	apiGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = v1alpha1storage

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	return s, nil
}
