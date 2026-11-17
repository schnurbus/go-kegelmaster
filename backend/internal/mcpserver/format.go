package mcpserver

import (
	"fmt"
	"net/url"
	"strings"
)

// euro formats cents as a German currency string, e.g. -350 → "-3,50 €".
func euro(cents int) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	whole := cents / 100
	// thousands separator
	s := fmt.Sprintf("%d", whole)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteRune('.')
		}
		b.WriteRune(r)
	}
	return fmt.Sprintf("%s%s,%02d €", sign, b.String(), cents%100)
}

func toValues(m map[string]string) url.Values {
	if len(m) == 0 {
		return nil
	}
	v := url.Values{}
	for k, val := range m {
		if val != "" {
			v.Set(k, val)
		}
	}
	return v
}

func pathEsc(s string) string { return url.PathEscape(s) }
