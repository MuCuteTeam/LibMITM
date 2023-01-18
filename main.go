package libmitm

import (
	"libmitm/endpoint"
	"net"
	"os"
	"strings"
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

	UdpRedirector Redirector
	TcpRedirector Redirector

	file  *os.File
	stack *stack.Stack
}

type Protector interface {
	Protect(fd int32) bool
}

type Redirector interface {
	Redirect(src string, srcPort int, dst string, dstPort int) string
}

func (t *TUN) AddLocalIP(ip string) {
	if t.LocalIP == nil {
		t.LocalIP = make([]string, 1)
		t.LocalIP[0] = ip
	} else {
		t.LocalIP = append(t.LocalIP, ip)
	}
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
	localIPs := t.LocalIP
	if localIPs == nil {
		localIPs = make([]string, 0)
	}
	dialer := &net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			if contains(localIPs, address[:strings.LastIndex(address, ":")]) {
				return nil
			}
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
	t.stack, err = createStack(opts, ep, dialer, t.TcpRedirector, t.UdpRedirector)

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
