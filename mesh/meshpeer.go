package mesh

import (
	"io/ioutil"
	"log"
	"net"
	"strconv"

	"os"

	"strings"

	"bytes"
	"encoding/gob"

	"fmt"

	"math/rand"

	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/mesh/meshconn"
	"github.com/FactomProject/factomd/p2p"
	"github.com/weaveworks/mesh"
)

type MeshNetwork struct {
	Config p2p.ControllerInit
	Logger *log.Logger
	router *mesh.Router
	send   *meshconn.Peer

	FromNetwork chan interface{} // Channel to the app for network data
	ToNetwork   chan interface{} // Parcels from the app for the network

}

func NewMeshPeer(ci p2p.ControllerInit) *MeshNetwork {
	p := new(MeshNetwork)
	p.Config = ci
	p.Logger = log.New(os.Stderr, "> ", log.LstdFlags)
	return p
}

func (m *MeshNetwork) Init() *MeshNetwork {
	password := "" // No password
	name := mesh.PeerNameFromBin(primitives.RandomHash().Bytes())

	nickname := name.String() // MustHostname()
	m.Config.Port = ":" + m.Config.Port
	host, portStr, err := net.SplitHostPort(m.Config.Port)
	if err != nil {
		m.Logger.Fatalf("mesh address: %s: %v", m.Config.Port, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		m.Logger.Fatalf("mesh address: %s: %v", m.Config.Port, err)
	}

	router, err := mesh.NewRouter(mesh.Config{
		Host:               host,
		Port:               port,
		ProtocolMinVersion: mesh.ProtocolMinVersion,
		Password:           []byte(password),
		ConnLimit:          64,
		PeerDiscovery:      true,
		TrustedSubnets:     []*net.IPNet{},
	}, name, nickname, mesh.NullOverlay{}, log.New(ioutil.Discard, "", 0))
	m.router = router
	m.router.Start()

	errs := m.router.ConnectionMaker.InitiateConnections(strings.Split(m.Config.CmdLinePeers, " "), true)
	for _, e := range errs {
		fmt.Println(e)
	}

	peer := meshconn.NewPeer(name, 0, m.Logger)
	gossip, err := router.NewGossip("factomd", peer)
	if err != nil {
		m.Logger.Fatalf("Could not create gossip: %v", err)
	}

	peer.Register(gossip)
	m.send = peer

	m.FromNetwork = make(chan interface{}, p2p.StandardChannelSize) // Channel to the app for network data
	m.ToNetwork = make(chan interface{}, p2p.StandardChannelSize)   // Parcels from the app for the network
	return m
}

func (m *MeshNetwork) StartNetwork() {
	go m.ManageOutChannel()
	go m.ManageInChannel()
}

func encode(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decode(d []byte) (*p2p.Parcel, error) {
	var p p2p.Parcel
	if err := gob.NewDecoder(bytes.NewBuffer(d)).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// manageOutChannel takes messages from the f.broadcastOut channel and sends them to the network.
func (f *MeshNetwork) ManageOutChannel() {
	for {
		select {
		case data := <-f.ToNetwork:
			switch data.(type) {
			case p2p.Parcel:
				parcel := data.(p2p.Parcel)
				parcel.Header.MeshSource = f.send.LocalAddr().(meshconn.MeshAddr)
				parcel.UpdateHeader()
				b, err := encode(parcel)
				if err != nil {
					f.Logger.Printf("Mesh (1) %s", err.Error())
				}

				if parcel.Header.TargetPeer == p2p.BroadcastFlag {
					f.send.Write(b)
				} else if parcel.Header.TargetPeer == p2p.RandomPeerFlag {
					all := f.router.Peers.Descriptions()
					i := rand.Intn(len(all))
					f.send.WriteTo(b, meshconn.MeshAddr{PeerName: all[i].Name, PeerUID: all[i].UID})
				} else {
					f.send.WriteTo(b, parcel.Header.MeshTarget)
				}
			default:
				f.Logger.Printf("Garbage on f.BrodcastOut. %+v", data)
			}
		}
	}
}

// manageInChannel takes messages from the network and stuffs it in the f.BroadcastIn channel
func (m *MeshNetwork) ManageInChannel() {
	for {
		d, from, err := m.send.ReadFrom()
		//m, err := p.OnGossip(b)
		if err != nil {
			m.Logger.Println(err)
			continue
		}

		p, err := decode(d)
		if err != nil {
			m.Logger.Println(err)
			continue
		}

		// m.Logger.Printf("Msg from %s\n", from.String())

		p.Header.PeerAddress = from.String()
		p.Header.MeshSource = from.(meshconn.MeshAddr)
		m.FromNetwork <- *p
	}
}

func (m *MeshNetwork) Close() {
	m.router.Stop()
}

func MustHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return hostname
}
