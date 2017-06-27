package engine_test

import (
	"bytes"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/engine"
	"github.com/hashicorp/go-plugin"
)

// FakeEtcdInstance is a fake etcd plugin
type FakeEtcdInstance struct{}

func (f *FakeEtcdInstance) SendIntoEtcd(msg []byte) error           { return nil }
func (f *FakeEtcdInstance) GetData() []byte                         { return nil }
func (f *FakeEtcdInstance) Reinitiate() error                       { return nil }
func (f *FakeEtcdInstance) NewBlockLease(blockHeight uint32) error  { return nil }
func (f *FakeEtcdInstance) PickUpFromHash(messageHash string) error { return nil }
func (f *FakeEtcdInstance) Ready() (bool, error)                    { return true, nil }

// PluginRPCConn returns a plugin RPC client and server that are connected
// together and configured.
func PluginRPCConn(t *testing.T, ps map[string]plugin.Plugin) (*plugin.RPCClient, *plugin.RPCServer) {
	// Create two net.Conns we can use to shuttle our control connection
	clientConn, serverConn := plugin.TestConn(t)

	// Start up the server
	server := &plugin.RPCServer{Plugins: ps, Stdout: new(bytes.Buffer), Stderr: new(bytes.Buffer)}
	go server.ServeConn(serverConn)

	// Connect the client to the server
	client, err := plugin.NewRPCClient(clientConn, ps)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return client, server
}

// TestEtcdImpl just checks the plugin implementation of the interface
func TestEtcdImpl(t *testing.T) {
	x := new(IEtcdPlugin)
	x.Impl = new(FakeEtcdInstance)
	client, _ := PluginRPCConn(t, map[string]plugin.Plugin{
		"etcd": x,
	})

	raw, err := client.Dispense("etcd")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Client working
	mc := raw.(interfaces.IEtcdManager)
	err = mc.SendIntoEtcd(nil)
	if err != nil {
		t.Error(err)
	}

	v := mc.GetData()
	if v != nil {
		t.Error("Should be nil")
	}

	err = mc.Reinitiate()
	if err != nil {
		t.Error(err)
	}

	err = mc.NewBlockLease(0)
	if err != nil {
		t.Error(err)
	}

	b, err := mc.Ready()
	if err != nil {
		t.Error(err)
	}
	if !b {
		t.Error("Should be true")
	}

	err = mc.PickUpFromHash("9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08")
	if err != nil {
		t.Error(err)
	}

	// Client closed
	client.Close()
	err = mc.SendIntoEtcd(nil)
	if err == nil {
		t.Error("Stream closed, SendIntoEtcd should fail")
	}

	v = mc.GetData()
	if v != nil {
		t.Error("Should be nil")
	}

	err = mc.Reinitiate()
	if err == nil {
		t.Error("Stream closed, Reinitiate should fail")
	}

	err = mc.NewBlockLease(0)
	if err == nil {
		t.Error("Stream closed, NewBlockLease should fail")
	}

	b, err = mc.Ready()
	if err == nil {
		t.Error("Stream closed, Ready should fail")
	}
	if b {
		t.Error("Should be false")
	}

	err = mc.PickUpFromHash("9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08")
	if err == nil {
		t.Error("Stream closed, PickUpFromHash should fail")
	}
}
