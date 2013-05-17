/*
Defines the basic types of ginta, with their methods. These types are placed in
a separate package to allow access to them from any package, without introducing
unneccessary dependencies. 
*/

package common

import "strings"

/*
Describes a single language. A has one code (usually 2-letter ISO),
and a human-readable name.  The requirement for a human-readable name
implies that usually, there needs to be some kind of bootstrapping for any 
language provider (to map the language code to that name). Each provider
is responsible for its own method.
*/
type Language struct {
	Code, DisplayName string
}

/*
A single resource entry. The (possibly) hierarchical key is mapped
to a simple string value
*/

type Resource struct {
	Key   string
	Value string
}

const (
	/*
		Internal separator for a hierarchical key
	*/
	ResourceKeySegmentSeparator = ":"
	/*
		Resource key for all "resource not found" errors
	*/
	ResourceNotFoundResourceKey = "errors:resource_not_found"
	/*
		Resource key for looking up a languages display name. Probably 
		part of some bootstrapping process for a language 
	*/
	DisplayNameResourceKey = "internal:DisplayName"
)

/*
 Basic "resource not found" error type. These errors, in their default handling,
 will  print out the name of the resource not found. 
*/
type ResourceNotFoundError string

/*
Implements error interface by returning the resource name not
found
*/
func (r ResourceNotFoundError) Error() string {
	return string(r)
}

/*
A hierarchical key. Hierarchical keys allow the grouping of resources into packages, and
the conditional overriding of some entries. All keys in a package shadow those in super
packages. This is useful to allow brief and easily-remembered identifiers. 

Example (german): 
key=Schlüßel
keyboard.key=Taste
music.key=Tonart
*/

type HierarchicalKey string

/*
returns the key part of a hierarchical resource key. The key is the last segment
after a :-seperator
*/
func (k HierarchicalKey) Key() string {
	str := string(k)
	if idx := strings.LastIndex(str, ResourceKeySegmentSeparator); idx > -1 {
		return str[idx+1:]
	}

	return str
}

/*
Returns the prefix of a hierarchical resource key. The prefix is any leading sequence
of :-separated path elements
*/

func (k HierarchicalKey) Prefix() string {
	str := string(k)
	if idx := strings.LastIndex(str, ResourceKeySegmentSeparator); idx > -1 {
		return str[:idx]
	}

	return ""
}

/* Returns the hierarchical key, converted back to its string
format
*/

func (k HierarchicalKey) String() string {
	return string(k)
}

/*
Splits the hierarchical key into a prefix and the local key. This method
is equivalent to calling Prefix() and Key() individually, but may have better
performance. 
*/

func (k HierarchicalKey) Split() (string, string) {
	str := string(k)
	if idx := strings.LastIndex(str, ResourceKeySegmentSeparator); idx > -1 {
		return str[:idx], str[idx+1:]
	}

	return "", str
}

/*
Calculates the parent key. The parent key is calculated by removing the second-to-last
component from the hierarchical key. This is equivalent to "moving up a package". In case
the name has no hierarchy (resides in the root package), the empty key is returned.

Examples: 
	Parent("some:key:path:element") = "some.key.element"
	Parent("some:element") = "element"
	Parent("element") = ""
*/
func (k HierarchicalKey) Parent() HierarchicalKey {
	str := ""
	if prefix, key := k.Split(); prefix != "" {
		if prefix2 := HierarchicalKey(prefix).Prefix(); prefix2 != "" {
			str = prefix2 + ResourceKeySegmentSeparator + key
		} else {
			str = key
		}
	}

	return HierarchicalKey(str)
}
