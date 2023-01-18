package libmitm

import (
	"libmitm/endpoint"
	"net"
	"os"
	"syscall"

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
	LocalIP       []string
	IPv6Config    int
	FdProtector   Protector

	file  *os.File
	stack *stack.Stack
}

type Protector interface {
	Protect(fd int32) bool
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

	prot := t.FdProtector
	dialer := &net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				// protect the socket to make it won't be forwarded into tun
				prot.Protect(int32(fd))
			})
		},
	}
	var err error
	ep, err := endpoint.NewRwEndpoint(t.FileDescriber, t.MTU)
	if err != nil {
		return err
	}
	t.stack, err = createStack(opts, ep, dialer)

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
