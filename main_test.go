package main

import "testing"

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
