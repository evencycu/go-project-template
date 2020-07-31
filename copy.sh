#!/bin/bash

if [ "$#" -ne 2 ]; then
cat << EOF
Need to specify target folder and project alias name

Usage:
	./copy.sh my-project project_alias

For example, if your project name is my-first-go-project, then your project alias could be mfgp
then

	./copy.sh my-first-go-project mfgp

    then you can find your project directory in ../my-first-go-project
EOF
    exit 1
fi

if [ -z "$1" ]
  then
cat << EOF
Need to specify target folder and project alias name

Usage:
	./copy.sh my-project project_alias

For example, if your project name is my-first-go-project, then your project alias could be mfgp
then

	./copy.sh my-first-go-project mfgp

    then you can find your project directory in ../my-first-go-project
EOF
    exit 1
fi

if [ -f ../$1 ]
then
    echo "destination is not a directory"
    exit 1
fi

if [ ! -d ../$1 ]
then
    mkdir -p ../$1
fi

copy_list=(apiserver
command
devops
gpt
local.toml
*.go
go.mod
go.sum
.gitignore
sonar-project.properties
Makefile)

for i in "${copy_list[@]}"
do
	cp -r $i ../$1
done

echo "# $1 project readme markup file

# markdown language guide
## https://guides.github.com/features/mastering-markdown/
## https://guides.github.com/pdfs/markdown-cheatsheet-online.pdf

Please make your documentation in this readme file
" > ../$1/README.md

cd ../$1
find . -type f -exec sed -i'' -e "s/go-project-template/$1/g" {} +
mv gpt $2
find . -type f -exec sed -i'' -e "s/gpt/$2/g" {} +

echo "Copy completed"
