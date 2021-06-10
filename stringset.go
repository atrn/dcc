// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

/*
 * A set of strings.
 */
type StringSet map[string]struct{}

/*
 * Insert an element into a StringSet
 */
func (s *StringSet) Insert(el string) {
	(*s)[el] = struct{}{}
}

/*
 * Return true if a StringSet contains n element
 */
func (s *StringSet) Contains(el string) bool {
	_, found := (*s)[el]
	return found
}

/*
 * Return true if the receiver is empty.
 */
func (s *StringSet) IsEmpty() bool {
	return len(*s) == 0
}
