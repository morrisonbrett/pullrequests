//
// Brett Morrison, July 2015
//
// Testing harness for pullrequest.go
//
package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"strings"
	"testing"
)

func TestLaunch(t *testing.T) {
	t.Logf("OK: Program Launches\n")
}

func TestFailAuth(t *testing.T) {
	err := rootRepos("dummyowner", "dummyuser", "dummypassword")
	if assert.Error(t, err) {
		t.Logf("OK: Expected Error and Received Error: %s", err)
	}
}

func TestUsage(t *testing.T) {
	cmd := exec.Command("go", "run", "pullrequests.go")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if assert.Error(t, err) {
		t.Logf("OK: Expected Error and Received Error: %s", err)
	}

	if strings.Contains(err.Error(), "exit status 1") {
		t.Logf("OK: Exit status should be 1 and is 1")
	} else {
		t.Errorf("ERROR: Exit status is not 1")
	}

	if strings.Contains(stderr.String(), "Usage of") {
		t.Logf("OK: Expected Usage Error and Usage Error Found")
	} else {
		t.Errorf("ERROR: Usage Error Not Found")
	}
}
