#!/bin/bash

arch=$(uname -m)
if [ "$arch" = "x86_64" ]; then
  wget -q -O backtrace.tar.gz  https://github.com/oneclickvirt/backtrace/releases/output/backtrace-linux-amd64
  mv backtrace-linux-amd64 backtrace
else
  wget -q -O backtrace.tar.gz  https://github.com/oneclickvirt/backtrace/releases/output/backtrace-linux-arm64
  mv backtrace-linux-arm64 backtrace
fi
mv backtrace /usr/bin/
backtrace