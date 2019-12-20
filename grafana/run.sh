#!/bin/bash

VERSION="6.4.4"

docker run -ti --rm -p 3000:3000 --name grafana \
  -v $PWD/etc/grafana:/etc/grafana \
  -v $PWD/var:/var \
  grafana/grafana:$VERSION
