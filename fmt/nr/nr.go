/*
	Detailed formatter for integers. This formatter allows output of signed or padded integers. Inputs must be valid for
	the fmt.Print... family of functions.

	The format has several flags that control layout. In addition to these flags (defined as constants here), a single numeric argument
	may be provided in an arbitrary position. The presence of a numeric arguments indicates to pad the output to the specified length. 

	Example:
		Format				Input		Output
		{0,nr}					4			4
		{0,nr}					-4			-4
		{0,nr,5}				4			    4
		{0,nr,5}				4			   -4
		{0,nr,5}				1234		 1234
		{0,nr,sign}				9			+9
		{0,nr,5,sign}			17			  +17
		{0,nr,5,padding}		1			00001
		{0,nr,padding}			28			28
		{0,nr,5,padding,sign}	432			+0432		
*/
package nr

import (
	"bytes"
	"github.com/beatgammit/ginta/fmt"
	"strconv"
)

const (
	// Format argument: zero-pad the output to the specified length
	PadZero = "padding"
	// Format argument: add a sign even for positive values
	Sign = "sign"
	// Format id
	Format = "nr"
)

// installs this format - should be called at the very start of the program, prior to registring
// the first provider.
func Install() {
	fmt.RegisterFormat(Format, fmt.FormatDefinitionFunc(parse))
}

func parse(args []string) (fmt.MessageInput, error) {
	b := new(bytes.Buffer)
	var length string
	var sign, pad bool

	for _, str := range args {
		lengthVal, err := strconv.Atoi(str)
		switch {
		case str == PadZero:
			pad = true
		case str == Sign:
			sign = true
		case err == nil && lengthVal > 0:
			length = str
		default:
			return nil, fmt.NewError(fmt.MalformedFormatSpecificationErrorResourceKey, Format, str)
		}
	}

	b.WriteRune('%')

	if sign {
		b.WriteRune('+')
	}

	if pad {
		b.WriteRune('0')
	}

	if length != "" {
		b.WriteString(length)
	}

	b.WriteRune('d')

	return fmt.SimpleMessageInput(b.String()), nil
}
