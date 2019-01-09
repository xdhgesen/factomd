#
# Remove any fastboot files from your M2 directory.  
# Run this script in your M2 directory
# Run factomd against the main net
# All the fast boot files will be preserved in m2/fastboots
#
rm -r ./fastboots
mkdir ./fastboots
FILE=FastBoot_MAIN_v10.db

while [ 1 ]; do
	if [ -f $FILE ]; then
 		((cnt+=1000))
		padded=$(echo $cnt | awk '{printf "%06d",$1}')			
		DEST=$FILE.$padded
		mv $FILE ./fastboots/$DEST
		sleep .1
		echo ----------------------------------------------------------
		ls fastboots
        else
                sleep .1
        fi
done
