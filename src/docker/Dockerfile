FROM golang:buster

RUN apt-get install subversion
ENV SVN_SERVER_IP=
ENV PROJECT=
ENV BRANCH=trunk
ENV PATH_TO_BIN=/home/build/app 
ENV PATH_TO_SRC=
ENV PATH_TO_MAIN=
ENV CROSS_COMPILE_PREFIX=
ENV USER=docker
ENV GROUP=docker
ADD ./src/docker_svn_go_build_entrypoint.sh /home/entrypoint.sh
ENTRYPOINT /bin/bash /home/entrypoint.sh
