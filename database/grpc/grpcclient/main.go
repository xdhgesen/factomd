package main

import (
	"context"
	"flag"

	log "github.com/sirupsen/logrus"

	"time"

	"io"

	"github.com/FactomProject/factomd/common/primitives"
	pb "github.com/FactomProject/factomd/database/grpc/shared"
	"google.golang.org/grpc"
)

var (
	//tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	//caFile             = flag.String("ca_file", "", "The file containning the CA root cert file")
	serverAddr = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")

//serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")

)

func get(client pb.DatabaseGrpcClient, key *pb.DBKey) {
	//log.Printf("Getting Key (%#v)", key)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	val, err := client.Retrieve(ctx, key)
	if err != nil {
		log.Fatalf("%v.Retrieve(_) = _, %v: ", client, err)
	}
	log.Println(val)
}

func getStream(client pb.DatabaseGrpcClient, key *pb.DBKey) {
	log.Printf("Getting Key (%x)", key.Key)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.RetrieveAllEntries(ctx, key)
	if err != nil {
		log.Fatalf("%v.RetrieveAllEntries(_) = _, %v", client, err)
	}
	for {
		val, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.RetrieveAllEntries(_) = _, %v", client, err)
		}
		log.Println(val)
	}
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	//if *tls {
	//	if *caFile == "" {
	//		*caFile = testdata.Path("ca.pem")
	//	}
	//	creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
	//	if err != nil {
	//		log.Fatalf("Failed to create TLS credentials %v", err)
	//	}
	//	opts = append(opts, grpc.WithTransportCredentials(creds))
	//} else {
	opts = append(opts, grpc.WithInsecure())
	//}
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewDatabaseGrpcClient(conn)

	hash, _ := primitives.HexToHash("a85ef783459a465572df902f3b909d76efdd9ba520c1c3ee7ad216bcfde33c9a")
	//get(client, &pb.DBKey{
	//	Key:     hash.Bytes(),
	//	KeyType: "entry",
	//})

	hash, _ = primitives.HexToHash("c6783746070aa3fdbc64fa3d29183e2675d9aa177a9b5c1c99a064bc522e5abd")
	getStream(client, &pb.DBKey{
		Key:     hash.Bytes(),
		KeyType: "dblock",
	})
}
