package main

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBot(t *testing.T) {
	val := NewTeleGBot(nil, reflect.TypeOf(&SharedExpensesBot{}))
	assert.NotNil(t, val, "unexpected error when constructing new bot")
}
