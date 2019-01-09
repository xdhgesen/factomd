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
