
arg1=$1

if [ ! -n "$arg1" ]; then
	echo "please ipa name"
	exit 0
else
  IPA_NAME=${arg1}
fi 

rm -rf Payload/

unzip  "${IPA_NAME}"   > /dev/null
 
./BundleUtils -r ./Payload/*.app 

rm -rf Payload/
#./genplist.py ${IPA_NAME%.*} "http://192.168.1.236/SampleApp"
#rm -rf info.json

