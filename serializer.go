package main

import (
	karpenterv1alpha5 "github.com/aws/karpenter-core/pkg/apis"
	karpenterapi "github.com/aws/karpenter/pkg/apis"
	certmanager "github.com/cert-manager/cert-manager/pkg/api"
	externalsecretsv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	kservev1alpha1 "github.com/kserve/kserve/pkg/apis/serving/v1alpha1"
	servingv1alpha1 "github.com/kserve/modelmesh-serving/apis/serving/v1alpha1"
	profilev1 "github.com/kubeflow/kubeflow/components/profile-controller/api/v1"
	profilev1beta1 "github.com/kubeflow/kubeflow/components/profile-controller/api/v1beta1"
	otelv1alpha1 "github.com/open-telemetry/opentelemetry-operator/apis/v1alpha1"
	slothv1alpha1 "github.com/slok/sloth/pkg/kubernetes/api/sloth/v1"
	tektonpipelinesv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	tektontriggersv1alpha1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	knativecachingalpha1 "knative.dev/caching/pkg/apis/caching/v1alpha1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	secretsstorev1 "sigs.k8s.io/secrets-store-csi-driver/apis/v1"
)

var (
	// deserializer = clientgoscheme.Codecs.UniversalDecoder(scheme.)
	// scheme = runtime.NewScheme()
	Scheme = runtime.NewScheme()
	Codecs = serializer.NewCodecFactory(Scheme)
)

func init() {
	// KUBERNETES
	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))
	utilruntime.Must(apiextensions.AddToScheme(Scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(Scheme))

	// CERT MANAGER
	utilruntime.Must(certmanager.AddToScheme(Scheme))

	// CILIUM
	// utilruntime.Must(ciliumv2.AddToScheme(Scheme)) // dependency ongithub.com/optiopay/kafka and has CVEs

	// EXTERNAL SECRETS
	utilruntime.Must(externalsecretsv1beta1.AddToScheme(Scheme))

	// GATEWAY API
	utilruntime.Must(gatewayv1alpha2.AddToScheme(Scheme))
	utilruntime.Must(gatewayv1beta1.AddToScheme(Scheme))

	// KARPENTER
	utilruntime.Must(karpenterv1alpha5.AddToScheme(Scheme))
	utilruntime.Must(karpenterapi.AddToScheme(Scheme))

	// KNATIVE
	utilruntime.Must(servingv1alpha1.AddToScheme(Scheme))
	utilruntime.Must(kservev1alpha1.AddToScheme(Scheme))
	utilruntime.Must(knativecachingalpha1.AddToScheme(Scheme))

	// KUBEFLOW
	utilruntime.Must(profilev1.AddToScheme(Scheme))
	utilruntime.Must(profilev1beta1.AddToScheme(Scheme))

	// ISTIO
	utilruntime.Must(istionetworkingv1beta1.AddToScheme(Scheme))
	utilruntime.Must(istiosecurityv1beta1.AddToScheme(Scheme))
	utilruntime.Must(istionetworkingv1alpha3.AddToScheme(Scheme))

	// OTEL
	utilruntime.Must(otelv1alpha1.AddToScheme(Scheme))

	// SECRET STORE
	utilruntime.Must(secretsstorev1.AddToScheme(Scheme))

	// SLOTH
	utilruntime.Must(slothv1alpha1.AddToScheme(Scheme))

	// TEKTON
	utilruntime.Must(tektonpipelinesv1.AddToScheme(Scheme))
	utilruntime.Must(tektontriggersv1alpha1.AddToScheme(Scheme))

}

func defaultSerializer() runtime.Decoder {
	// KUBERNETES
	utilruntime.Must(clientgoscheme.AddToScheme(clientgoscheme.Scheme))
	utilruntime.Must(apiextensions.AddToScheme(clientgoscheme.Scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(clientgoscheme.Scheme))

	// CERT MANAGER
	utilruntime.Must(certmanager.AddToScheme(clientgoscheme.Scheme))

	// CILIUM
	// utilruntime.Must(ciliumv2.AddToScheme(clientgoscheme.Scheme))

	// EXTERNAL SECRETS
	utilruntime.Must(externalsecretsv1beta1.AddToScheme(clientgoscheme.Scheme))

	// GATEWAY API
	utilruntime.Must(gatewayv1alpha2.AddToScheme(clientgoscheme.Scheme))
	utilruntime.Must(gatewayv1beta1.AddToScheme(clientgoscheme.Scheme))

	// KARPENTER
	utilruntime.Must(karpenterv1alpha5.AddToScheme(clientgoscheme.Scheme))
	utilruntime.Must(karpenterapi.AddToScheme(clientgoscheme.Scheme))

	// KNATIVE
	utilruntime.Must(servingv1alpha1.AddToScheme(clientgoscheme.Scheme))
	utilruntime.Must(kservev1alpha1.AddToScheme(clientgoscheme.Scheme))
	utilruntime.Must(knativecachingalpha1.AddToScheme(clientgoscheme.Scheme))

	// KUBEFLOW
	utilruntime.Must(profilev1.AddToScheme(clientgoscheme.Scheme))
	utilruntime.Must(profilev1beta1.AddToScheme(clientgoscheme.Scheme))

	// ISTIO
	utilruntime.Must(istionetworkingv1beta1.AddToScheme(clientgoscheme.Scheme))
	utilruntime.Must(istiosecurityv1beta1.AddToScheme(clientgoscheme.Scheme))
	utilruntime.Must(istionetworkingv1alpha3.AddToScheme(clientgoscheme.Scheme))

	// OTEL
	utilruntime.Must(otelv1alpha1.AddToScheme(clientgoscheme.Scheme))

	// SECRET STORE
	utilruntime.Must(secretsstorev1.AddToScheme(clientgoscheme.Scheme))

	// SLOTH
	utilruntime.Must(slothv1alpha1.AddToScheme(clientgoscheme.Scheme))

	// TEKTON
	utilruntime.Must(tektonpipelinesv1.AddToScheme(clientgoscheme.Scheme))
	utilruntime.Must(tektontriggersv1alpha1.AddToScheme(clientgoscheme.Scheme))

	return clientgoscheme.Codecs.UniversalDeserializer()
}
