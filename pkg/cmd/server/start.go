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

package server

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/aenix.io/cozystack/cozystack-api/pkg/apis/apps/v1alpha1"
	"github.com/aenix.io/cozystack/cozystack-api/pkg/apiserver"
	"github.com/aenix.io/cozystack/cozystack-api/pkg/config"
	sampleopenapi "github.com/aenix.io/cozystack/cozystack-api/pkg/generated/openapi"
	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/apiserver/pkg/endpoints/openapi"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	utilversionpkg "k8s.io/apiserver/pkg/util/version"
	"k8s.io/component-base/featuregate"
	baseversion "k8s.io/component-base/version"
	netutils "k8s.io/utils/net"
)

// AppsServerOptions содержит состояние для master/api server
type AppsServerOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions

	StdOut io.Writer
	StdErr io.Writer

	AlternateDNS []string

	// Добавляем поле для хранения пути к конфигурации
	ResourceConfigPath string

	// Добавляем поле для хранения конфигурации
	ResourceConfig *config.ResourceConfig
}

// NewAppsServerOptions возвращает новый экземпляр AppsServerOptions
func NewAppsServerOptions(out, errOut io.Writer) *AppsServerOptions {
	o := &AppsServerOptions{
		RecommendedOptions: genericoptions.NewRecommendedOptions(
			"",
			apiserver.Codecs.LegacyCodec(v1alpha1.SchemeGroupVersion),
		),

		StdOut: out,
		StdErr: errOut,
	}
	o.RecommendedOptions.Etcd = nil
	return o
}

// NewCommandStartAppsServer предоставляет обработчик CLI для команды 'start apps-server'
func NewCommandStartAppsServer(ctx context.Context, defaults *AppsServerOptions) *cobra.Command {
	o := *defaults
	cmd := &cobra.Command{
		Short: "Launch an Apps API server",
		Long:  "Launch an Apps API server",
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return utilversionpkg.DefaultComponentGlobalsRegistry.Set()
		},
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.RunAppsServer(c.Context()); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.SetContext(ctx)

	flags := cmd.Flags()
	o.RecommendedOptions.AddFlags(flags)

	// Добавляем флаг для пути к конфигу
	flags.StringVar(&o.ResourceConfigPath, "config", "config.yaml", "Path to the resource configuration file")

	// Следующие строки демонстрируют, как настроить совместимость версий и feature gates
	// для компонента "Apps" в соответствии с KEP-4330.

	// Создаём объект версии по умолчанию для компонента "Apps".
	defaultAppsVersion := "1.1"
	// Регистрируем компонент "Apps" в глобальном реестре компонентов,
	// связывая его с эффективной версией и конфигурацией feature gates.
	_, appsFeatureGate := utilversionpkg.DefaultComponentGlobalsRegistry.ComponentGlobalsOrRegister(
		apiserver.AppsComponentName, utilversionpkg.NewEffectiveVersion(defaultAppsVersion),
		featuregate.NewVersionedFeatureGate(version.MustParse(defaultAppsVersion)),
	)

	// Добавляем спецификации feature gates для компонента "Apps".
	utilruntime.Must(appsFeatureGate.AddVersioned(map[featuregate.Feature]featuregate.VersionedSpecs{
		// Пример добавления feature gates:
		// "FeatureName": {{"v1", true}, {"v2", false}},
	}))

	// Регистрируем стандартный kube компонент, если он ещё не зарегистрирован в глобальном реестре.
	_, _ = utilversionpkg.DefaultComponentGlobalsRegistry.ComponentGlobalsOrRegister(
		utilversionpkg.DefaultKubeComponent,
		utilversionpkg.NewEffectiveVersion(baseversion.DefaultKubeBinaryVersion),
		utilfeature.DefaultMutableFeatureGate,
	)

	// Устанавливаем сопоставление эмуляции версии от компонента "Apps" к kube компоненту.
	utilruntime.Must(utilversionpkg.DefaultComponentGlobalsRegistry.SetEmulationVersionMapping(
		apiserver.AppsComponentName, utilversionpkg.DefaultKubeComponent, AppsVersionToKubeVersion,
	))

	// Добавляем флаги из глобального реестра компонентов.
	utilversionpkg.DefaultComponentGlobalsRegistry.AddFlags(flags)

	return cmd
}

// Complete заполняет поля, которые не установлены
func (o *AppsServerOptions) Complete() error {
	// Загружаем конфигурационный файл
	cfg, err := config.LoadConfig(o.ResourceConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config from %s: %v", o.ResourceConfigPath, err)
	}
	o.ResourceConfig = cfg
	return nil
}

// Validate проверяет правильность опций
func (o AppsServerOptions) Validate(args []string) error {
	var allErrors []error
	allErrors = append(allErrors, o.RecommendedOptions.Validate()...)
	allErrors = append(allErrors, utilversionpkg.DefaultComponentGlobalsRegistry.Validate()...)
	return utilerrors.NewAggregate(allErrors)
}

// Config возвращает конфигурацию для API сервера на основе AppsServerOptions
func (o *AppsServerOptions) Config() (*apiserver.Config, error) {
	// TODO: установите "реальный" внешний адрес
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts(
		"localhost", o.AlternateDNS, []net.IP{netutils.ParseIPSloppy("127.0.0.1")},
	); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(apiserver.Codecs)

	serverConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(
		sampleopenapi.GetOpenAPIDefinitions, openapi.NewDefinitionNamer(apiserver.Scheme),
	)
	serverConfig.OpenAPIConfig.Info.Title = "Apps"
	serverConfig.OpenAPIConfig.Info.Version = "0.1"

	serverConfig.OpenAPIV3Config = genericapiserver.DefaultOpenAPIV3Config(
		sampleopenapi.GetOpenAPIDefinitions, openapi.NewDefinitionNamer(apiserver.Scheme),
	)
	serverConfig.OpenAPIV3Config.Info.Title = "Apps"
	serverConfig.OpenAPIV3Config.Info.Version = "0.1"

	serverConfig.FeatureGate = utilversionpkg.DefaultComponentGlobalsRegistry.FeatureGateFor(
		utilversionpkg.DefaultKubeComponent,
	)
	serverConfig.EffectiveVersion = utilversionpkg.DefaultComponentGlobalsRegistry.EffectiveVersionFor(
		apiserver.AppsComponentName,
	)

	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	config := &apiserver.Config{
		GenericConfig:  serverConfig,
		ResourceConfig: o.ResourceConfig,
	}
	return config, nil
}

// RunAppsServer запускает новый AppsServer на основе AppsServerOptions
func (o AppsServerOptions) RunAppsServer(ctx context.Context) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	server.GenericAPIServer.AddPostStartHookOrDie("start-sample-server-informers", func(context genericapiserver.PostStartHookContext) error {
		config.GenericConfig.SharedInformerFactory.Start(context.Done())
		return nil
	})

	return server.GenericAPIServer.PrepareRun().RunWithContext(ctx)
}

// AppsVersionToKubeVersion определяет соответствие версий между компонентом Apps и kube
func AppsVersionToKubeVersion(ver *version.Version) *version.Version {
	if ver.Major() != 1 {
		return nil
	}
	kubeVer := utilversionpkg.DefaultKubeEffectiveVersion().BinaryVersion()
	// "1.2" соответствует kubeVer
	offset := int(ver.Minor()) - 2
	mappedVer := kubeVer.OffsetMinor(offset)
	if mappedVer.GreaterThan(kubeVer) {
		return kubeVer
	}
	return mappedVer
}
