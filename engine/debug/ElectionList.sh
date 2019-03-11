#/bin/sh
f=$1
shift
pattern="$1"
shift

if [ -z "$f" ]; then
    f="fnode0_election.txt"
fi
if [ -z "$pattern" ]; then
    pattern="."
fi

################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'
BEGIN{  PROCINFO["sorted_in"]="@ind_num_asc";}

 {if (NR%1024 == 0) {printf("%40s:%d\\r", FILENAME, NR)>"/dev/stderr";}}

function print_list(t, height,feds_array,auds_array){
    PROCINFO["sorted_in"]="@ind_num_asc";
    str = "";
    
    str = sprintf("feds[ ");
#    for(i in feds_array) {str = str sprintf("%d:%s, ", i-1, feds_array[i]);} 
    for(i in feds_array) {str = str sprintf("%s ", feds_array[i]);} 
    str = str sprintf("] auds[ ");
#    for(i in auds_array) {str = str sprintf("%d:%s, ", i-1, auds_array[i]);} 
    for(i in auds_array) {str = str sprintf("%s ", auds_array[i]);} 
    str = str sprintf("]");
    
    if(str != prev ) {
 #     print ">"prev;
 #     print "<"str;
      print t, height, str;
      prev = str;
    }
}
 
 # 553197715 08:24:00.991 180791-:-0 exec -1                                 M-??????|R-??????|H-??????|0xc000e02a80           INTERNALAUTHLIST[35]:AuthorityListInternal DBH 180791 fed [0180b0, ... ff0fa6, ] aud[04081f,..., f5b2bb, ]  

 /AuthorityListInternal/ {
     x = index($0, "fed [")+5;
     y = index($0, ", ] aud[")+8;
     z = index($0,"DBH ")+4;
    feds = substr($0,x, y-x-8);
    auds = substr($0,y, length($0)-y-4);
    height = substr($0,z,x-z-5) + 0
#    print "--->", $0;
    split(feds,feds_array,/, /)
    split(auds,auds_array,/, /)
    
#    print "values",x,y,z, height;
#    print "**feds <"feds">", length(feds_array);
#    print "**auds <"auds">", length(auds_array);
#    exit(1);

    print_list($2, height"-:-0",feds_array,auds_array)
    
#    feds_list[prev_height] = feds_array
#    auds_list[prev_height] = auds_array
 }

 #554651406 08:27:34.231 180791-:-0 **** FedVoteLevelMsg       FNode0 Swapping Fed: 15(7529d6) Audit: 17(b02c99) 
 /Swapping Fed/ {

  height = $3;
  x = index($9,"(");
  fi = substr($9,0,x-1);
  fid = substr($9,x+1,6);
  y = index($11,"(");
  ai = substr($11,0,y-1);
  aid = substr($11,y+1,6);
  
  this = sprintf("%s:%d(%s) %d(%s)",height,fi,fid, ai,aid)
  
  ## ignore duplicates
  if(this == prev){next;}
  
#  print "-->", $0;
#  
#  print height,fi,fid, ai,aid;
#  
#  print "feds";   for(i in feds_array) {print "f " i-1 " : "feds_array[i];}
#  print "auds";   for(i in auds_array) {print "a " i-1 " : "auds_array[i];}
# 
  if(feds_array[fi+1] != fid) {
   print;
   print  "bad fed", height, fi, fid, feds_array[fi+1];
   print this
   print prev, this == prev
  }
  if(auds_array[ai+1] != aid) {
   print;
   print  "bad aud", height, ai, aid, auds_array[fi+1];
  }
  
  #swap them  
  feds_array[fi+1]=aid
  auds_array[ai+1]=fid
  
  print_list($2, height,feds_array,auds_array);
  prev = this;
  
  
 }

EOF
################################
# End of AWK Scripts           #
################################
 grep -E "Auth|Swapping"  $f | grep -E "$pattern" | gawk "$scriptVariable"  
