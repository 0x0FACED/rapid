package mdnss

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/0x0FACED/rapid/internal/model"
	"github.com/grandcat/zeroconf"
)

type MDNSScanner struct {
	_uuid     string // uuid of our service
	service   *zeroconf.Server
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
					// TODO: remove
					LastSeen: time.Now(),
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
	for _, ip := range ips {
		curr := ip.To4().String()
		if strings.HasPrefix(curr, "192.168") {
			return curr
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
