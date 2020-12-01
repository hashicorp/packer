package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/url"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/mitchellh/go-vnc"
	"golang.org/x/net/websocket"
)

type StepVNCConnect struct {
	VNCEnabled         bool
	VNCOverWebsocket   bool
	InsecureConnection bool
	DriverConfig       *DriverConfig
}

func (s *StepVNCConnect) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if !s.VNCOverWebsocket && !s.VNCEnabled {
		return multistep.ActionContinue
	}
	ui := state.Get("ui").(packersdk.Ui)

	var c *vnc.ClientConn
	var err error

	if s.VNCOverWebsocket {
		ui.Say("Connecting to VNC over websocket...")
		c, err = s.ConnectVNCOverWebsocketClient(state)
	} else {
		ui.Say("Connecting to VNC...")
		c, err = s.ConnectVNC(state)
	}

	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("vnc_conn", c)
	return multistep.ActionContinue
}

func (s *StepVNCConnect) ConnectVNCOverWebsocketClient(state multistep.StateBag) (*vnc.ClientConn, error) {
	driver := state.Get("driver").(*ESX5Driver)

	// Acquire websocket ticket
	ticket, err := driver.AcquireVNCOverWebsocketTicket()
	if err != nil {
		err := fmt.Errorf("Error acquiring vnc over websocket ticket: %s", err)
		state.Put("error", err)
		return nil, err
	}
	host := ticket.Host
	if len(host) == 0 {
		host = s.DriverConfig.RemoteHost
	}
	port := ticket.Port
	if port == 0 {
		port = 443
	}

	websocketUrl := fmt.Sprintf("wss://%s:%d/ticket/%s", host, port, ticket.Ticket)
	log.Printf("[DEBUG] websocket url: %s", websocketUrl)
	u, err := url.Parse(websocketUrl)
	if err != nil {
		err := fmt.Errorf("Error parsing websocket url: %s\n", err)
		state.Put("error", err)
		return nil, err
	}
	origin, err := url.Parse("http://localhost")
	if err != nil {
		err := fmt.Errorf("Error parsing websocket origin url: %s\n", err)
		state.Put("error", err)
		return nil, err
	}

	// Create the websocket connection and set it to a BinaryFrame
	websocketConfig := &websocket.Config{
		Location:  u,
		Origin:    origin,
		TlsConfig: &tls.Config{InsecureSkipVerify: s.InsecureConnection},
		Version:   websocket.ProtocolVersionHybi13,
		Protocol:  []string{"binary"},
	}
	nc, err := websocket.DialConfig(websocketConfig)
	if err != nil {
		err := fmt.Errorf("Error Dialing: %s\n", err)
		state.Put("error", err)
		return nil, err
	}
	nc.PayloadType = websocket.BinaryFrame

	// Setup the VNC connection over the websocket
	ccconfig := &vnc.ClientConfig{
		Auth:      []vnc.ClientAuth{new(vnc.ClientAuthNone)},
		Exclusive: false,
	}
	c, err := vnc.Client(nc, ccconfig)
	if err != nil {
		err := fmt.Errorf("Error setting the VNC over websocket client: %s\n", err)
		state.Put("error", err)
		return nil, err
	}
	return c, nil
}

func (s *StepVNCConnect) ConnectVNC(state multistep.StateBag) (*vnc.ClientConn, error) {
	vncIp := state.Get("vnc_ip").(string)
	vncPort := state.Get("vnc_port").(int)
	vncPassword := state.Get("vnc_password")

	nc, err := net.Dial("tcp", fmt.Sprintf("%s:%d", vncIp, vncPort))
	if err != nil {
		err := fmt.Errorf("Error connecting to VNC: %s", err)
		state.Put("error", err)
		return nil, err
	}

	auth := []vnc.ClientAuth{new(vnc.ClientAuthNone)}
	if vncPassword != nil && len(vncPassword.(string)) > 0 {
		auth = []vnc.ClientAuth{&vnc.PasswordAuth{Password: vncPassword.(string)}}
	}

	c, err := vnc.Client(nc, &vnc.ClientConfig{Auth: auth, Exclusive: true})
	if err != nil {
		err := fmt.Errorf("Error handshaking with VNC: %s", err)
		state.Put("error", err)
		return nil, err
	}
	return c, nil
}

func (s *StepVNCConnect) Cleanup(multistep.StateBag) {}
