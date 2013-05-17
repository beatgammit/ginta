/*
Provides function for messages containing variable, formatted values. Such
messages contain data that is both locale and execution-dependant. Therefore,
value strings may contain a variable-replacement marker, which is interpreted
by this package. This marker consists of the following
	{<variable-nr>[,format-id[,args...]]}

The variable nr is freely selectable, but may be no higher than the nr of arguments
provided to the invocation. The format ID is either omitted, or must refer to a format
passed to RegisterFormat. Formats defined in subpackages of this package are automatically 
registered. It is recommended that custom implementations also invoke RegisterFormat during
their initialization  

This package contains both the primitives for such formatted messages (Compile
and ApplyFormat), and some abstractions that increase quality-of-life for using
the package.
*/
package fmt

import (
	"bytes"
	i18n "code.google.com/p/ginta"
	"code.google.com/p/ginta/common"
	sysfmt "fmt"
	"strconv"
	"strings"
)

const (
	// Internal separator of format specifier parts
	FormatSegmentSeparator = ","
	// Resource key for errors "Bad Format specified"
	BadFormatResourceKey = "errors:bad_msg_format"
	// Referenced formatter is not registered
	UnknownFormatterResourceKey = "errors:unknown_msg_format"
	// Format was found, but the arguments were malformed in some way
	MalformedFormatSpecificationErrorResourceKey = "errors:bad_format_specification"
)

// An error message that, in addition to its normal string conversion, can be displayed in 
// different languages
type TranslatableError interface {
	error
	LocalError(i18n.Locale) string
}

// Converters modify format input values in arbitrary ways to conform to the needs of a format
type Converter interface {
	Convert(i18n.Locale, interface{}) interface{}
}

// Convenience declaration to allow passing plain functions as converters
type ConverterFunc func(i18n.Locale, interface{}) interface{}

func (c ConverterFunc) Convert(locale i18n.Locale, i interface{}) interface{} {
	return c(locale, i)
}

/*
The input to the system formatter specified by one format definition
*/
type MessageInput interface {
	FormatString() string
	Converter() Converter
}

// Simple message inputs provide no converter, and only a plain, immutable format string
// They are most suitable for any formats easily expressed in terms of SPrintf-verbs
type SimpleMessageInput string

// Returns the format string value
func (s SimpleMessageInput) FormatString() string {
	return string(s)
}

// Returns nil
func (s SimpleMessageInput) Converter() Converter {
	return nil
}

/*
A compiled message string. A message format can be used to format
any number of messages. An effort should be made to reuse these objects
when possible since parsing message definitions can add some degree of
unnecessary overhead.
*/
type MessageFormat struct {
	format          string
	argumentIndices []int
	converters      map[int]Converter
}

/*
Executes this format template, for a given input locale and arguments
*/
func (m *MessageFormat) Format(locale i18n.Locale, args ...interface{}) string {
	fmtArgs := make([]interface{}, len(m.argumentIndices))
	for i := range fmtArgs {
		arg := args[m.argumentIndices[i]]

		if converter, ok := m.converters[i]; ok {
			arg = converter.Convert(locale, arg)
		}

		fmtArgs[i] = arg
	}

	return sysfmt.Sprintf(m.format, fmtArgs...)
}

type errorResource struct {
	key       string
	arguments []interface{}
}

/*
returns a new error. The returned instance is guaranteed to 
TranslatableError in addition to the plain error interface.

The error created will display the formatted value (As per ApplyFormat)
of its resource, if such a resource exists. If it does not exist,
the translated error value for common.ResourceNotFoundResourceKey will be
displayed. I Feven that fails, a generic "resource not found" error is
printed
*/
func NewError(resourceKey string, args ...interface{}) error {
	rv := new(errorResource)
	rv.key = resourceKey
	rv.arguments = args

	return rv
}

func (err *errorResource) Error() string {
	return err.LocalError(i18n.DefaultLocale)
}

func (err *errorResource) LocalError(loc i18n.Locale) string {
	var str string
	var lookupErr error
	if str, lookupErr = loc.GetResource(err.key); err == nil {
		str = ApplyFormat(loc, str, err.arguments...)
	} else {
		if notFound, ok := lookupErr.(common.ResourceNotFoundError); ok {
			if notFoundTemplate, err2 := loc.GetResource(common.ResourceNotFoundResourceKey); err2 == nil {
				str = ApplyFormat(loc, notFoundTemplate, string(notFound))
			} else {
				str = notFound.Error()
			}
		} else {
			str = lookupErr.Error()
		}
	}

	return str
}

/*
Compiles and executes a format template with the given locale and arguments. This method
should be avoided in favor of Compile whenever possible since each invocation re-parses
the message template
*/
func ApplyFormat(locale i18n.Locale, template string, args ...interface{}) string {
	if len(args) > 0 {
		if t, err := Compile(template); err == nil {
			template = t.Format(locale, args)
		}
	}

	return template
}

/*
Compiles a format template into a format ready for execution. The
result may be saved and executed any number of times
*/
func Compile(template string) (*MessageFormat, error) {
	formatString := new(bytes.Buffer)
	argumentString := new(bytes.Buffer)
	buffer := formatString

	argumentMapping := make([]int, 0)
	converterMapping := make(map[int]Converter)

	for _, next := range []rune(template) {
		switch next {
		case '%':
			if buffer == formatString {
				buffer.WriteRune('%')
			}
		case '{':
			if buffer == formatString {
				buffer = argumentString
				continue
			}
		case '}':
			if buffer == argumentString {
				argumentDefinition := buffer.String()
				buffer.Reset()
				idx, input, err := parseArgument(argumentDefinition)

				if err != nil {
					return nil, err
				}

				argumentMapping = append(argumentMapping, idx)

				if converter := input.Converter(); converter != nil {
					converterMapping[idx] = converter
				}

				formatString.WriteString(input.FormatString())

				buffer = formatString

				continue
			}
		}

		buffer.WriteRune(next)
	}

	return &MessageFormat{formatString.String(), argumentMapping, converterMapping}, nil
}

func parseArgument(def string) (int, MessageInput, error) {

	if parts := strings.Split(def, FormatSegmentSeparator); len(parts) > 0 {
		for i, val := range parts {
			parts[i] = strings.Trim(val, " ")
		}

		if pos, err := strconv.Atoi(parts[0]); err == nil {
			if len(parts) > 1 {
				formatterName := parts[1]
				if factory, ok := registry[formatterName]; ok {
					result, err := factory.Compile(parts[2:])
					return pos, result, err
				} else {
					return -1, nil, NewError(UnknownFormatterResourceKey, parts[1])
				}
			} else {
				return pos, SimpleMessageInput("%v"), nil
			}
		}
	}

	return -1, nil, NewError(BadFormatResourceKey, def)
}

var registry map[string]FormatDefinition = make(map[string]FormatDefinition)

// Message formatters implement this interface.
type FormatDefinition interface {
	// given any extra arguments (after variable nr and format name), init a MessageInput
	// as defined by these options. Returns an error when bad options have been provided.
	Compile([]string) (MessageInput, error)
}

// Convenience to use plain functors as format definitions
type FormatDefinitionFunc func([]string) (MessageInput, error)

func (f FormatDefinitionFunc) Compile(args []string) (MessageInput, error) {
	return f(args)
}

// Registers a new format, given a unique name and a parser callback
func RegisterFormat(name string, def FormatDefinition) {
	registry[name] = def
}
