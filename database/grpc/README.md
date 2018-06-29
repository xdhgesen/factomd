# Database GRPC

Ability to access raw database from other languages.

# Requirements

- GoLang
- Glide
- A Factomd database

# Running the GRPC Server

**Run `glide install` in the `factomd` directory before continuing**

The grpc server opens a factomd database and exposes grpc functions for a client. To run the server:
 - You may provide a different path

```
go run grpcserver/*.go -path=$HOME/.factom/m2/main-database/ldb/MAIN/factoid_level.db -port 10000
```

You can also use `go install` to avoid needing to run from this directory.

## GRPC Code Gen

```
protoc -I . db.proto --go_out=plugins=grpc:shared
```

```
protoc --elixir_out=. db.proto

# With services
protoc --elixir_out=plugins=grpc:./lib/ db.proto
```