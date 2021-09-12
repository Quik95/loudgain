package loudgain

import (
	"fmt"
	"testing"
)

func TestEscapeQuotes(t *testing.T) {
	var tests = []struct {
		in    string
		wants string
	}{
		{"test", "test"},
		{"test'", `test'\''`},
		{"'test'", `'\''test'\''`},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s, %s", tt.in, tt.wants)
		t.Run(testname, func(t *testing.T) {
			ans := escapeQuotes(tt.in)
			if ans != tt.wants {
				t.Errorf("got %s, wants: %s", ans, tt.wants)
			}
		})
	}
}
