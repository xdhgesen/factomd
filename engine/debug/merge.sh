#!/usr/bin/env bash
# merge multiple input log files
################################
# AWK scripts                  #
################################

#2019-07-03 12:38:49.387237662 -0500 CDT m=+0.016776816
#        1 12:38:49.387       0-:-0 SetLeaderTimestamp(2019-07-03 12:38:49) 
#     1749 12:38:56.040       1-:-1 from API, Enqueue2                     M-bb003c|R-bb003c|H-757dc8|0xc4216163c0               Commit Chain[ 5]:CChain-VM  0: entryhash[757dc8] hash[757dc8]...
#123456789 

read -d '' scriptVariable << 'EOF'
/^[0-9]{4}-[0-9]{2}-[0-9]{2} / {next;} # drop file time stamp

NR==2 {if(!($1~/^[a-zA-Z0-9_\.]+:/)) {printf("Please grep with -H\\n<%s>\\n",$1)>"/dev/stderr"; exit(1);}}

{
   sub(/from /,"")
   l = index($0,":") # find the end of the file name
   fname = substr($0,1,l); #seperate that
   seq =   substr($0,l+1,9); #grab the sequence number
   rest = substr($0,l+11) 
   gsub(/^ +/,"",rest); # compress leading spaces 
   
#   printf("%d <%s><%s><%s>\\n", l, fname, seq, rest);
   
   m = index(rest,"M-") # find the message hash
   if(m==0) {
     note = rest;
     msg ="";
   } else {
     note = substr(rest,1,m-1); # seperate the note
     msg = substr(rest,m);      # from the message
   }
   printf("%-30s %s  %-40s %s\\n", fname, seq, note, msg);
}

EOF
################################
# End of AWK Scripts           #
################################

if [ "$#" -ne 0 ]
then
grep -HE . "$@"  | awk  "$scriptVariable" | sort -nk2 | less -R
else
awk  "$scriptVariable" | sort -nk2 | less -R
fi
