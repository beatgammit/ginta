/*
	The root of the ginta i18n library. Ginta is an i18n solution that provides both simple (name=value) lookups, context-dependant
	translations and support for formatted output.  
	
	Ginta is designed to be easy to use and extend. Easy to use insofar as that the user needs not worry about internals such
	as resource loading, synchronization and lifetime. Easy to extend insofar as that it should be easy to "plug in" new ways
	to acquire resources or formats.
	
	Ginta has two modes of query for a resource. A Resource can be queried, or it can be resolved. When a resource is
	queried, a direct lookup of the resource name is performed. If no resource with the specified name is known, an
	error is returned. In contrast, if a resource is resolved, its name is interpreted as a "hierarchical key" (See
	package common). Such a hierarchical key defines one or more super-keys. If a resource is not found under its
	hierarchical key, the system automatically walks through the super-keys until either a matching resource is located,
	or resolution fails even on the "root key", in which case an error is returned.
	
	This package contains the basic primitive functions of ginta. These functions are used to query the translation database for resource 
	entries, either individually or in bulk.	
*/

package ginta

import (
	types "code.google.com/p/ginta/common"
	"code.google.com/p/ginta/internal"
)

/*
	An Enumerator is any type that can enumerate one or more languages for which resources can be loaded. Result channel
	must be filled with all available languages, then closed.  An empty Enumerator should therefore return a closed channel.
*/

type Enumerator interface {
	/* Returns all languages available on this enumerator. The returned channel must be closed after the last language has been send */
	Enumerate() <-chan types.Language
}

/*
	A Lister is any type that can list the available resources (in some backing storage) for a given language code. After
	all available resources have been transmitted, the returned channel must be closed.  
*/
type Lister interface {
	/* Returns all resources for the language code available on this source. The returned channel must be closed after the last resource has been send */
	List(code string) <-chan types.Resource
}

/*
	The full service provider interface. Implementions are able to provide a number of lanugages (and their contents). See the 
	package providers for some samples. 
*/

type LanguageProvider interface {
	// enumeration functionality
	Enumerator
	// list functionality
	Lister
}

/*
	Locale defines methods to access resources for a language
*/

type Locale string

/*
	The default locale of the application - may be changed
	by the application. 
*/
var DefaultLocale Locale = Locale("en")

/*
	Adds a language provider to the system. Often there will be 
	a single provider, but there may be more. In case of multiple
	providers defining the same language, their definitions will
	be merged (that is, their resource sets combined). In this case,
	providers registered later will overwrite these defined
	earlier. 
*/

func Register(p LanguageProvider) {
	internal.Register(p.Enumerate(), func(code string) <-chan types.Resource {
		resource := p.List(code)
		return resource
	})
}

/*
	Lists all currently known languages 
*/

func List() []*types.Language {
	return internal.List()
}

/*
	Resolves a resource by its hierarchical key. 
*/
func (l Locale) ResolveResource(k types.HierarchicalKey) (string, error) {
	locale := string(l)
	internal.Activate(locale)
	return internal.Request(locale, string(k), true)
}

/*
	Returns a resource by simple name matching
*/

func (l Locale) GetResource(key string) (string, error) {
	locale := string(l)
	internal.Activate(locale)
	return internal.Request(locale, key, false)
}

/*
	Returns a "resource bundle". This bundle is contains all resource whose hierarchical key has
	exactly the specified prefix - but no resources with shorter or longer prefix paths.
*/
func (l Locale) GetResourceBundle(prefix string) map[string]string {
	locale := string(l)
	internal.Activate(locale)
	return internal.RequestBundle(locale, prefix, false)
}

/*
	Returns the combination of this resource bundle and all its parent bundles. The resources
	defined in a child are not overwritten by its parent.
*/

func (l Locale) ResolveResourceBundle(prefix string) map[string]string {
	locale := string(l)
	internal.Activate(locale)
	return internal.RequestBundle(locale, prefix, true)
}
