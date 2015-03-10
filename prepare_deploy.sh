#!/bin/bash
#
# 
# Description : Prepare deploy vorimport package. 
# Author : vassilux
#

set -e


VERSION=$(cat VERSION)

cp main.go main.go.bkp

sed -i "/VERSION = \"X.X.X\"/c\VERSION = \"${VERSION}\"" main.go

make clean

make fmt

make

if [ ! -f ./bin/vorimport ]; then
	echo "Can not find compiled  project file ./bin/vorimport."
	echo "Please cheque make output."
    exit 1
fi

mv main.go.bkp main.go 

DEPLOY_DIR="vorimport_${VERSION}"
DEPLOY_FILE_NAME="vorimport_${VERSION}.tar.gz"

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
pandoc -o "$DEPLOY_DIR/docs/ReleaseNotes.html" ./docs/ReleaseNotes.md
cp "$DEPLOY_DIR/docs/INSTALL.html" .
cp "$DEPLOY_DIR/docs/ReleaseNotes.html" .

tar cvzf "${DEPLOY_FILE_NAME}" "${DEPLOY_DIR}"

if [ ! -f "$DEPLOY_FILE_NAME" ]; then
    echo "Deploy build failed."
    exit 1
fi

if [ ! -d releases ]; then
        mkdir releases
fi

mv ${DEPLOY_FILE_NAME} ./releases
mv INSTALL.* ./releases
mv ReleaseNotes.* ./releases

rm -rf "$DEPLOY_DIR"

echo "Deploy build complete."
echo "Live well."
