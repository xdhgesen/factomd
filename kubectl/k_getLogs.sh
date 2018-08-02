#/bin/sh
pod="$1"
shift
kubectl exec -it $pod -c factomd  -- sh -c "tar cvzf logs.tgz *.txt" 
mkdir -p $pod
kubectl cp  $pod:logs.tgz $pod
