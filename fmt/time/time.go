/*
Allows formatted output of time values. 

The input to this format needs to be of type time.Time, but its formatting
is performed in a locale-sensitive manner. This package registers three formats: date, time
and dateTime, exported as the ...Format constants. Each format provides three different
manners of formatting for each locale: A short form, a normal (default) form, and
a verbose form. The forms are selected with the Option... constants exported
by this package. 

The package expects, for each combination of format and length, a resource entry 
(named TimeFormatRoot : <type> : <length>). This entry must have a form that is valid
to feed into Time.Format(). 

Also, the SubstitutionsResourceBundle may define a number of fixed strings that are
replaced in the formatted output. Its chief use is to translate the names of months,
and days of week.
*/
package time

import (
	"github.com/beatgammit/ginta"
	"github.com/beatgammit/ginta/common"
	"github.com/beatgammit/ginta/fmt"
	"strings"
	"time"
)

const (
	// Format ID: date format
	DateFormat = "date"
	// Format ID: time format
	TimeFormat = "time"
	// Format ID: date-Time format
	DateTimeFormat = "dateTime"

	// Format should select the short form
	OptionShort = "short"
	// Format should select a verbose form
	OptionLong = "long"
	// Format should select the "default" form. Optional
	OptionDefault = "default"

	// Root resource bundle for all formatting resources
	TimeFormatRoot = "time_format"
	// Resource bundle for string replacements
	SubstitutionsResourceBundle = TimeFormatRoot + ":substitutions"
)

type dateFormatType string

func (typ dateFormatType) Compile(args []string) (fmt.MessageInput, error) {
	l := len(args)
	var res string

	if l == 0 || (l == 1 && args[0] == OptionDefault) {
		res = TimeFormatRoot + ":" + string(typ) + ":" + OptionDefault
	} else if l == 1 && (args[0] == OptionShort || args[0] == OptionLong) {
		res = TimeFormatRoot + ":" + string(typ) + ":" + args[0]
	}

	if res != "" {
		return dateFormat(res), nil
	}

	return nil, fmt.NewError(fmt.MalformedFormatSpecificationErrorResourceKey, args)
}

// Registers the date formats with the format package
func Install() {
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

// Evaluates the format stored under the provided hierarchical key, performing formatting and substitutions
// as defined in the current locale
func EvaluateFormat(format common.HierarchicalKey, locale ginta.Locale, instant time.Time) string {
	fmtString, err := locale.ResolveResource(format)
	if err == nil {
		result := instant.Format(fmtString)

		bundle := locale.ResolveResourceBundle(SubstitutionsResourceBundle)

		// now perform substitutions for strings (wednesday -> miércoles)
		for from, to := range bundle {
			result = strings.Replace(result, from, to, 1)
		}

		return result
	}

	return "?" + string(format)
}
