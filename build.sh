#! /bin/bash

APP_DIR=/go/src/github.com/${GITHUB_REPOSITORY}/
mkdir -p $(APP_DIR) && cp -r ./ $(APP_DIR) && cd $(APP_DIR)
make
