// The setup package makes it easy and convenient to bootstrap the ginta library
package setup

import (
	"github.com/beatgammit/ginta"
	"github.com/beatgammit/ginta/fmt/nr"
	"github.com/beatgammit/ginta/fmt/plural"
	"github.com/beatgammit/ginta/fmt/quoted"
	"github.com/beatgammit/ginta/fmt/time"
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
