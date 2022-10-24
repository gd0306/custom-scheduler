package pkg

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func InitClientSet(masterUrl string, configPath string) (*kubernetes.Clientset, error) {
	var kubeConfig *rest.Config
	// in-cluster config
	inClusterConfig, err := rest.InClusterConfig()
	if err != nil {
	} else {
		kubeConfig = inClusterConfig
		if masterUrl != "" {
			kubeConfig.Host = masterUrl
		}
	}
	// out-cluster configuration
	if masterUrl == "" || kubeConfig == nil {
		configFromFlags, err := clientcmd.BuildConfigFromFlags(masterUrl, configPath)
		if err != nil {
			return nil, err
		}
		kubeConfig = configFromFlags
	}
	return kubernetes.NewForConfig(kubeConfig)
}
