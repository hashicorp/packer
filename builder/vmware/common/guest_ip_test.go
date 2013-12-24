package common

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestDHCPLeaseGuestLookup_impl(t *testing.T) {
	var _ GuestIPFinder = new(DHCPLeaseGuestLookup)
}

func TestDHCPLeaseGuestLookup(t *testing.T) {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if _, err := tf.Write([]byte(testLeaseContents)); err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Close()
	defer os.Remove(tf.Name())

	driver := new(DriverMock)
	driver.DhcpLeasesPathResult = tf.Name()

	finder := &DHCPLeaseGuestLookup{
		Driver:     driver,
		Device:     "vmnet8",
		MACAddress: "00:0c:29:59:91:02",
	}

	ip, err := finder.GuestIP()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !driver.DhcpLeasesPathCalled {
		t.Fatal("should ask for DHCP leases path")
	}
	if driver.DhcpLeasesPathDevice != "vmnet8" {
		t.Fatal("should be vmnet8")
	}

	if ip != "192.168.126.130" {
		t.Fatalf("bad: %#v", ip)
	}
}

const testLeaseContents = `
# All times in this file are in UTC (GMT), not your local timezone.   This is
# not a bug, so please don't ask about it.   There is no portable way to
# store leases in the local timezone, so please don't request this as a
# feature.   If this is inconvenient or confusing to you, we sincerely
# apologize.   Seriously, though - don't ask.
# The format of this file is documented in the dhcpd.leases(5) manual page.

lease 192.168.126.129 {
	starts 0 2013/09/15 23:58:51;
	ends 1 2013/09/16 00:28:51;
	hardware ethernet 00:0c:29:59:91:02;
	client-hostname "precise64";
}
lease 192.168.126.130 {
	starts 2 2013/09/17 21:39:07;
	ends 2 2013/09/17 22:09:07;
	hardware ethernet 00:0c:29:59:91:02;
	client-hostname "precise64";
}
lease 192.168.126.128 {
	starts 0 2013/09/15 20:09:59;
	ends 0 2013/09/15 20:21:58;
	hardware ethernet 00:0c:29:59:91:02;
	client-hostname "precise64";
}
lease 192.168.126.127 {
	starts 0 2013/09/15 20:09:59;
	ends 0 2013/09/15 20:21:58;
	hardware ethernet 01:0c:29:59:91:02;
	client-hostname "precise64";

`
