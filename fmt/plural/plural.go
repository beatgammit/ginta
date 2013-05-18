/*
The plural package substitutes numeric values for "plural words". While english has relatively
straightforward rules on pluralization, other languages may be more involved, involving special 
cases, or complex rules. This package aims to support most languages reasonably well.

Towards this end, this package allows to define "plural bundles", which are resources with special 
format and semantics. A plural bundle is a hierarchical prefix located under a common prefix. Each resource
of such a bundle has a path, a condition, and a value.

The path is a normal hierarchical key path (one or more components). The Value likewise is a string constant with
no further formatting. The condition is the final part of the resource name (the Key). This key is not  a plain string,
but one of the following operators (listed in order of precedence):

	eq<value>                - simple equality check
	ge<value>                - greater-equals check
	le<value>                - less-equals check
	gt<value>                - greater check
	lt<value>                - less check
	range(<value1>,<value2>) - checks if the value is greater or equal to value 1, but less than value 2
	modEq(<value1>,<value2>) - checks if the input value mod value1 is equal to value2
	default                  - matches without condition

When a plural format is encountered, it is expected to have a single argument, which is the path of a plural bundle (relative 
to "plurals:"). This argument determines which plural bundle is loaded. The highest-priority matching condition determines the
output.

The input value to a plural converter may be of any numeric basic type. Comparisons and calculations are done using float64 arithmetic,
so extremely low fractions or very high values should usually be expressed in ranges rather than single values. In addition to numeric
basic types, any type may be used as an input, provided it implements the IntValuer or FloatValuer interface.

The following example should illustrate this:

Resources:
	plurals:key1:default=unknown
	plurals:key1:eq0=zero
	plurals:key1:ge19=many
	plurals:key1:gt9=some
	plurals:key1:range(8,9)=significant
	plurals:key1:modEq(2,0)=Even
	plurals:key1:modEq(2,1)=Odd

Format:
	{0,plural,key1}

Conversions:
	-2.0: Even
	0: zero
	1: Odd
	2: Even
	3: Odd
	4: Even
	0.5: unknown
	8: significant
	8.32: significant
	9: some
	12.292: some
	20: many
*/
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
	// Format ID
	Format = "plural"

	// default operation
	DefaultOperation = "default"
	// module equal to operation
	ModuloOperation = "modEq"
	// greater-than operation
	GreaterOperation = "gt"
	// greater-equal operation
	GreaterEqualOperation = "ge"
	// less-than operation
	LessOperation = "lt"
	// less-equals operation
	LessEqualOperation = "le"
	// equality operation
	EqualOperation = "eq"
	// range check operation
	RangeOperation = "range"

	// The root of the plural resource tree. Added automatically
	PluralStemResourcesPath = "plurals:"
	nothingFoundGlobal      = 0x7fffffff
)

type pluralStem string

// Implement this to allow your custom type to be used as a value
type IntValuer interface {
	IntValue() int32
}

// Implement this to allow your custom type to be used as a value
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
	base := PluralStemResourcesPath + string(p)
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

			case key == DefaultOperation:
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

	if got, err := sysfmt.Sscanf(input, "(%f,%f)", &module, &expect); got == 2 && err == nil {
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

	if got, err := sysfmt.Sscanf(input, "(%f,%f)", &start, &end); got == 2 && err == nil {
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
