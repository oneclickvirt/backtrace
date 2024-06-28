#!/bin/bash
#From https://github.com/oneclickvirt/backtrace
#2024.06.28

rm -rf /usr/bin/backtrace
os=$(uname -s)
arch=$(uname -m)

case $os in
  Linux)
    case $arch in
      "x86_64" | "x86" | "amd64" | "x64")
        wget -O backtrace https://github.com/oneclickvirt/backtrace/releases/download/output/backtrace-linux-amd64
        ;;
      "i386" | "i686")
        wget -O backtrace https://github.com/oneclickvirt/backtrace/releases/download/output/backtrace-linux-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O backtrace https://github.com/oneclickvirt/backtrace/releases/download/output/backtrace-linux-arm64
        ;;
      *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
    esac
    ;;
  Darwin)
    case $arch in
      "x86_64" | "x86" | "amd64" | "x64")
        wget -O backtrace https://github.com/oneclickvirt/backtrace/releases/download/output/backtrace-darwin-amd64
        ;;
      "i386" | "i686")
        wget -O backtrace https://github.com/oneclickvirt/backtrace/releases/download/output/backtrace-darwin-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O backtrace https://github.com/oneclickvirt/backtrace/releases/download/output/backtrace-darwin-arm64
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
        wget -O backtrace https://github.com/oneclickvirt/backtrace/releases/download/output/backtrace-freebsd-amd64
        ;;
      "i386" | "i686")
        wget -O backtrace https://github.com/oneclickvirt/backtrace/releases/download/output/backtrace-freebsd-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O backtrace https://github.com/oneclickvirt/backtrace/releases/download/output/backtrace-freebsd-arm64
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
        wget -O backtrace https://github.com/oneclickvirt/backtrace/releases/download/output/backtrace-openbsd-amd64
        ;;
      "i386" | "i686")
        wget -O backtrace https://github.com/oneclickvirt/backtrace/releases/download/output/backtrace-openbsd-386
        ;;
      "armv7l" | "armv8" | "armv8l" | "aarch64" | "arm64")
        wget -O backtrace https://github.com/oneclickvirt/backtrace/releases/download/output/backtrace-openbsd-arm64
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
cp backtrace /usr/bin/backtrace
