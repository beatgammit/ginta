package quoted

import (
	"code.google.com/p/ginta/fmt"
)

const FormatName = "quoted"

func init() {
	fmt.RegisterFormat(FormatName, fmt.FormatDefinitionFunc(parse))
}

func parse(args []string) (fmt.MessageInput, error) {
	if len(args) != 0 {
		return nil, fmt.NewError(fmt.MalformedFormatSpecificationErrorResourceKey, args)
	}

	return fmt.SimpleMessageInput("%#v"), nil
}
