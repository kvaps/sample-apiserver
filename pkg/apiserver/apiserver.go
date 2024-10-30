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

	"github.com/aenix.io/cozystack/cozystack-api/pkg/apis/apps"
	"github.com/aenix.io/cozystack/cozystack-api/pkg/apis/apps/install"
	appsv1alpha1 "github.com/aenix.io/cozystack/cozystack-api/pkg/apis/apps/v1alpha1"
	"github.com/aenix.io/cozystack/cozystack-api/pkg/config"
	appsregistry "github.com/aenix.io/cozystack/cozystack-api/pkg/registry"
	applicationstorage "github.com/aenix.io/cozystack/cozystack-api/pkg/registry/apps/application"
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

// Config defines the config for the apiserver
type Config struct {
	GenericConfig  *genericapiserver.RecommendedConfig
	ResourceConfig *config.ResourceConfig
}

// AppsServer содержит состояние для Kubernetes master/api server.
type AppsServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

type completedConfig struct {
	GenericConfig  genericapiserver.CompletedConfig
	ResourceConfig *config.ResourceConfig
}

// CompletedConfig внедряет приватный указатель, который нельзя создать за пределами этого пакета.
type CompletedConfig struct {
	*completedConfig
}

// Complete заполняет любые поля, которые не установлены, но необходимы для корректной работы.
func (cfg *Config) Complete() CompletedConfig {
	c := completedConfig{
		cfg.GenericConfig.Complete(),
		cfg.ResourceConfig,
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

	// Динамическая регистрация типов на основе конфигурации
	err = appsv1alpha1.RegisterDynamicTypes(Scheme, c.ResourceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to register dynamic types: %v", err)
	}

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

	for _, res := range c.ResourceConfig.Resources {
		kind := res.Application.Kind
		plural := res.Application.Plural

		gvr := schema.GroupVersionResource{
			Group:    apps.GroupName,
			Version:  "v1alpha1",
			Resource: plural,
		}

		storage := applicationstorage.NewREST(dynamicClient, gvr, kind)
		v1alpha1storage[plural] = appsregistry.RESTInPeace(storage)
	}

	apiGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = v1alpha1storage

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	return s, nil
}
