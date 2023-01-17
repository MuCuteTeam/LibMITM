package androtun

import (
	"androtun/option"
	"log"
	"time"

	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
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

func withTCPHandler(protector Protector) option.Option {
	return func(s *stack.Stack) error {
		tcpForwarder := tcp.NewForwarder(s, defaultWndSize, maxConnAttempts, func(r *tcp.ForwarderRequest) {
			var (
				// wq  waiter.Queue
				// ep  tcpip.Endpoint
				// err tcpip.Error
				id = r.ID()
			)
			log.Println("forward tcp request %s:%d->%s:%d:",
				id.RemoteAddress, id.RemotePort, id.LocalAddress, id.LocalPort)

			// // Perform a TCP three-way handshake.
			// ep, err = r.CreateEndpoint(&wq)
			// if err != nil {
			// 	// RST: prevent potential half-open TCP connection leak.
			// 	r.Complete(true)
			// 	return
			// }
			// defer r.Complete(false)

			// err = setSocketOptions(s, ep)

			// conn := &tcpConn{
			// 	TCPConn: gonet.NewTCPConn(&wq, ep),
			// 	id:      id,
			// }
			// handle(conn)
		})
		s.SetTransportProtocolHandler(tcp.ProtocolNumber, tcpForwarder.HandlePacket)
		return nil
	}
}

func withUDPHandler(protector Protector) option.Option {
	return func(s *stack.Stack) error {
		udpForwarder := udp.NewForwarder(s, func(r *udp.ForwarderRequest) {
			var (
				// wq waiter.Queue
				id = r.ID()
			)
			log.Println("udp forwarder request %s:%d->%s:%d: %s",
				id.RemoteAddress, id.RemotePort, id.LocalAddress, id.LocalPort)

			// ep, err := r.CreateEndpoint(&wq)
			// if err != nil {
			// 	printf("udp forwarder request %s:%d->%s:%d: %s",
			// 		id.RemoteAddress, id.RemotePort, id.LocalAddress, id.LocalPort, err)
			// 	return
			// }

			// conn := &udpConn{
			// 	UDPConn: gonet.NewUDPConn(s, &wq, ep),
			// 	id:      id,
			// }
			// handle(conn)
		})
		s.SetTransportProtocolHandler(udp.ProtocolNumber, udpForwarder.HandlePacket)
		return nil
	}
}
