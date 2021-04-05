package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const kubeconfigPath = "/home/leandro/.kube/config"

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
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	labels := map[string]string {"app.kubernetes.io/managed-by": "Helm"}
	podNames, err := retrieveResourceNamesByLabel(clientset, labels)
	if err != nil {
		return err
	}

	fmt.Println(podNames)

	fieldSelector := fmt.Sprintf("involvedObject.name in (%s)", strings.Join(podNames, ","))
	fmt.Println(fieldSelector)
	events, _ := clientset.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{FieldSelector: fieldSelector})
	fmt.Println(events)
	return nil
}

func retrieveResourceNamesByLabel(clientset *kubernetes.Clientset, labels map[string]string) ([]string, error) {
	labelSelector := formatLabelQuery(labels)
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}

	return mapPod(pods.Items, func(p corev1.Pod) string {
		return p.Name
	}), nil
}

func formatLabelQuery(labels map[string]string) string {
	var list []string
	for k, v := range labels {
		list = append(list, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(list, ",")
}

func mapPod(items []corev1.Pod, mapping func(corev1.Pod) string) []string {
	var result = make([]string, len((items)))
	for i, item := range items {
		result[i] = mapping(item)
	}
	return result
}
