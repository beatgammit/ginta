package time

import (
	"code.google.com/p/ginta"
	"code.google.com/p/ginta/common"
	"code.google.com/p/ginta/fmt"
	"time"
)

const (
	DateFormat     = "date"
	TimeFormat     = "time"
	DateTimeFormat = "dateTime"

	OptionShort = "short"
	OptionLong  = "long"
)

type dateFormatType string

func (typ dateFormatType) Compile(args []string) (fmt.MessageInput, error) {
	l := len(args)
	var res string

	if l == 0 {
		res = "time_format:" + string(typ) + ":default"
	} else if l == 1 && (args[0] == OptionShort || args[0] == OptionLong) {
		res = "time_format:" + string(typ) + ":" + args[0]
	}

	if res != "" {
		return dateFormat(res), nil
	}

	return nil, fmt.NewError(fmt.MalformedFormatSpecificationErrorResourceKey, args)
}

func init() {
	fmt.RegisterFormat(DateFormat, dateFormatType(DateFormat))
	fmt.RegisterFormat(TimeFormat, dateFormatType(TimeFormat))
	fmt.RegisterFormat(DateTimeFormat, dateFormatType(DateTimeFormat))
}

type dateFormat string

func (d dateFormat) FormatString() string {
	return "%v"
}

func (d dateFormat) Converter() fmt.Converter {
	return d
}

func (d dateFormat) Convert(locale ginta.Locale, arg interface{}) interface{} {
	if time, ok := arg.(time.Time); ok {
		return EvaluateFormat(common.HierarchicalKey(d), locale, time)
	}

	return arg
}

func EvaluateFormat(format common.HierarchicalKey, locale ginta.Locale, instant time.Time) string {
	if fmtString, err := locale.ResolveResource(format); err == nil {
		result := instant.Format(fmtString)

		// TODO now perform substitutions for strings (wednesday -> miércoles)

		return result
	}

	return "?" + string(format)
}
