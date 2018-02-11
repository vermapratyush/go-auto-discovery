package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	discovery1, err := New(24040, "")
	assert.Nil(t, err, "unable to start server2")
	discovery1.Start()
	var closeChan chan struct{}
	go discovery1.PeriodicNotify(closeChan)

	time.Sleep(100 * time.Second)

	close(closeChan)
}
