package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAccount(test *testing.T) {
	account, err := NewAccount("a", "b", "hunter")
	assert.Nil(test, err)
	fmt.Printf("%+v\n", account)
}
