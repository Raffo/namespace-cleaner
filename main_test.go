package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes/fake"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestContains(t *testing.T) {
	s := "kube-system"
	list := []string{"kube-system", "foo", "bar"}
	b := contains(list, s)
	if !b {
		t.Fatalf("expected true, got false")
	}

	s = "baz"
	b = contains(list, s)
	if b {
		t.Fatalf("expected false, got true")
	}
}

func TestNamespacesToDelete(tt *testing.T) {
	for _, test := range []struct {
		name       string
		namespaces []v1.Namespace
		toRetain   []string
		expected   []string
	}{
		{
			name: "delete all namespaces",
			namespaces: []v1.Namespace{
				v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom",
					},
				},
			},
			toRetain: []string{},
			expected: []string{"custom"},
		},
		{
			name: "retain one from one",
			namespaces: []v1.Namespace{
				v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom",
					},
				},
			},
			toRetain: []string{"custom"},
			expected: []string{},
		},
		{
			name: "retain one from multiple",
			namespaces: []v1.Namespace{
				v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom",
					},
				},
				v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "kube-system",
					},
				},
			},
			toRetain: []string{"kube-system"},
			expected: []string{"custom"},
		},
		{
			name: "do not delete labelled namespaces",
			namespaces: []v1.Namespace{
				v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom",
						Labels: map[string]string{
							"preserve": "true",
						},
					},
				},
				v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "kube-system",
					},
				},
			},
			toRetain: []string{"kube-system"},
			expected: []string{},
		},
	} {
		tt.Run(test.name, func(t *testing.T) {
			toDelete := namespacesToDelete(test.namespaces, test.toRetain)
			if !reflect.DeepEqual(toDelete, test.expected) {
				t.Fatalf("expected %v, have %v", test.expected, toDelete)
			}
		})
	}

}

func TestDeleteNamespaces(t *testing.T) {
	client := fake.NewSimpleClientset()

	client.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "custom",
		},
	})
	err := deleteNamespaces(client, []string{"custom"})
	if err != nil {
		t.Fatalf("got error while deleting namespace: %v", err)
	}
	ns, _ := client.CoreV1().Namespaces().List(metav1.ListOptions{})
	if len(ns.Items) > 0 {
		t.Fatalf("expected empty, got len %d, with items: %v", len(ns.Items), ns.Items)
	}
}

func TestCombineArray(t *testing.T) {
	a := []string{"a"}
	b := []string{"b"}
	c := combineArray(a, b)
	if !reflect.DeepEqual(c, []string{"a", "b"}) {
		t.Fatalf("expected {a, b}, have %v", c)
	}
}

func TestDo(tt *testing.T) {
	client := fake.NewSimpleClientset()

	for _, test := range []struct {
		name               string
		existingNamespaces []v1.Namespace
		yes                bool
		neverDelete        []string
		doNotDelete        []string
		expected           []string
	}{
		{
			name: "delete all namespaces",
			existingNamespaces: []v1.Namespace{
				v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom",
					},
				},
			},
			yes:         true,
			neverDelete: []string{"kube-system"},
			doNotDelete: []string{},
			expected:    []string{},
		},
		{
			name: "do not delete custom",
			existingNamespaces: []v1.Namespace{
				v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom",
					},
				},
			},
			yes:         true,
			neverDelete: []string{"kube-system"},
			doNotDelete: []string{"custom"},
			expected:    []string{"custom"},
		},
		{
			name: "do not delete custom",
			existingNamespaces: []v1.Namespace{
				v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom",
					},
				},
				v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "kube-system",
					},
				},
			},
			yes:         true,
			neverDelete: []string{"kube-system"},
			doNotDelete: []string{},
			expected:    []string{"kube-system"},
		},
		{
			name: "do not delete anything in dry run",
			existingNamespaces: []v1.Namespace{
				v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "custom",
					},
				},
				v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "kube-system",
					},
				},
			},
			yes:         false,
			neverDelete: []string{"kube-system"},
			doNotDelete: []string{},
			expected:    []string{"kube-system", "custom"},
		},
	} {
		tt.Run(test.name, func(t *testing.T) {
			for _, n := range test.existingNamespaces {
				client.CoreV1().Namespaces().Create(&n)
			}
			do(client, test.yes, test.neverDelete, test.doNotDelete)
			ns, _ := client.CoreV1().Namespaces().List(metav1.ListOptions{})
			val := []string{}
			for _, n := range ns.Items {
				val = append(val, n.Name)
			}
			if !reflect.DeepEqual(test.expected, val) {
				t.Fatalf("expected %v, have %v", test.expected, val)
			}
		})
	}
}

