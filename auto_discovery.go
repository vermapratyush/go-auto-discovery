package main

import (
	"errors"
	"fmt"
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
}

func New(port int, serviceDiscoveryAddr string) (*AutoDiscovery, error) {
	interfaceWithMultiCast := supportMultiCase()
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
	//fmt.Print(lAddr)
	//fmt.Print(rAddr)
	return &AutoDiscovery{
		Port:                   port,
		lAddr:                  lAddr,
		rAddr:                  rAddr,
		updateFrequency:        time.Second,
		interfaceWithMultiCast: interfaceWithMultiCast,
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
	return nil, errors.New("are you connected to the network?")
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
			_, udpAddr, err := udpConn.ReadFromUDP(b)
			if udpAddr.IP.IsLoopback() {
				continue
			}
			if err != nil {
				panic(err)
			}
			fmt.Println("peer: ", udpAddr)
		}
	}()
}

func (discovery *AutoDiscovery) PeriodicNotify(closeChan chan struct{}) {
	for {
		select {
		case <-time.After(time.Second):
			go discovery.NotifyAll()
		case <-closeChan:
			return
		}
	}
}

func (discovery *AutoDiscovery) NotifyAll() {
	conn, err := net.DialUDP("udp4", discovery.lAddr, discovery.rAddr)
	defer conn.Close()
	if err != nil {
		panic(err)
	}
	conn.Write([]byte("hi"))

}
