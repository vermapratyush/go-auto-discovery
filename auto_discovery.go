package main

import (
	"errors"
	"log"
	"net"
	"strconv"
	"time"
)

type AutoDiscovery struct {
	Port                   int `inject:"port"`
	updateFrequency        time.Duration
	interfaceWithMultiCast *net.Interface
	lAddr                  *net.UDPAddr
	rAddr                  *net.UDPAddr
	onNewPeerListener      []OnNewPeerListener
	groupName              string
}

type OnNewPeerListener interface {
	OnNewPeer(peerAddr net.UDPAddr)
}

func New(groupName string, port int, serviceDiscoveryAddr string) (*AutoDiscovery, error) {
	interfaceWithMultiCast := supportMultiCast()
	if interfaceWithMultiCast == nil {
		return nil, errors.New("no multi-cast interface detected")
	}
	if serviceDiscoveryAddr == "" {
		serviceDiscoveryAddr = "239.255.255.250:1900"
	}
	ip, err := externalIP()
	if err != nil || ip == nil {
		return nil, err
	}
	lAddr, err := net.ResolveUDPAddr("udp4", ip.String()+":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	rAddr, err := net.ResolveUDPAddr("udp4", serviceDiscoveryAddr)
	if err != nil {
		return nil, err
	}
	return &AutoDiscovery{
		Port:                   port,
		lAddr:                  lAddr,
		rAddr:                  rAddr,
		updateFrequency:        time.Second,
		interfaceWithMultiCast: interfaceWithMultiCast,
		groupName:              groupName,
	}, nil
}

func (discovery *AutoDiscovery) Start() {
	udpConn, err := net.ListenMulticastUDP("udp4", discovery.interfaceWithMultiCast, discovery.rAddr)
	if err != nil {
		log.Fatalf("unable to listen for multicastUDP conn %s", err)
		return
	}
	// start listening
	go func() {
		for {
			b := make([]byte, len(discovery.groupName))
			_, udpAddr, err := udpConn.ReadFromUDP(b)
			if udpAddr.IP.IsLoopback() || string(b) != discovery.groupName {
				continue
			}
			if err != nil {
				log.Fatalf("unable to read from UDPConn %s", err)
			}
			go discovery.fireCallback(*udpAddr)
		}
	}()
}

func (discovery *AutoDiscovery) NotifyAll() {
	conn, err := net.DialUDP("udp4", discovery.lAddr, discovery.rAddr)
	if err != nil {
		log.Fatalf("unable to dial udp %s", err)
		return
	}
	defer conn.Close()

	conn.Write([]byte(discovery.groupName))
}

func (discovery *AutoDiscovery) SetOnJoinListener(listener ...OnNewPeerListener) {
	discovery.onNewPeerListener = append(discovery.onNewPeerListener, listener...)
}

func (discovery *AutoDiscovery) fireCallback(peer net.UDPAddr) {
	for _, listener := range discovery.onNewPeerListener {
		go listener.OnNewPeer(peer)
	}
}
