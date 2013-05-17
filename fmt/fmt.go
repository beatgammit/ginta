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
	FormatSegmentSeparator                       = ","
	BadFormatResourceKey                         = "errors:bad_msg_format"
	UnknownFormatterResourceKey                  = "errors:unknown_msg_format"
	MalformedFormatSpecificationErrorResourceKey = "errors:bad_format_specification"
)

type TranslatableError interface {
	error
	LocalError(i18n.Locale) string
}

type Converter interface {
	Convert(i18n.Locale, interface{}) interface{}
}

type ConverterFunc func(i18n.Locale, interface{}) interface{}

func (c ConverterFunc) Convert(locale i18n.Locale, i interface{}) interface{} {
	return c(locale, i)
}

type MessageInput interface {
	FormatString() string
	Converter() Converter
}

type SimpleMessageInput string

func (s SimpleMessageInput) FormatString() string {
	return string(s)
}

func (s SimpleMessageInput) Converter() Converter {
	return nil
}

type MessageFormat struct {
	format          string
	argumentIndices []int
	converters      map[int]Converter
}

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

func ApplyFormat(locale i18n.Locale, template string, args ...interface{}) string {
	if len(args) > 0 {
		if t, err := Compile(template); err == nil {
			template = t.Format(locale, args)
		}
	}

	return template
}

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

type FormatDefinition interface {
	Compile([]string) (MessageInput, error)
}

type FormatDefinitionFunc func([]string) (MessageInput, error)

func (f FormatDefinitionFunc) Compile(args []string) (MessageInput, error) {
	return f(args)
}

func RegisterFormat(name string, def FormatDefinition) {
	registry[name] = def
}
