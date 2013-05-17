package fs

import (
	"bytes"
	"code.google.com/p/ginta"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

const (
	language               = "/en"
	path1                  = language + "/test"
	path2                  = language + "/a/longer/path"
	file1                  = path1 + "/errors1.txt"
	file2                  = path1 + "/errors2.txt"
	file3                  = path2 + "/content.txt"
	bootstrap_path         = language + bootstrapExtension
	errors1_txt_contents   = "err_something_went_wrong=General Error\n"
	errors2_txt_contents   = "err_file_not_found=Its gone!\nerr_no_space_left=No space left on device\n"
	content_txt_contents   = "greeting=Hello World\n"
	bootstrap_txt_contents = "internal:DisplayName=English\n"
	filePermissions        = os.FileMode(0600)
	dirPermissions         = os.FileMode(0700)
)

/*	creates the following layout:
	en/
	+- bootstrap.txt (1 assignment - the display name)
	+- test/
	 +- errors1.txt (1 assignment)
	 +- errors2.txt (2 assignments)
	+- a/
	 +- longer/
	  +- path/
	   +- content.txt (1 assignments)
*/

func dumpFile(path, contents string) error {
	var ptr *os.File
	var err error
	if ptr, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY, filePermissions); err == nil {
		defer ptr.Close()

		_, err = io.Copy(ptr, bytes.NewBuffer([]byte(contents)))
	}

	return err
}

func prepare(prefix string, t *testing.T) string {
	root, err := ioutil.TempDir("", prefix)
	t.Logf("Building entry root@%s", root)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if err = os.MkdirAll(root+path1, dirPermissions); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if err = os.MkdirAll(root+path2, dirPermissions); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if err = dumpFile(root+bootstrap_path, bootstrap_txt_contents); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if err = dumpFile(root+file1, errors1_txt_contents); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if err = dumpFile(root+file2, errors2_txt_contents); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if err = dumpFile(root+file3, content_txt_contents); err != nil {
		t.Error(err)
		t.FailNow()
	}

	return root
}

func scrub(prefix string, t *testing.T) {
	if err := os.RemoveAll(prefix); err != nil {
		t.Error(err)
	}
}

func TestBootstrapFile(t *testing.T) {
	dir := prepare("t1", t)
	defer scrub(dir, t)

	c := New(dir).Enumerate()

	lang, ok := <-c

	if !ok || lang.Code != "en" || lang.DisplayName != "English" {
		t.Error(lang, ok)
	}
}

func TestMissingBootstrapFile(t *testing.T) {
	dir := prepare("t2", t)
	defer scrub(dir, t)

	os.Remove(dir + bootstrap_path)

	c := New(dir).Enumerate()

	lang, ok := <-c

	if !ok || lang.Code != "en" || lang.DisplayName != "en" {
		t.Error(lang, ok)
	}
}

func TestDescendDirTree(t *testing.T) {
	dir := prepare("t3", t)
	defer scrub(dir, t)

	expect := map[string]string{
		"test:err_something_went_wrong": "General Error",
		"test:err_file_not_found":       "Its gone!",
		"test:err_no_space_left":        "No space left on device",
		"a:longer:path:greeting":        "Hello World",
		"internal:DisplayName":          "English",
	}

	ginta.Register(New(dir))
	locale := ginta.Locale("en")

	for key, expectedVal := range expect {
		if actualVal, err := locale.GetResource(key); err != nil {
			t.Error(err)
		} else if expectedVal != actualVal {
			t.Errorf("Expected %s but got %s", expectedVal, actualVal)
		}
	}
}
