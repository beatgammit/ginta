// Simple static resource provider. Only really suitable for
// very small applications or testing scenarios
package simple

import (
	types "github.com/beatgammit/ginta/common"
)

// Language provider type 
type Provider map[string]*language
type language struct {
	DisplayName string
	Entries     map[string]string
}

// Allocates a new language provider
func New() Provider {
	return make(map[string]*language)
}

// Adds a language (by means of a key->value map) to the provider, and returns
// itself for call chaining
func (p Provider) AddLanguage(code, name string, entries map[string]string) Provider {
	p[code] = &language{
		DisplayName: name,
		Entries:     entries,
	}

	return p
}

func (f Provider) Enumerate() <-chan types.Language {
	c := make(chan types.Language)

	go func() {
		defer close(c)
		for code, val := range f {
			c <- types.Language{
				Code:        code,
				DisplayName: val.DisplayName,
			}
		}

	}()

	return c
}

func (f Provider) List(code string) <-chan types.Resource {
	c := make(chan types.Resource)

	if lang, ok := f[code]; ok {
		go func() {
			defer close(c)
			for key, val := range lang.Entries {
				c <- types.Resource{key, val}
			}
		}()
	} else {
		close(c)
	}

	return c
}
