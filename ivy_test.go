// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"robpike.io/ivy/config"
	"robpike.io/ivy/exec"
	"robpike.io/ivy/run"
)

const verbose = false

var testConf config.Config

// Note: These tests share some infrastructure and cannot run in parallel.

func TestAll(t *testing.T) {
	var err error
	check := func() {
		if err != nil {
			t.Fatal(err)
		}
	}
	dir, err := os.Open("testdata")
	check()
	names, err := dir.Readdirnames(0)
	check()
	for _, name := range names {
		if !strings.HasSuffix(name, ".ivy") {
			continue
		}
		t.Log(name)
		var data []byte
		path := filepath.Join("testdata", name)
		data, err = ioutil.ReadFile(path)
		check()
		text := string(data)
		lines := strings.Split(text, "\n")
		// Will have a trailing empty string.
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		lineNum := 1
		errCount := 0
		for len(lines) > 0 {
			// Assemble the input to one example.
			input, output, length := getText(t, path, lineNum, lines)
			if input == nil {
				break
			}
			if verbose {
				fmt.Printf("%s:%d: %s\n", path, lineNum, input)
			}
			if !runTest(t, path, lineNum, input, output) {
				errCount++
				if errCount > 3 {
					t.Fatal("too many errors")
				}
			}
			lines = lines[length:]
			lineNum += length
		}
	}
}

func runTest(t *testing.T, name string, lineNum int, input, output []string) bool {
	shouldFail := strings.HasSuffix(name, "_fail.ivy")
	reset()
	in := strings.Join(input, "\n")
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	run.Ivy(exec.NewContext(&testConf), in, stdout, stderr)
	if shouldFail {
		if stderr.Len() == 0 {
			t.Fatalf("\nexpected execution failure at %s:%d:\n%s", name, lineNum, in)
		}
		return true
	}
	if stderr.Len() != 0 {
		t.Fatalf("\nexecution failure (%s) at %s:%d:\n%s", stderr, name, lineNum, in)
	}
	if shouldFail {
		return true
	}
	result := stdout.String()
	if !equal(strings.Split(result, "\n"), output) {
		t.Errorf("\n%s:%d:\n%s\ngot:\n%swant:\n%s",
			name, lineNum,
			strings.Join(input, "\n"), result, strings.Join(output, "\n"))
		return false
	}
	return true
}

func equal(a, b []string) bool {
	// Split leaves an empty traililng line.
	if len(a) > 0 && a[len(a)-1] == "" {
		a = a[:len(a)-1]
	}
	if len(a) != len(b) {
		return false
	}
	for i, s := range a {
		if strings.TrimSpace(s) != strings.TrimSpace(b[i]) {
			return false
		}
	}
	return true
}

func getText(t *testing.T, fileName string, lineNum int, lines []string) (input, output []string, length int) {
	// Skip blank and initial comment lines.
	for _, line := range lines {
		if len(line) > 0 && !strings.HasPrefix(line, "#") {
			break
		}
		length++
	}
	// Input starts in left column.
	for _, line := range lines[length:] {
		if len(line) == 0 {
			t.Fatalf("%s:%d: unexpected empty line", fileName, lineNum+length)
		}
		if strings.HasPrefix(line, "\t") {
			break
		}
		input = append(input, line)
		length++
	}
	// Output is indented by a tab.
	for _, line := range lines[length:] {
		length++
		if len(line) == 0 {
			break
		}
		if !strings.HasPrefix(line, "\t") {
			t.Fatalf("%s:%d: output not indented", fileName, lineNum+length)
		}
		output = append(output, line[1:])
	}
	return // Will return nil if no more tests exist.
}

func reset() {
	testConf.SetFormat("")
	testConf.SetMaxBits(1e9)
	testConf.SetMaxDigits(1e4)
	testConf.SetOrigin(1)
	testConf.SetPrompt("")
	testConf.SetBase(0, 0)
	testConf.SetRandomSeed(0)
}
