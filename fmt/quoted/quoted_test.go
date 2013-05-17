package quoted

import (
	"code.google.com/p/ginta/trunk/ginta"
	"code.google.com/p/ginta/trunk/ginta/fmt"
	"testing"
)

const (
	simpleArg    = "Radomal"
	expectSimple = "\"Radomal\""
)

func TestQuoteParseOk(t *testing.T) {
	in, err := parse([]string{})
	if in.FormatString() != "%#v" || err != nil {
		t.Error(in, err)
	}
}

func TestQuoteParseFail(t *testing.T) {
	in, err := parse([]string{"1"})
	if in != nil || err == nil {
		t.Error(in, err)
	}
}

func TestQuotedSimple(t *testing.T) {
	format, err := fmt.Compile("{0,quoted}")

	if err != nil {
		t.Error(format, err)
	}

	if str := format.Format(ginta.DefaultLocale, simpleArg); err != nil || str != expectSimple {
		t.Errorf("Got:\n%s\nExp:\n%s\nErr:%v\n", str, expectSimple, err)
	}
}
