// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package utils

import (
	"net"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

const XForwardedHeader = "X-Forwarded-For"
const XRealIPHeader = "X-Real-Ip"

// GetProtocol will return the protocol used for an HTTP request.
func GetProtocol(r *http.Request) string {
	if r.Header.Get("X-Header-Forwarded-Proto") == "https" || r.TLS != nil {
		return "https"
	}
	return "http"
}

// IsSecure returns true if the HTTP request is using HTTPS
func IsSecure(r *http.Request) bool {
	return GetProtocol(r) == "https"
}

// GetIPFromRequest gets the client's ip address in the following
// order:
//   1. X-Forwarded-For header
//   2. X-Real-IP header
//   3. From the RemoteAddr field from the request
func GetIPFromRequest(r *http.Request) (net.IP, error) {
	realIPHeader := r.Header.Get(XRealIPHeader)
	forwardedForHeader := r.Header.Get(XForwardedHeader)

	if forwardedForHeader != "" {
		ips := strings.Split(forwardedForHeader, ",")
		ipAddress := net.ParseIP(ips[0])
		if ipAddress == nil {
			return nil, errors.Errorf("ip address %s in X-Forwarded-For header isn't valid", ipAddress)
		}

		return ipAddress, nil
	}

	if realIPHeader != "" {
		ipAddress := net.ParseIP(realIPHeader)
		if ipAddress == nil {
			return nil, errors.Errorf("ip address %s in X-Real-IP header isn't valid", ipAddress)
		}

		return ipAddress, nil
	}

	remoteIPAddr := strings.Split(r.RemoteAddr, ":")
	ipAddress := net.ParseIP(remoteIPAddr[0])
	if ipAddress == nil {
		return nil, errors.Errorf("ip address %s in request.RemoteIPAddr isn't valid", remoteIPAddr)
	}

	return ipAddress, nil
}

// IPAddressInCIDRBlock returns true if the passed ip address belongs to one of the
// passed CIDR blocks.
func IPAddressInCIDRBlock(ipAddress net.IP, cidrBlocks []string) (bool, error) {
	for _, cidr := range cidrBlocks {
		_, cidrParsed, err := net.ParseCIDR(cidr)
		if err != nil {
			return false, errors.Errorf("cidr block %s passed in the white list isn't valid", cidr)
		}
		if cidrParsed.Contains(ipAddress) {
			return true, nil
		}
	}

	return false, nil
}
