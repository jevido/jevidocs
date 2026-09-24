package controllers

import "testing"

func TestRightmostHop(t *testing.T) {
	cases := map[[2]string]string{
		{"1.1.1.1, 2.2.2.2", "10.0.0.1:5"}: "2.2.2.2",
		{"spoofed", "10.0.0.1:5"}:          "spoofed",
		{"", "10.0.0.1:5"}:                 "10.0.0.1",
		{"", "weird"}:                      "weird",
	}
	for in, want := range cases {
		if got := rightmostHop(in[0], in[1]); got != want {
			t.Errorf("rightmostHop(%q, %q) = %q, want %q", in[0], in[1], got, want)
		}
	}
}
