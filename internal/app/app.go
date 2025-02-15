package app

import (
	"github.com/0x0FACED/rapid/internal/lan/mdnss"
	lanserver "github.com/0x0FACED/rapid/internal/lan/server"
)

type Rapid struct {
	lan   *lanserver.LANServer
	mdnss *mdnss.MDNSScanner
	
}
