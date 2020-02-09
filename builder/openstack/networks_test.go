package openstack

import (
	"net"
	"testing"
)

func testYes(t *testing.T, a, b string) {
	var m, n *net.IPNet
	_, m, _ = net.ParseCIDR(a)
	_, n, _ = net.ParseCIDR(b)
	if !containsNet(m, n) {
		t.Errorf("%s expected to contain %s", m, n)
	}
}

func testNot(t *testing.T, a, b string) {
	var m, n *net.IPNet
	_, m, _ = net.ParseCIDR(a)
	_, n, _ = net.ParseCIDR(b)
	if containsNet(m, n) {
		t.Errorf("%s expected to not contain %s", m, n)
	}
}

func TestNetworkDiscovery_SubnetContainsGood_IPv4(t *testing.T) {
	testYes(t, "192.168.0.0/23", "192.168.0.0/24")
	testYes(t, "192.168.0.0/24", "192.168.0.0/24")
	testNot(t, "192.168.0.0/25", "192.168.0.0/24")

	testYes(t, "192.168.101.202/16", "192.168.202.101/16")
	testNot(t, "192.168.101.202/24", "192.168.202.101/24")
	testNot(t, "192.168.202.101/24", "192.168.101.202/24")

	testYes(t, "0.0.0.0/0", "192.168.0.0/24")
	testYes(t, "0.0.0.0/0", "0.0.0.0/1")
	testNot(t, "192.168.0.0/24", "0.0.0.0/0")
	testNot(t, "0.0.0.0/1", "0.0.0.0/0")
}

func TestNetworkDiscovery_SubnetContainsGood_IPv6(t *testing.T) {
	testYes(t, "2001:db8::/63", "2001:db8::/64")
	testYes(t, "2001:db8::/64", "2001:db8::/64")
	testNot(t, "2001:db8::/65", "2001:db8::/64")

	testYes(t, "2001:db8:fefe:b00b::/32", "2001:db8:b00b:fefe::/32")
	testNot(t, "2001:db8:fefe:b00b::/64", "2001:db8:b00b:fefe::/64")
	testNot(t, "2001:db8:b00b:fefe::/64", "2001:db8:fefe:b00b::/64")

	testYes(t, "::/0", "2001:db8::/64")
	testYes(t, "::/0", "::/1")
	testNot(t, "2001:db8::/64", "::/0")
	testNot(t, "::/1", "::/0")
}
