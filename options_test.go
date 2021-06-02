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

func readOptionsFromString(t *testing.T, data string) *Options {
	options := new(Options)
	r := strings.NewReader(data)
	if ok, err := options.ReadFromReader(r, "<data>", nil); !ok {
		t.Fatalf("reading options failed for data %q", data)
	} else if err != nil {
		t.Fatal(err)
	}
	return options
}

func testOptions(t *testing.T, data string, expectedValues []string) {
	options := readOptionsFromString(t, data)
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
