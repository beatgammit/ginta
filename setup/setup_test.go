package setup

import (
	"code.google.com/p/ginta/fmt"
	"testing"
)

func InstalledDefaultFormats(t *testing.T) {
	if _, err := fmt.Compile("{0,nr} {1,quoted} {2,plural,stem}"); err != nil {
		t.Error(err)
	}
}
