package yandex

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"strings"
)

const (
	defaultContentType = "text/cloud-config"
	shellContentType   = "text/x-shellscript"
)

const (
	cloudInitIPv6Config = `#cloud-config
bootcmd:
- [ sh, -c, '/usr/bin/env dhclient -6 -D LL -nw -pf /run/dhclient_ipv6.eth0.pid -lf /var/lib/dhcp/dhclient_ipv6.eth0.leases eth0' ]
`
)

// MergeCloudUserMetaData allow merge some user-data sections
func MergeCloudUserMetaData(usersData ...string) (string, error) {
	buff := new(bytes.Buffer)
	data := multipart.NewWriter(buff)
	_, err := buff.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", data.Boundary()))
	if err != nil {
		return "", err
	}
	_, err = buff.WriteString("MIME-Version: 1.0\r\n\r\n")
	if err != nil {
		return "", err
	}

	for i, userData := range usersData {
		w, err := data.CreatePart(textproto.MIMEHeader{
			"Content-Disposition": {fmt.Sprintf("attachment; filename=\"user-data-%d\"", i)},
			"Content-Type":        {detectContentType(userData)},
		})
		if err != nil {
			return "", err
		}
		_, err = w.Write([]byte(userData))
		if err != nil {
			return "", err
		}
	}
	return buff.String(), nil
}

func detectContentType(content string) string {
	switch {
	case strings.HasPrefix(content, "#!"):
		return shellContentType
	case strings.HasPrefix(content, "#cloud-config"):
		return defaultContentType
	}

	return defaultContentType
}
