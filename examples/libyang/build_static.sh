#!/bin/bash

# build docker
cd ./docker
docker build -t sysrepo/sysrepo-netopeer2:golang -f Dockerfile .
cd -

# compile golang code
pwd_dir=$(pwd)
docker run -i -t -v $pwd_dir:/opt/yang --rm sysrepo/sysrepo-netopeer2:golang bash -c /opt/yang/docker/static_entry_point.sh
