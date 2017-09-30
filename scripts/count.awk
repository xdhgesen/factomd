/So far, we/ {
	partition = $6
	stalled = $10
	if (lp > partition) {
		printf("+")
		tp += lp
		ts += ls
	}
	lp = partition
	ls = stalled
}

/^[ ]*[0-9]+\/[0-9]+\/[0-9]/ {
	gsub(/\/.*/,"") 
	gsub(/[ ]*/,"")
	v = $1
	if (last > v ) {
		printf (".")
		total += last
		reboot++
	}
	last = v
}
END{
	tp += partition
	ts += stalled
	total += last
	print
	print "Number of reboots:         " reboot
	print "Total Factoid Transactions " total
	print "Total Partitions:          " tp
	print "Total Stalled Nodes:       " ts
}
