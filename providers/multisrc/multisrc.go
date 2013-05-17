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

type ResourceSource struct {
	Reader io.ReadCloser
	Prefix string
}

type Walker interface {
	Walk(code string) <-chan ResourceSource
}

type WalkerFunc func(string) <-chan ResourceSource

func (f WalkerFunc) Walk(code string) <-chan ResourceSource {
	return f(code)
}

type Provider struct {
	Enumerator ginta.Enumerator
	Walker     Walker
}

func (p *Provider) Enumerate() <-chan common.Language {
	return p.Enumerator.Enumerate()
}

func (p *Provider) List(code string) <-chan common.Resource {
	c := make(chan common.Resource)
	go list(p.Walker.Walk(code), c)
	return c
}

func list(in <-chan ResourceSource, target chan<- common.Resource) {
	defer close(target)
	for input := range in {
		ParseTo(input.Reader, input.Prefix, target)
	}
}

type runeWriter interface {
	WriteRune(rune) (int, error)
}

type runeDrop int

func (_ runeDrop) WriteRune(_ rune) (int, error) {
	return 0, nil
}

func ParseTo(inRaw io.ReadCloser, prefix string, target chan<- common.Resource) {
	defer inRaw.Close()
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
