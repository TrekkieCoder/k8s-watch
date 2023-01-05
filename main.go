package main

    import (
        "fmt"
        "time"

         "k8s.io/api/core/v1"
         "k8s.io/apimachinery/pkg/fields"
         "k8s.io/client-go/kubernetes"
         "k8s.io/client-go/tools/cache"
         "k8s.io/client-go/tools/clientcmd"
    )

    func main() {
        config, err := clientcmd.BuildConfigFromFlags("", "/etc/rancher/k3s/k3s.yaml")
        if err != nil {
            fmt.Println(err)
        }
        clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
			fmt.Println(err)
        }

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
                    fmt.Printf("service added: %s \n", obj)
                },
                DeleteFunc: func(obj interface{}) {
                    fmt.Printf("service deleted: %s \n", obj)
                },
                UpdateFunc: func(oldObj, newObj interface{}) {
                    fmt.Printf("service changed \n")
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