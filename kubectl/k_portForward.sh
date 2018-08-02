#/bin/sh
################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'
BEGIN {
   print "remove old errors file"
#   system("rm -f errors")
   print "kill all old port forwards"
#   system("killall kubectl >> errors");
}
 {print;}

/factomd/ {
 id = substr($1,index($1,"-")+1);
 cmd = sprintf("kubectl port-forward %s %d:8090 %d:8093 %d:8088 >> errors &", $1, 8091+id, 8101+id, 8111+id)
 print cmd;
 system(cmd);
 listOfNodes = sprintf("%s http://localhost:%d", listOfNodes, 8091+id);
 
 }
END {
   startChome = "google-chrome --new-window " listOfNodes;
   print startChome;
   # wait till the ports forwards are up...
   system("trap 'exit 1' 2; sleep 5")
   system(startChrome);
   # when chrome exits kill the port forwards
#   print "kill all old port forwards"
#   system("killall kubectl");
}
EOF
################################
# End of AWK Scripts           #
################################
kubectl get pods -l role=member | awk "$scriptVariable"
