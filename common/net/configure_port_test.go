package net

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestListenRangeConfig_Listen(t *testing.T) {
	topCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	var lockedListener *Listener
	{ // open a random port in range
		ctx, cancel := context.WithTimeout(topCtx, time.Second*5)

		lockedListener, err = ListenRangeConfig{
			Min:  800,
			Max:  10000,
			Addr: "localhost",
		}.Listen(ctx)
		if err != nil {
			t.Fatalf("could not open first port")
		}
		cancel()
		defer lockedListener.Close() // in case
	}

	{ // open a second random port in range
		ctx, cancel := context.WithTimeout(topCtx, time.Second*5)

		listener, err := ListenRangeConfig{
			Min:  800,
			Max:  10000,
			Addr: "localhost",
		}.Listen(ctx)
		if err != nil {
			t.Fatalf("could not open first port")
		}
		cancel()
		if err := listener.Close(); err != nil { // in case
			t.Fatal("failed to close second random port")
		}
	}

	{ // test that opened port cannot be openned using min/max
		ctx, cancel := context.WithTimeout(topCtx, 250*time.Millisecond)

		l, err := ListenRangeConfig{
			Min: lockedListener.Port,
			Max: lockedListener.Port,
		}.Listen(ctx)
		if err == nil {
			l.Close()
			t.Fatal("port should be taken, this should fail")
		}
		if p := int(err.(ErrPortFileLocked)); p != lockedListener.Port {
			t.Fatalf("wrong fileport: %d", p)
		}
		cancel()
	}

	{ // test that opened port cannot be openned using min only
		ctx, cancel := context.WithTimeout(topCtx, 250*time.Millisecond)

		l, err := ListenRangeConfig{
			Min: lockedListener.Port,
		}.Listen(ctx)
		if err == nil {
			l.Close()
			t.Fatalf("port should be taken, this should timeout.")
		}
		if p := int(err.(ErrPortFileLocked)); p != lockedListener.Port {
			t.Fatalf("wrong fileport: %d", p)
		}
		cancel()
	}

	err = lockedListener.Close() // close port and release lock file
	if err != nil {
		t.Fatalf("could not release lockfile or port: %v", err)
	}

	{ // test that closed port can be reopenned.
		ctx, cancel := context.WithTimeout(topCtx, 250*time.Millisecond)

		lockedListener, err = ListenRangeConfig{
			Min: lockedListener.Port,
		}.Listen(ctx)
		if err != nil {
			t.Fatalf("port should have been freed: %v", err)
		}
		cancel()
		defer lockedListener.Close() // in case
	}

	err = lockedListener.Listener.Close() // close listener, keep lockfile only
	if err != nil {
		t.Fatalf("could not release lockfile or port: %v", err)
	}

	{ // test that file locked port cannot be opened
		ctx, cancel := context.WithTimeout(topCtx, 250*time.Millisecond)

		l, err := ListenRangeConfig{
			Min: lockedListener.Port,
		}.Listen(ctx)
		if err == nil {
			l.Close()
			t.Fatalf("port should be file locked, this should timeout")
		}
		cancel()
	}

	var netListener net.Listener
	{ // test that the closed network port can be reopened using net.Listen
		netListener, err = net.Listen("tcp", lockedListener.Addr().String())
		if err != nil {
			t.Fatalf("listen on freed port failed: %v", err)
		}

	}

	if err := lockedListener.lock.Unlock(); err != nil {
		t.Fatalf("error closing port: %v", err)
	}

	{ // test that busy port cannot be opened
		ctx, cancel := context.WithTimeout(topCtx, 250*time.Millisecond)

		l, err := ListenRangeConfig{
			Min: lockedListener.Port,
		}.Listen(ctx)
		if err == nil {
			l.Close()
			t.Fatalf("port should be file locked, this should timeout")
		}
		busyErr := err.(*ErrPortBusy)
		if busyErr.Port != lockedListener.Port {
			t.Fatal("wrong port")
		}
		// error types vary depending on OS and it might get quickly
		// complicated to test for the error we want.
		cancel()
	}

	if err := netListener.Close(); err != nil { // free port
		t.Fatalf("close failed: %v", err)
	}

	{ // test that freed port can be opened
		ctx, cancel := context.WithTimeout(topCtx, 250*time.Minute)

		lockedListener, err = ListenRangeConfig{
			Min: lockedListener.Port,
		}.Listen(ctx)
		if err != nil {
			t.Fatalf("port should have been freed: %v", err)
		}
		cancel()
		defer lockedListener.Close() // in case
	}

}
