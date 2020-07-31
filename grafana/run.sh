#!/bin/bash

VERSION="7.1.1"

docker run -ti --rm -p 3000:3000 --name grafana \
  grafana/grafana:$VERSION
