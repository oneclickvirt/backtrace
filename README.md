# backtrace

[![Hits](https://hits.spiritlhl.net/backtrace.svg?action=hit&title=Hits&title_bg=%23555555&count_bg=%230eecf8&edge_flat=false)](https://hits.spiritlhl.net)

[![Build and Release](https://github.com/oneclickvirt/backtrace/actions/workflows/main.yaml/badge.svg)](https://github.com/oneclickvirt/backtrace/actions/workflows/main.yaml)

三网回程路由线路测试

路由的线路判断最终还是得人工判断的才准确，本项目测试结果仅供参考

## 功能

- [x] 检测回程显示IPV4/IPV6地址时的线路(使用1500字节的包)，不显示IP地址时显示ASN检测不到
- [x] 支持对```9929```、```4837```和```163```线路的判断
- [x] 支持对```CTGNET```、```CN2GIA```和```CN2GT```线路的判断
- [x] 支持对```CMIN2```和```CMI```线路的判断
- [x] 支持对整个回程路由进行线路分析，一个目标IP可能会分析出多种线路
- [x] 支持对主流接入点的线路检测，方便分析国际互联能力
- [x] 增加对全平台的编译支持，原版[backtrace](https://github.com/zhanghanyun/backtrace)仅支持linux平台的amd64和arm64架构
- [x] 兼容额外的ICMP地址获取，若当前目标IP无法查询路由尝试额外的IP地址

## 使用

下载、安装、更新

```shell
curl https://raw.githubusercontent.com/oneclickvirt/backtrace/main/backtrace_install.sh -sSf | bash
```

或

```
curl https://cdn.spiritlhl.net/https://raw.githubusercontent.com/oneclickvirt/backtrace/main/backtrace_install.sh -sSf | bash
```

使用

```
backtrace
```

或

```
./backtrace
```

进行测试

无环境依赖，理论上适配所有系统和主流架构，更多架构请查看 https://github.com/oneclickvirt/backtrace/releases/tag/output

```
Usage: backtrace [options]
  -h    Show help information
  -ipv6
        Enable ipv6 testing
  -log
        Enable logging
  -s    Disabe show ip info (default true)
  -v    Show version
```

## 卸载

```
rm -rf /root/backtrace
rm -rf /usr/bin/backtrace
```

## 在Golang中使用

```
go get github.com/oneclickvirt/backtrace@v0.0.6-20250805091811
```

## 概览图

![图片](https://github.com/oneclickvirt/backtrace/assets/103393591/4688f99f-0f02-486f-8ffc-78d30f2c2f95)

![图片](https://github.com/oneclickvirt/backtrace/assets/103393591/2812a47d-4e6b-4091-9bb9-596af6c3c8bc)

![图片](https://github.com/oneclickvirt/backtrace/assets/103393591/2e5cc625-e0da-41ff-85ff-9d21c01114a3)

## Thanks

部分代码基于 https://github.com/zhanghanyun/backtrace 的重构和优化，与原版存在很大不同

IPV4/IPV6可ICMP进行ping测试的 https://github.com/spiritLHLS/icmp_targets 收集仓库
