package main

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/labstack/echo/v4"
)

const (
	kubeconfigPath   = "/home/leandro/.kube/config"
	defaultNamespace = "charles"
)

// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
// https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-api/#directly-accessing-the-rest-api
// https://kubernetes.io/docs/reference/kubernetes-api/workloads-resources/
func main() {
	if err := runServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func runServer() error {
	clientset, err := configKubeClient(kubeconfigPath)
	if err != nil {
		fmt.Println(err)
		return err
	}

	e := echo.New()
	e.GET("/logs", logsHandler(clientset))
	e.Logger.Fatal(e.Start(":1323"))
	return nil
}

func configKubeClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
