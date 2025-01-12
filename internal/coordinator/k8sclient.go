package coordinator

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewK8sClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	appConfig := GetConfig()
	if appConfig.IsInDevMode() {
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
