package main

import (
	"reflect"
	"testing"

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
