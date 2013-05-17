package fmt

import (
	i18n "code.google.com/p/ginta"
	"code.google.com/p/ginta/common"
)

/*
Resolvers intended for advanced message operations. They provide a simplified access to resources, by
defining a base hierarchical path relative to which all resources will be located. Additionally, they
automatically perform formatting operations on invocation. 
*/

type Resolver struct {
	// The base path relative to which all resources will be located. 
	Base string
	// This type contains unexported fields
	locale i18n.Locale
	cache  map[string]*MessageFormat
}

/*
Initializes a new resolver with the specified locale, and base path
*/
func NewResolver(locale i18n.Locale, base string) *Resolver {
	return &Resolver{base, locale, map[string]*MessageFormat{}}
}

/*
Retrieves a resource, relative to the base path of this resolver, and performs variable
expansion on the retrieved string, using the specified argument list.
*/
func (r *Resolver) Format(key string, args ...interface{}) (result string, err error) {
	key = r.Base + common.ResourceKeySegmentSeparator + key
	fmt, ok := r.cache[key]
	if !ok {
		var str string
		if str, err = r.locale.GetResource(key); err == nil {
			fmt, err = Compile(str)
			r.cache[key] = fmt
		} else {
			return
		}
	}

	result = fmt.Format(r.locale, args...)
	return
}
