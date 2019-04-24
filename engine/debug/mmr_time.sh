S#/bin/sh
################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'
func time2sec(t) {	
  x = split(t,ary,":");
  if(x!=3) {print "bad time", t,NR, $0; exit;}
  sec = (ary[1]*60+ary[2])*60+ary[3];
#  printf("time2sec(%s) %02d:%02d:%02d= %d\\n",t, ary[1]+0, ary[2]+0,ary[3]+0,sec);
  return sec;
}

#   8285 16:41:39 3-:-0 Enqueue             M-f8a39e|R-f8a39e|H-f8a39e                Missing Msg[16]:MissingMsg --> 455b7b<FNode0> asking for DBh/VMh/h[3/1/1, ] Sys: 0 msgHash[f8a39e]

#debug print lines associated with a height
 {debug = 0;}
#/2\\/0\\/1[^0-9]/ {print time2sec($2), $0;debug=1}
 
 # 7993030 15:52:44.690      10-:-0 Ask 10/2/2 1 
/Ask/         { 
      v = $5;
      if (!(v in asks)) {
          asks[v] = time2sec($2); 
          lasks[v] = asks[v];
          if(debug) {print "Ask", "<"v">", lasks[v];}
      } else {
          lasks[v] = time2sec($2); 
          if(debug) {print "lAsk", "<"v">", lasks[v];}
          if(debug){for( i in asks){print i, "<"asks[i]">";}}
      }
}

#  7991512 15:52:44.585      10-:-0 Add 11/2/0 0 
/Add/         {
     adds[$5] = time2sec($2)
     if(debug) {print "Add", "<"$5">", adds[$5];}
}

/sendout/     {
    cnt = match($0,/([0-9]+\\/[0-9]\\/[0-9]+, )+/,ary);
    list = substr($0,RSTART,RLENGTH);
    total_request_msgs_a++
 #   print "          1         2         3         4         5         6         7         8         9         0         1         2         3         4         5         6         7         8         9         0"
 #   print "012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
    #print $0
    ts =  time2sec($2);
    n = split(list,ary,", ");
    for( i in ary ) {
	if(ary[i] ~ /[0-9]+\\/[0-9]\\/[0-9]+/) {
	   total_requests_a++;
           v = ary[i];
	 
#	   if(v in asks) {print v, "ask to sendout", ts-asks[v]; } else {{print v, "sendout no prior ask";}
           if(!(v in firstsendout)) {
#              print "first", v, firstsendout[v], (v in firstsendout)
               firstsendout[v] = ts;
               lastsendout[v] = ts;
           } else {
#              print "last", v
               lastsendout[v] = ts;
           }
	    sendcnt[v]++
        }
        if(debug) {print "sendout", "<"v">", lastsendout[v];}
    }
}

#   8286 16:41:39 3-:-0 Send P2P FNode02    M-f8a39e|R-f8a39e|H-f8a39e                Missing Msg[16]:MissingMsg --> 455b7b<FNode0> asking for DBh/VMh/h[3/1/1, ] Sys: 0 msgHash[f8a39e]
/Send P2P.*MissingMsgSend P2P.*MissingMsg"/ {
#    print "MM", $2, $0
    ts =  time2sec($2);
    total_request_msgs_b++
    cnt = match($0,/([0-9]+\\/[0-9]\\/[0-9]+, )+/,ary);
    list = substr($0,RSTART,RLENGTH);
     n = split(list,ary,", ");
    for( i in ary ) {
        v = ary[i]
#	print i, v
	if(v ~ /[0-9]+\\/[0-9]\\/[0-9]+/) {
	   total_requests_b++;
	    asking[v][substr($6,1)]++
            if(!(v in firstp2p)) {
               firstp2p[v] = ts;
               lastp2p[v] = ts;
#               if(v in asks) {print "firstp2p["v"] =", firstp2p[v], asks[v];} else {print "firstp2p["v"] no prior ask"}
              } else {
               lastp2p[v] = ts;
           }
           if(debug) {print "sendP2P", "<"v">", lastp2p[v];}
        }
    }

}

