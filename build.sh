#go install -ldflags "-X github.com/FactomProject/factomd/engine.Build=`git rev-parse HEAD` -X github.com/FactomProject/factomd/engine.FactomdVersion=`cat VERSION`" -v

# create a static shared obj
go build -v -o factomd.so -buildmode=c-shared factomd.go
