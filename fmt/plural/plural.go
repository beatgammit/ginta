package plural

import (
	"code.google.com/p/ginta"
	"code.google.com/p/ginta/fmt"
	sysfmt "fmt"
	"math"
	"strconv"
	"strings"
)

const (
	Format      = "plural"
	DefaultSpec = "default"

	ModuloOperation       = "%"
	GreaterOperation      = ">"
	GreaterEqualOperation = ">="
	LessOperation         = "<"
	LessEqualOperation    = "<="
	EqualOperation        = "=="
	RangeOperation        = "["

	pluralStemResourcesPath = "plurals:"
	nothingFoundGlobal      = 0x7fffffff
)

type pluralStem string

type IntValuer interface {
	IntValue() int32
}

type FloatValuer interface {
	FloatValue() float64
}

func init() {
	fmt.RegisterFormat(Format, fmt.FormatDefinitionFunc(parse))
}

func parse(args []string) (fmt.MessageInput, error) {
	if len(args) == 1 {
		return pluralStem(args[0]), nil
	}
	return nil, fmt.NewError(fmt.MalformedFormatSpecificationErrorResourceKey, args)
}

func (p pluralStem) Converter() fmt.Converter {
	return p
}

func (p pluralStem) Convert(l ginta.Locale, input interface{}) interface{} {
	base := pluralStemResourcesPath + string(p)
	bundle := l.GetResourceBundle(base)

	if f, convert := value(input); convert {
		priority := nothingFoundGlobal
		var translation string

		for key, val := range bundle {
			var nextPriority int
			var op func(float64) bool
			switch {
			case strings.HasPrefix(key, EqualOperation):
				op = equals(key[len(EqualOperation):])
				nextPriority = 0
			case strings.HasPrefix(key, GreaterEqualOperation):
				op = greaterEqual(key[len(GreaterEqualOperation):])
				nextPriority = 1
			case strings.HasPrefix(key, LessEqualOperation):
				op = lessEqual(key[len(LessEqualOperation):])
				nextPriority = 2
			case strings.HasPrefix(key, GreaterOperation):
				op = greater(key[len(GreaterOperation):])
				nextPriority = 3
			case strings.HasPrefix(key, LessOperation):
				op = less(key[len(LessOperation):])
				nextPriority = 4
			case strings.HasPrefix(key, RangeOperation):
				op = inRange(key[len(RangeOperation):])
				nextPriority = 5
			case strings.HasPrefix(key, ModuloOperation):
				op = moduleVal(key[len(ModuloOperation):])
				nextPriority = 6

			case key == DefaultSpec:
				op = _true
				nextPriority = nothingFoundGlobal - 1
			}

			if op != nil && nextPriority < priority && op(f) {
				priority = nextPriority
				translation = val
			}
		}

		if priority < nothingFoundGlobal {
			return translation
		}
	}

	return input

}
func _true(float64) bool {
	return true
}

func moduleVal(input string) func(float64) bool {
	var module, expect float64

	if got, err := sysfmt.Sscanf(input, "%f==%f", &module, &expect); got == 2 && err == nil {
		return func(x float64) bool {
			result := math.Mod(x, module)
			return result == expect
		}
	} else {
		sysfmt.Println(err)
	}

	return nil
}

func inRange(input string) func(float64) bool {
	var start, end float64

	if got, err := sysfmt.Sscanf(input, "%f,%f", &start, &end); got == 2 && err == nil {
		return func(x float64) bool {
			return x >= start && x < end
		}
	}

	return nil
}

func lessEqual(input string) func(float64) bool {
	if cmp, err := strconv.ParseFloat(input, 64); err == nil {
		return func(in float64) bool {
			return in <= cmp
		}
	}

	return nil
}

func less(input string) func(float64) bool {
	if cmp, err := strconv.ParseFloat(input, 64); err == nil {
		return func(in float64) bool {
			return in < cmp
		}
	}

	return nil
}

func greaterEqual(input string) func(float64) bool {
	if cmp, err := strconv.ParseFloat(input, 64); err == nil {
		return func(in float64) bool {
			return in >= cmp
		}
	}

	return nil
}

func greater(input string) func(float64) bool {
	if cmp, err := strconv.ParseFloat(input, 64); err == nil {
		return func(in float64) bool {
			return in > cmp
		}
	}

	return nil
}

func equals(input string) func(float64) bool {
	if cmp, err := strconv.ParseFloat(input, 64); err == nil {
		return func(in float64) bool {
			return in == cmp
		}
	}

	return nil
}

func (p pluralStem) FormatString() string {
	return "%s"
}

func value(in interface{}) (float64, bool) {
	switch in.(type) {
	case FloatValuer:
		return in.(FloatValuer).FloatValue(), true
	case IntValuer:
		return float64(in.(IntValuer).IntValue()), true
	case float64:
		return float64(in.(float64)), true
	case float32:
		return float64(in.(float32)), true
	case uint:
		return float64(in.(uint)), true
	case int:
		return float64(in.(int)), true
	case int8:
		return float64(in.(int8)), true
	case int16:
		return float64(in.(int16)), true
	case int32:
		return float64(in.(int32)), true
	case int64:
		return float64(in.(int64)), true
	case uint8:
		return float64(in.(uint8)), true
	case uint16:
		return float64(in.(uint16)), true
	case uint32:
		return float64(in.(uint32)), true
	case uint64:
		return float64(in.(uint64)), true
	}

	return 0, false
}
