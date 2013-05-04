package rpc

import (
	"cgl.tideland.biz/asserts"
	"net/rpc"
	"testing"
)

func TestServer_Address_PanicIfNotStarted(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	defer func() {
		p := recover()
		assert.NotNil(p, "should panic")
		assert.Equal(p.(string), "Server not listening.", "right panic")
	}()

	NewServer().Address()
}

func TestServer_Start(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	s := NewServer()

	// Verify it can start
	err := s.Start()
	assert.Nil(err, "should start without err")
	addr := s.Address()

	// Verify we can connect to it!
	_, err = rpc.Dial("tcp", addr)
	assert.Nil(err, "should be able to connect to RPC")

	// Verify it stops
	s.Stop()
	_, err = rpc.Dial("tcp", addr)
	assert.NotNil(err, "should NOT be able to connect to RPC")
}

func TestServer_RegisterUi(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	ui := &testUi{}

	// Start the server with a UI
	s := NewServer()
	s.RegisterUi(ui)
	assert.Nil(s.Start(), "should start properly")
	defer s.Stop()

	// Verify it works
	client, err := rpc.Dial("tcp", s.Address())
	assert.Nil(err, "should connect via RPC")

	uiClient := &Ui{client}
	uiClient.Say("format")

	assert.Equal(ui.sayFormat, "format", "format should be correct")
}
