// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

import "testing"

func TestMakeStringSet(t *testing.T) {
	empty := MakeStringSet()
	if !empty.IsEmpty() {
		t.Fail()
	}

	elements := []string{"one", "two", "three"}
	set := MakeStringSet(elements...)
	if set.IsEmpty() {
		t.Fail()
	}
	if len(set) != len(elements) {
		t.Fail()
	}
	for _, el := range elements {
		if !set.Contains(el) {
			t.Fail()
		}
	}
}
