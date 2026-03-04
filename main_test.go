package main

import (
	"strings"
	"testing"
)

func TestConvertToToon(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		key      string
		expected []string // use slices to check for containing strings (order of maps is random in Go)
	}{
		{
			name:  "Simple Object",
			input: map[string]interface{}{"temp": 22.5, "status": "ok"},
			key:   "",
			expected: []string{
				"temp: 22.5",
				"status: ok",
			},
		},
		{
			name: "Uniform Tabular Array",
			input: []interface{}{
				map[string]interface{}{"id": 1, "val": "A"},
				map[string]interface{}{"id": 2, "val": "B"},
			},
			key: "users",
			expected: []string{
				"users[2]{id,val}:",
				"  1,A",
				"  2,B",
			},
		},
		{
			name: "Non-Uniform Array Fallback",
			input: []interface{}{
				map[string]interface{}{"id": 1},
				"string-item",
			},
			key: "mixed",
			expected: []string{
				"mixed[2]:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToToon(tt.input, 0, tt.key)
			for _, exp := range tt.expected {
				if !strings.Contains(result, exp) {
					t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
				}
			}
		})
	}
}

func TestGetUniformKeys(t *testing.T) {
	t.Run("Valid Uniform Keys", func(t *testing.T) {
		input := []interface{}{
			map[string]interface{}{"a": 1, "b": 2},
			map[string]interface{}{"a": 3, "b": 4},
		}
		keys, uniform := getUniformKeys(input)
		if !uniform || len(keys) != 2 {
			t.Errorf("Expected uniform=true and 2 keys, got %v and %d", uniform, len(keys))
		}
	})

	t.Run("Mismatched Keys", func(t *testing.T) {
		input := []interface{}{
			map[string]interface{}{"a": 1},
			map[string]interface{}{"b": 2},
		}
		_, uniform := getUniformKeys(input)
		if uniform {
			t.Error("Expected uniform=false for mismatched keys")
		}
	})
}