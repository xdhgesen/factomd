#!/usr/bin/env bash

# run same tests as specified in .circleci/config.yml
#PACKAGES=$(glide nv | grep -v Utilities | grep -v LongTests)
PACKAGES=('./simTest/...')
FAIL=""

for PKG in ${PACKAGES[*]} ; do
  go test -v -vet=off $PKG &> ./test.out
  if [[ $? != 0 ]] ;  then
    FAIL=1
  fi
done

if [[ "${FAIL}x" != "x" ]] ; then
  echo "TESTS FAIL"
  cat ./test.out
  exit 1
else
  echo "-------------------DETAILED LOG-------------------"
  cat simTest/FNode0_simTest.txt
  echo "ALL TESTS PASS"
  exit 0
fi
