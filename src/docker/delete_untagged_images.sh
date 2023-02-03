docker image rm $(docker images | grep '<none>' | grep -E -o '[0-9a-z]{12}') 
