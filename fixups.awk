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



# Use debugMutexes instead of syncMutex
FILENAME!~/(atomic)|(rateCalculator)/ { fixes += gsub(/sync.Mutex/,"atomic.DebugMutex");}
/import \(/   {inimport=1}
/import \(\)/ {inimport=0} # HANDLE FILES TAHT DON'T IMPORT ANYTING
$0~"github.com/FactomProject/factomd/util/atomic" {inimport=0} # don't duplicate the atomic import
FILENAME~/(atomic)|(rateCalculator)/ && inimport {inimport=0} # don't import atomic in these files that use sync on purpose
/^)$/ && inimport {inimport=0; 	print "\"github.com/FactomProject/factomd/util/atomic\"";}# add atomic to import, goimport will toss it if it's not needed
# End Use debugMutexes instead of syncMutex

# EntryDBHeightComplete atomic.AtomicUint32
# wsapiNode, ListenTo atomic.AtomicInt
# DBFinished, OutputAllowed, NetStateOff, ControlPanelDataRequest atomic.AtomicBool


# do the work for atomic stores
func dostore(name) { # assumes EOL commnets are stripped
 sub("*" name, name);
 match($0, / *= */);
 rest = substr($0,RSTART+RLENGTH)
 $0 = substr($0,1,RSTART-1) ".Store(" rest ")"
 fixes++
}


 {comment =""}
# strip eol comment
/[/][/]/ {start = index($0,"//"); comment = substr($0,start); $0 = substr($0,1,start-1)}

# atomic status access
/Status +uint8/ 			    {fixes += gsub(/uint8/,"atomic.AtomicUint8")}
/((oneID)|(newAuth)|(.*[iI]dent.*)|(.*Auth.*)|([^a-z][abei])|(auth)|(id))\.Status +=[^=]/ 		    {dostore()}
/((oneID)|(newAuth)|(.*[iI]dent.*)|(.*Auth.*)|([^a-z][abei])|(auth)|(id))\.Status +((==)|[><)]|(<=)|(>=))/ {fixes += gsub(/Status /,"Status.Load() ")}
/[ \t(]((oneID)|(newAuth)|(.*[iI]dent.*)|(.*Auth.*)|([^a-z][abei])|(auth)|(id))\.Status[)]/ {fixes += gsub(/Status[)]/,"Status.Load())")}
/[(][abei].Status[)]/ {fixes += gsub(/Status[)]/,"Status.Load())")}
/status :=.*\.Status/ {fixes += gsub(/Status$/,"Status.Load()")}
# end atomic status access


# EntryDBHeightComplete atomic.AtomicUint32

/!.*\.DBFinished/     	{fixes += gsub(/DBFinished/,"DBFinished.Load() ")}
/DBFinished +bool/ 	{fixes += gsub(/bool/,"atomic.AtomicBool")}
/DBFinished +=[^=]/ 	{dostore("DBFinished")}
/DBFinished == true/    {fixes += gsub(/DBFinished == true/,"DBFinished.Load() ")}
#
/OutputAllowed +bool/ 	{fixes += gsub(/bool/,"atomic.AtomicBool")}
/OutputAllowed +=[^=]/ 	{dostore("OutputAllowed")}
/[^ \t] *OutputAllowed([^.]|$)/ {fixes += gsub(/OutputAllowed/,"OutputAllowed.Load() ")}

#
/[^t]NetStateOff +bool/ 	{fixes += gsub(/bool/,"atomic.AtomicBool")}
/[^t]NetStateOff +=[^=]/ 	{dostore()}
/[^t\"]NetStateOff([^.]|$)/ {fixes += gsub(/\.NetStateOff/,".NetStateOff.Load() ")}

#
/[^t]ControlPanelDataRequest +bool/ 	{fixes += gsub(/bool/,"atomic.AtomicBool")}
/[^t]ControlPanelDataRequest +=[^=]/ 	{dostore("ControlPanelDataRequest")}
/[^t\"]ControlPanelDataRequest(( [{])|( *[^. a])|$)/ {fixes += gsub(/\.ControlPanelDataRequest/,".ControlPanelDataRequest.Load() ")}
#
/[^t]wsapiNode +int/ 	{fixes += gsub(/int/,"atomic.AtomicInt")}
/[^t]wsapiNode +=[^=]/ 	{dostore("wsapiNode")}
/\[wsapiNode/ {fixes += gsub(/wsapiNode/,"wsapiNode.Load() ")}
/wsapiNode \*int/ {fixes += gsub(/wsapiNode \*int/,"wsapiNode *atomic.AtomicInt")}
/\*wsapiNode/ {fixes += gsub(/\*wsapiNode/,"wsapiNode.Load()")}

#
/\[ListenTo/ {fixes += gsub(/ListenTo/,"ListenTo.Load()")}
/[^t]ListenTo +=[^=]/ 	{dostore("ListenTo")}
/[^t]ListenTo +int/ 	{fixes += gsub(/int/,"atomic.AtomicInt")}

/[<>] ListenTo/ {fixes += gsub(/ListenTo/,"ListenTo.Load()")}
/ListenTo [<>]/ {fixes += gsub(/ListenTo/,"ListenTo.Load()")}
/[^&]ListenTo[,)]/ {fixes += gsub(/ListenTo/,"ListenTo.Load()")}
/ListenTo\+\+/ {fixes += gsub(/ListenTo\+\+/,"ListenTo.Store(ListenTo.Load()+1)")}

#exclude the ISR routine
/func.*((InstantaneousStatusReport)|(SimControl))/ {inISR=1;}
/^}/ {inISR=0}

/listenTo \*int/        {if(inISR==0){fixes += gsub(/listenTo \*int/,"listenTo *atomic.AtomicInt")}}
/[^.]listenTo +=[^=]/ 	{if(inISR==0){dostore("listenTo")}}
/\*listenTo[^P.]/       {if(inISR==0){fixes += gsub(/\*listenTo/,"listenTo.Load()")}}
/\[listenTo[^.]/        {if(inISR==0){fixes += gsub(/listenTo/,"listenTo.Load()")}}


#
 { printf("%s%s\n", $0, comment); 
   comment =""
   if(fixes!=lfixes){ 
	printf("\r%3d", fixes) > "/dev/stderr";
   } 
 }

