package zip

import (
	"code.google.com/p/ginta"
	"code.google.com/p/ginta/common"
	"code.google.com/p/ginta/providers/multisrc"
	
	"archive/zip"
	"strings"
)

const (
	boostrapFile = "bootstrap.txt"
	pathSeps = "/\\:"
)

type zipFilePath string

func New(path string) ginta.LanguageProvider {
	zPath := zipFilePath(path)
	return &multisrc.Provider {
		zPath,
		zPath,	
	}
}

func (p zipFilePath) Enumerate() <-chan common.Language {
	c := make(chan common.Language)
	go func() {
		defer close(c)
		if r, err := zip.OpenReader(string(p)); err == nil {
			defer r.Close()
			
			for _, file := range r.File {
				if  file.Name == boostrapFile {
					if stream, err := file.Open(); err == nil {
						defer stream.Close()
						languages := make(chan common.Resource)
						go func () {
							defer close(languages)
							multisrc.ParseTo(stream, "", languages)
						}()
					
						for lang := range languages {
							c <- common.Language {
								Code: lang.Key,
								DisplayName: lang.Value,
							}
						}				
					} 
				} 
			}
		}	
	}()
	
	return c
} 

func (p zipFilePath) Walk(code string) <-chan *multisrc.ResourceSource {
	c:= make(chan *multisrc.ResourceSource)
	
	go func() {
		defer close(c)
		codeLen := len(code)
		
		if r, err := zip.OpenReader(string(p)); err == nil {
			for _, file := range r.File {
				if len(file.Name) > codeLen + 1 && file.Name[:codeLen] == code && strings.Contains(pathSeps, file.Name[codeLen:codeLen + 1]) {
					if rc, err := file.Open(); err == nil {
						prefix := file.Name[codeLen + 1:]
						
						if last := strings.LastIndexAny(prefix, pathSeps); last >= 0 {
							prefix = prefix[:last]
							
							for _, r := range []rune(pathSeps) {
								prefix = strings.Replace(prefix, string([]rune{r}), common.ResourceKeySegmentSeparator, -1)
							}
							
							prefix = prefix + common.ResourceKeySegmentSeparator
						} else {
							prefix = ""
						}
						
						
						c <- &multisrc.ResourceSource {
							Reader: rc,
							Prefix: prefix,
						}
					}
				} 
			}
		}
	}()
	
	return c
}