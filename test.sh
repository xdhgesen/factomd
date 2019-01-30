#!/usr/bin/env bash

# NOTE: this script modified from mainline factomd and is used for a simulator benchmark

# you can set some ENV vars to tune this test
# GIVEN_NODES=L MAX_BLOCKS=200 BATCHES=10 ENTRIES=1000 DELAY_BLOCKS=1 DROP_RATE=0 ./test.sh

# GIVEN_NODES: encodes a networks setup LLF is 2 leaders and 1 follower
# MAX_BLOCKS: how many blocks do we wait for all batches to clear
# DELAY_BLOCKS: is the number of blocks to wait after we see empty holding
# ENTRIES: how many entries to write during each batch
# BATCHES: how many times to send ENTRIES & wait for holding to empty + repeats
# DROP_RATE: number 1-1000 to specify message drop rate

# run same tests as specified in .circleci/config.yml
#PACKAGES=$(glide nv | grep -v Utilities | grep -v LongTests)
PACKAGES=('./simTest/...')
FAIL=""

for PKG in ${PACKAGES[*]} ; do
  go test -v -vet=off -timeout 99m $PKG &>> ./test.out
  if [[ $? != 0 ]] ;  then
    FAIL=1
  fi
done

if [[ "${FAIL}x" != "x" ]] ; then
  echo "TESTS FAIL"
  #cat ./test.out
  exit 1
else
  echo "-------------------DETAILED LOG-------------------"
  #cat simTest/fnode0_simtest.txt
  echo "ALL TESTS PASS"
  exit 0
fi

