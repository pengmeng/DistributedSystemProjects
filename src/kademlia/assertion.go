package kademlia

import (
	"strings"
	"testing"
)

func assertStringEqual(expect, actual, msg string, t *testing.T) {
	if actual != expect {
		t.Error(msg)
		t.Error("Expect: " + expect)
		t.Error("Actual: " + actual)
	}
}

func assertContains(universe, subset, msg string, t *testing.T) {
	if !strings.Contains(universe, subset) {
		t.Error(msg)
		t.Error("Universe: " + universe)
		t.Error("Subset: " + subset)
	}
}

func assertNotContains(universe, subset, msg string, t *testing.T) {
	if strings.Contains(universe, subset) {
		t.Error(msg)
		t.Error("Universe: " + universe)
		t.Error("Subset: " + subset)
	}
}

func assertIntEqual(expect, actual int, msg string, t *testing.T) {
	if actual != expect {
		t.Error(msg)
		t.Errorf("Expect: %d\n", expect)
		t.Errorf("Actual: %d\n", actual)
	}
}

func assertTrue(flag bool, msg string, t *testing.T) {
	if !flag {
		t.Error(msg)
	}
}

func assertFalse(flag bool, msg string, t *testing.T) {
	if flag {
		t.Error(msg)
	}
}
