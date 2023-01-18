package libmitm

import (
	"fmt"
	"io"
	"libmitm/option"
	"log"
	"net"
	"time"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
)

const (
	// defaultWndSize if set to zero, the default
	// receive window buffer size is used instead.
	defaultWndSize = 0

	// maxConnAttempts specifies the maximum number
	// of in-flight tcp connection attempts.
	maxConnAttempts = 2 << 10

	// tcpKeepaliveCount is the maximum number of
	// TCP keep-alive probes to send before giving up
	// and killing the connection if no response is
	// obtained from the other end.
	tcpKeepaliveCount = 9

	// tcpKeepaliveIdle specifies the time a connection
	// must remain idle before the first TCP keepalive
	// packet is sent. Once this time is reached,
	// tcpKeepaliveInterval option is used instead.
	tcpKeepaliveIdle = 60 * time.Second

	// tcpKeepaliveInterval specifies the interval
	// time between sending TCP keepalive packets.
	tcpKeepaliveInterval = 30 * time.Second
)

func withTCPHandler(dialer *net.Dialer) option.Option {
	return func(s *stack.Stack) error {
		tcpForwarder := tcp.NewForwarder(s, defaultWndSize, maxConnAttempts, func(r *tcp.ForwarderRequest) {
			var (
				wq waiter.Queue
				// 	ep  tcpip.Endpoint
				// 	err tcpip.Error
				id = r.ID()
			)

			// Perform a TCP three-way handshake.
			ep, err := r.CreateEndpoint(&wq)
			if err != nil {
				log.Println("create endpoint failed:", err)
				r.Complete(true)
				return
			}
			r.Complete(false)

			if err = setSocketOptions(s, ep); err != nil {
				log.Printf("set socket options failed: %v\n", err)
			}

			go func(local net.Conn, dialer *net.Dialer, addr string) {
				defer local.Close()

				remote, err := dialer.Dial("tcp", addr)
				if err != nil {
					log.Println("dial failed:", err)
					return
				}
				defer remote.Close()

				go func() {
					io.Copy(local, remote)
				}()
				io.Copy(remote, local)
			}(gonet.NewTCPConn(&wq, ep), dialer, addressId(id))
		})
		s.SetTransportProtocolHandler(tcp.ProtocolNumber, tcpForwarder.HandlePacket)
		return nil
	}
}

func withUDPHandler(dialer *net.Dialer) option.Option {
	return func(s *stack.Stack) error {
		udpForwarder := udp.NewForwarder(s, func(r *udp.ForwarderRequest) {
			var (
				wq waiter.Queue
				id = r.ID()
			)
			log.Printf("udp forwarder request %s:%d->%s:%d\n",
				id.RemoteAddress, id.RemotePort, id.LocalAddress, id.LocalPort)

			ep, err := r.CreateEndpoint(&wq)
			if err != nil {
				log.Println(err.String())
				return
			}

			go func(local net.Conn, dialer *net.Dialer, addr string) {
				defer local.Close()

				remote, err := dialer.Dial("udp", addr)
				if err != nil {
					log.Println("dial failed:", err)
					return
				}
				defer remote.Close()

				go func() {
					io.Copy(local, remote)
				}()
				io.Copy(remote, local)
			}(gonet.NewUDPConn(s, &wq, ep), dialer, addressId(id))
		})
		s.SetTransportProtocolHandler(udp.ProtocolNumber, udpForwarder.HandlePacket)
		return nil
	}
}

func setSocketOptions(s *stack.Stack, ep tcpip.Endpoint) tcpip.Error {
	{ /* TCP keepalive options */
		ep.SocketOptions().SetKeepAlive(true)

		idle := tcpip.KeepaliveIdleOption(tcpKeepaliveIdle)
		if err := ep.SetSockOpt(&idle); err != nil {
			return err
		}

		interval := tcpip.KeepaliveIntervalOption(tcpKeepaliveInterval)
		if err := ep.SetSockOpt(&interval); err != nil {
			return err
		}

		if err := ep.SetSockOptInt(tcpip.KeepaliveCountOption, tcpKeepaliveCount); err != nil {
			return err
		}
	}
	{ /* TCP recv/send buffer size */
		var ss tcpip.TCPSendBufferSizeRangeOption
		if err := s.TransportProtocolOption(header.TCPProtocolNumber, &ss); err == nil {
			ep.SocketOptions().SetReceiveBufferSize(int64(ss.Default), false)
		}

		var rs tcpip.TCPReceiveBufferSizeRangeOption
		if err := s.TransportProtocolOption(header.TCPProtocolNumber, &rs); err == nil {
			ep.SocketOptions().SetReceiveBufferSize(int64(rs.Default), false)
		}
	}
	return nil
}

func addressId(id stack.TransportEndpointID) string {
	if len(id.LocalAddress) == 4 {
		return fmt.Sprintf("%s:%d", id.LocalAddress.String(), id.LocalPort)
	} else {
		return fmt.Sprintf("[%s]:%d", id.LocalAddress.String(), id.LocalPort)
	}
}
