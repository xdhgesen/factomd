#   139 P   [d2bc4e] REntry-VM  9: Min:   3          -- Leader[d2bc4e] Entry[45bc9f] ChainID[2cc45d5914] hash[45bc9f]
#   140 P   [d2bc4e] REntry-VM  9: Min:   3          -- Leader[d2bc4e] Entry[e07b33] ChainID[56417b1161] hash[e07b33]
#   141 P   [d2bc4e] REntry-VM  9: Min:   3          -- Leader[d2bc4e] Entry[62133f] ChainID[4b0db1a007] hash[62133f]
#   142 P   [d2bc4e] REntry-VM  9: Min:   3          -- Leader[d2bc4e] Entry[286122] ChainID[56417b1161] hash[286122]
#
#Federated Servers:
#    38bab1455b7bd7e5efd1
#    8888881570f89283f3a5
#    88888832ceba14177e9c
#    888888435ae7a3c1494f
#    8888887f03e531e68922
#    8888888da6ed14ec63e6
#    8888889b844de72a15f8
#    888888aeaac80d825ac9
#    888888c0bc99166c1419
#    888888d2bc4ed232378c
#Audit Servers:
#    88888867ee42e8b22134 online
#    8888887020255b631764 online
#    888888f0b7e308974afc online
#
#FNode03 #VMs 10 Complete true DBHeight 5 DBSig false EOM true p-dbstate = signed Entries Complete 5
#  VM 0  vMin 10 vHeight 759 len(List)759 Syncing true Synced true EOMProcessed 9 DBSigProcessed 0
#   0 P   [9b844d]  DBSig-VM  0:          DBHt:    5 -- Signer=889b844d PrevDBKeyMR[:3]=a84f8e hash=3c9b83
# 158     [d2bc4e]    EOM-     DBh/VMh/h 6/9/-- minute 3 FF  0 --Leader[d2bc4e] hash[e77a9f] 
#  172     <nil>
    
/^FNode/ { node=$1 " "}
/.*VM.*vMin/ { VM = $1 $2 " "}
/.*EOM.*DBh|[<]nil[>]/   { 
           if (node != lastnode) {
              lastnode = node
              leoms[eomcnt++] = "\n ================== " node " ======================="
           }
           if( eoms[node VM $0] == 0) {
               eoms[node VM $0] = node VM $0
	      leoms[eomcnt++] = node VM $0	
           }
         }
END{ print "EOMs:"
     for(i=0;i<eomcnt;i++){ print leoms[i] }
}
