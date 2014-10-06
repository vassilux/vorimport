#!/bin/bash
#
# 
# Description : Prepare deploy revor package. 
# Author : vassilux
# Last modified : 2014-06-03 14:53:54  
#

set -e

VER_MAJOR="1"
VER_MINOR="0"
VER_PATCH="0"

DEPLOY_DIR="vorimport_${VER_MAJOR}.${VER_MINOR}.${VER_PATCH}"
DEPLOY_FILE_NAME="vorimport_${VER_MAJOR}.${VER_MINOR}.${VER_PATCH}.tar.gz"

if [ -d "$DEPLOY_DIR" ]; then
    rm -rf  "$DEPLOY_DIR"
fi
#
#
mkdir "$DEPLOY_DIR"

cp -aR ./bin/* "$DEPLOY_DIR"
cp -aR ./samples/* "$DEPLOY_DIR"
#
mkdir "$DEPLOY_DIR/docs"
pandoc -o "$DEPLOY_DIR/docs/INSTALL.html" ./docs/INSTALL.md
cp "$DEPLOY_DIR/docs/INSTALL.html" .

tar cvzf "${DEPLOY_FILE_NAME}" "${DEPLOY_DIR}"

if [ ! -f "$DEPLOY_FILE_NAME" ]; then
    echo "Deploy build failed."
    exit 1
fi

rm -rf "$DEPLOY_DIR"

echo "Deploy build complete."
