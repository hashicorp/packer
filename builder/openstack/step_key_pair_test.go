package openstack

import (
	"bytes"
	"github.com/mitchellh/packer/packer"
	"golang.org/x/crypto/ssh"
	"os/exec"
	"testing"
)

var ber_encoded_key = `
-----BEGIN RSA PRIVATE KEY-----
MIIEqQIBAAKCAQEAo1C/iQtu7iSUJROyQgKm5VtJbaPg5GEP7IC+7TTZ2kVnWa9V
MU37sXKWk0J7oRHcSZ/dhRyNDcsvscFAU6V3FRpZxiTEvCqIBgLyoV3g+wkan+gD
AqNmqm77ldUIwMXT7n3AeWipw0uOBzZNjANhAf8qHIT2PXoT6LfaCof4PlPucCTx
eC+oT5zA/MbQoGneNkHnR26ijMac2vRC90+1WpZ5KwCVE6qd2kERlb9gbAGsNLE/
WrqR7bx4d8esLtE3l5zwkBTB63KdojMCrX/ZBiHT15TBQVsFPQqhdBT3BXfssnut
MkCf6+X0PIQkQcW6RqKR5nOHMFdB1kEChaKYMwIDAQABAoIBAQCRBPv/opJvjzWp
stLAgQBYe/Y5EKN7yKDOPwjLM/obMzPx1JqOvJO6X2lL/GYxgd2d1wJq2A586CdC
7brETBLxP0PmifHUsOO2itmO5wEHiW8F/Yzmw9g/kWuAAfrSyxhFF49Zf9H3ZFkL
GHJF2R5EGqP3TS4nKwcQyGkqntCV7p+QmPvlB05jT3vsc2jLn2yXXVEbu1X5Zj82
cdFwH4ZSc2BDv8ixBHy+zOGLx0TMF0hxEHdNIeAjGTyYfiyDr5mgMP4w6igGvP8q
pB5fE60ZKCEPLGyxMw69nDbXJK6YAKyNCVAD79FEl1yLYJlWfYAtOI6UeE0RUPeD
i42vINyBAoIAgQC9zZzl4kJUTdzYLYrxQ7f51B+eOF621kleTYLprr8I4VZUExht
KtsHdHWzMkxVtd7Q7UcG9L+Pqa3DVnD3S2C1IuUkO3eCpa93jhiwCIRYQfR7EzuR
ntVrQHfYaCgr8ahmWdoeZUr8lSDDp0/h+MamEksNEjZA+XwFWg6ycLgGwQKCAIEA
3EYz5iHqaOdekBYXcFpkq8uk+Sn27Cmnd7pLfK/YWEAb3Pie9X0pftA1O8z0tW9p
vEBS+pA27S55a38avmZSkZAiXMJOZ3HV99VZi5WU0Pg9A9oIiQvW6td9b1uuVl5q
o9fCK/r17E7NCFimtqSNtrNR/X08g+l4Dp1Ourgm7/MCggCAIGMohbWhGd+bcqv6
zIaAqzm+F3KI/uv74wKY9yUhZfOFlp0XivFIJLKDrwtDKVD6b249s3sqAOq0QuPK
LPiIzP/iV9dp4jpBgcYWgltBsgm3HRVAEe4nfsCmcp/7UtxOnwBwDsW8EPOlfp1b
LTUVOJtggR99cILh3cvrPBmt3UECggCBALsRH7g8a1+lxmglashu6/n+G1/DZMER
avjCDKOajuf7oe4acpzXK6tX1S2xFM0VDj3iftXuLcdl5ZYGPsceDNc0CgqutXki
cu1jkgV6BgUmHGMuAnuow19znEI7ISaWTohQjsVc/wctsPB6oTKRMwzK40Gc3wzD
9MKsk5T9GYxDAoIAgAFW9agGS4th1RrVwngjKHCPenQR1QSvHDHv3E3EQta1d0K3
Cx7RCdBpDNeNIpHVq8uTaWjSZ+Em6YDja7CqIfPj0HcykizxS4L7Dh0gt7I5+kuy
yBgJhEgIw3mK7U47nm9MmR/67RI4IS978ekVXP2oGdYqVUhHvMuOix8a72xk
-----END RSA PRIVATE KEY-----
`

