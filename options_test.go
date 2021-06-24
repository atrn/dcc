// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func expectValues(t *testing.T, options *Options, expectedValues []string) {
	if len(options.Values) != len(expectedValues) {
		t.Fatalf("read %d option values but expected %d\nactual: %#v\nexpected: %#v", len(options.Values), len(expectedValues), options.Values, expectedValues)
	} else {
		for index, value := range expectedValues {
			if value != expectedValues[index] {
				t.Fatalf("option value %d - %q but expected %q", index, value, expectedValues[index])
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

func readOptionsFromFile(t *testing.T, path string) (*Options, error) {
	options := NewOptions()
	ok, err := options.ReadFromFile(path, nil)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("failed to read options from file %q", path)
	}
	return options, err
}

func mustReadOptionsFromFile(t *testing.T, path string) *Options {
	options, err := readOptionsFromFile(t, path)
	if err != nil {
		t.Fatal(err)
	}
	return options
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

func TestComments(t *testing.T) {
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

func TestConditionals(t *testing.T) {
	data := `
!ifdef PATH
a-value
!endif
`
	testOptions(t, data, []string{"a-value"})

	data = `
!ifdef NOTDEF
a-value
!endif
`
	testOptions(t, data, []string{})

	data = `
!ifndef NOTDEF
a-value
!endif
`
	testOptions(t, data, []string{"a-value"})
}

func TestElse(t *testing.T) {
	data := `
!ifdef NOTDEF
  a-value
!else
  b-value
!endif
c-value
`
	testOptions(t, data, []string{"b-value", "c-value"})
}

func TestNestedIfdef(t *testing.T) {
	data := `
!ifdef PATH
  !ifdef NOTDEF
    a-value
  !else
    b-value
  !endif
  c-value
!endif
d-value
`
	testOptions(t, data, []string{"b-value", "c-value", "d-value"})

	data = `
!ifdef PATH
  !ifndef NOTDEF
    a-value
  !else
    b-value
  !endif
  c-value
!endif
d-value
`
	testOptions(t, data, []string{"a-value", "c-value", "d-value"})
}

func TestError(t *testing.T) {
	errorMessage := "Something bad happened"

	data := fmt.Sprintf(`
!error %s
`,
		errorMessage,
	)

	_, err := readOptionsFromString(data)
	if err == nil {
		t.Fail()
	}
	if !strings.HasSuffix(err.Error(), errorMessage) {
		t.Fatalf("!error produced unexpected error %q, expected %q", err.Error(), errorMessage)
	}

	data = fmt.Sprintf(`
!ifdef NOTDEF
  !error %s
!endif
`,
		errorMessage,
	)
	_, err = readOptionsFromString(data)
	if err != nil {
		t.Fatalf("!error within false condition triggered with message %q", err.Error())
	}

	data = fmt.Sprintf(`
!ifdef HOME
  !error %s
!endif
`,
		errorMessage,
	)
	_, err = readOptionsFromString(data)
	if err == nil {
		t.Fatal("!error within true condition did not return expected error")
	}
	if !strings.HasSuffix(err.Error(), errorMessage) {
		t.Fatalf("!error within true condition produced unexpected error %q, expected %q", err.Error(), errorMessage)
	}
}

func TestInclude(t *testing.T) {
	setupTest(t)
	defer removeTestDirs(t)
	includeFilename := "included.options"
	includedValue := "included-value"

	includedFile := fmt.Sprintf(
		`# This file is included
%s
`,
		includedValue,
	)

	includingFile := fmt.Sprintf(`# This file includes another
!include %q
`,
		includeFilename,
	)

	filename := filepath.Join(testProjectChildDir, testFilename)
	makeFileWithContent(t, filename, includingFile)
	makeFileWithContent(t, filepath.Join(testProjectChildDir, includeFilename), includedFile)
	options := mustReadOptionsFromFile(t, filename)
	expectValues(t, options, []string{includedValue})
}

func TestIncludeNonExistentFile(t *testing.T) {
	setupTest(t)
	defer removeTestDirs(t)
	includeFilename := "included.options"

	includingFile := fmt.Sprintf(`# This file includes another
!include %q
`,
		includeFilename,
	)

	filename := filepath.Join(testProjectChildDir, testFilename)
	makeFileWithContent(t, filename, includingFile)
	_, err := readOptionsFromFile(t, filename)
	if err == nil {
		t.Fatalf("!include of non-existent file did not fail")
	}
}

func TestInherit(t *testing.T) {
	setupTest(t)
	defer removeTestDirs(t)

	directValue := "direct-value"
	inheritedValue := "inherited-value"

	fileInChildDir := filepath.Join(testProjectChildDir, testFilename)
	fileInParentDir := filepath.Join(testProjectRootDir, testFilename)

	makeFileWithContent(t, fileInParentDir, inheritedValue)
	makeFileWithContent(t, fileInChildDir, fmt.Sprintf("%s\n!inherit\n", directValue))

	options := mustReadOptionsFromFile(t, fileInChildDir)
	expectValues(t, options, []string{directValue, inheritedValue})
}

func TestInheritFromPlatformSpecificFile(t *testing.T) {
	setupTest(t)
	defer removeTestDirs(t)

	directValue := "direct-value"
	inheritedValue := "inherited-value"

	fileInChildDir := filepath.Join(testProjectChildDir, OsSpecificFilename(testFilename))
	fileInParentDir := filepath.Join(testProjectRootDir, testFilename)

	makeFileWithContent(t, fileInParentDir, inheritedValue)
	makeFileWithContent(t, fileInChildDir, fmt.Sprintf("%s\n!inherit\n", directValue))

	options := mustReadOptionsFromFile(t, fileInChildDir)
	expectValues(t, options, []string{directValue, inheritedValue})
}

func TestInheritPlatformSpecificFileFromPlainFile(t *testing.T) {
	setupTest(t)
	defer removeTestDirs(t)

	directValue := "direct-value"
	inheritedValue := "inherited-value"

	fileInChildDir := filepath.Join(testProjectChildDir, testFilename)
	fileInParentDir := filepath.Join(testProjectRootDir, OsSpecificFilename(testFilename))

	makeFileWithContent(t, fileInParentDir, inheritedValue)
	makeFileWithContent(t, fileInChildDir, fmt.Sprintf("%s\n!inherit\n", directValue))

	options := mustReadOptionsFromFile(t, fileInChildDir)
	expectValues(t, options, []string{directValue, inheritedValue})
}

func TestInheritFile(t *testing.T) {
	setupTest(t)
	defer removeTestDirs(t)

	directValue := "direct-value"
	inheritedValue := "inherited-value"

	inheritedFilename := "inherited.file"

	fileInChildDir := filepath.Join(testProjectChildDir, testFilename)
	fileInParentDir := filepath.Join(testProjectRootDir, inheritedFilename)

	makeFileWithContent(t, fileInParentDir, inheritedValue)
	makeFileWithContent(t, fileInChildDir, fmt.Sprintf("%s\n!inherit %s\n", directValue, filepath.Base(fileInParentDir)))

	options := mustReadOptionsFromFile(t, fileInChildDir)
	expectValues(t, options, []string{directValue, inheritedValue})
}

func TestInheritPlatformSpecificFile(t *testing.T) {
	setupTest(t)
	defer removeTestDirs(t)

	directValue := "direct-value"
	inheritedValue := "inherited-value"

	inheritedFilename := "inherited.file"

	fileInChildDir := filepath.Join(testProjectChildDir, testFilename)
	fileInParentDir := filepath.Join(testProjectRootDir, OsSpecificFilename(inheritedFilename))

	makeFileWithContent(t, fileInParentDir, inheritedValue)
	makeFileWithContent(t, fileInChildDir, fmt.Sprintf("%s\n!inherit %s\n", directValue, filepath.Base(fileInParentDir)))

	options := mustReadOptionsFromFile(t, fileInChildDir)
	expectValues(t, options, []string{directValue, inheritedValue})
}
