package nr

import (
	"testing"
)

func check(t *testing.T, expect string, args ...string) {
	f, err := parse(args)

	if err != nil {
		t.Error(err)
	} else if f.Converter() != nil || f.FormatString() != expect {
		t.Error(f.Converter(), f.FormatString(), "vs", expect)
	}
}

func TestFormatSimple(t *testing.T) {
	check(t, "%d")
}

func TestFormatLength(t *testing.T) {
	check(t, "%12d", "12")
}

func TestFormatSign(t *testing.T) {
	check(t, "%+d", Sign)
}

func TestFormatLengthSign(t *testing.T) {
	check(t, "%+21d", Sign, "21")
	check(t, "%+5d", "5", Sign)
}

func TestFormatPadZero(t *testing.T) {
	check(t, "%09d", "9", PadZero)
	check(t, "%04d", PadZero, "4")
}

func TestFormatPadZeroSign(t *testing.T) {
	check(t, "%+01d", "1", PadZero, Sign)
	check(t, "%+04d", "4", Sign, PadZero)
	check(t, "%+05d", PadZero, "5", Sign)
	check(t, "%+06d", Sign, "6", PadZero)
	check(t, "%+02d", PadZero, Sign, "2")
	check(t, "%+03d", Sign, PadZero, "3")
}

func TestFormatInvalid(t *testing.T) {
	if f, err := parse([]string{"no"}); err == nil {
		t.Error(f)
	}

	if f, err := parse([]string{"192.1020"}); err == nil {
		t.Error(f)
	}
}
