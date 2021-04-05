package main

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	kubeconfigPath = "/home/leandro/.kube/config"
	namespace      = "charles"
)

// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
// https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-api/#directly-accessing-the-rest-api
// https://kubernetes.io/docs/reference/kubernetes-api/workloads-resources/
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func run() error {
	clientset, err := configKubeClient(kubeconfigPath)
	if err != nil {
		return err
	}

	labels := map[string]string{"app.kubernetes.io/managed-by": "Helm"}
	podNames, err := filterPodsByLabel(clientset, labels, namespace)
	if err != nil {
		return err
	}

	//log.Println(podNames)

	for _, podName := range podNames {
		events, err := filterEventsByPodName(clientset, podName, namespace)
		if err != nil {
			return err
		}
		// data, err := json.Marshal(events)
		// if err != nil {
		// 	return nil
		// }
		printEvents(events)
	}
	return nil
}

func printEvents(events []*EventLog) {
	for _, event := range events {
		fmt.Printf("%s\t\t%s\t\t%s\t\t%s\n", event.Type, event.Reason, event.Object, event.Message)
	}
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
