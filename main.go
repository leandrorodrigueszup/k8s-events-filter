package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/labstack/echo/v4"
)

const (
	kubeconfigPath   = "/home/leandro/.kube/config"
	defaultNamespace = "charles"
)

type label struct {
	name  string
	value string
}

// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
// https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-api/#directly-accessing-the-rest-api
// https://kubernetes.io/docs/reference/kubernetes-api/workloads-resources/
func main() {
	runServer()
}

func runServer() {
	e := echo.New()
	e.GET("/logs", func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c.Response().WriteHeader(http.StatusOK)
		enc := json.NewEncoder(c.Response())

		clientset, err := configKubeClient(kubeconfigPath)
		if err != nil {
			fmt.Println(err)
			return err
		}

		label := extractLabel(c.QueryParam("label"))

		namespace := getNamespaceOr(c, defaultNamespace)
		watch, err := watchEventsByPodName(clientset, namespace)
		if err != nil {
			return err
		}

		for {
			if e, ok := <-watch.ResultChan(); ok {
				event := toEvent(e.Object)
				if event.InvolvedObject.Kind != "Pod" {
					fmt.Printf("Resource Type '%s' Not Supported Yet\n", event.InvolvedObject.Kind)
					continue
				}

				pod, err := findPod(clientset, namespace, event.InvolvedObject.Name)
				if err != nil {
					fmt.Println(err)
					continue
				}

				eventLog := toEventLog(e.Object)

				if !podHasLabel(pod, label) {
					fmt.Printf("DISCART: %v\n", eventLog)
				} else {
					fmt.Println(eventLog)

					if err := enc.Encode(eventLog); err != nil {
						fmt.Println(err)
						continue
					}
				}
				c.Response().Flush()
			} else {
				fmt.Println("Stop watching")
				break
			}
		}
		return nil
	})
	e.Logger.Fatal(e.Start(":1323"))
}

func podHasLabel(pod *corev1.Pod, l label) bool {
	labels := pod.Labels
	return labels[l.name] == l.value
}

func toEvent(eventObject runtime.Object) *corev1.Event {
	bytes, _ := json.Marshal(eventObject)
	var event corev1.Event
	json.Unmarshal(bytes, &event)
	return &event
}

func toEventLog(eventObject runtime.Object) *EventLog {
	return newEventLog(toEvent(eventObject))
}

func extractLabel(l string) label {
	keyValue := strings.Split(l, "=")
	return label{keyValue[0], keyValue[1]}
}

func getNamespaceOr(c echo.Context, defaultNamespace string) string {
	namespace := c.QueryParam("namespace")
	if namespace == "" {
		return defaultNamespace
	}
	return namespace
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
