package net

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/hashicorp/packer/common/filelock"
	"github.com/hashicorp/packer/common/retry"
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
	lock    *filelock.Flock
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
	// like "tcp" or "udp". defaults to "tcp".
	Network  string
	Addr     string
	Min, Max int
	net.ListenConfig
}

// Listen tries to Listen to a random open TCP port in the [min, max) range
// until ctx is cancelled.
// Listen uses net.ListenConfig.Listen internally.
func (lc ListenRangeConfig) Listen(ctx context.Context) (*Listener, error) {
	if lc.Network == "" {
		lc.Network = "tcp"
	}
	portRange := lc.Max - lc.Min

	var listener *Listener

	err := retry.Config{
		RetryDelay: func() time.Duration { return 20 * time.Millisecond },
	}.Run(ctx, func(context.Context) error {
		port := lc.Min
		if portRange > 0 {
			port += rand.Intn(portRange)
		}

		lockFilePath, err := packer.CachePath("port", strconv.Itoa(port))
		if err != nil {
			return err
		}

		lock := filelock.New(lockFilePath)
		locked, err := lock.TryLock()
		if err != nil {
			return err
		}
		if !locked {
			return fmt.Errorf("Port %d is file locked", port)
		}

		log.Printf("Trying port: %d", port)

		l, err := lc.ListenConfig.Listen(ctx, lc.Network, fmt.Sprintf("%s:%d", lc.Addr, port))
		if err != nil {
			if err := lock.Unlock(); err != nil {
				log.Fatalf("Could not unlock file lock for port %d: %v", port, err)
			}
			return fmt.Errorf("Port %d cannot be opened: %v", port, err)
		}

		log.Printf("Found available port: %d on IP: %s", port, lc.Addr)
		listener = &Listener{
			Address:  lc.Addr,
			Port:     port,
			Listener: l,
			lock:     lock,
		}
		return nil
	})
	return listener, err
}
