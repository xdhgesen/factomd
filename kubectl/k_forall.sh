#/bin/sh
################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'
BEGIN {\\
    for(i=1;i<ARGC;i++) {
       if (ARGV[i] == "-p" || ARGV[i]=="--parallel") {
         parallel = "&";
         delete ARGV[i];
       } else if (ARGV[i] == "-h" || ARGV[i]=="--help") {
         help = 1;
         break;
       } else if (ARGV[i] == "-c" || ARGV[i]=="--container") {
         container = "-c " ARGV[i+1];
         delete ARGV[i];
         delete ARGV[i+1];
         i++;
       } else if(pattern == "") {
         pattern = ARGV[i];
         delete ARGV[i];
       } else {
         command = command ARGV[i] " ";
         delete ARGV[i];
       }
    }
    print length(ARGV), container, "\\"" command "\\"", parallel, help
    if (help || length(ARGV) > 1 || command == "") {
      print "Execute a shell command on your pods";
      print "k_forall.sh [-p|--parallel] [-c conatiner] <regex to select pods> <command to execute>"
      exit(0);
    }
}
 
func execOnPod(pod, container, cmd, parallel) {
    cmd=sprintf("kubectl exec -it %s %s  -- sh -c \\"%s\\" %s", pod, container, cmd, parallel);
    print cmd; 
    system(cmd);
}

 {
    execOnPod($1, container, command, parallel);
}

END {
   
}
EOF
################################
# End of AWK Scripts           #
################################
kubectl get pods -l role=member | awk "$scriptVariable" $@
