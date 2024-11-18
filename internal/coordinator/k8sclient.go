package coordinator

import (
	"os"
	"slices"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewK8sClient() (*kubernetes.Clientset, error) {
	args := os.Args[1:]
	var config *rest.Config
	var err error
	if slices.Contains(args, "--dev") {
		kubeConfigPath := GetConfig().GetKubeConfigPath()
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
