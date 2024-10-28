package conversion

import (
	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1alpha1 "k8s.io/sample-apiserver/pkg/apis/apps/v1alpha1"
)

// ConvertHelmReleaseToApplication преобразует HelmRelease в Application
func ConvertHelmReleaseToApplication(hr *helmv2.HelmRelease) (*appsv1alpha1.Application, error) {
	app := &appsv1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.cozystack.io/v1alpha1",
			Kind:       hr.Spec.Chart.Spec.Chart,
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

// ConvertApplicationToHelmRelease преобразует Application в HelmRelease
func ConvertApplicationToHelmRelease(app *appsv1alpha1.Application) (*helmv2.HelmRelease, error) {
	hr := &helmv2.HelmRelease{
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
			Values: &apiextensionsv1.JSON{
				Raw: app.Spec.Values,
			},
		},
	}
	return hr, nil
}
