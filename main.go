package main

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/alecthomas/kingpin"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var neverDelete = []string{"kube-system", "default", "kube-public"}

const preserveLabel = "preserve"

func contains(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}

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

func controlLoop(client kubernetes.Interface, config *config) {
	nextTimeToDelete := nextDeleteTime(time.Now(), config.d.Weekday().String(), config.d.Hour())
	for {
		select {
		case <-time.After(config.interval):
			logrus.Infof("next delete time: %v", nextTimeToDelete)
			now := time.Now()
			if nextTimeToDelete.Before(time.Now()) {
				logrus.Infof("it's time to delete %v", now)
				do(client, config.yes, config.neverDelete, config.nsToRetain)
				if config.oneShot {
					return
				}
				nextTimeToDelete = nextDeleteTime(now, config.d.Weekday().String(), config.d.Hour())
			} else {
				logrus.Infof("It's not time to delete: %v", now)
			}
		case <-config.c:
			return
		}
	}
}

var daysOfWeek = map[string]time.Weekday{
	"Sunday":    time.Sunday,
	"Monday":    time.Monday,
	"Tuesday":   time.Tuesday,
	"Wednesday": time.Wednesday,
	"Thursday":  time.Thursday,
	"Friday":    time.Friday,
	"Saturday":  time.Saturday,
}

func nextDeleteTime(now time.Time, day string, hour int) time.Time {
	var futureDay int
	currentWeekDay := int(now.Weekday())
	if currentWeekDay < int(daysOfWeek[day]) {
		futureDay = int(daysOfWeek[day]) - currentWeekDay
	} else if currentWeekDay == int(daysOfWeek[day]) {
		if now.Hour() >= hour {
			futureDay = 7 - currentWeekDay + int(daysOfWeek[day])
		} else {
			futureDay = 0
		}
	} else {
		futureDay = 7 - currentWeekDay + int(daysOfWeek[day])
	}
	//adding the days that I need to add to go the next day
	date := now.AddDate(0, 0, futureDay)
	//removing minutes and seconds to start at the exact hour
	date = date.Add(-(time.Duration(now.Minute()) * time.Minute))
	date = date.Add(-(time.Duration(now.Second()) * time.Second))
	//set the right hour to exec
	return date.Add(time.Duration(hour-now.Hour()) * time.Hour)
}

type config struct {
	oneShot     bool
	c           chan int
	yes         bool
	neverDelete []string
	nsToRetain  []string
	d           time.Time
	interval    time.Duration
}

func main() {
	nsToRetain := kingpin.Flag("namespaces-to-retain", "List of namespaces to retain.").Strings()
	kubeconfig := kingpin.Flag("kubeconfig", "path to kubeconfig file.").String()
	yes := kingpin.Flag("yes", "Set this flag if you want to delete the namespace otherwise it will only run in dry run mode.").Bool()
	day := kingpin.Flag("day", "Set the value of the day of the week during which to execute the cleaning operation.").Required().String()
	t := kingpin.Flag("time", "Set the value of the time of the day in the week during which to execute the cleaning operation.").Required().Int()
	oneShot := kingpin.Flag("oneShot", "Run in one shot mode or control loop.").Bool()
	kingpin.Parse()

	clientConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		logrus.Fatalf("cannot build config: %v", err)
	}

	client, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		logrus.Fatalf("cannot build kubeclient: %v", err)
	}

	d := nextDeleteTime(time.Now(), *day, *t) //TODO refactor

	c := make(chan int)
	controlLoop(client, &config{
		oneShot:     *oneShot,
		c:           c,
		yes:         *yes,
		neverDelete: neverDelete,
		nsToRetain:  *nsToRetain,
		d:           d,
		interval:    30 * time.Second,
	})

}
