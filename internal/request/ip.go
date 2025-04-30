package request

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
)

func ParseIP(ipAddr netip.Addr) (string, error) {
	var target string

	ip := net.ParseIP(ipAddr.String())
	if ip == nil {
		return "", errors.New("failed to parse ip address")
	}
	if ip.To4() != nil {
		target = fmt.Sprintf("%s:50051", ipAddr)
	} else {
		target = fmt.Sprintf("[%s]:50051", ipAddr)
	}

	return target, nil
}
