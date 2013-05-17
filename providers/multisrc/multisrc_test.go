package multisrc

import (
	"bytes"
	"code.google.com/p/ginta/trunk/ginta/common"
	"io"
	"io/ioutil"
	"testing"
)

const (
	simpleContent   = "key1=val1\nkey2=val2\nkey3=val3\n"
	commentLines    = "#k1 is the primary key\nk1=nix\n\n#some\n#more\n#comments\n"
	inLineComment   = "k1=v1#First key\nk2=various\n"
	lineTermination = "long=This is a really long\\\n text\\nwith an internal line break"
)

func TestParseSimple(t *testing.T) {
	buff := ioutil.NopCloser(bytes.NewBuffer([]byte(simpleContent)))
	c := make(chan common.Resource)
	defer close(c)

	go ParseTo(buff, "", c)

	res := <-c

	if res.Key != "key1" || res.Value != "val1" {
		t.Error(res)
	}

	res = <-c

	if res.Key != "key2" || res.Value != "val2" {
		t.Error(res)
	}

	res = <-c

	if res.Key != "key3" || res.Value != "val3" {
		t.Error(res)
	}
}

func TestPrefix(t *testing.T) {
	buff := ioutil.NopCloser(bytes.NewBuffer([]byte(simpleContent)))
	c := make(chan common.Resource)
	defer close(c)

	go ParseTo(buff, "my:prefix:", c)

	res := <-c

	if res.Key != "my:prefix:key1" || res.Value != "val1" {
		t.Error(res)
	}

	res = <-c

	if res.Key != "my:prefix:key2" || res.Value != "val2" {
		t.Error(res)
	}

	res = <-c

	if res.Key != "my:prefix:key3" || res.Value != "val3" {
		t.Error(res)
	}
}

func TestParseMissingLastLineTerminator(t *testing.T) {
	buff := ioutil.NopCloser(bytes.NewBuffer([]byte(simpleContent[:len(simpleContent)-1])))
	c := make(chan common.Resource)
	defer close(c)

	go ParseTo(buff, "", c)

	res := <-c

	if res.Key != "key1" || res.Value != "val1" {
		t.Error(res)
	}

	res = <-c

	if res.Key != "key2" || res.Value != "val2" {
		t.Error(res)
	}

	res = <-c

	if res.Key != "key3" || res.Value != "val3" {
		t.Error(res)
	}
}

func TestParseCommentsAndEmptyLines(t *testing.T) {
	buff := ioutil.NopCloser(bytes.NewBuffer([]byte(commentLines)))
	c := make(chan common.Resource)
	defer close(c)

	go ParseTo(buff, "", c)

	res := <-c

	if res.Key != "k1" || res.Value != "nix" {
		t.Error(res)
	}
}

func TestParseInLineComments(t *testing.T) {
	buff := ioutil.NopCloser(bytes.NewBuffer([]byte(inLineComment)))
	c := make(chan common.Resource)
	defer close(c)

	go ParseTo(buff, "", c)

	res := <-c

	if res.Key != "k1" || res.Value != "v1" {
		t.Error(res)
	}

	res = <-c

	if res.Key != "k2" || res.Value != "various" {
		t.Error(res)
	}
}

func TestLineTermination(t *testing.T) {
	buff := ioutil.NopCloser(bytes.NewBuffer([]byte(lineTermination)))
	c := make(chan common.Resource)
	defer close(c)

	go ParseTo(buff, "", c)

	res := <-c

	if res.Key != "long" || res.Value != "This is a really long text\nwith an internal line break" {
		t.Error(res)
	}
}

func TestScanPipeline(t *testing.T) {
	out := make(chan common.Resource)
	in := make(chan ResourceSource)
	buffers := []io.ReadCloser{
		ioutil.NopCloser(bytes.NewBuffer([]byte(simpleContent))),
		ioutil.NopCloser(bytes.NewBuffer([]byte(commentLines))),
		ioutil.NopCloser(bytes.NewBuffer([]byte(inLineComment))),
		ioutil.NopCloser(bytes.NewBuffer([]byte(lineTermination))),
	}

	go func() {
		for _, buffer := range buffers {
			in <- ResourceSource{buffer, ""}
		}

		close(in)
	}()

	go list(in, out)

	res := <-out

	if res.Key != "key1" || res.Value != "val1" {
		t.Error(res)
	}

	res = <-out
	if res.Key != "key2" || res.Value != "val2" {
		t.Error(res)
	}

	res = <-out
	if res.Key != "key3" || res.Value != "val3" {
		t.Error(res)
	}

	res = <-out
	if res.Key != "k1" || res.Value != "nix" {
		t.Error(res)
	}

	res = <-out
	if res.Key != "k1" || res.Value != "v1" {
		t.Error(res)
	}

	res = <-out
	if res.Key != "k2" || res.Value != "various" {
		t.Error(res)
	}

	res = <-out
	if res.Key != "long" || res.Value != "This is a really long text\nwith an internal line break" {
		t.Error(res)
	}

	if _, ok := <-out; ok {
		t.Error("channel should be closed after last reader")
	}
}
