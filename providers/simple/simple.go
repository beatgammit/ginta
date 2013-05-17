package simple

import (
	types "code.google.com/p/ginta/trunk/ginta/common"
)

type Provider map[string]*Language
type Language struct {
	DisplayName string
	Entries     map[string]string
}

func New() Provider {
	return make(map[string]*Language)
}

func (p Provider) AddLanguage(code, name string, entries map[string]string) Provider {
	p[code] = &Language{
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
