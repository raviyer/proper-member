package util

import (
	"testing"
)

func TestUUID(t *testing.T) {
	s, err := NewUUID()
	if err != nil {
		t.Error(err)
	}
	t.Log(s)
	s1, err := NewUUID()
	if err != nil {
		t.Error(err)
	}

	t.Log(s1)
	if s == s1 {
		t.Error("Same Guid Generated")
	}
}

func TestGetName(t *testing.T) {
	n, err := GetRandomName()
	if err != nil {
		t.Error(err)
	}
	if len(n) == 0 {
		t.Error("Failed to get a name")
	}
}

func TestFailpoints(t *testing.T) {
	fpmap := NewFailPointMap("foo")
	fpmap.AddFailpoint("Bla", false, 100, "Bad ju")
	err := fpmap.Save("foo")
	if err != nil {
		t.Errorf("Failed to save: %v", err)
	}
}