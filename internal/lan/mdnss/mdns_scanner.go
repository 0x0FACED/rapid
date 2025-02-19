package mdnss

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

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

func (s *MDNSScanner) DiscoverPeers(ctx context.Context, ch chan model.ServiceInstance) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("failed init resolver:", err.Error())
	}

	entries := make(chan *zeroconf.ServiceEntry)

	go func(results <-chan *zeroconf.ServiceEntry) {
		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-results:
				if !ok {
					return
				}
				if entry.Instance != s._uuid {
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
		}
	}(entries)

	err = resolver.Browse(ctx, model.SERVICE_NAME, ".local.", entries)
	if err != nil {
		log.Fatalln("failed browse servers:", err.Error())
	}
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
