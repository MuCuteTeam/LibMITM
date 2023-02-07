package libmitm

import (
	"libmitm/endpoint"
	"net"
	"os"

	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

const (
	IPv6Enable  = 1
	IPv6Disable = 2
	IPv6Only    = 0
)

type TUN struct {
	FileDescriber int32
	MTU           int32
	IPv6Config    int

	TcpRedirector       Redirector
	UdpRedirector       Redirector
	TcpEstablishHandler EstablishHandler
	UdpEstablishHandler EstablishHandler

	file  *os.File
	stack *stack.Stack
}

type Redirector interface {
	Redirect(src string, srcPort int, dst string, dstPort int) string
}

type EstablishHandler interface {
	Handle(localAddr string, originalRemoteIp string)
}

func (t *TUN) Start() error {
	var opts stack.Options
	switch t.IPv6Config {
	case IPv6Disable:
		opts = stack.Options{
			NetworkProtocols: []stack.NetworkProtocolFactory{
				ipv4.NewProtocol,
			},
			TransportProtocols: []stack.TransportProtocolFactory{
				tcp.NewProtocol,
				udp.NewProtocol,
				icmp.NewProtocol4,
			},
		}
	case IPv6Only:
		opts = stack.Options{
			NetworkProtocols: []stack.NetworkProtocolFactory{
				ipv6.NewProtocol,
			},
			TransportProtocols: []stack.TransportProtocolFactory{
				tcp.NewProtocol,
				udp.NewProtocol,
				icmp.NewProtocol6,
			},
		}
	default:
		opts = stack.Options{
			NetworkProtocols: []stack.NetworkProtocolFactory{
				ipv4.NewProtocol,
				ipv6.NewProtocol,
			},
			TransportProtocols: []stack.TransportProtocolFactory{
				tcp.NewProtocol,
				udp.NewProtocol,
				icmp.NewProtocol4,
				icmp.NewProtocol6,
			},
		}
	}

	dialer := &net.Dialer{}
	var err error
	ep, err := endpoint.NewEndpoint(t.FileDescriber, t.MTU)
	if err != nil {
		return err
	}
	t.stack, err = t.createStack(opts, ep, dialer)

	return err
}

func (t *TUN) Close() {
	if t.file != nil {
		t.file.Close()
	}
	if t.stack != nil {
		t.stack.Close()
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
