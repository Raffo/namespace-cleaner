package main

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/alecthomas/kingpin"
	"k8s.io/api/core/v1"
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

func namespacesToDelete(namespaces []v1.Namespace, toRetain []string) []string {
	toDelete := []string{}
	for _, n := range namespaces {
		if !contains(toRetain, n.Name) {
			if _, ok := n.Labels[preserveLabel]; !ok {
				toDelete = append(toDelete, n.Name)
			}
		}
	}
	return toDelete
}

func deleteNamespaces(client kubernetes.Interface, ns []string) error {
	for _, n := range ns {
		logrus.Infof("deleting namespace %s", n)
		err := client.CoreV1().Namespaces().Delete(n, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func combineArray(a, b []string) []string {
	return append(a, b...)
}

func do(client kubernetes.Interface, yes bool, neverDelete, doNotDelete []string) error {
	toRetain := combineArray(neverDelete, doNotDelete)

	ns, err := client.CoreV1().Namespaces().List(metav1.ListOptions{})

	if err != nil {
		return fmt.Errorf("cannot list namespaces: %v", err)
	}

	toDelete := namespacesToDelete(ns.Items, toRetain)

	if yes {
		err = deleteNamespaces(client, toDelete)
		if err != nil {
			return fmt.Errorf("cannot delete namespaces %v", err)
		}
	} else {
		logrus.Infof("Dry run mode, I would have deleted the following namespaces: %v", toDelete)
	}

	return nil
}

func main() {
	nsToRetain := kingpin.Flag("namespaces-to-retain", "List of namespaces to retain.").Strings()
	kubeconfig := kingpin.Flag("kubeconfig", "path to kubeconfig file.").String()
	yes := kingpin.Flag("yes", "Set this flag if you want to delete the namespace otherwise it will only run in dry run mode.").Bool()
	kingpin.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		logrus.Fatalf("cannot build config: %v", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("cannot build kubeclient: %v", err)
	}

	do(client, *yes, neverDelete, *nsToRetain)
}
