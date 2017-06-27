package engine

// All plugins we can intiate

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/hashicorp/go-plugin"
)

// How often to check the plugin if it has messages ready
var CHECK_BUFFER time.Duration = 2 * time.Second

var _ log.Logger
var _ = ioutil.Discard

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"etcd": &IEtcdPlugin{},
}

var etcdHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "Etcd_Manager",
	MagicCookieValue: "factom_etcd",
}

func LaunchEtcdPlugin(path, addr, uid, prefix string) (interfaces.IEtcdManager, error) {
	// So we don't get debug logs. Comment this out if you want to keep plugin
	// logs
	//log.SetOutput(ioutil.Discard)

	// We're a host! Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: etcdHandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(path+"etcd-manager", "plugin", addr, uid, prefix),
	})

	// Make sure we close our client on close
	AddInterruptHandler(func() {
		fmt.Println("Etcd plugin is now closing...")
		client.Kill()
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Println("RpcClient (etcd plugin) connect issue:", err)
		return nil, err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("etcd")
	if err != nil {
		log.Println("RpcClient (etcd plugin) dispense issue:", err)
		return nil, err
	}

	etcdeManager := raw.(interfaces.IEtcdManager)

	return etcdeManager, nil
}
