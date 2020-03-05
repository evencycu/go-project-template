#!/bin/bash

if [ -z "$1" ]
  then
cat << EOF
Need to specify target folder

Usage:
	./copy.sh ../my-project
EOF
    exit 1
fi

copy_list=(apiserver
command
devops
gpt
local.toml
*.go
Makefile)

for i in "${copy_list[@]}"
do
	cp -r $i $1
done
