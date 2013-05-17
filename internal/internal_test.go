package internal

import (
	"code.google.com/p/ginta/common"
	"testing"
	"time"
)

func sendTestLanguage(t *testing.T, l chan<- common.Language, code, name string) {
	l <- common.Language{
		Code:        code,
		DisplayName: name,
	}
	close(l)
}

func sendMap(t *testing.T, m map[string]string) func(key string) <-chan common.Resource {
	local := m

	fill := func(c chan<- common.Resource) {
		defer close(c)
		if local != nil {
			for key, entry := range local {
				c <- common.Resource{key, entry}
			}
		}
	}

	f := func(_ string) <-chan common.Resource {
		c := make(chan common.Resource)
		go fill(c)
		return c
	}

	return f
}

func TestSingleRegisterLanguage(t *testing.T) {
	l := make(chan common.Language)

	go sendTestLanguage(t, l, "t1", "Testing 1")

	Register(l, func(key string) <-chan common.Resource {
		t.FailNow()
		return nil
	})

	internalPtr := universe["t1"]
	if internalPtr == nil ||
		internalPtr.displayName != "Testing 1" ||
		len(internalPtr.pendingFetches) != 1 ||
		internalPtr.runningFetches != 0 {
		t.FailNow()
	}
}

func TestRegisterAndActivateEmpty(t *testing.T) {
	l := make(chan common.Language)

	go sendTestLanguage(t, l, "t2", "Testing 2")

	Register(l, sendMap(t, nil))
	Activate("t2")

	internalPtr := universe["t2"]
	if internalPtr == nil ||
		internalPtr.displayName != "Testing 2" ||
		len(internalPtr.pendingFetches) != 0 ||
		internalPtr.runningFetches != 0 {
		t.Log("Bad internal result: ", internalPtr)
		t.FailNow()
	}
}

func TestRegisterAndActivateWithSomeValues(t *testing.T) {
	l := make(chan common.Language)

	go sendTestLanguage(t, l, "t3", "Testing 3")

	Register(l, sendMap(t, map[string]string{
		"a": "aaa",
		"b": "abc",
	}))

	l = make(chan common.Language)

	go sendTestLanguage(t, l, "t3", "Testing 3")

	Register(l, sendMap(t, map[string]string{
		"c": "xxx",
		"d": "xyz",
	}))

	Activate("t3")

	internalPtr := universe["t3"]
	if internalPtr == nil ||
		internalPtr.displayName != "Testing 3" ||
		len(internalPtr.pendingFetches) != 0 ||
		internalPtr.runningFetches != 0 {
		t.Log("Bad internal result: ", internalPtr)
		t.FailNow()
	}

	if internalPtr.entries[""]["a"] != "aaa" ||
		internalPtr.entries[""]["b"] != "abc" ||
		internalPtr.entries[""]["c"] != "xxx" ||
		internalPtr.entries[""]["d"] != "xyz" {
		t.Log("Entries are ", internalPtr.entries)
		t.FailNow()
	}
}

func TestActivateTwice(t *testing.T) {
	l := make(chan common.Language)

	go sendTestLanguage(t, l, "t4", "Testing 4")

	Register(l, sendMap(t, map[string]string{
		"a": "aaa",
		"b": "abc",
	}))

	Activate("t4")
	Activate("t4")

	internalPtr := universe["t4"]
	if internalPtr == nil ||
		internalPtr.displayName != "Testing 4" ||
		len(internalPtr.pendingFetches) != 0 ||
		internalPtr.runningFetches != 0 {
		t.Log("Bad internal result: ", internalPtr)
		t.FailNow()
	}

	if internalPtr.entries[""]["a"] != "aaa" ||
		internalPtr.entries[""]["b"] != "abc" {
		t.Log("Entries are ", internalPtr.entries)
		t.FailNow()
	}
}

func TestBlockConcurringActivate(t *testing.T) {
	l := make(chan common.Language)

	go sendTestLanguage(t, l, "t4a", "Testing 4a")

	f := func(key string) <-chan common.Resource {
		c := make(chan common.Resource)

		go func() {
			c <- common.Resource{"q", "b"}
			time.Sleep(time.Second / 10)
			c <- common.Resource{"w", "c"}
			time.Sleep(time.Second / 10)
			c <- common.Resource{"e", "d"}
			time.Sleep(time.Second / 10)
			close(c)
		}()

		return c
	}

	for i := 0; i < 100; i++ {

		if i%50 == 0 {
			Register(l, f)
		}

		go Activate("t4a")

		time.Sleep(10 * time.Millisecond)
	}
}

func TestGetResource(t *testing.T) {
	l := make(chan common.Language)

	go sendTestLanguage(t, l, "t5", "Testing 5")

	Register(l, sendMap(t, map[string]string{
		"a": "aaa",
		"b": "abc",
	}))

	var val string
	var err error

	Activate("t5")

	val, err = Request("t5", "a", false)

	if val != "aaa" || err != nil {
		t.Errorf("Got %v:%v, with entry %#v\n", val, err, universe["t5"])
	}

	val, err = Request("t5", "b", false)

	if val != "abc" || err != nil {
		t.Errorf("Got %v:%v, with entry %#v\n", val, err, universe["t5"])
	}
}

func TestUpdateResource(t *testing.T) {
	l := make(chan common.Language)

	go sendTestLanguage(t, l, "t6", "Testing 6")

	Register(l, sendMap(t, map[string]string{
		"a": "aaa",
		"b": "abc",
	}))
	Activate("t6")
	val, err := Request("t6", "a", false)

	if val != "aaa" || err != nil {
		t.Errorf("Got %v:%v, with entry %#v\n", val, err, universe["t5"])
	}

	Update("t6", "b", "any")

	val, err = Request("t6", "b", false)

	if val != "any" || err != nil {
		t.Errorf("Got %v:%v, with entry %#v\n", val, err, universe["t5"])
	}
}

func TestList(t *testing.T) {
	m := make(map[string]string)
	for _, value := range List() {
		m[value.Code] = value.DisplayName
	}

	if len(m) != 7 {
		t.Log(len(m))
		t.FailNow()
	}

	for key := range m {
		if key != "t1" &&
			key != "t2" &&
			key != "t3" &&
			key != "t4" &&
			key != "t4a" &&
			key != "t5" &&
			key != "t6" {
			t.Log(key)
			t.FailNow()
		}
	}
}
