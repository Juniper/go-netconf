#!/bin/bash

cd /opt/yang/example
go build --ldflags '-extldflags "-static"'
cd /opt/yang/xpath_example
go build --ldflags '-extldflags "-static"'
