package main

import (
	"flag"
	"fmt"
	"net"

	pb "github.com/FactomProject/factomd/database/grpc/shared"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"strings"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/boltdb"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/leveldb"
	log "github.com/sirupsen/logrus"
)

var (
	//tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	//certFile   = flag.String("cert_file", "", "The TLS cert file")
	//keyFile    = flag.String("key_file", "", "The TLS key file")
	port   = flag.Int("port", 10000, "The server port")
	dbpath = flag.String("path", "", "Path to database")
	dbtype = flag.String("db", "level", "DBType of level or bolt")
)

var _ pb.DatabaseGrpcServer = (*databaseGrpcServer)(nil)

type databaseGrpcServer struct {
	Raw     interfaces.IDatabase
	Overlay interfaces.DBOverlay
}

func getHash(key *pb.DBKey) (interfaces.IHash, error) {
	if len(key.Key) == 32 {
		return primitives.NewHash(key.Key), nil
	}
	return nil, fmt.Errorf("Key must be 32 bytes, found %d", len(key.Key))
}

func marshalDBResponse(marshallable interfaces.BinaryMarshallable, err error) ([]byte, error) {
	if err != nil {
		return []byte{}, err
	}
	return marshallable.MarshalBinary()
}

func (db *databaseGrpcServer) Retrieve(ctx context.Context, key *pb.DBKey) (*pb.DBValue, error) {
	var err error
	h := &pb.EmptyUnmarshaler{}
	var hash interfaces.IHash
	switch strings.ToLower(key.KeyType) {
	case "entry", "ent":
		var data []byte
		hash, err = getHash(key)
		if err == nil {
			data, err = marshalDBResponse(db.Overlay.FetchEntry(hash))
			h.Data = data
		}
	default:
		// Default resorts to raw key lookup
		_, err = db.Raw.Get(key.Bucket, key.Key, h)
	}
	return &pb.DBValue{
		Value: h.Data,
	}, err
}

func (db *databaseGrpcServer) RetrieveAllEntries(key *pb.DBKey, stream pb.DatabaseGrpc_RetrieveAllEntriesServer) error {
	hash, err := getHash(key)
	if err != nil {
		return err
	}
	var eblocks []interfaces.IEntryBlock

	switch strings.ToLower(key.KeyType) {
	case "dblock":
		dblock, err := db.Overlay.FetchDBlock(hash)
		if err != nil {
			return err
		}

		eblocksEntries := dblock.GetEBlockDBEntries()
		for _, ebe := range eblocksEntries {
			eblock, err := db.Overlay.FetchEBlock(ebe.GetKeyMR())
			if err != nil {
				return err
			}

			if eblock == nil {
				return fmt.Errorf("EntryBlock %x not found from DirectoryBlock", ebe.GetKeyMR())
			}
			eblocks = append(eblocks, eblock)
		}
	case "eblock":
		eblock, err := db.Overlay.FetchEBlock(hash)
		if err != nil {
			return err
		}

		if eblock == nil {
			return fmt.Errorf("EntryBlock %x not found", hash.Fixed())
		}

		eblocks = append(eblocks, eblock)
	default:
		return fmt.Errorf("Must provide type of 'dblock' or 'eblock'")

	}

	// Fetch all entries
	for _, eb := range eblocks {
		entryHashes := eb.GetEntryHashes()
		eblockKeyMr, _ := eb.KeyMR()
		container := eblockKeyMr.Bytes()
		for i, ent := range entryHashes {
			if ent.IsMinuteMarker() {
				continue
			}
			val, err := db.Overlay.FetchEntry(ent)
			if err != nil {
				stream.Send(&pb.DBValue{
					Error: err.Error(),
				})
			} else if val == nil {
				stream.Send(&pb.DBValue{
					Error: fmt.Sprintf("Entry %x not found", ent.Fixed()),
				})
			} else {
				data, _ := val.MarshalBinary()
				stream.Send(&pb.DBValue{
					Value:       data,
					ValType:     "entry",
					Sequence:    int32(i),
					ContainedIn: container,
				})
			}
		}
	}

	return nil
}

func newServer() *databaseGrpcServer {
	s := &databaseGrpcServer{}
	return s
}

func main() {
	flag.Parse()
	log.Infof("Running server on port %d", *port)
	server := newServer()
	var raw interfaces.IDatabase
	var err error
	if *dbpath == "" {
		usage()
		log.Fatalf("Expect path to db")
	}

	// Open Database
	switch *dbtype {
	case "level", "lvl":
		raw, err = leveldb.NewLevelDB(*dbpath, false)
	case "bolt":
		raw = boltdb.NewBoltDB([][]byte{}, *dbpath)
	default:
		log.Fatalf("Expect 'level' or 'bolt', found %s", *dbtype)
	}
	if err != nil {
		log.Fatalf(err.Error())
	}
	server.Raw = raw
	server.Overlay = databaseOverlay.NewOverlay(raw)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	//if *tls {
	//	if *certFile == "" {
	//		*certFile = testdata.Path("server1.pem")
	//	}
	//	if *keyFile == "" {
	//		*keyFile = testdata.Path("server1.key")
	//	}
	//	creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
	//	if err != nil {
	//		log.Fatalf("Failed to generate credentials %v", err)
	//	}
	//	opts = []grpc.ServerOption{grpc.Creds(creds)}
	//}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterDatabaseGrpcServer(grpcServer, server)
	grpcServer.Serve(lis)
}

func usage() {
	fmt.Printf("grpcserver -dbtype level -path=PATH/TO/DB")
}
