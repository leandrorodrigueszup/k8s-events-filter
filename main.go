package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/labstack/echo/v4"
)

const (
	kubeconfigPath 	 = "/home/leandro/.kube/config"
	defaultNamespace = "charles"
)

// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
// https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-api/#directly-accessing-the-rest-api
// https://kubernetes.io/docs/reference/kubernetes-api/workloads-resources/
func main() {
	// if err := run(); err != nil {
	// 	fmt.Fprintf(os.Stderr, "%v", err)
	// 	os.Exit(1)
	// }
	newServer()
}

func newServer() {
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

		// labels := map[string]string{"app.kubernetes.io/managed-by": "Helm"}
		// label := c.QueryParam("label")
		//labels := extractLabel(label)
		// app.kubernetes.io%2Fmanaged-by%3DHelm
		//labels := map[string]string{"app.kubernetes.io/managed-by": "Helm"}
		// podNames, err := filterPodsByLabel(clientset, nil, namespace)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return err
		// }

		namespace := getNamespaceOr(c, defaultNamespace)
		watch, err := watchEventsByPodName(clientset, "", namespace)
		if err != nil {
			return err
		}

		for {
			if event, ok := <-watch.ResultChan(); ok {
				eventData,_ := json.Marshal(event.Object)
				var coreEvent corev1.Event
				json.Unmarshal(eventData, &coreEvent)
				
				eventLog := newEventLog(coreEvent)

				fmt.Println(eventLog)

				if err := enc.Encode(eventLog); err != nil {
					fmt.Println(err)
					return err
				}
				c.Response().Flush()
			} else {
				break
			}
		}

		// enc := json.NewEncoder(c.Response())
		// for _, l := range locations {
		// 	if err := enc.Encode(l); err != nil {
		// 		return err
		// 	}
		// 	c.Response().Flush()
		// }
		return nil
	})
	e.Logger.Fatal(e.Start(":1323"))
}

func extractLabel(label string) map[string]string {
	keyValue := strings.Split(label, "=")
	labels := make(map[string]string, len(keyValue))
	labels[keyValue[0]] = keyValue[1]
	return labels
}

func getNamespaceOr(c echo.Context, defaultNamespace string) string {
	namespace := c.QueryParam("namespace")
	if namespace == "" {
		return defaultNamespace
	}
	return namespace
}

func run() error {
	clientset, err := configKubeClient(kubeconfigPath)
	if err != nil {
		return err
	}

	labels := map[string]string{"app.kubernetes.io/managed-by": "Helm"}
	podNames, err := filterPodsByLabel(clientset, labels, defaultNamespace)
	if err != nil {
		return err
	}

	//log.Println(podNames)

	for _, podName := range podNames {
		events, err := filterEventsByPodName(clientset, podName, defaultNamespace)
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
