package zip

import (
    "testing"
)

func TestLanguageDetected(t *testing.T) {
	load := New("test_content.zip")
	
	hit := false
	for l := range load.Enumerate() {
		if l.Code != "de" || l.DisplayName != "Deutsch" {
			t.Error(l)
		} else {
			hit = true
		}
	}
	
	if !hit {
		t.Error("de not provided")
	}
}

func TestLoadContents(t *testing.T) {
	load := New("test_content.zip")
	
	for r := range load.List("de") {
		switch {
		case r.Key == "basic": 
			if r.Value != "Grundlegend" {
				t.Error(r)
			}
		case r.Key == "advanced:advanced":
			if r.Value != "Fortgeschritten" {
				t.Error(r)
			} 
		default:
			t.Error(r)
		}
	}
}

