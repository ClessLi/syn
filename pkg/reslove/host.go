package reslove

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var (
	ipv4Match = regexp.MustCompile(ipv4AddressReg).FindStringSubmatch

	// Error
	InvalidIpv4Address = fmt.Errorf("invalid ipv4 address")
	NoHostRecord       = fmt.Errorf("no record of this host")
)

type VisitHosts struct {
	visits map[string]*Host
	//alarm  map[string]bool
	alarm Queue
}

func NewHosts() *VisitHosts {
	//return &VisitHosts{visits: make(map[string]*Host), alarm: make(map[string]bool)}
	return &VisitHosts{visits: make(map[string]*Host), alarm: NewQueue()}
}

func (hs *VisitHosts) AuthFailed(ipv4 string, threshold int) error {
	if hs.visits[ipv4] == nil {
		host, hostErr := NewHost(ipv4)
		if hostErr != nil {
			return hostErr
		}
		hs.visits[ipv4] = host
	}
	hs.visits[ipv4].AddAuthFailCount()
	//hs.alarm[ipv4] = hs.visits[ipv4].ThresholdReached(threshold)
	if hs.visits[ipv4].ThresholdReached(threshold) && !hs.visits[ipv4].isAlarmed {
		//hs.alarm.Add(hs.visits[ipv4])
		hs.alarm.Add(ipv4)
	}
	return nil
}

func (hs *VisitHosts) ResetHostFailCount(ipv4 string) error {
	if hs.visits[ipv4] != nil {
		//hs.alarm[ipv4] = false
		hs.visits[ipv4].ResetAuthFailCount()
		hs.visits[ipv4].ResetMarkTime()
		return nil
	}
	return fmt.Errorf("%s: %s", NoHostRecord, ipv4)
}

func (hs *VisitHosts) ResetHostRecord(ipv4 string) error {
	if hs.visits[ipv4] != nil {
		hs.visits[ipv4].ResetAuthFailCount()
		hs.visits[ipv4].ResetMarkTime()
		hs.visits[ipv4].isAlarmed = false
		return nil
	}
	return fmt.Errorf("%s: %s", NoHostRecord, ipv4)
}

type Host struct {
	ipv4Address      [4]byte
	authFailureCount int
	timestamp        time.Time
	isAlarmed        bool
}

func NewHost(ipv4 string) (*Host, error) {
	ipv4Address, err := strToIpv4(ipv4)
	if err != nil {
		return nil, err
	}
	return &Host{
		ipv4Address:      ipv4Address,
		authFailureCount: 0,
		//timestamp:        time.Unix(0, 0),
		timestamp: time.Now(),
	}, nil
}

func strToIpv4(ipv4 string) ([4]byte, error) {
	address := ipv4Match(ipv4)
	ipv4Address := [4]byte{}
	isInvalidAddress := false
	if len(address) != 5 {
		isInvalidAddress = true
	} else {
		for i := 1; i < len(address); i++ {
			num, strErr := strconv.Atoi(address[i])
			if strErr != nil {
				return ipv4Address, fmt.Errorf("ipv4 address resolv failed: %s", strErr)
			}
			if num > 255 || num < 0 {
				isInvalidAddress = true
				break
			}
			ipv4Address[i-1] = byte(num)
		}
	}
	if isInvalidAddress {
		return ipv4Address, fmt.Errorf("%s: %s", InvalidIpv4Address, ipv4)
	}
	return ipv4Address, nil
}

func (h *Host) AddAuthFailCount() {
	h.authFailureCount++
}

func (h *Host) ResetAuthFailCount() {
	h.authFailureCount = 0
}

func (h *Host) ResetMarkTime() {
	h.timestamp = time.Now()
}

func (h Host) ThresholdReached(num int) bool {
	if num <= 0 {
		return false
	}
	return h.authFailureCount >= num
}
