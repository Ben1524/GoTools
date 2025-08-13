### 注意事项
1. 容器的localhost地址与宿主机的localhost地址不同，容器内的localhost地址是容器本身的地址。
2. 如果需要访问宿主机的服务，可以使用宿主机的IP地址或配置Docker网络。
3. 尽量使用容器名称来访问服务，而不是使用IP地址，这样可以避免IP地址变化带来的问题。


### 客户端通过提供`/metrics`路径来暴露指标给Prometheus抓取
- 例如：`http://<host>:<port>/metrics`
- Prometheus会定期抓取这个路径上的指标数据。
- 确保客户端应用程序正确实现了指标暴露的逻辑。

### Prometheus配置文件示例
```yaml
global:
  scrape_interval: 15s # 抓取间隔时间
  evaluation_interval: 15s # 规则评估间隔时间
scrape_configs:
    - job_name: 'my_service' # 服务名称
        static_configs:
        - targets: ['<host>:<port>'] # 服务地址和端口
        labels:
            env: 'production' # 可选标签
    - job_name: 'another_service'
        static_configs:
        - targets: ['<host>:<port>']
        labels:
            env: 'staging'
```

### Prometheus
1. 默认使用pull方式抓取指标,由Prometheus主动去抓取客户端暴露的指标。
2. 可以主动提交指标到Pushgateway,由Pushgateway保存指标数据,Prometheus定期抓取Pushgateway的指标数据。