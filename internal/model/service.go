package model

type ServiceInstance struct {
	InstanceName string
	ServiceName  string
	Domain       string
	HostName     string
	Port         int
	IPv4         string
}
