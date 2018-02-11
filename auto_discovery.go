package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

type AutoDiscovery struct {
	Port                   int `inject:"port"`
	interfaceWithMultiCast *net.Interface
	lAddr                  *net.UDPAddr
	rAddr                  *net.UDPAddr
}

func New(port int, serviceDiscoveryAddr string) (*AutoDiscovery, error) {
	interfaceWithMultiCast := supportMultiCase()
	if interfaceWithMultiCast == nil {
		return nil, errors.New("no multi-cast interface detected")
	}
	if serviceDiscoveryAddr == "" {
		serviceDiscoveryAddr = "239.255.255.250:1900"
	}
	lAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	rAddr, err := net.ResolveUDPAddr("udp4", serviceDiscoveryAddr)
	if err != nil {
		return nil, err
	}
	//fmt.Print(lAddr)
	//fmt.Print(rAddr)
	return &AutoDiscovery{
		Port:  port,
		lAddr: lAddr,
		rAddr: rAddr,
		interfaceWithMultiCast: interfaceWithMultiCast,
	}, nil
}

func supportMultiCase() *net.Interface {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagBroadcast != 0 {
			return &iface
		}
	}
	return nil
}

func (discovery *AutoDiscovery) Start() {
	udpConn, err := net.ListenMulticastUDP("udp4", discovery.interfaceWithMultiCast, discovery.rAddr)
	if err != nil {
		panic(err)
	}
	// start listening
	go func() {
		for {
			b := make([]byte, 1024*4)
			readBytes, udpAddr, err := udpConn.ReadFromUDP(b)
			if err != nil {
				panic(err)
			}
			fmt.Println("reading...")
			fmt.Println(readBytes)
			fmt.Println(udpAddr)
		}
	}()
}

func (discovery *AutoDiscovery) NotifyAll() {
	conn, err := net.DialUDP("udp4", discovery.lAddr, discovery.rAddr)
	if err != nil {
		panic(err)
	}
	conn.Write([]byte("hi"))
}
