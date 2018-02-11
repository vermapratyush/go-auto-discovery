package main

import (
	"errors"
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

func externalIP() (*net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return &ip, nil
		}
	}
	return nil, errors.New("are you connected to the network")
}

func supportMultiCast() *net.Interface {
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
			b := make([]byte, 1024)
			_, udpAddr, err := udpConn.ReadFromUDP(b)
			if udpAddr.IP.IsLoopback() || string(b) != discovery.groupName {
				continue
			}
			if err != nil {
				panic(err)
			}
			go discovery.fireCallback(*udpAddr)
		}
	}()
}

func (discovery *AutoDiscovery) NotifyAll() {
	conn, err := net.DialUDP("udp4", discovery.lAddr, discovery.rAddr)
	defer conn.Close()
	if err != nil {
		panic(err)
	}
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
