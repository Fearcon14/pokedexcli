package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			input:    "HELLO WORLD",
			expected: []string{"hello", "world"},
		},
		{
			input:    "hello   world",
			expected: []string{"hello", "world"},
		},
		{
			input:    "hEllo woRld",
			expected: []string{"hello", "world"},
		},
		{
			input:    "hello WORLD",
			expected: []string{"hello", "world"},
		},
	}

	for _, test := range tests {
		actual := cleanInput(test.input)
		if len(actual) != len(test.expected) {
			t.Errorf("Input: %q - Expected length %d (%v), got length %d (%v)",
				test.input, len(test.expected), test.expected, len(actual), actual)
			continue
		}
		for i, word := range actual {
			if word != test.expected[i] {
				t.Errorf("Input: %q - At index %d: Expected %q, got %q",
					test.input, i, test.expected[i], word)
			}
		}
	}
}
