// model/service_instance.go
package model

import (
	"fmt"
	"math"
)

type ServiceInstance struct {
	InstanceName string
	ServiceName  string
	Domain       string
	HostName     string
	Port         int
	IPv4         string
}

func (si ServiceInstance) Key() string {
	return si.InstanceName
}

func (si ServiceInstance) Address() string {
	return fmt.Sprintf("%s:%d", si.IPv4, si.Port)
}

func (si ServiceInstance) String() string {
	return fmt.Sprintf(
		"ServiceInstance[Name: %s, Address: %s, Host: %s]",
		si.InstanceName,
		si.Address(),
		si.HostName,
	)
}

func (si ServiceInstance) Validate() error {
	if si.InstanceName == "" {
		return fmt.Errorf("instance name is required")
	}
	if si.IPv4 == "" {
		return fmt.Errorf("IPv4 address is required")
	}
	if si.Port <= 0 || si.Port > math.MaxUint16 {
		return fmt.Errorf("invalid port number: %d", si.Port)
	}
	return nil
}
