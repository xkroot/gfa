# gfa (高性能、低延迟的流量采集与拆包工具，可实现网络流量可视化、安全实时事件分析)

# 编译

```bash
make build
```

# 使用

```bash
yum install tcpdump
./bin/gfa -conf core.config.toml
```

# 拆包(可根据需求自行定义其他字段)

```json
{
  "tcp_flags": "tcp三次握手标志位(SYN/ACK/FIN...)",
  "source_port": "源端口",
  "pcap": "pcap源文件",
  "location": "流量来源位置",
  "request_http_header": "http请求头(json)",
  "frame_length": "报文大小(bit)",
  "request_http_body": "http请求体",
  "dest_ip": "目标IP",
  "request_http_path": "http请求路径",
  "source_ip": "源IP",
  "timestamp": "时间戳",
  "service": "服务(mysql/ssh...)",
  "protocol": "应用层协议(TCP/UDP..)",
  "request_http_url": "http请求地址",
  "request_http_method": "http请求方式",
  "dest_port": "目标端口"
}
```

# 设计

![img](doc/traffic.jpg)