#   4159 16:41:26 1-:-4 Send P2P FNode02    M-b86815|R-b86815|H-b86815       Missing Msg Response[19]:MissingMsgResponse <-- DBh/VMh/h[         1/0/45] msgHash[b86815] EmbeddedMsg: REntry-VM  0: Min:   4          -- Leader[455b7b<FNode0MissingMsgResponse |>] Entry[c1c4d4] ChainID[888888dc44] hash[c1c4d4] |    ACK-    DBh/VMh/h 1/0/45        -- Leader[455b7b<FNode0>] hash[c1c4d4]
/MissingMsgResponse/{
#    sub(/:/," ");
#   print "MMR           1         2         3         4         5         6         7         8         9         0         1         2         3         4         5         6         7         8         9         0"
#   print "MMR 012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"##
#    print "MMR " $0
    ts =  time2sec($2);
    total_responces++
    cnt = match($0,/([0-9]+\\/[0-9]\\/[0-9]+)+/,ary);
    list = substr($0,RSTART,RLENGTH);
    n = split(list,ary,", ");
    v = ary[1];
#	print list, n, v
    peer = substr($6,1);
#    print "MMR", v, peer;
	if(v ~ /[0-9]+\\/[0-9]\\/[0-9]+/) {
           if(!(v in firstMR)) {
               firstMR[v] = ts;
               lastMR[v] = ts;
#              if(v in asks) {print "firstMR["v"] =", firstMR[v], asks[v];} else {print "firstMR["v"] no prior ask", firstMR[v]}
              } else {
               lastMR[v] = ts;
           }
           peers[v][peer]++;
	   countMR[v]++;
           if(debug) {print "MMR", "<"v">", lastMR[v];}
        }
}


END {
   printf("%10s %10s %10s %10s %10s %10s %10s %10s %10s %10s %10s %-10s\\n", "loc","ask2LAsk","ask2send", "ask2last", "ask2add", "askcount", "ask2p2p", "ask2Lp2p","firstMR","lastMR", "replies","peers");
   PROCINFO["sorted_in"] ="@ind_num_asc";
   for(i in firstsendout) {
     ask = asks[i];

#     if (i=="2/0/1") {
#       print i, "ask="asks[i],"lask="lasks[i], "1send="firstsendout[i], "lsend="lastsendout[i], "add="adds[i], "1p2p="firstp2p[i], "lp2p="lastp2p[i], "1MR="firstMR[i], "lMR="lastMR[i];
#       print i, "ask="asks[i],"lask="lasks[i]-ask,  "1send="firstsendout[i]-ask, "lsend="lastsendout[i]-ask, "add="adds[i]-ask,  "1MR="firstMR[i]-ask, "lMR="lastMR[i]-ask;
#     }
     lask = lasks[i]-ask;
     fs = firstsendout[i]-ask;
     ls = lastsendout[i]-ask;
     if(i in adds)    {add = adds[i]-ask;} else {add = "never"}
     if(i in firstp2p){ fp = firstp2p[i]-ask;} else {fp = "NA";}
     if(i in lastp2p) { lp = lastp2p[i]-ask;}  else {lp = "NA";}
     if(i in firstMR) { fr = firstMR[i]-ask;} else {fr = "NA";}
     if(i in lastMR)  { lr = lastMR[i]-ask;}  else {lr = "NA";}
     delete peerCnt
     replies = countMR[i]+0
     PROCINFO["sorted_in"] ="@ind_str_asc";
     peerStr = ""
     if(i in peers) {
       
        for(j in peers[i]) {
 #         print "<"i"><"j">["peers[i][j]"]";
          peerStr = peerStr " "  j "-" peers[i][j];
        }
        peerStr = substr(peerStr,2)
     } else {peerStr = "NA";}

     if (fs > 5 || ls > 5 || add > 5|| fp > 5 )	 {
        printf("%10s %10s %10s %10s %10s %10s %10s %10s %10s %10s %10s %-10s\\n", i, lask, fs, ls, add, sendcnt[i], fp, lp, fr,lr, replies,  peerStr);
     }
   }

	printf("Total requests messages %d/%d, total requests %d/%d, total responces %d\\n", total_request_msgs_a,total_request_msgs_b,total_requests_a,total_requests_b,  total_responces);
}
EOF
################################
# End of AWK Scripts           #
################################MissingMsg

 
(grep -h "." $1_missing_messages.txt; grep -hE "MissingMsg " $1_networkoutputs.txt; grep -hE "Send P2P.*MissingMsgResponse" fnode*_networkoutputs.txt) | sort -n | awk "$scriptVariable"

#(grep -h "." $1_missing_messages.txt; grep -hE "MissingMsg " $1_networkoutputs.txt; grep -hE "Send P2P.*MissingMsgResponse" fnode*_networkoutputs.txt) | sort -n | grep -E "2/0/1[^0-9]" |  awk "$scriptVariable"
