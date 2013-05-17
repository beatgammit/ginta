package simple

import (
	"code.google.com/p/ginta/trunk/ginta"
	"testing"
)

const (
	e1      = "entry1"
	e2      = "entry2"
	hello   = "Hello World"
	feature = "class 1 feature"
	missing = "Missing from other locale!"
)

var (
	l1Map = map[string]string{e1: hello}
	l2Map = map[string]string{e1: feature, e2: missing}
)

func TestSimple(t *testing.T) {
	ginta.Register(New().AddLanguage("l1", "Language 1", l1Map).AddLanguage("l2", "Language 2", l2Map))

	lang1 := ginta.Locale("l1")
	lang2 := ginta.Locale("l2")

	if str, err := lang1.GetResource(e1); err != nil || str != hello {
		t.Error(str, err)
	}

	if str, err := lang1.GetResource(e2); err == nil {
		t.Error(str, err)
	}

	if str, err := lang2.GetResource(e1); err != nil || str != feature {
		t.Error(str, err)
	}

	if str, err := lang2.GetResource(e2); err != nil || str != missing {
		t.Error(str, err)
	}
}
