package time

import (
	"code.google.com/p/ginta"
	"code.google.com/p/ginta/internal"
	"code.google.com/p/ginta/providers/simple"
	"testing"
	"time"
)

func en() {
	p := simple.New().AddLanguage("en", "English", map[string]string{
		TimeFormatRoot + ":" + DateFormat + ":" + OptionDefault: "Jan 02, 2006",
	})

	ginta.Register(p)
	internal.Activate("en")
}

func de() {
	p := simple.New().AddLanguage("de", "Deutsch", map[string]string{
		TimeFormatRoot + ":" + DateTimeFormat + ":" + OptionLong: "Monday, 02. January 2006, um 15:04",
		SubstitutionsResourceBundle + ":Wednesday":               "Mittwoch",
		SubstitutionsResourceBundle + ":May":                     "Mai",
	})

	ginta.Register(p)
	internal.Activate("de")
}

func TestSimpleDate(t *testing.T) {
	en()
	format := dateFormatType(DateFormat)
	input, err := format.Compile([]string{})

	if err != nil {
		t.Error(err)
	}

	if input.Converter() == nil {
		t.Error("No converter function")
	}

	example := time.Date(2013, 5, 20, 17, 25, 29, 0, time.UTC)

	if conv := input.Converter().Convert(ginta.Locale("en"), example); conv != "May 20, 2013" {
		t.Error(conv)
	}
}

func TestLongWithSubstitutions(t *testing.T) {
	de()
	format := dateFormatType(DateTimeFormat)
	input, err := format.Compile([]string{OptionLong})

	if err != nil {
		t.Error(err)
	}

	if input.Converter() == nil {
		t.Error("No converter function")
	}

	example := time.Date(2013, 5, 22, 22, 07, 12, 0, time.UTC)

	if conv := input.Converter().Convert(ginta.Locale("de"), example); conv != "Mittwoch, 22. Mai 2013, um 22:07" {
		t.Error(conv)
	}
}
