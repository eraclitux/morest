package main

import (
	"fmt"
	"testing"
)

type decodeSortCase struct {
	Case     string
	Expected []string
}

func compareSortSlices(o, p []string) bool {
	// FIXME use DeepEqual
	if len(o) != len(p) {
		return false
	}
	for i, v := range o {
		if v != p[i] {
			return false
		}
	}
	return true
}

func TestDecodeSortArgs(t *testing.T) {
	// TODO add $natural case
	cases := []decodeSortCase{}
	oneCase := decodeSortCase{`{"name":-1,"age":1}`, []string{"-name", "age"}}
	cases = append(cases, oneCase)
	for _, singleCase := range cases {
		got := decodeSortArgs(singleCase.Case)
		if !compareSortSlices(got, singleCase.Expected) && testing.Verbose() {
			fmt.Printf("expected: %+v\n", singleCase)
			fmt.Printf("got: %+v\n", got)
			t.Fail()
		}
	}
}
