package time

import (
	"code.google.com/p/ginta/fmt"
	"code.google.com/p/ginta"
	"time"
)

const (
	DateFormat = "date"
	TimeFormat = "time"
	DateTimeFormat = "dateTime"
)

type dateFormatType string

func (typ dateFormatType) Compile(args []string) (fmt.MessageInput, error) {
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
		return EvaluateFormat(string(d), locale, time)
	}
	
	return arg
}

func EvaluateFormat(format string, locale ginta.Locale, instant time.Time) string {
	return ""
}