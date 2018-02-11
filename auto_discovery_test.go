package main

import (
	"testing"
	"time"

	"fmt"
	"net"

	"github.com/stretchr/testify/assert"
)

type autoConnectListener struct{}

func (listener *autoConnectListener) OnNewPeer(peerAddr net.UDPAddr) {
	fmt.Println("New peer joined ", peerAddr.IP.To4())
}

func TestNew(t *testing.T) {
	discovery1, err := New("test-server", 24040, "")
	assert.Nil(t, err, "unable to start server2")
	discovery1.SetOnJoinListener(&autoConnectListener{})

	discovery1.Start()
	discovery1.NotifyAll()

	time.Sleep(100 * time.Second)
}
