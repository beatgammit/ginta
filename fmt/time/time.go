package time

import (
	"code.google.com/p/ginta/fmt"
	"code.google.com/p/ginta"
	"time"
)

func init() {
	fmt.RegisterFormat("date", fmt.FormatDefinitionFunc(nil))
}

type dateFormat func (ginta.Locale, time.Time) string

func (d dateFormat) FormatString() string {
	return "%v"
}

func (d dateFormat) Converter() fmt.Converter {
	return fmt.ConverterFunc(func (locale ginta.Locale, arg interface{}) interface{} {
		if time, ok := arg.(time.Time); ok {
			return d(locale, time)
		} 
		
		return arg
	})
}