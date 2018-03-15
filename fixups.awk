  {  
 fixes += gsub(/Recieve/,"Receive");
 fixes += gsub(/recieve/,"receive");
 fixes += gsub(/Journalling/,"Journaling");
 fixes += gsub(/detatched/,"detached")
 fixes += gsub(/sync.Mutex/,"atomic.DebugMutex");
 fixes += gsub(/Acknowledgemets/,"Acknowledgements");
 fixes += gsub(/identitiy/,"identity");
 fixes += gsub(/Controler/,"Controller");
 fixes += gsub(/controler/,"controller");
 fixes += gsub(/ factom /," Factom ");
 fixes += gsub(/cant/,"can't");
 fixes += gsub(/initial/,"inital");
 fixes += gsub(/caculates/,"calculates");
 fixes += gsub(/werid/,"weird")
 fixes += gsub(/agains /,"against");
  }

# EntryDBHeightComplete atomic.AtomicUint32
# wsapiNode, listenTo atomic.AtomicInt
# DBFinished, OutputAllowed, NetStateOff, ControlPanelDataRequest atomic.AtomicBool



/Status +uint8/ 			{fixes += gsub(/uint8/,"atomic.Uint8")}
/((auth)|(id))\.Status +=[^=]/ 		{value = $3; fixes += sub(/ += .*/,".Store("value")")}



	{print; if(fixes>0) print fixes > "/dev/stderr";}

