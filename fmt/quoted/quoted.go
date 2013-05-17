// allows output of quoted-literal (%#v) values. Mainly useful for logging...
package quoted

import (
	"code.google.com/p/ginta/fmt"
)

// Format ID
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
