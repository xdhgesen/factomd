#/bin/sh
# set -x

if [ -z "$1" ]
  then
    echo excluding long test_resultss
    pattern='(?<!_long)$'
  else
    pattern="$1"
fi
if [ -z "$2" ]
  then
    echo excluding debug test_results and long test_resultss
    npattern="TestPass|TestFail|TestRandom|long"
  else
    npattern="$2|TestPass|TestFail|TestRandom"
fi



echo preparing to run: -$pattern- -$npattern-
grep -Eo " Test[^( ]+" simTest/*_test.go | grep -P "$pattern" | grep -Pv "$npattern" | sort
sleep 3

mkdir -p test_results
#remove old logs
grep -hEo " Test[^( ]+" simTest/*_test.go | grep -P "$pattern" | grep -Pv "$npattern" |  xargs -n 1 -I test_name rm -rf test_results/test_name

#compile the tests
go test -c github.com/FactomProject/factomd/simTest -o test_results/factomd_test

#run the tests
grep -hEo " Test[^( ]+" simTest/*_test.go | grep -P "$pattern" | grep -Pv "$npattern" | sort | xargs -I TestMakeALeader -n1 bash -c  'echo "Run TestMakeALeader"; mkdir -p test_results/TestMakeALeader; cd test_results/TestMakeALeader; ../factomd_test --test.v --test.timeout 30m  --test.run "^TestMakeALeader$" &> test_log.txt; pwd; grep -EH "PASS:|FAIL:|panic|bind| Timeout "  test_log.txt'

echo "Results:"
find test_results -name test_log.txt | sort | xargs grep -EHm1 "PASS:"
echo ""
find test_results -name test_log.txt | sort | xargs grep -EHm1 "FAIL:|panic|bind| Timeout "



#(echo git checkout git rev-parse HEAD; find . -name test_log.txt | xargs grep -EH "PASS:|FAIL:|panic") | mail -s "Test results `date`" `whoami`@factom.com