func TestControlLoop(t *testing.T) {
	loop := true
	c := make(chan int)
	go func() {
		time.Sleep(2 * time.Second)
		c <- 0
	}()
	client := fake.NewSimpleClientset()
	config := config{
		oneShot:     loop,
		c:           c,
		yes:         false,
		neverDelete: neverDelete,
		nsToRetain:  []string{},
		d:           nextDeleteTime(time.Date(2019, time.January, 20, 10, 0, 0, 0, time.UTC), "Friday", 20),
		interval:    4 * time.Second,
	}
	controlLoop(client, &config)
}

func TestControlLoopWithDelete(t *testing.T) {
	loop := true
	c := make(chan int)
	go func() {
		time.Sleep(2 * time.Second)
		c <- 0
	}()
	client := fake.NewSimpleClientset()
	config := config{
		oneShot:     loop,
		c:           c,
		yes:         false,
		neverDelete: neverDelete,
		nsToRetain:  []string{},
		d:           nextDeleteTime(time.Date(2019, time.January, 25, 10, 0, 0, 0, time.UTC), "Friday", 20),
		interval:    4 * time.Second,
	}
	controlLoop(client, &config)
}

func TestControlLoopLoop(t *testing.T) {
	oneShot := true
	c := make(chan int)
	go func() {
		time.Sleep(10 * time.Second)
		c <- 0
	}()
	client := fake.NewSimpleClientset()
	config := config{
		oneShot:     oneShot,
		c:           c,
		yes:         false,
		neverDelete: neverDelete, nsToRetain: []string{},
		d:        nextDeleteTime(time.Now(), "Friday", 20),
		interval: 4 * time.Second,
	}
	controlLoop(client, &config)
}

func TestNextDeleteTime(tt *testing.T) {
	for _, test := range []struct {
		name       string
		now        time.Time
		deleteDay  string
		deleteHour int
		expected   time.Time
	}{
		{
			name:       "now is before the delete time",
			now:        time.Date(2019, time.January, 20, 10, 0, 0, 0, time.UTC),
			deleteDay:  "Friday",
			deleteHour: 22,
			expected:   time.Date(2019, time.January, 25, 22, 0, 0, 0, time.UTC),
		},
		{
			name:       "now is after the delete day",
			now:        time.Date(2019, time.January, 26, 10, 0, 0, 0, time.UTC),
			deleteDay:  "Friday",
			deleteHour: 22,
			expected:   time.Date(2019, time.February, 1, 22, 0, 0, 0, time.UTC),
		},
		{
			name:       "now is after the delete day for a few minutes",
			now:        time.Date(2019, time.January, 25, 10, 0, 0, 0, time.UTC),
			deleteDay:  "Friday",
			deleteHour: 10,
			expected:   time.Date(2019, time.February, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			name:       "now is the same day as the delete day but earlier",
			now:        time.Date(2019, time.January, 25, 10, 0, 0, 0, time.UTC),
			deleteDay:  "Friday",
			deleteHour: 22,
			expected:   time.Date(2019, time.January, 25, 22, 0, 0, 0, time.UTC),
		},
		{
			name:       "now is the same day as the delete day but later",
			now:        time.Date(2019, time.January, 25, 23, 0, 0, 0, time.UTC),
			deleteDay:  "Friday",
			deleteHour: 22,
			expected:   time.Date(2019, time.February, 1, 22, 0, 0, 0, time.UTC),
		},
		{
			name:       "first delete of the new year",
			now:        time.Date(2018, time.December, 30, 10, 0, 0, 0, time.UTC),
			deleteDay:  "Friday",
			deleteHour: 22,
			expected:   time.Date(2019, time.January, 4, 22, 0, 0, 0, time.UTC),
		},
		{
			name:       "end of February (no leap year)",
			now:        time.Date(2019, time.February, 26, 10, 0, 0, 0, time.UTC),
			deleteDay:  "Friday",
			deleteHour: 22,
			expected:   time.Date(2019, time.March, 1, 22, 0, 0, 0, time.UTC),
		},
		{
			name:       "end of February (lep year)",
			now:        time.Date(2020, time.February, 29, 10, 0, 0, 0, time.UTC),
			deleteDay:  "Friday",
			deleteHour: 22,
			expected:   time.Date(2020, time.March, 6, 22, 0, 0, 0, time.UTC),
		},
	} {
		tt.Run(test.name, func(t *testing.T) {
			deleteTime := nextDeleteTime(test.now, test.deleteDay, test.deleteHour)
			fmt.Println(test.now)
			fmt.Println(deleteTime)
			fmt.Println("------")
			if !deleteTime.Equal(test.expected) {
				t.Fatalf("expected time %v, have %v", test.expected, deleteTime)
			}
		})
	}
}
