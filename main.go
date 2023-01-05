package main

import (
	"fmt"
	"context"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
  const namespace = "default"

	// AUTHENTICATE
	// /etc/rancher/k3s/k3s.yaml
	var kubeconfig = "/etc/rancher/k3s/k3s.yaml"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// GET ENDPOINTS RESOURCE VERSION
	ctx := context.Background()

	var api = clientset.CoreV1().Endpoints(namespace)
	endpoints, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	resourceVersion := endpoints.ListMeta.ResourceVersion

	// SETUP WATCHER CHANNEL
	watcher, err := api.Watch(ctx, metav1.ListOptions{ResourceVersion: resourceVersion})
	if err != nil {
		panic(err.Error())
	}
	ch := watcher.ResultChan()

	// LISTEN TO CHANNEL
	for {
		event := <-ch
		endpoints, ok := event.Object.(*coreV1.Endpoints)
		if !ok {
			panic("Could not cast to Endpoint")
		}
		fmt.Printf("%v\n", endpoints.ObjectMeta.Name)
		for _, endpoint := range endpoints.Subsets {
			for _, address := range endpoint.Addresses {
				fmt.Printf("%v\n", address.IP)
			}
		}
	}
}
