#!/usr/bin/env sh

docker run -it --rm -u "$(id -u):$(id -g)" -v "$(pwd):/src" --network host --workdir /src/webui node:lts /bin/bash
