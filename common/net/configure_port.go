package net

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"

	"github.com/gofrs/flock"

	"github.com/hashicorp/packer/packer"
)

var _ net.Listener = &Listener{}

// Listener wraps a net.Lister with some magic packer capabilies. For example
// until you call Listener.Close, any call to ListenRangeConfig.Listen cannot
// bind to Port. Packer tries tells moving parts which port they can use, but
// often the port has to be released before a 3rd party is started, like a VNC
// server.
type Listener struct {
	// Listener can be closed but Port will be file locked by packer until
	// Close is called.
	net.Listener
	Port    int
	Address string
	lock    *flock.Flock
}

func (l *Listener) Close() error {
	err := l.lock.Unlock()
	if err != nil {
		log.Printf("cannot unlock lockfile %#v: %v", l, err)
	}
	return l.Listener.Close()
}

// ListenRangeConfig contains options for listening to a free address [Min,Max)
// range. ListenRangeConfig wraps a net.ListenConfig.
type ListenRangeConfig struct {
	// tcp", "udp"
	Network  string
	Addr     string
	Min, Max int
	net.ListenConfig
}

// Listen tries to Listen to a random open TCP port in the [min, max) range
// until ctx is cancelled.
// Listen uses net.ListenConfig.Listen internally.
func (lc ListenRangeConfig) Listen(ctx context.Context) (*Listener, error) {
	portRange := lc.Max - lc.Min
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		port := lc.Min
		if portRange > 0 {
			port += rand.Intn(portRange)
		}

		log.Printf("Trying port: %d", port)

		lockFilePath, err := packer.CachePath("port", strconv.Itoa(port))
		if err != nil {
			return nil, err
		}

		lock := flock.New(lockFilePath)
		locked, err := lock.TryLock()
		if err != nil {
			return nil, err
		}
		if !locked {
			continue // this port seems to be locked by another packer goroutine
		}

		l, err := lc.ListenConfig.Listen(ctx, lc.Network, fmt.Sprintf("%s:%d", lc.Addr, port))
		if err != nil {
			if err := lock.Unlock(); err != nil {
				log.Printf("Could not unlock file lock for port %d: %v", port, err)
			}

			continue // this port is most likely already open
		}

		log.Printf("Found available port: %d on IP: %s", port, lc.Addr)
		return &Listener{
			Address:  lc.Addr,
			Port:     port,
			Listener: l,
			lock:     lock,
		}, err

	}
}
