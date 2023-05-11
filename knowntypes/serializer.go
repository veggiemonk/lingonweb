package knowntypes

import (
	karpenterv1alpha5 "github.com/aws/karpenter-core/pkg/apis"
	karpenterapi "github.com/aws/karpenter/pkg/apis"
	certmanager "github.com/cert-manager/cert-manager/pkg/api"
	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	ciliumv2alpha1 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2alpha1"
	externalsecretsv1beta1 "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	imageautov1 "github.com/fluxcd/image-automation-controller/api/v1beta1"
	imagereflectv1beta2 "github.com/fluxcd/image-reflector-controller/api/v1beta2"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	notificationv1b2 "github.com/fluxcd/notification-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	grafanav1beta1 "github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	kservev1alpha1 "github.com/kserve/kserve/pkg/apis/serving/v1alpha1"
	servingv1alpha1 "github.com/kserve/modelmesh-serving/apis/serving/v1alpha1"
	profilev1 "github.com/kubeflow/kubeflow/components/profile-controller/api/v1"
	profilev1beta1 "github.com/kubeflow/kubeflow/components/profile-controller/api/v1beta1"
	jetstreamv1beta2 "github.com/nats-io/nack/pkg/jetstream/apis/jetstream/v1beta2"
	otelv1alpha1 "github.com/open-telemetry/opentelemetry-operator/apis/v1alpha1"
	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	slothv1alpha1 "github.com/slok/sloth/pkg/kubernetes/api/sloth/v1"
	tektonpipelinesv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	tektontriggersv1alpha1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	istionetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	apidiscoveryv2beta1 "k8s.io/api/apidiscovery/v2beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	apiserverconfig "k8s.io/apiserver/pkg/apis/config"
	apiserverconfigv1 "k8s.io/apiserver/pkg/apis/config/v1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	aggregatorclientsetscheme "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/scheme"
	knativecachingalpha1 "knative.dev/caching/pkg/apis/caching/v1alpha1"
	knativeservingv1 "knative.dev/serving/pkg/apis/serving/v1"
	capiahelm "sigs.k8s.io/cluster-api-addon-provider-helm/api/v1alpha1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	secretsstorev1 "sigs.k8s.io/secrets-store-csi-driver/apis/v1"
)

var (
	Scheme = runtime.NewScheme()
	Codecs = serializer.NewCodecFactory(Scheme)
)

// ADD MORE TYPES / CRDs HERE
func init() {
	// KUBERNETES
	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))
	utilruntime.Must(apiextensions.AddToScheme(Scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(Scheme))
	utilruntime.Must(aggregatorclientsetscheme.AddToScheme(Scheme))
	utilruntime.Must(apiserverconfig.AddToScheme(Scheme))
	utilruntime.Must(apiserverconfigv1.AddToScheme(Scheme))
	utilruntime.Must(apidiscoveryv2beta1.AddToScheme(Scheme))

	// CERT MANAGER
	utilruntime.Must(certmanager.AddToScheme(Scheme))

	// CILIUM
	utilruntime.Must(ciliumv2.AddToScheme(Scheme))
	utilruntime.Must(ciliumv2alpha1.AddToScheme(Scheme))

	// EXTERNAL SECRETS
	utilruntime.Must(externalsecretsv1beta1.AddToScheme(Scheme))

	// GATEWAY API
	utilruntime.Must(gatewayv1alpha2.AddToScheme(Scheme))
	utilruntime.Must(gatewayv1beta1.AddToScheme(Scheme))

	// NATS JETSTREAM
	utilruntime.Must(jetstreamv1beta2.AddToScheme(Scheme))

	// KARPENTER
	utilruntime.Must(karpenterv1alpha5.AddToScheme(Scheme))
	utilruntime.Must(karpenterapi.AddToScheme(Scheme))

	// KNATIVE
	utilruntime.Must(servingv1alpha1.AddToScheme(Scheme))
	utilruntime.Must(kservev1alpha1.AddToScheme(Scheme))
	utilruntime.Must(knativecachingalpha1.AddToScheme(Scheme))
	utilruntime.Must(knativeservingv1.AddToScheme(Scheme))

	// KUBEFLOW
	utilruntime.Must(profilev1.AddToScheme(Scheme))
	utilruntime.Must(profilev1beta1.AddToScheme(Scheme))

	// FLUXCD
	utilruntime.Must(kustomizev1.AddToScheme(Scheme))
	utilruntime.Must(helmv2.AddToScheme(Scheme))
	utilruntime.Must(notificationv1b2.AddToScheme(Scheme))
	utilruntime.Must(imagereflectv1beta2.AddToScheme(Scheme))
	utilruntime.Must(imageautov1.AddToScheme(Scheme))
	utilruntime.Must(sourcev1.AddToScheme(Scheme))

	utilruntime.Must(capiahelm.AddToScheme(Scheme))

	// ISTIO
	utilruntime.Must(istionetworkingv1beta1.AddToScheme(Scheme))
	utilruntime.Must(istiosecurityv1beta1.AddToScheme(Scheme))
	utilruntime.Must(istionetworkingv1alpha3.AddToScheme(Scheme))

	// GRAFANA
	utilruntime.Must(grafanav1beta1.AddToScheme(Scheme))

	// PROMETHEUS
	utilruntime.Must(prometheusv1.AddToScheme(Scheme))

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
