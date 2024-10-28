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

package wardle

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupName is the group name used in this package
const GroupName = "apps.cozystack.io"

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: runtime.APIVersionInternal}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns back a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	// SchemeBuilder is the scheme builder with scheme init functions to run for this API package
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme is a common registration function for mapping packaged scoped group & version keys to a scheme
	AddToScheme = SchemeBuilder.AddToScheme
)

// Adds the list of known types to the given scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Application{},
		&ApplicationList{},
	)

	// Register internal types
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("Application"), &Application{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("ApplicationList"), &ApplicationList{})

	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("Kubernetes"), &Application{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("KubernetesList"), &ApplicationList{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("Postgres"), &Application{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("PostgresList"), &ApplicationList{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("Redis"), &Application{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("RedisList"), &ApplicationList{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("Kafka"), &Application{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("KafkaList"), &ApplicationList{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("RabbitMQ"), &Application{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("RabbitMQList"), &ApplicationList{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("FerretDB"), &Application{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("FerretDBList"), &ApplicationList{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("VMDisk"), &Application{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("VMDiskList"), &ApplicationList{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("VMInstance"), &Application{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("VMInstanceList"), &ApplicationList{})

	return nil
}
