package utils

import (
	"fmt"
	"testing"
)

func TestNestedMapLookup(t *testing.T) {
	m := map[string]any{
		"a": map[string]any{
			"b": 1,
			"c": map[string]any{
				"d": "foo",
			},
		},
	}

	tests := []struct {
		ks   []string
		want any
		ok   bool
	}{
		{ks: []string{"a", "b"}, want: 1, ok: true},
		{ks: []string{"a", "c", "d"}, want: "foo", ok: true},
		{ks: []string{"a", "x", "y"}, want: nil, ok: false},
		{ks: []string{"a", "c", "x"}, want: nil, ok: false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.ks), func(t *testing.T) {
			val, ok := NestedMapLookup(m, tt.ks...)
			if val != tt.want || ok != tt.ok {
				t.Errorf("want %v, got %v", tt.want, val)
			}
		})
	}
}
