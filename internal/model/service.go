package model

import "time"

type ServiceInstance struct {
	InstanceName string
	ServiceName  string
	Domain       string
	HostName     string
	Port         int
	IPv4         string
	// TODO: remove
	LastSeen time.Time
}
