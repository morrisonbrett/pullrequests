//
// Brett Morrison, July 2015
//
// Testing harness for pullrequest.go
//
package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLaunch(t *testing.T) {
	t.Logf("Program Launches\n")
}

func TestFailAuth(t *testing.T) {
	err := rootRepos("dummyowner", "dummyuser", "dummypassword")
	if assert.Error(t, err) {
		t.Logf("Expected Error, Received Error: %s", err)
	}
}
