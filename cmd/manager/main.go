package manager

import (
	"flag"
	"github.com/wenwenxiong/kubeipfixed/pkg/manager"
	"github.com/wenwenxiong/kubeipfixed/pkg/names"
	"os"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {

	runKubeipfixedManager()
}

func runKubeipfixedManager() {
	var logType, metricsAddr string
	var waitingTime int

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&logType, "v", "production", "Log type (debug/production).")
	flag.IntVar(&waitingTime, names.WAIT_TIME_ARG, 600, "waiting time to release the mac if object was not created")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(logType != "production")))

	log := ctrl.Log.WithName("runKubeipfixedManager")

	podNamespace, ok := os.LookupEnv("POD_NAMESPACE")
	if !ok {
		log.Error(nil, "Failed to load pod namespace from environment variable")
		os.Exit(1)
	}

	podName, ok := os.LookupEnv("POD_NAME")
	if !ok {
		log.Error(nil, "Failed to load pod name from environment variable")
		os.Exit(1)
	}

	kubeippoolManager := manager.NewKubeIPPoolManager(podNamespace, podName, metricsAddr, waitingTime)

	err = kubeippoolManager.Run()
	if err != nil {
		log.Error(err, "Failed to run the kubemacpool manager")
		os.Exit(1)
	}

}
