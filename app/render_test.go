package app

import (
	"fmt"
	"testing"
)

var compareIntputCases = []struct {
	text  string
	input string
	done  string
	wrong string
	todo  string
}{
	{"", "", "", "", ""},
	{"", "to much", "", "to much", ""},
	{"type this", "type ", "type ", "", "this"},
	{"type this", "typer", "type", "r", "this"},
	{"type this", "type that", "type th", "at", ""},
	{"type this", "type this!", "type this", "!", ""},
}

func Test_compareInput(t *testing.T) {
	for _, tc := range compareIntputCases {
		t.Run(fmt.Sprintf("%#v+%#v", tc.text, tc.input), func(t *testing.T) {
			done, wrong, todo := compareInput(tc.text, tc.input)
			if done != tc.done || wrong != tc.wrong || todo != tc.todo {
				t.Errorf("compareInput(%#v, %#v) returned %#v, %#v, %#v, want %#v %#v %#v",
					tc.text, tc.input, done, wrong, todo, tc.done, tc.wrong, tc.todo,
				)
			}
		})
	}
}
