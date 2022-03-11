package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertSub(t *testing.T) {
	testSender := "Eric"
	msg := Msg{Sender: testSender}
	subs := make(subscribers)

	AddSubscriber(&msg, subs)

	_, ok := subs[testSender]
	assert.Equal(t, ok, true, "Sender should be in list of subscribers")
}

func TestRemoveSub(t *testing.T) {
	testSender := "Eric"
	msg := Msg{Sender: testSender}
	subs := make(subscribers)

	AddSubscriber(&msg, subs)
	RemoveSubscriber(&msg, subs)

	_, ok := subs[testSender]
	assert.Equal(t, ok, false, "Sender should not be in list of subscribers")
}
