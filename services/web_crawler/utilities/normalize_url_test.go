package utilities

import (
	"fmt"
	"testing"
)

func TestNormalizeLink(t *testing.T) {
	s, err := NormalizeLink("https://email.osu.edu/", "mailto:president@osu.edu")
	if err != nil {
		t.Error(err)
	}
	s2, err := NormalizeLink("https://github.com", "/Liontree97")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(s)
	fmt.Println(s2)
}
