# Database GRPC

Ability to access raw database from other languages.

# Running the GRPC Server

The grpc server opens a factomd database and exposes grpc functions for a client. To run the server:
 - You may provide a different path

```
go run *.go -path=$HOME/.factom/m2/main-database/ldb/MAIN/factoid_level.db -port 10000
```

## GRPC Code Gen

```
protoc -I . db.proto --go_out=plugins=grpc:shared
```

```
protoc --elixir_out=. db.proto

# With services
protoc --elixir_out=plugins=grpc:./lib/ db.proto
```