// allows output of quoted-literal (%#v) values. Mainly useful for logging...
package quoted

import (
	"github.com/beatgammit/ginta/fmt"
)

// Format ID
const FormatName = "quoted"

// installs this format - should be called at the very start of the program, prior to registring
// the first provider.
func Install() {
	fmt.RegisterFormat(FormatName, fmt.FormatDefinitionFunc(parse))
}

func parse(args []string) (fmt.MessageInput, error) {
	if len(args) != 0 {
		return nil, fmt.NewError(fmt.MalformedFormatSpecificationErrorResourceKey, args)
	}

	return fmt.SimpleMessageInput("%#v"), nil
}
