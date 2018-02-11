package main

import (
	"fmt"
	"net"
	"time"

	"github.com/vermapratyush/go-auto-discovery"
)

type autoConnectListener struct{}

func (listener *autoConnectListener) OnNewPeer(peerAddr net.UDPAddr) {
	fmt.Println("New peer joined ", peerAddr)
}

func main() {

	discovery1, _ := main1.New("test-server", 24040, "")
	discovery1.SetOnJoinListener(&autoConnectListener{})

	discovery1.Start()
	discovery1.NotifyAll()

	time.Sleep(100 * time.Second)

}
