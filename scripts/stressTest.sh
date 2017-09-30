#!/bin/bash

date > out.txt
date > err.txt
date > trans.txt

runtrans () {
	echo Start Transactions
	sleep 20
	scripts/multifactoidtrans.sh >> trans.txt
}

for ((i=0; i<100000; i++)); do

	runtrans &
	gawk 'BEGIN {print"s";system("sleep 15");print"S50";system("sleep 5");print"F50";system("sleep 5");print"r";system("sleep 5");print"Vt"}' | g factomd  -count=25 -net=alot+  -blktime=60 -faulttimeout=20 -enablenet=true -network=LOCAL -startdelay=1 $@ >> out.txt 2>> err.txt
	echo "Restarting simulation"
	mypid=$(ps -ef | grep multifactoidtrans.sh | grep bash | awk '!o{o=1;print $2}')
	kill -9 $mypid
	DatabaseIntegrityCheck level ~/.factom/m2/local-database/ldb/LOCAL/factoid_level.db/
done

