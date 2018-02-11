package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	discovery1, err := New(24040, "")
	assert.Nil(t, err, "unable to start server1")
	discovery2, err := New(24041, "")
	assert.Nil(t, err, "unable to start server2")
	discovery1.Start()
	//discovery2.Start()
	time.Sleep(time.Second)
	discovery1.NotifyAll()
	discovery2.NotifyAll()
	time.Sleep(time.Second)
}
