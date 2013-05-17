package plural

import (
	"bytes"
	"code.google.com/p/ginta"
	"code.google.com/p/ginta/providers/simple"
	"testing"
)

func TestOneTwoThree(t *testing.T) {
	expect := map[int64]string{
		1: "one",
		2: "two",
		3: "three",
	}

	ginta.Register(simple.New().AddLanguage("l1", "Language 1", map[string]string{
		"plurals:t0:==1": "one",
		"plurals:t0:==2": "two",
		"plurals:t0:==3": "three",
	}))

	if p, err := parse([]string{"t0"}); err == nil {
		for key, val := range expect {
			if str := p.Converter().Convert(ginta.Locale("l1"), key).(string); str != val {
				t.Error(key, val, str)
			}
		}
	} else {
		t.Error(p, err)
	}
}

func doTest(t *testing.T, expect map[int64]string, contents map[string]string, code string) {
	ginta.Register(simple.New().AddLanguage(code, code, contents))

	if p, err := parse([]string{"t0"}); err == nil {
		for key, val := range expect {
			if str := p.Converter().Convert(ginta.Locale(code), key).(string); str != val {
				t.Error(key, val, str)
			}
		}
	} else {
		t.Error(p, err)
	}
}

func TestOneTwoMany(t *testing.T) {
	expect := map[int64]string{
		1:  "one",
		2:  "two",
		3:  "many",
		48: "many",
	}

	contents := map[string]string{
		"plurals:t0:==1":     "one",
		"plurals:t0:==2":     "two",
		"plurals:t0:default": "many",
	}

	doTest(t, expect, contents, "l2")
}

func TestLessEqualGreaterThan(t *testing.T) {
	expect := map[int64]string{
		0: "lessEqual",
		1: "lessEqual",
		2: "greater",
	}

	contents := map[string]string{
		"plurals:t0:<=1": "lessEqual",
		"plurals:t0:>1":  "greater",
	}

	doTest(t, expect, contents, "l3")
}

func TestLessThanGreaterEqual(t *testing.T) {
	expect := map[int64]string{
		0: "less",
		1: "greaterEqual",
		2: "greaterEqual",
	}

	contents := map[string]string{
		"plurals:t0:<1":  "less",
		"plurals:t0:>=1": "greaterEqual",
	}

	doTest(t, expect, contents, "l4")
}

func TestModule(t *testing.T) {
	expect := map[int64]string{
		1:  "st",
		2:  "nd",
		3:  "rd",
		4:  "th",
		5:  "th",
		6:  "th",
		7:  "th",
		8:  "th",
		9:  "th",
		10: "th",
		11: "th",
		12: "th",
		13: "th",
		14: "th",
		15: "th",
		16: "th",
		17: "th",
		18: "th",
		19: "th",
		20: "th",
		21: "st",
		22: "nd",
		23: "rd",
		24: "th",
		25: "th",
		26: "th",
		27: "th",
		28: "th",
		29: "th",
	}

	contents := map[string]string{
		"plurals:t0:%10==1":  "st",
		"plurals:t0:%10==2":  "nd",
		"plurals:t0:%10==3":  "rd",
		"plurals:t0:==11":    "th",
		"plurals:t0:==12":    "th",
		"plurals:t0:==13":    "th",
		"plurals:t0:default": "th",
	}

	doTest(t, expect, contents, "l5")
}

func TestInterval(t *testing.T) {
	expect := map[float64]string{
		-1e-30: "outside",
		0:      "inside",
		.5:     "inside",
		.99999: "inside",
		1:      "outside",
	}

	contents := map[string]string{
		"plurals:t0:[0,1[":   "inside",
		"plurals:t0:default": "outside",
	}

	code := "l6"

	ginta.Register(simple.New().AddLanguage(code, code, contents))

	if p, err := parse([]string{"t0"}); err == nil {
		for key, val := range expect {
			if str := p.Converter().Convert(ginta.Locale(code), key).(string); str != val {
				t.Error(key, val, str)
			}
		}
	} else {
		t.Error(p, err)
	}
}

type inter struct{}

func (x inter) IntValue() int32 {
	return 0
}

func TestInputTypes(t *testing.T) {
	expect := []interface{}{
		inter{},
		int(1),
		int8(2),
		int16(3),
		int32(4),
		int64(5),
		uint(6),
		uint8(7),
		uint16(8),
		uint32(9),
		uint64(10),
		float32(11),
		float64(12),
	}

	ginta.Register(simple.New().AddLanguage("lx", "Language 3", map[string]string{
		"plurals:t0:==1":  "1",
		"plurals:t0:==2":  "2",
		"plurals:t0:==3":  "3",
		"plurals:t0:==4":  "4",
		"plurals:t0:==5":  "5",
		"plurals:t0:==6":  "6",
		"plurals:t0:==7":  "7",
		"plurals:t0:==8":  "8",
		"plurals:t0:==9":  "9",
		"plurals:t0:==10": "10",
		"plurals:t0:==11": "11",
		"plurals:t0:==12": "12",
		"plurals:t0:==0":  "0",
	}))

	if p, err := parse([]string{"t0"}); err == nil {
		var b bytes.Buffer
		for _, val := range expect {
			str := p.Converter().Convert(ginta.Locale("lx"), val).(string)
			b.WriteString(str)
		}

		if b.String() != "0123456789101112" {
			t.Error(b.String())
		}
	} else {
		t.Error(p, err)
	}
}
