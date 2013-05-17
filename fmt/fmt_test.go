package fmt

import (
	"code.google.com/p/ginta/trunk/ginta"
	"testing"
)

const (
	qbf                = "The quick brown fox jumped over the lazy dog"
	qbfWithVars        = "The %v brown %v jumped over the %v %v"
	arguments          = "The {0} brown {1} jumped over the {2} {3}"
	outOfOrder         = "The {2} brown {3} jumped over the {0} {1}"
	customFormatResult = "be-silly"
	customFormatName   = "silly"
)

func TestParseNone(t *testing.T) {
	format, err := Compile(qbf)

	if err != nil {
		t.Error(err)
	}

	if format.format != qbf {
		t.Error(format.format)
	}

	if len(format.argumentIndices) != 0 {
		t.Error(format.argumentIndices)
	}

	if len(format.converters) != 0 {
		t.Error(format.converters)
	}

	if str := format.Format(ginta.DefaultLocale); str != qbf {
		t.Error(str)
	}
}

func TestParseSimple(t *testing.T) {
	format, err := Compile(arguments)

	if err != nil {
		t.Error(err)
	}

	if format.format != qbfWithVars {
		t.Error(format.format)
	}

	if len(format.argumentIndices) != 4 {
		t.Error(format.argumentIndices)
	}

	if len(format.converters) != 0 {
		t.Error(format.converters)
	}

	if str := format.Format(ginta.DefaultLocale, "quick", "fox", "lazy", "dog"); str != qbf {
		t.Error(str)
	}
}

func TestParseReorder(t *testing.T) {
	format, err := Compile(outOfOrder)

	if err != nil {
		t.Error(err)
	}

	if format.format != qbfWithVars {
		t.Error(format.format)
	}

	if len(format.argumentIndices) != 4 {
		t.Error(format.argumentIndices)
	}

	if len(format.converters) != 0 {
		t.Error(format.converters)
	}

	if str := format.Format(ginta.DefaultLocale, "lazy", "dog", "quick", "fox"); str != qbf {
		t.Error(str)
	}
}

type t1Fmt struct{}

func (_ t1Fmt) FormatString() string {
	return "%s"
}

func (x t1Fmt) Compile([]string) (MessageInput, error) {
	return x, nil
}

func (_ t1Fmt) Converter() Converter {

	return ConverterFunc(func(_ ginta.Locale, _ interface{}) interface{} {
		return customFormatResult
	})
}

func TestRegisterNewFormatter(t *testing.T) {
	RegisterFormat(customFormatName, t1Fmt{})
	fmt, err := Compile("{0," + customFormatName + "}")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if str := fmt.Format(ginta.DefaultLocale, "abc"); str != customFormatResult {
		t.Error(str, fmt)
	}
}
