package conversion

import (
	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	appsv1alpha1 "k8s.io/sample-apiserver/pkg/apis/apps/v1alpha1"
)

// ConvertHelmReleaseToApplication конвертирует HelmRelease в Application
func ConvertHelmReleaseToApplication(hr *helmv2.HelmRelease) (*appsv1alpha1.Application, error) {
	app := &appsv1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.cozystack.io/v1alpha1",
			Kind:       "Application",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hr.Name,
			Namespace: hr.Namespace,
		},
		Spec: appsv1alpha1.ApplicationSpec{
			Version: hr.Spec.Chart.Spec.Version,
			Values:  hr.Spec.Values,
		},
	}
	return app, nil
}

// ConvertApplicationToHelmRelease конвертирует Application в HelmRelease
func ConvertApplicationToHelmRelease(app *appsv1alpha1.Application) (*helmv2.HelmRelease, error) {
	helmRelease := &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "helm.toolkit.fluxcd.io/v2",
			Kind:       "HelmRelease",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			Labels: map[string]string{
				"cozystack.io/ui": "true",
			},
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: &helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:             "kubernetes",
					Version:           app.Spec.Version,
					ReconcileStrategy: "Revision",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      "HelmRepository",
						Name:      "cozystack-apps",
						Namespace: "cozy-public",
					},
				},
			},
			Values: app.Spec.Values,
		},
	}
	return helmRelease, nil
}

// ConvertHelmReleaseToUnstructured конвертирует HelmRelease в unstructured.Unstructured Application
func ConvertHelmReleaseToUnstructured(hr *helmv2.HelmRelease) (*unstructured.Unstructured, error) {
	app, err := ConvertHelmReleaseToApplication(hr)
	if err != nil {
		return nil, err
	}

	unstrApp, err := runtime.DefaultUnstructuredConverter.ToUnstructured(app)
	if err != nil {
		return nil, err
	}

	unstructuredApp := &unstructured.Unstructured{Object: unstrApp}
	unstructuredApp.SetAPIVersion(appsv1alpha1.SchemeGroupVersion.String())
	unstructuredApp.SetKind("Application")

	return unstructuredApp, nil
}

// ConvertUnstructuredToApplication конвертирует unstructured.Unstructured Application в Application struct
func ConvertUnstructuredToApplication(u *unstructured.Unstructured) (*appsv1alpha1.Application, error) {
	var app appsv1alpha1.Application
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &app)
	if err != nil {
		return nil, err
	}

	return &app, nil
}

// ConvertUnstructuredToHelmRelease конвертирует unstructured.Unstructured HelmRelease в HelmRelease struct
func ConvertUnstructuredToHelmRelease(u *unstructured.Unstructured) (*helmv2.HelmRelease, error) {
	var hr helmv2.HelmRelease
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &hr)
	if err != nil {
		return nil, err
	}
	return &hr, nil
}

// ConvertApplicationToUnstructured конвертирует Application struct в unstructured.Unstructured
func ConvertApplicationToUnstructured(app *appsv1alpha1.Application) (*unstructured.Unstructured, error) {
	unstrApp, err := runtime.DefaultUnstructuredConverter.ToUnstructured(app)
	if err != nil {
		return nil, err
	}

	unstructuredApp := &unstructured.Unstructured{Object: unstrApp}
	unstructuredApp.SetAPIVersion(appsv1alpha1.SchemeGroupVersion.String())
	unstructuredApp.SetKind("Application")

	return unstructuredApp, nil
}
