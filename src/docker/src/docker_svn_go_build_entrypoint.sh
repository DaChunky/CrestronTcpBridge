svn checkout svn://$SVN_SERVER_IP/$PROJECT/$BRANCH 
pathArr=(${BRANCH///// })
pathLen=${#pathArr[@]}
if [ $pathLen -eq 1 ]
then
	cd ./$BRANCH/$PATH_TO_SRC 
else
	cd ./${pathArr[1]}/$PATH_TO_SRC
fi
echo "build app from $PATH_TO_MAIN to $PATH_TO_BIN"
$CROSS_COMPILE_PREFIX go build -o $PATH_TO_BIN ./$PATH_TO_MAIN
echo "successfully build" 
chmod 777 $PATH_TO_BIN 
echo "make it an executable"
chown -v $(id -u):$(id -g) $PATH_TO_BIN
echo "change ownership"

