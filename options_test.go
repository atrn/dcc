package main

import (
	"os"
	"strings"
	"testing"
)

func expectValues(t *testing.T, options *Options, expectedValues []string) {
	if len(options.Values) != len(expectedValues) {
		t.Fatalf("%d values but expected %d\nactual: %#v\nexpected: %#v", len(options.Values), len(expectedValues), options.Values, expectedValues)
	} else {
		for index, value := range expectedValues {
			if value != expectedValues[index] {
				t.Fatalf("value %d - %q but expected %q", index, value, expectedValues[index])
			}
		}
	}
}

func readOptionsFromString(data string) (*Options, error) {
	options := new(Options)
	r := strings.NewReader(data)
	_, err := options.ReadFromReader(r, "<data>", nil)
	return options, err
}

func mustReadOptionsFromString(t *testing.T, data string) *Options {
	options, err := readOptionsFromString(data)
	if err != nil {
		t.Fatal(err)
	}
	return options
}

func testOptions(t *testing.T, data string, expectedValues []string) {
	options := mustReadOptionsFromString(t, data)
	expectValues(t, options, expectedValues)
}

func TestReading(t *testing.T) {
	data := `
# This is a comment
value1
value2
    # This is not a comment
# This is another comment
`
	testOptions(t, data, []string{"value1", "value2", "#", "This", "is", "not", "a", "comment"})
}

func TestEnvVars(t *testing.T) {
	data := `
# comment
$AVAR $BVAR
`
	testOptions(t, data, []string{})

	os.Setenv("AVAR", "a-var-value")
	testOptions(t, data, []string{"a-var-value"})

	os.Setenv("AVAR", "value1 value2")
	testOptions(t, data, []string{"value1", "value2"})

	os.Setenv("BVAR", "")
	testOptions(t, data, []string{"value1", "value2"})

	os.Setenv("BVAR", "value3")
	testOptions(t, data, []string{"value1", "value2", "value3"})
}

func TestIfdef(t *testing.T) {
	data := `
#ifdef PATH
a-value
#endif
`
	testOptions(t, data, []string{"a-value"})

	data = `
#ifdef NOTDEF
a-value
#endif
`
	testOptions(t, data, []string{})

	data = `
#ifndef NOTDEF
a-value
#endif
`
	testOptions(t, data, []string{"a-value"})
}

func TestElse(t *testing.T) {
	data := `
#ifdef NOTDEF
a-value
#else
b-value
#endif
c-value
`
	testOptions(t, data, []string{"b-value", "c-value"})
}

func TestNestedIfdef(t *testing.T) {
	data := `
#ifdef PATH
#ifdef NOTDEF
a-value
#else
b-value
#endif
c-value
#endif
d-value
`
	testOptions(t, data, []string{"b-value", "c-value", "d-value"})

	data = `
#ifdef PATH
#ifndef NOTDEF
a-value
#else
b-value
#endif
c-value
#endif
d-value
`
	testOptions(t, data, []string{"a-value", "c-value", "d-value"})
}

func TestErrors(t *testing.T) {
	data := `
#error Some thing bad happened
`
	_, err := readOptionsFromString(data)
	if err == nil {
		t.Fail()
	}
	if !strings.HasSuffix(err.Error(), "Some thing bad happened") {
		t.Fatalf("#error produced unexpected error %q", err.Error())
	}

	data = `
#ifdef NOTDEF
#error Some thing bad happened
#endif
`
	_, err = readOptionsFromString(data)
	if err != nil {
		t.Fatalf("conditional #error returned unexpected error %q", err.Error())
	}

	data = `
#ifdef HOME
#error Some thing bad happened
#endif
`
	_, err = readOptionsFromString(data)
	if err == nil {
		t.Fatal("conditional #error did not return expected error")
	}
}
