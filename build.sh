#!/bin/bash

platforms=(
  "windows/amd64"
  "darwin/arm64"
)

for platform in "${platforms[@]}"
do
  GOOS=${platform%/*}
  GOARCH=${platform#*/}
  output_name=Neon
  if [ $GOOS = "windows" ]; then
    output_name+='_win.exe'
  fi
  
  echo "Building for $GOOS/$GOARCH..."
  env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name main.go
  if [ $? -ne 0 ]; then
    echo "An error has occurred! Aborting the script execution..."
    exit 1
  fi
done
