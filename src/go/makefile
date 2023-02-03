build:
	go build -o ../../bin/crebrid ./cmd/crebrid/main.go
	go build -o ../../bin/crebri ./cmd/crebri/main.go

build_pi:
	env GOOS=linux GOARCH=arm GOARM=7 go build -o ../../bin/pi/crebrid ./cmd/crebrid/main.go
	env GOOS=linux GOARCH=arm GOARM=7 go build -o ../../bin/pi/crebri ./cmd/crebri/main.go

docker_linux:
	mkdir -p $(PATH_TO_BIN)
	chown $(U):$(G) $(PATH_TO_BIN)
	docker run --rm -e SVN_SERVER_IP=192.168.178.56 -e PROJECT=CrestronTcpBridge -e PATH_TO_SRC=src/go/cmd/crebrid -e PATH_TO_MAIN=main.go -v $(PATH_TO_BIN):/home/build -e USER=$(U) -e GROUP=$(G) go_build:latest
	mv $(PATH_TO_BIN)/app $(PATH_TO_BIN)/crebrid
	docker run --rm -e SVN_SERVER_IP=192.168.178.56 -e PROJECT=CrestronTcpBridge -e PATH_TO_SRC=src/go/cmd/crebri -e PATH_TO_MAIN=main.go -v $(PATH_TO_BIN):/home/build -e USER=$(U) -e GROUP=$(G) go_build:latest
	mv $(PATH_TO_BIN)/app $(PATH_TO_BIN)/crebri

docker_pi:
	mkdir -p $(PATH_TO_BIN)
	mkdir -p $(PATH_TO_BIN)/pi
	chown -vR $(U):$(G) $(PATH_TO_BIN)
	docker run --rm -e CROSS_COMPILE_PREFIX="env GOOS=linux GOARCH=arm GOARM=7" -e SVN_SERVER_IP=192.168.178.56 -e PROJECT=CrestronTcpBridge -e PATH_TO_SRC=src/go/cmd/crebrid -e PATH_TO_MAIN=main.go -v $(PATH_TO_BIN):/home/build -e USER=$(U) -e GROUP=$(G) go_build:latest
	mv $(PATH_TO_BIN)/app $(PATH_TO_BIN)/pi/crebrid
	docker run --rm  -e CROSS_COMPILE_PREFIX="env GOOS=linux GOARCH=arm GOARM=7" -e SVN_SERVER_IP=192.168.178.56 -e PROJECT=CrestronTcpBridge -e PATH_TO_SRC=src/go/cmd/crebri -e PATH_TO_MAIN=main.go -v $(PATH_TO_BIN):/home/build -e USER=$(U) -e GROUP=$(G) go_build:latest
	mv $(PATH_TO_BIN)/app $(PATH_TO_BIN)/pi/crebri
