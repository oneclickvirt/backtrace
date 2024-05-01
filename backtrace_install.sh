#!/bin/bash
#From https://github.com/oneclickvirt/backtrace
#2024.05.01

os=$(uname -s)
arch=$(uname -m)

case $os in
  Linux)
    case $arch in
      x86_64)
        wget -q -O backtrace https://github.com/oneclickvirt/backtrace/releases/output/backtrace-linux-amd64
        ;;
      arm*)
        wget -q -O backtrace https://github.com/oneclickvirt/backtrace/releases/output/backtrace-linux-arm64
        ;;
      *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  Darwin)
    case $arch in
      x86_64)
        wget -q -O backtrace https://github.com/oneclickvirt/backtrace/releases/output/backtrace-darwin-amd64
        ;;
      arm64)
        wget -q -O backtrace https://github.com/oneclickvirt/backtrace/releases/output/backtrace-darwin-arm64
        ;;
      *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  FreeBSD)
    case $arch in
      amd64)
        wget -q -O backtrace https://github.com/oneclickvirt/backtrace/releases/output/backtrace-freebsd-amd64
        ;;
      i386)
        wget -q -O backtrace https://github.com/oneclickvirt/backtrace/releases/output/backtrace-freebsd-386
        ;;
      arm*)
        wget -q -O backtrace https://github.com/oneclickvirt/backtrace/releases/output/backtrace-freebsd-arm64
        ;;
      *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  OpenBSD)
    case $arch in
      amd64)
        wget -q -O backtrace https://github.com/oneclickvirt/backtrace/releases/output/backtrace-openbsd-amd64
        ;;
      i386)
        wget -q -O backtrace https://github.com/oneclickvirt/backtrace/releases/output/backtrace-openbsd-386
        ;;
      arm*)
        wget -q -O backtrace https://github.com/oneclickvirt/backtrace/releases/output/backtrace-openbsd-arm64
        ;;
      *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  *)
    echo "Unsupported operating system: $os"
    exit 1
    ;;
esac

chmod 777 backtrace
if [ -f /usr/bin/ ]; then
  mv backtrace /usr/bin/
  backtrace
else
  ./backtrace
fi
