package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "  salut   ",
			expected: []string{"salut"},
		},
		{
			input:    "  hi  sandy cool ",
			expected: []string{"hi", "sandy", "cool"},
		},
	}
	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("length mismatch\nexpected len : %d\n actual len: %d\n", len(c.expected), len(actual))
			continue
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("words mismatch\nexpected : %s\actual: %s\n", expectedWord, word)
				t.Fail()
			}

		}
	}

}
