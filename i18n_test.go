package ginta

import (
	types "code.google.com/p/ginta/trunk/ginta/common"
	"reflect"
	"testing"
)

type mockProviderEmpty string
type mockProviderSingle struct {
	code, key, value string
}

type mockProviderMap struct {
	code string
	vals map[string]string
}

func (m mockProviderEmpty) Enumerate() <-chan types.Language {
	c := make(chan types.Language)

	go func() {
		c <- types.Language{string(m), string(m)}
		close(c)
	}()

	return c
}

func (m mockProviderEmpty) List(_ string) <-chan types.Resource {
	c := make(chan types.Resource)
	close(c)

	return c
}

func (m *mockProviderSingle) Enumerate() <-chan types.Language {
	c := make(chan types.Language)

	go func() {
		c <- types.Language{m.code, m.code}
		close(c)
	}()

	return c
}

func (m *mockProviderSingle) List(_ string) <-chan types.Resource {
	c := make(chan types.Resource)

	go func() {
		c <- types.Resource{m.key, m.value}
		close(c)
	}()

	return c
}

func (m *mockProviderMap) Enumerate() <-chan types.Language {
	c := make(chan types.Language)

	go func() {
		c <- types.Language{m.code, m.code}
		close(c)
	}()

	return c
}

func (m *mockProviderMap) List(_ string) <-chan types.Resource {
	c := make(chan types.Resource)

	go func() {
		for key, val := range m.vals {
			c <- types.Resource{key, val}
		}
		close(c)
	}()

	return c
}

func TestRegisterLanguage(t *testing.T) {
	Register(mockProviderEmpty("l1"))

	if list := List(); len(list) != 1 || list[0].Code != "l1" {
		t.Error(list)
	}
}

func TestSimpleGet(t *testing.T) {
	Register(&mockProviderSingle{"l2", "key1", "val1"})

	if result, err := Locale("l2").GetResource("key1"); err != nil || result != "val1" {
		t.Error(result, err)
	}
}

func TestRegisterDoesNotEvict(t *testing.T) {
	Register(&mockProviderSingle{"l3", "key1", "val1"})

	Register(mockProviderEmpty("l3"))

	if result, err := Locale("l3").GetResource("key1"); err != nil || result != "val1" {
		t.Error(result, err)
	}
}

func TestRegisterDoesNotBleedOver(t *testing.T) {
	Register(&mockProviderSingle{"l4", "key1", "val1"})

	Register(mockProviderEmpty("l5"))

	result, err := Locale("l5").GetResource("key1")
	if err == nil {
		t.Error(result, err)
	}

	if _, ok := err.(types.ResourceNotFoundError); !ok {
		t.Error(reflect.TypeOf(err))
	}
}

func TestProvidersAdditive(t *testing.T) {
	keyValMap := map[string]string{
		"key1":         "val1",
		"key2":         "fruits",
		"k3":           "herbs",
		"fjfjjfjf":     "panem et circensis",
		"lucky Ümläut": "€uro",
	}

	for key, val := range keyValMap {
		Register(&mockProviderSingle{"l6", key, val})
	}

	l := Locale("l6")
	for key, val := range keyValMap {
		if str, err := l.GetResource(key); err != nil || str != val {
			t.Error(str, err)
		}
	}
}

func TestProviderMultiContent(t *testing.T) {
	keyValMap := map[string]string{
		"aswoj":       "akpfwe+",
		"aesogjwaeoj": "#wef",
		"we4u":        "34tz08",
		"dejejo":      "$T§;",
		"wegrojg":     "€egr",
	}

	Register(&mockProviderMap{"l7", keyValMap})

	l := Locale("l7")
	for key, val := range keyValMap {
		if str, err := l.GetResource(key); err != nil || str != val {
			t.Error(str, err)
		}
	}
}

func TestGetNonHierarchical(t *testing.T) {
	keyValMap := map[string]string{
		"x":   "efhwefeh",
		"a:x": "akpfwe+",
		"b:x": "#wef",
	}

	Register(&mockProviderMap{"l8", keyValMap})

	if str, err := Locale("l8").GetResource("c:x"); err == nil {
		t.Error(str, err)
	}
}

func TestGetHierarchical(t *testing.T) {
	keyValMap := map[string]string{
		"x":   "efhwefeh",
		"a:x": "akpfwe+",
		"b:x": "#wef",
	}

	Register(&mockProviderMap{"l9", keyValMap})

	if str, err := Locale("l9").ResolveResource("c:x"); err != nil || str != "efhwefeh" {
		t.Error(str, err)
	}
}

func TestListNonHierarchical(t *testing.T) {
	keyValMap := map[string]string{
		"x":           "efhwefeh",
		"jim":         "bob",
		"a:x":         "akpfwe+",
		"a:y":         "hewfehieihwewif",
		"a:z":         "hfeewfewifewihe",
		"a:nox:poo":   "e39u2u29u",
		"b:something": "#wef",
	}

	expect := map[string]string{
		"x": "akpfwe+",
		"y": "hewfehieihwewif",
		"z": "hfeewfewifewihe",
	}

	Register(&mockProviderMap{"l10", keyValMap})

	l := Locale("l10")

	bundle := l.GetResourceBundle("a")

	if len(expect) != len(bundle) {
		t.Error(bundle)
	}

	for key, val := range expect {
		if v2, ok := bundle[key]; !ok || v2 != val {
			t.Error(key, val, v2, ok)
		}
	}

	bundle = l.GetResourceBundle("q")

	if 0 != len(bundle) {
		t.Error(bundle)
	}
}

func TestListHierarchical(t *testing.T) {
	keyValMap := map[string]string{
		"x":           "2efhwefeh",
		"jim":         "bob",
		"a:x":         "2akpfwe+",
		"a:y":         "2hewfehieihwewif",
		"a:z":         "2hfeewfewifewihe",
		"a:nox:poo":   "2e39u2u29u",
		"b:something": "2#wef",
	}

	expect := map[string]string{
		"x":   "2akpfwe+",
		"jim": "bob",
		"y":   "2hewfehieihwewif",
		"z":   "2hfeewfewifewihe",
	}

	Register(&mockProviderMap{"l11", keyValMap})

	l := Locale("l11")

	bundle := l.ResolveResourceBundle("a")

	if len(expect) != len(bundle) {
		t.Error(bundle)
	}

	for key, val := range expect {
		if v2, ok := bundle[key]; !ok || v2 != val {
			t.Error(key, val, v2, ok)
		}
	}

	expect = map[string]string{"x": "2efhwefeh", "jim": "bob"}

	bundle = l.ResolveResourceBundle("q")

	if len(expect) != len(bundle) {
		t.Error(bundle)
	}

	for key, val := range expect {
		if v2, ok := bundle[key]; !ok || v2 != val {
			t.Error(key, val, v2, ok)
		}
	}
}
