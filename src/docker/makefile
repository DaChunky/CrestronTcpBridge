BUILD:=$(shell svn info | grep -E -o 'Rev: [0-9]*' | grep -E -o '[0-9]*')
DO_UPDATE?=1

crebrid_image:
ifeq ($(DO_UPDATE),1)
	svn update ../..
endif
	make -f ../go/makefile docker_linux PATH_TO_BIN=$(PATH_TO_BIN)
	cd ..
	cp -r ../../bin/crebri* ./src
	docker build --tag crebrid:rev$(BUILD) -f Dockerfile.crebrid .
	docker tag crebrid:rev$(BUILD) crebrid:latest
	docker build --tag crebri:rev$(BUILD) -f Dockerfile.crebri .
	docker tag crebri:rev$(BUILD) crebri:latest
