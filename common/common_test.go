package common

import (
	"testing"
)

func TestErrorImplemented(t *testing.T) {
	var err error = ResourceNotFoundError("some:resource")
	err.Error()
}

func TestError(t *testing.T) {
	if ResourceNotFoundError("some:resource").Error() != "some:resource" {
		t.Fail()
	}
}

func TestHierarchicalKeySplit(t *testing.T) {
	key := HierarchicalKey("some:path:key")

	prefix, local := key.Split()

	if prefix != "some:path" {
		t.Error(prefix)
	}

	if local != "key" {
		t.Error(local)
	}
}

func TestHierarchyWalk(t *testing.T) {
	key := HierarchicalKey("some:long:path:key")
	expect := []string{"some:long:path:key", "some:long:key", "some:key", "key", ""}

	for _, expectStr := range expect {
		if key.String() != expectStr {
			t.Error(key)
		}

		key = key.Parent()
	}
}
