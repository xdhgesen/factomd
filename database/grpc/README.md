# Database GRPC

Ability to access raw database from other languages.

## Shared Compile

```
protoc -I . db.proto --go_out=plugins=grpc:shared
```

```
protoc --elixir_out=. db.proto

# With services
protoc --elixir_out=plugins=grpc:./lib/ db.proto
```