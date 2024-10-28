//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	unsafe "unsafe"

	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
	apps "k8s.io/sample-apiserver/pkg/apis/apps"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*Application)(nil), (*apps.Application)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_Application_To_apps_Application(a.(*Application), b.(*apps.Application), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*apps.Application)(nil), (*Application)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_apps_Application_To_v1alpha1_Application(a.(*apps.Application), b.(*Application), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ApplicationList)(nil), (*apps.ApplicationList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_ApplicationList_To_apps_ApplicationList(a.(*ApplicationList), b.(*apps.ApplicationList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*apps.ApplicationList)(nil), (*ApplicationList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_apps_ApplicationList_To_v1alpha1_ApplicationList(a.(*apps.ApplicationList), b.(*ApplicationList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ApplicationSpec)(nil), (*apps.ApplicationSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_ApplicationSpec_To_apps_ApplicationSpec(a.(*ApplicationSpec), b.(*apps.ApplicationSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*apps.ApplicationSpec)(nil), (*ApplicationSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_apps_ApplicationSpec_To_v1alpha1_ApplicationSpec(a.(*apps.ApplicationSpec), b.(*ApplicationSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ApplicationStatus)(nil), (*apps.ApplicationStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_ApplicationStatus_To_apps_ApplicationStatus(a.(*ApplicationStatus), b.(*apps.ApplicationStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*apps.ApplicationStatus)(nil), (*ApplicationStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_apps_ApplicationStatus_To_v1alpha1_ApplicationStatus(a.(*apps.ApplicationStatus), b.(*ApplicationStatus), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_Application_To_apps_Application(in *Application, out *apps.Application, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_v1alpha1_ApplicationSpec_To_apps_ApplicationSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1alpha1_ApplicationStatus_To_apps_ApplicationStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1alpha1_Application_To_apps_Application is an autogenerated conversion function.
func Convert_v1alpha1_Application_To_apps_Application(in *Application, out *apps.Application, s conversion.Scope) error {
	return autoConvert_v1alpha1_Application_To_apps_Application(in, out, s)
}

func autoConvert_apps_Application_To_v1alpha1_Application(in *apps.Application, out *Application, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_apps_ApplicationSpec_To_v1alpha1_ApplicationSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_apps_ApplicationStatus_To_v1alpha1_ApplicationStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_apps_Application_To_v1alpha1_Application is an autogenerated conversion function.
func Convert_apps_Application_To_v1alpha1_Application(in *apps.Application, out *Application, s conversion.Scope) error {
	return autoConvert_apps_Application_To_v1alpha1_Application(in, out, s)
}

func autoConvert_v1alpha1_ApplicationList_To_apps_ApplicationList(in *ApplicationList, out *apps.ApplicationList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]apps.Application)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_v1alpha1_ApplicationList_To_apps_ApplicationList is an autogenerated conversion function.
func Convert_v1alpha1_ApplicationList_To_apps_ApplicationList(in *ApplicationList, out *apps.ApplicationList, s conversion.Scope) error {
	return autoConvert_v1alpha1_ApplicationList_To_apps_ApplicationList(in, out, s)
}

func autoConvert_apps_ApplicationList_To_v1alpha1_ApplicationList(in *apps.ApplicationList, out *ApplicationList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]Application)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_apps_ApplicationList_To_v1alpha1_ApplicationList is an autogenerated conversion function.
func Convert_apps_ApplicationList_To_v1alpha1_ApplicationList(in *apps.ApplicationList, out *ApplicationList, s conversion.Scope) error {
	return autoConvert_apps_ApplicationList_To_v1alpha1_ApplicationList(in, out, s)
}

func autoConvert_v1alpha1_ApplicationSpec_To_apps_ApplicationSpec(in *ApplicationSpec, out *apps.ApplicationSpec, s conversion.Scope) error {
	out.Version = in.Version
	out.Values = in.Values
	return nil
}

// Convert_v1alpha1_ApplicationSpec_To_apps_ApplicationSpec is an autogenerated conversion function.
func Convert_v1alpha1_ApplicationSpec_To_apps_ApplicationSpec(in *ApplicationSpec, out *apps.ApplicationSpec, s conversion.Scope) error {
	return autoConvert_v1alpha1_ApplicationSpec_To_apps_ApplicationSpec(in, out, s)
}

func autoConvert_apps_ApplicationSpec_To_v1alpha1_ApplicationSpec(in *apps.ApplicationSpec, out *ApplicationSpec, s conversion.Scope) error {
	out.Version = in.Version
	out.Values = in.Values
	return nil
}

// Convert_apps_ApplicationSpec_To_v1alpha1_ApplicationSpec is an autogenerated conversion function.
func Convert_apps_ApplicationSpec_To_v1alpha1_ApplicationSpec(in *apps.ApplicationSpec, out *ApplicationSpec, s conversion.Scope) error {
	return autoConvert_apps_ApplicationSpec_To_v1alpha1_ApplicationSpec(in, out, s)
}

func autoConvert_v1alpha1_ApplicationStatus_To_apps_ApplicationStatus(in *ApplicationStatus, out *apps.ApplicationStatus, s conversion.Scope) error {
	return nil
}

// Convert_v1alpha1_ApplicationStatus_To_apps_ApplicationStatus is an autogenerated conversion function.
func Convert_v1alpha1_ApplicationStatus_To_apps_ApplicationStatus(in *ApplicationStatus, out *apps.ApplicationStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_ApplicationStatus_To_apps_ApplicationStatus(in, out, s)
}

func autoConvert_apps_ApplicationStatus_To_v1alpha1_ApplicationStatus(in *apps.ApplicationStatus, out *ApplicationStatus, s conversion.Scope) error {
	return nil
}

// Convert_apps_ApplicationStatus_To_v1alpha1_ApplicationStatus is an autogenerated conversion function.
func Convert_apps_ApplicationStatus_To_v1alpha1_ApplicationStatus(in *apps.ApplicationStatus, out *ApplicationStatus, s conversion.Scope) error {
	return autoConvert_apps_ApplicationStatus_To_v1alpha1_ApplicationStatus(in, out, s)
}
