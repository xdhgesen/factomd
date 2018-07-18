#/bin/sh
grep -Eo " Test[^( ]+" factomd_test.go | xargs -I testname -n1 sh -c "mkdir -p test/testname; cd test/testname; go test -v github.com/FactomProject/factomd/engine -run testname" 
