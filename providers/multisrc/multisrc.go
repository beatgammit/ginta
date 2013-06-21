/*
Parser for "resource-file" like inputs. This package does not define an own 
resource provider, but instead exposes functions that can be used by different
providers that read some lower-level format.
*/
package multisrc

import (
	"bufio"
	"bytes"
	"code.google.com/p/ginta"
	"code.google.com/p/ginta/common"
	"io"
	"strings"
)

const trim = " \t\r\n"

/*
A resource source is defined by being able to be read, and an info about a
common prefix to all resources encountered
*/
type ResourceSource struct {
	Reader io.ReadCloser
	Prefix string
}

/*
A walker can enumerate resource sources for a given language code
*/
type Walker interface {
	// List all resource sources for a language. Once the last resource
	// has been transferred, the channel must be closed 
	Walk(code string) <-chan *ResourceSource
}

// Convenience walker for simple functions
type WalkerFunc func(string) <-chan *ResourceSource

func (f WalkerFunc) Walk(code string) <-chan *ResourceSource {
	return f(code)
}

// Simplified provider interface - needs only provide a walker and
// a resource enumerator
type Provider struct {
	Enumerator ginta.Enumerator
	Walker     Walker
}

func (p *Provider) Enumerate() <-chan common.Language {
	return p.Enumerator.Enumerate()
}

// imports each resource in the providers Walker,  and calls ParseTo
// for each resource file found
func (p *Provider) List(code string) <-chan common.Resource {
	c := make(chan common.Resource)
	go list(p.Walker.Walk(code), c)
	return c
}

func list(in <-chan *ResourceSource, target chan<- common.Resource) {
	defer close(target)
	
	i := 0
	done := make(chan int)
	defer close(done)
	
	for input := range in {
		input := input
		i++
		
		go func() {
			defer input.Reader.Close()
			ParseTo(input.Reader, input.Prefix, target)
			done <- 0
		}()
	}
	
	for i > 0 {
		<- done
		i--
	}
}

type runeWriter interface {
	WriteRune(rune) (int, error)
}

type runeDrop int

func (_ runeDrop) WriteRune(_ rune) (int, error) {
	return 0, nil
}

/*
Reads an input, and parses it into key-value pairs that define resources. The file format
is interpreted the following way:
	\n newline in the string value
	\  join lines if trailing
	\\ literal backslash
	#  treat the rest of the line as a comment

Lines are terminated by newlines (0x0a). Resource key names are separated from their values
by = characters. Keys may not contain additional equals characters, but values may.
*/
func ParseTo(inRaw io.ReadCloser, prefix string, target chan<- common.Resource) {
	in := bufio.NewReader(inRaw)

	var key, val bytes.Buffer
	var err error
	var nextRune rune

	// all possible targets never fail at a write, so checking writes is not necessary
	var buffer runeWriter = &key
	backslash := false
	for nextRune, _, err = in.ReadRune(); err == nil; nextRune, _, err = in.ReadRune() {
		if backslash {
			backslash = false
			switch nextRune {
			case 'n':
				nextRune = '\n'
			// this allows the trailing backslash to join lines 
			case '\n':
				continue
			default:
				buffer.WriteRune('\\')
			}
		} else {
			switch nextRune {
			case '\\':
				backslash = true
				continue
			case '#':
				buffer = runeDrop(0)
				continue
			case '=':
				if buffer == &key {
					buffer = &val
				}
				continue
			case '\n':
				transmitValid(prefix, &key, &val, target)
				key.Reset()
				val.Reset()
				buffer = &key
				continue
			}
		}

		buffer.WriteRune(nextRune)
	}

	if err == io.EOF {
		transmitValid(prefix, &key, &val, target)
	}
}

func transmitValid(prefix string, key, val *bytes.Buffer, target chan<- common.Resource) {

	keyStr := strings.Trim(key.String(), trim)
	valStr := strings.Trim(val.String(), trim)

	if keyStr != "" && valStr != "" {
		target <- common.Resource{prefix + keyStr, valStr}
	}
}
