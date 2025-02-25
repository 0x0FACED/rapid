package mdnss

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/0x0FACED/rapid/internal/model"
	"github.com/grandcat/zeroconf"
)

type MDNSScanner struct {
	_uuid   string // uuid of our service
	service *zeroconf.Server
	// redundant
	entriesCh chan *zeroconf.ServiceEntry
}

// Создание mDNS-сканера
func New(serviceName string, port int) (*MDNSScanner, error) {
	service, err := zeroconf.Register(serviceName, model.SERVICE_NAME, "local.", port, []string{"txtv=0", "lo=1", "la=2"}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed register mdns service: %w", err)
	}

	return &MDNSScanner{
		_uuid:     serviceName,
		service:   service,
		entriesCh: make(chan *zeroconf.ServiceEntry, 10),
	}, nil
}

// Infinite loop
func (s *MDNSScanner) DiscoverPeers(ctx context.Context, ch chan model.ServiceInstance) error {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry)
	defer close(entries)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-entries:
				if !ok {
					return
				}
				if entry.Instance == s._uuid {
					continue
				}
				inst := model.ServiceInstance{
					InstanceName: entry.Instance,
					ServiceName:  entry.Service,
					Domain:       entry.Domain,
					HostName:     entry.HostName,
					Port:         entry.Port,
					IPv4:         extractLocalIP(entry.AddrIPv4),
				}
				ch <- inst
			}
		}
	}()

	err = resolver.Browse(ctx, model.SERVICE_NAME, ".local.", entries)
	if err != nil {
		return fmt.Errorf("failed to browse servers: %w", err)
	}

	wg.Wait()
	return nil
}

func extractLocalIP(ips []net.IP) string {
	privateBlocks := []*net.IPNet{
		{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)},
		{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)},
		{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)},
	}

	for _, ip := range ips {
		ipv4 := ip.To4()
		if ipv4 == nil {
			continue
		}

		for _, block := range privateBlocks {
			if block.Contains(ipv4) {
				return ipv4.String()
			}
		}
	}

	return ""
}

func (s *MDNSScanner) Stop() {
	if s.service != nil {
		s.service.Shutdown()
	}
	fmt.Println("mDNS остановлен.")
}
