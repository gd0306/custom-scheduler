package pkg

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

func InitClientSet(masterUrl string, configPath string) (*kubernetes.Clientset, error) {
	var kubeConfig *rest.Config
	// in-cluster config
	inClusterConfig, err := rest.InClusterConfig()
	if err != nil {
		klog.V(4).Infof("not running in-cluster, falling back to out-cluster config: %v", err)
	} else {
		kubeConfig = inClusterConfig
		if masterUrl != "" {
			kubeConfig.Host = masterUrl
		}
	}
	// out-cluster configuration
	if kubeConfig == nil {
		configFromFlags, err := clientcmd.BuildConfigFromFlags(masterUrl, configPath)
		if err != nil {
			return nil, err
		}
		kubeConfig = configFromFlags
	}
	return kubernetes.NewForConfig(kubeConfig)
}
