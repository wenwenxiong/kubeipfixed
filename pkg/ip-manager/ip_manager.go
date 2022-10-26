package ip_manager

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sync"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	sriovNetworksAnnotation        = "k8s.v1.cni.cncf.io/sriovnetworks"
	NetworksAnnotation             = "k8s.v1.cni.cncf.io/networks"
	TransactionTimestampAnnotation = "kubeippool.io/transaction-timestamp"
	mutatingWebhookConfigName      = "kubeippool-mutator"
	virtualMachnesWebhookName      = "mutatevirtualmachines.kubeippool.io"
	podsWebhookName                = "mutatepods.kubeippool.io"
	defaultNameservers             = "114.114.114.114"
)

var log = logf.Log.WithName("IPManager")

// now is an artifact to do some unit testing without waiting for expiration time.
var now = func() time.Time { return time.Now() }

type IPManager struct {
	cachedKubeClient client.Client
	Scheme           *runtime.Scheme
	kubeClient       client.Client
	managerNamespace string
	poolMutex        sync.Mutex // mutex for allocation an release
	isKubevirt       bool       // bool if kubevirt virtualmachine crd exist in the cluster
	waitTime         int        // Duration in second to free macs of allocated vms that failed to start.
}

func NewIPManager(kubeClient, cachedKubeClient client.Client, managerNamespace string, kubevirtExist bool, waitTime int, Scheme *runtime.Scheme) (*IPManager, error) {

	ipManger := &IPManager{
		cachedKubeClient: cachedKubeClient,
		kubeClient:       kubeClient,
		isKubevirt:       kubevirtExist,
		managerNamespace: managerNamespace,
		poolMutex:        sync.Mutex{},
		waitTime:         waitTime,
		Scheme:           Scheme,
	}

	return ipManger, nil
}

func (p *IPManager) Start() error {
	return nil
}

func (p *IPManager) IsKubevirtEnabled() bool {
	return p.isKubevirt
}
