#!/bin/bash

arch=$(uname -m)
if [ "$arch" = "x86_64" ]; then
  wget -q -O backtrace  https://github.com/oneclickvirt/backtrace/releases/output/backtrace-linux-amd64
else
  wget -q -O backtrace  https://github.com/oneclickvirt/backtrace/releases/output/backtrace-linux-arm64
fi
mv backtrace /usr/bin/
backtrace