var der_encoded_key = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAmfhARpOHdm5i551GEtOO2NEH36Dm2NjJp6eHQpnrwDW0bqd6
qHJSmhocSb/3r2WwPOsfBKa135aCEOqXKcrc2gY+IsgGMF8bFG//UVcHA0KGG9A8
mWwYBIU4lpb7DY54IllDDO11BERrxOEsPqmNgGzE0yOYApC2t8OcXsUE15kznhLn
d4V53kN//jRGR44KrGMcL8nrKDIBudz90ae7tRGBHYiMWPB89yYdUt06dXPLqqbf
JIUhuOBwle3IOuwJPmPEwsXFXkxKCYzKDAyi+weCDKtqeUl3HNbCB0f7g/uxUDhQ
wRzZ1v4g35LHh5zQR26dizl4Uj2dNLL6K5xkPwIDAQABAoIBABdpTPSuSAG1BSrs
mhQQwP6swgK553//bqIkcgepedRPFjFhG+BzCaZO5BA+tT2hO6v3oE7Hvo3Rx9Mk
qHl9VBl+q4IEYhSG0YpJAUxv7CwNuHCQODan3fsJ+rHDIUdNa2zln7FehdVxReW4
y05334EwiLkGB34UXQQSJTuvv228rF5F7wCs0luZWJ9IpGNWpwI2K6pg3ME/bIb6
xFMf6jKWUUZht1m+40UG2bhsMPPknde2zAq4DjJTl7bJuuhceqp3ylm5h+oUDCl2
twyGhn8gpHrMXVCcEvxwno8aJfNJ18iZu+lIQTT1DVuRrFprGDc9dFW6zVFauv6W
ZG5AUXkCgYEAyoBUPVFaf3E6nDqq26J5x6+DBe6afS7tqNn5E/z2cI54KLiUDwjs
P9axjdVMqqUiNw0jJRi8nUe1psDLguSZo4tFSgsWbsdnZQutQz2fy87jIgZOzZpj
fqud64O/fMm8fpGBz5LQjo27FLRfQcr23KVjWeQgBSMfMk2oe8Mw8+sCgYEAwqWc
abQBpG5TqtgvR3M9W0tAK51SSZF5okKif8KMoDJ67pXX2iTgh4bnUugaYObcG3qV
5VH9xlOuN4td4RfBybzRcKJDN/oinwZhMjQKKjyhpjLInDqnF4ogLPevT6cCq8I+
rHfK6DhtqLUG4BYDD7QWOBZpCTwd3n+7Xk39v/0CgYEAoaM9mpRNgFyJRBswNpDC
VDosg5epiTLkUVtsDiBlNgMCtr5esIGW0n40y9nukGevn/HEk9/i7khHHwvVZm3C
lWCdtjSTe2l/hpCDhKCz5KMHeik+za7mrD2gmFVZi+obo4vR6jZuctt+8U/omUPB
OO5rF12YkYEvbZ+/VMrBUHECgYEArWGpvvpJ0Dc6Ld9d1e5PxCd2pKMBLmj4CNIE
P3uDmhr9J9KvsC/TFMXU/iOjg5eAjrWWGev7+pKFiBKLcDqiMtoPUZ4n9A/KkQ60
u2xhdZgGga2QxqD0P+KYoJWMQo5IschX3XbjdhD1lSaTVj4lQfKvLAzCSSiUjqIG
u40LL90CgYB9t3CISlVJ+l1zy7QPIvAul5VuhwjxBHaotdLuXzvXFJXZH2dQ9j/a
/p2GMUq+VZbKq45x4I5Z3ax462qMRDmpx9yjtrwNavDV47uf34pQrNpuWO7kQfsM
mKMH6Gf6COfSIbLuejdzSOUAmjkFpm+nwBkka1eHdAy4ALn9wNQz3w==
-----END RSA PRIVATE KEY-----
`

func TestBerToDer(t *testing.T) {
	_, err := exec.LookPath("openssl")
	if err != nil {
		t.Skipf("OpenSSL not availible skippint test.")
	}

	msg := new(bytes.Buffer)
	ui := &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: msg,
	}

	// Test - a DER encoded key commes back unchanged.
	newKey := berToDer(der_encoded_key, ui)
	if newKey != der_encoded_key {
		t.Errorf("Trying to convert a DER encoded key should return the same key.")
	}
	if string(msg.Bytes()) != "" {
		t.Errorf("Doing nothing with a DER encoded key result in no messages to the UI .")
	}

	// Test - a BER encoded key should be converted to DER.
	newKey = berToDer(ber_encoded_key, ui)
	_, err = ssh.ParsePrivateKey([]byte(newKey))
	if err != nil {
		t.Errorf("Trying to convert a BER encoded key should return a DER encoded key parsable by Go.")
	}
	if string(msg.Bytes()) != "Successfully converted BER encoded SSH key to DER encoding.\n" {
		t.Errorf("Trying to convert a BER encoded key should tell the UI about the success.")
	}
}
