package fs

import (
	"code.google.com/p/ginta"
	types "code.google.com/p/ginta/common"
	multi "code.google.com/p/ginta/providers/multisrc"
	"io/ioutil"
	"os"
	"path/filepath"
)

type provider string

func New(path string) ginta.LanguageProvider {
	return &multi.Provider{provider(path), provider(path)}
}

const (
	cutChars           = " \r\n"
	keyValueSep        = '='
	lineCommentChar    = '#'
	bootstrapExtension = "/bootstrap.txt"
)

func (f provider) Enumerate() <-chan types.Language {
	c := make(chan types.Language)

	go enumerate(string(f), c)

	return c
}

func enumerate(baseDir string, target chan<- types.Language) {
	defer close(target)

	if entries, err := ioutil.ReadDir(baseDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				displayName, err := parseBootstrap(baseDir + "/" + entry.Name())

				if err != nil {
					displayName = entry.Name()
				}
				target <- types.Language{entry.Name(), displayName}
			}
		}
	}
}

func parseBootstrap(dir string) (string, error) {
	c := make(chan types.Resource)
	go func() {
		if file, err := open(dir + bootstrapExtension); err == nil {
			multi.ParseTo(file, "", c)
		}

		close(c)
	}()

	for res := range c {
		if res.Key == types.DisplayNameResourceKey {
			return res.Value, nil
		}
	}

	return "", types.ResourceNotFoundError(types.DisplayNameResourceKey)
}

func (f provider) Walk(code string) <-chan multi.ResourceSource {
	c := make(chan multi.ResourceSource)
	go func() {
		defer close(c)
		list(string(f)+"/"+code, "", c)
	}()

	return c
}

func list(dirPath string, prefix string, target chan<- multi.ResourceSource) {
	if entries, err := ioutil.ReadDir(filepath.FromSlash(dirPath)); err == nil {
		for _, file := range entries {
			name := dirPath + "/" + file.Name()
			if file.IsDir() {
				list(name, prefix+file.Name()+types.ResourceKeySegmentSeparator, target)
			} else if file, err := open(name); err == nil {
				target <- multi.ResourceSource{file, prefix}
			}
		}
	}
}

func open(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY, os.FileMode(0666))
}
