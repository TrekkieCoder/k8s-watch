package main

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
	"time"
)

type svcPair struct {
	IP   string
	Port int32
}

type dnsIf interface{}

type networkStatus struct {
	Name  string   `json:"name"`
	Iface string   `json:"interface"`
	Ips   []string `json:"ips"`
	Mac   string   `json:"mac"`
	Dflt  bool     `json:"default"`
	Dns   dnsIf    `json:"dns"`
}

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", "/etc/rancher/k3s/k3s.yaml")
	if err != nil {
		fmt.Println(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Second))
	//eps, err := clientset.CoreV1().Endpoints("default").List(ctx, metav1.ListOptions{FieldSelector:"metadata.name=kubernetes"})
	selector := fmt.Sprintf("%s=%s", "metadata.name", "multus-service")
	networkSelector := "macvlan-conf-1"
	eps, err := clientset.CoreV1().Endpoints("default").List(ctx, metav1.ListOptions{FieldSelector: selector})

	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("%v\n", eps)
	cancel()

	var svcpairs []svcPair
	var svcMpairs []svcPair
	for _, item := range eps.Items {
		for _, epSS := range item.Subsets {
			for idx, addr := range epSS.Addresses {
				port := epSS.Ports[idx]
				fmt.Printf("EP IP %v:%v\n\n", addr.IP, port.Port)
				svcpairs = append(svcpairs, svcPair{IP: addr.IP, Port: port.Port})
			}
		}
	}

	ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	pods, err := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, pod := range pods.Items {
		fmt.Printf("Service POD name %s:%s\n", pod.Name, pod.Status.PodIP)

		for _, svcpair := range svcpairs {
			if pod.Status.PodIP == svcpair.IP {
				fmt.Printf("Service POD name %s:%s (Found) \n", pod.Name, pod.Spec.NodeName)

				if na := pod.Annotations["k8s.v1.cni.cncf.io/networks"]; na != "" {
					fmt.Printf("Networks %s\n", na)
					if na == networkSelector {
						if ns := pod.Annotations["k8s.v1.cni.cncf.io/networks-status"]; ns != "" {

							fmt.Printf("Status %s\n", ns)
							data := []networkStatus{}
							err := json.Unmarshal([]byte(ns), &data)
							if err != nil {
								fmt.Printf("Status unmarshall fail %s\n", err)
								continue
							}

							fmt.Printf("Status unmarshall  %v\n", data)

							for _, dat := range data {
								fmt.Printf("Dat  %s\n", dat.Name)
								if strings.Contains(dat.Name, networkSelector) {
									for _, ip := range dat.Ips {
										svcMpairs = append(svcMpairs, svcPair{IP: ip, Port: svcpair.Port})
									}
								}
							}

							fmt.Printf("Status struct %v\n", data)
						}
					}
				}
			}
		}
	}
	cancel()

	fmt.Printf("svcMpair %v\n", svcMpairs)

	watchlist := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		string(v1.ResourceServices),
		v1.NamespaceAll,
		fields.Everything(),
	)
	_, controller := cache.NewInformer( // also take a look at NewSharedIndexInformer
		watchlist,
		&v1.Service{},
		0, //Duration is int64
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				//fmt.Printf("service added: %s \n", obj)
			},
			DeleteFunc: func(obj interface{}) {
				//fmt.Printf("service deleted: %s \n", obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				//fmt.Printf("service changed \n")
			},
		},
	)
	// I found it in k8s scheduler module. Maybe it's help if you interested in.
	// serviceInformer := cache.NewSharedIndexInformer(watchlist, &v1.Service{}, 0, cache.Indexers{
	//     cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
	// })
	// go serviceInformer.Run(stop)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)
	for {
		time.Sleep(time.Second)
	}
}
