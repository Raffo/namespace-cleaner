package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/alecthomas/kingpin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const preserveLabel = "preserve"

func contains(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}

var neverDelete = []string{"kube-system", "default", "kube-public"}

func main() {
	namespaces := kingpin.Flag("namespaces", "List of namespaces").Strings()
	kubeconfig := kingpin.Flag("kubeconfig", "path to kubeconfig file").Default("~/.kube/config").String()
	kingpin.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	toRetain := append(neverDelete, *namespaces...)

	// get all namespaces.
	ns, err := client.CoreV1().Namespaces().List(metav1.ListOptions{})

	for _, n := range ns.Items {
		if !contains(toRetain, n.Name) {
			if _, ok := n.Labels[preserveLabel]; !ok {
				logrus.Infof("deleting namespace %s", n.Name)
				err := client.CoreV1().Namespaces().Delete(n.Name, &metav1.DeleteOptions{})
				if err != nil {
					logrus.Errorf("cannot delete namespace %s, error: %v", n.Name, err)
				}
			} else {
				logrus.Infof("skipping delete for namespace %s cause it's market with preserve label", n.Name)
			}
		}
	}
}
