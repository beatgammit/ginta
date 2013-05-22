// The setup package makes it easy and convenient to bootstrap the ginta library
package setup

import (
	"code.google.com/p/ginta"
	"code.google.com/p/ginta/fmt/nr"
	"code.google.com/p/ginta/fmt/plural"
	"code.google.com/p/ginta/fmt/quoted"
	"code.google.com/p/ginta/fmt/time"
)

func Setup(providers ...ginta.LanguageProvider) {
	nr.Install()
	quoted.Install()
	plural.Install()
	time.Install()

	for _, p := range providers {
		ginta.Register(p)
	}
}
