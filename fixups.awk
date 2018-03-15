# find . -name vendor -prune -o -name .git -prune -o -name \*.go | grep -v vendor | grep -v git | xargs awk -i inplace -f ~/fixups.awk
# find . -name vendor -prune -o -name .git -prune -o -name \*.go | grep -v vendor | grep -v git | xargs goimports -w
# find . -name vendor -prune -o -name .git -prune -o -name \*.go | grep -v vendor | grep -v git | xargs gofmt -w
# git reset --hard

  {  
 fixes += gsub(/Recieve/,"Receive");
 fixes += gsub(/recieve/,"receive");
 fixes += gsub(/Journalling/,"Journaling");
 fixes += gsub(/detatched/,"detached")
 fixes += gsub(/Acknowledgemets/,"Acknowledgements");
 fixes += gsub(/identitiy/,"identity");
 fixes += gsub(/Controler/,"Controller");
 fixes += gsub(/controler/,"controller");
 fixes += gsub(/ factom /," Factom ");
 fixes += gsub(/ cant /," can't ");
 fixes += gsub(/inital/,"initial");
 fixes += gsub(/caculates/,"calculates");
 fixes += gsub(/werid/,"weird")
 fixes += gsub(/agains /,"against");
 fixes += gsub(/readible /,"readable");
 fixes += gsub(/caluclate/,"calculate");
 fixes += gsub(/signiture/,"signature");

  }


FILENAME!~/atomic/ { fixes += gsub(/sync.Mutex/,"atomic.DebugMutex");}

/import \(/   {inimport=1}
/import \(\)/ {inimport=0} # HANDLE FILES TAHT DON'T IMPORT ANYTING

# add atomic to import, goimport will toss it if it's not needed
/^)$/ && inimport {
	inimport=0; 
	if(FILENAME!~/(atomic)|(rateCalculator)/){
		print "\"github.com/FactomProject/factomd/util/atomic\"";
	}
}


# EntryDBHeightComplete atomic.AtomicUint32
# wsapiNode, listenTo atomic.AtomicInt
# DBFinished, OutputAllowed, NetStateOff, ControlPanelDataRequest atomic.AtomicBool


# do the work for atomic stores
func dostore() { # assumes EOL commnets are stripped
 match($0, / *= */);
 rest = substr($0,RSTART+RLENGTH)
 $0 = substr($0,1,RSTART-1) ".Store(" rest ")"
}


 {comment =""}
# strip eol comment
/[/][/]/ {start = index($0,"//"); comment = substr($0,start); $0 = substr($0,1,start-1)}

/Status +uint8/ 			    {fixes += gsub(/uint8/,"atomic.AtomicUint8")}
/((oneID)|(newAuth)|(.*[iI]dent.*)|(.*Auth.*)|([^a-z][abei])|(auth)|(id))\.Status +=[^=]/ 		    {dostore()}
/((oneID)|(newAuth)|(.*[iI]dent.*)|(.*Auth.*)|([^a-z][abei])|(auth)|(id))\.Status +((==)|[><)]|(<=)|(>=))/ {fixes += gsub(/Status /,"Status.Load() ")}
/[ \t(]((oneID)|(newAuth)|(.*[iI]dent.*)|(.*Auth.*)|([^a-z][abei])|(auth)|(id))\.Status[)]/ {fixes += gsub(/Status[)]/,"Status.Load())")}
/[(][abei].Status[)]/ {fixes += gsub(/Status[)]/,"Status.Load())")}

/status :=.*\.Status/ {fixes += gsub(/Status$/,"Status.Load()")}

 { printf("%s%s\n", $0, comment); 
   comment =""
   if(fixes!=lfixes){ 
	printf("\r%3d", fixes) > "/dev/stderr";
   } 
 }

