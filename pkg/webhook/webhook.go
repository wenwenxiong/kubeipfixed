package webhook

import (
	ip_manager "github.com/wenwenxiong/kubeipfixed/pkg/ip-manager"
	"os"
	"strings"

	"github.com/pkg/errors"

	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	kawwebhook "github.com/qinqon/kube-admission-webhook/pkg/webhook"
)

const (
	WebhookServerPort = 8000
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations;validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;update;create;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;create;update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;create;update;patch;list;watch
// +kubebuilder:rbac:groups="kubevirt.io",resources=virtualmachines,verbs=get;list;watch;create;update;patch
var AddToWebhookFuncs []func(*kawwebhook.Server, *ip_manager.IPManager) error

// AddToManager adds all Controllers to the Manager
func AddToManager(mgr manager.Manager, ipManager *ip_manager.IPManager) error {

	s := &kawwebhook.Server{
		Port:          WebhookServerPort,
		TLSMinVersion: tlsMinVersion(),
		CipherSuites:  cipherSuites(),
	}
	s.Register("/readyz", healthz.CheckHandler{Checker: healthz.Ping})

	for _, f := range AddToWebhookFuncs {
		if err := f(s, ipManager); err != nil {
			return err
		}
	}

	err := mgr.Add(s)
	if err != nil {
		return errors.Wrap(err, "failed adding webhook server to manager")
	}
	return nil
}

// cipherSuites read the TLS handshake ciphers from a environment variable if
// empty the decision is delegated to go tls package.
func cipherSuites() []string {
	cipherSuitesEnv := os.Getenv("TLS_CIPHERS")
	if cipherSuitesEnv == "" {
		return nil
	}
	return strings.Split(cipherSuitesEnv, ",")
}

// tlsMinVersion read the TLS minimal version from environment a environment
// variable, if it's empty the webhook server will fallback to "1.0"
func tlsMinVersion() string {
	return os.Getenv("TLS_MIN_VERSION")
}
