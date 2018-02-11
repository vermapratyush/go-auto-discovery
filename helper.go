package main

import (
	"errors"
	"log"
	"net"
)

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
		log.Fatalf("unable to read interfaces %s", err)
		return nil
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagBroadcast != 0 {
			return &iface
		}
	}
	return nil
}
