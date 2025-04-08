#!/bin/bash

echo "Installing dependencies..."

echo "Installing kopia..."

RELEASE=$(wget -q https://github.com/kopia/kopia/releases/latest -O - | grep "title>Release" | cut -d " " -f 4 | sed 's/^v//')

wget -P ./data/dependencies/ -q https://github.com/kopia/kopia/releases/download/v$RELEASE/kopia-$RELEASE-linux-x64.tar.gz

tar -xzf ./data/dependencies/kopia-$RELEASE-linux-x64.tar.gz -C ./data/dependencies/

rm -rf ./data/dependencies/kopia-$RELEASE-linux-x64*

mv ./data/dependencies/kopia /usr/bin

echo "Kopia installed successfully."