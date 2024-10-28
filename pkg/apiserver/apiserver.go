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
	clientrest "k8s.io/client-go/rest" // Присваиваем псевдоним для client-go rest пакета

	"k8s.io/sample-apiserver/pkg/apis/apps"
	"k8s.io/sample-apiserver/pkg/apis/apps/install"
	appsregistry "k8s.io/sample-apiserver/pkg/registry"
	applicationstorage "k8s.io/sample-apiserver/pkg/registry/apps/application"
)

var (
	// Scheme определяет методы для сериализации и десериализации API объектов.
	Scheme = runtime.NewScheme()
	// Codecs предоставляет методы для получения кодеков и сериализаторов для определенных
	// версий и типов контента.
	Codecs            = serializer.NewCodecFactory(Scheme)
	AppsComponentName = "apps"
)

func init() {
	install.Install(Scheme)

	// Регистрация типов HelmRelease
	if err := helmv2.AddToScheme(Scheme); err != nil {
		panic(fmt.Sprintf("Failed to add HelmRelease types to scheme: %v", err))
	}

	// Регистрация unversioned типов
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	// Регистрация базовых типов
	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}

// ExtraConfig содержит пользовательскую конфигурацию apiserver.
type ExtraConfig struct {
	// Здесь можно разместить пользовательскую конфигурацию.
}

// Config определяет конфигурацию для apiserver.
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
	inClusterConfig, err := clientrest.InClusterConfig() // Используем псевдоним clientrest
	if err != nil {
		return nil, fmt.Errorf("unable to get in-cluster config: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(inClusterConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create dynamic client: %v", err)
	}

	v1alpha1storage := map[string]rest.Storage{}
	v1alpha1storage["applications"] = appsregistry.RESTInPeace(
		applicationstorage.NewREST(
			dynamicClient,
			schema.GroupVersionResource{Group: "apps.cozystack.io", Version: "v1alpha1", Resource: "applications"},
			"Application",
			Scheme,
		),
	)
	// Добавьте другие ресурсы, если необходимо

	apiGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = v1alpha1storage

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	return s, nil
}
