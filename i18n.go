package ginta

import (
	types "code.google.com/p/ginta/common"
	"code.google.com/p/ginta/internal"
)

type Enumerator interface {
	Enumerate() <-chan types.Language
}

type Lister interface {
	List(code string) <-chan types.Resource
}

/*
Service provider interface. 
*/

type LanguageProvider interface {
	Enumerator
	Lister
}

/*
Locale defines methods to access resources for a language
*/

type Locale string

/*
The default locale of the application - may be changed
by the application
*/
var DefaultLocale Locale = Locale("en")

/*
Adds a language provider to the system. Often there will be 
a single provider, but there may be more. In case of multiple
providers defining the same language, their definitions will
be merged (that is, their resource sets combined). In this case,
providers registered later will overwrite these defined
earlier 
*/

func Register(p LanguageProvider) {
	internal.Register(p.Enumerate(), func(code string) <-chan types.Resource {
		resource := p.List(code)
		return resource
	})
}

func List() []*types.Language {
	return internal.List()
}

func (l Locale) ResolveResource(k types.HierarchicalKey) (string, error) {
	locale := string(l)
	internal.Activate(locale)
	return internal.Request(locale, string(k), true)
}

func (l Locale) GetResource(key string) (string, error) {
	locale := string(l)
	internal.Activate(locale)
	return internal.Request(locale, key, false)
}

func (l Locale) GetResourceBundle(prefix string) map[string]string {
	locale := string(l)
	internal.Activate(locale)
	return internal.RequestBundle(locale, prefix, false)
}

func (l Locale) ResolveResourceBundle(prefix string) map[string]string {
	locale := string(l)
	internal.Activate(locale)
	return internal.RequestBundle(locale, prefix, true)
}
