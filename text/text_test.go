// Copyright (c) 2022 Levi Gruspe
// License: GNU AGPLv3 or later

package text

import (
	"testing"
)

func TestCasefold(t *testing.T) {
	t.Parallel()

	a := Casefold("Foo")
	b := Casefold("fOo")
	c := Casefold("foO")

	if a != "foo" {
		t.Fatal("expected `a` to be 'foo':", a)
	}
	if b != "foo" {
		t.Fatal("expected `b` to be 'foo':", b)
	}
	if c != "foo" {
		t.Fatal("expected `c` to be 'foo':", c)
	}
}
