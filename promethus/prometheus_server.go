//package promethus
//
//import (
//	"fmt"
//	"github.com/prometheus/client_golang/prometheus"
//	"github.com/prometheus/client_golang/prometheus/promhttp"
//	"math/rand"
//	"net/http"
//	"runtime"
//	"time"
//)
//
//// 定义指标（常见类型示例）
//// 计数器：统计请求总数（只增不减）
//var httpReqCnt = prometheus.NewCounterVec(
//	prometheus.CounterOpts{
//		Name: "http_requests_cnt",
//		Help: "Number of HTTP requests",
//	},
//	[]string{"method", "path", "endpoint"}) // 标签：请求方法、路径、端点
//
//var orderNum = prometheus.NewGaugeVec(
//	prometheus.GaugeOpts{
//		Name: "order_num",
//		Help: "Number of orders",
//	}, []string{"status", "endpoint"}) // 标签：订单状态、端点
//
//var goruntineNum = prometheus.NewGauge(
//	prometheus.GaugeOpts{
//		Name: "goruntine_num",
//		Help: "Number of goruntines",
//	}) // 无标签，表示当前运行的 goroutine 数量
//
//// 直方图：统计请求延迟分布
//var httpReqLatency = prometheus.NewHistogramVec(
//	prometheus.HistogramOpts{
//		Name: "http_request_latency_seconds",
//		Help: "HTTP request latency distributions",
//	}, []string{"method", "path", "endpoint"}) // 标签：请求方法、路径、端点
//// 摘要：统计请求延迟的分位数
//var httpReqLatencySummary = prometheus.NewSummaryVec(
//	prometheus.SummaryOpts{
//		Name: "http_request_latency_summary_seconds",
//		Help: "HTTP request latency summary",
//	},
//	[]string{"method", "path", "endpoint"}) // 标签：请求方法、路径、端点
//
//func RecordHTTPRequest(method, path, endpoint string, duration time.Duration) {
//	// 记录 HTTP 请求计数
//	httpReqCnt.WithLabelValues(method, path, endpoint).Inc()
//	// 记录 HTTP 请求延迟
//	httpReqLatency.WithLabelValues(method, path, endpoint).Observe(duration.Seconds())
//	// 记录 HTTP 请求延迟摘要
//	httpReqLatencySummary.WithLabelValues(method, path, endpoint).Observe(duration.Seconds())
//}
//
//func RecordOrderStatus(status, endpoint string) {
//	// 设置订单状态
//	orderNum.WithLabelValues(status, endpoint).Inc()
//}
//
//func RecordOrderNum(status, endpoint string, num float64) {
//	// 设置订单数量
//	orderNum.WithLabelValues(status, endpoint).Set(num)
//}
//
//func StartPrometheusServer(addr string) error {
//	// 启动 Prometheus HTTP 服务器
//	http.Handle("/metrics", promhttp.Handler())
//	go func() {
//		if err := http.ListenAndServe(addr, nil); err != nil {
//			fmt.Printf("Failed to start Prometheus server: %v\n", err)
//		}
//	}()
//
//	// 定时更新 goroutine 数量
//	go func() {
//		for {
//			goruntineNum.Set(float64(runtime.NumGoroutine()))
//			time.Sleep(10 * time.Second) // 每10秒更新一次
//		}
//	}()
//
//	return nil
//}
//
//// 初始化指标
//func init() {
//	// 注册指标到默认注册表
//	prometheus.MustRegister(httpReqCnt)
//	prometheus.MustRegister(orderNum)
//	prometheus.MustRegister(goruntineNum)
//	prometheus.MustRegister(httpReqLatency)
//	prometheus.MustRegister(httpReqLatencySummary)
//}
//func main() {
//	// 启动 Prometheus 服务器
//	if err := StartPrometheusServer(":8080"); err != nil {
//		fmt.Printf("Error starting Prometheus server: %v\n", err)
//		return
//	}
//
//	// 模拟 HTTP 请求和订单状态更新
//	for {
//		method := "GET"
//		i := rand.Intn(100) // 模拟随机资源 ID
//		// 模拟请求路径和端点
//		path := fmt.Sprintf("/api/v1/resource/%d", i)
//		endpoint := "example_endpoint" + fmt.Sprintf("%d", i%5) // 模拟不同的端点
//		// 模拟请求方法和延迟
//		if rand.Intn(2) == 0 {
//			method = "POST"
//			path = fmt.Sprintf("/api/v1/resource/%d/order", i) // 模拟订单创建路径
//		} else {
//			method = "GET"
//			path = fmt.Sprintf("/api/v1/resource/%d/status", i) // 模拟订单状态查询路径
//		}
//		duration := time.Duration(rand.Intn(1000)) * time.Millisecond
//
//		RecordHTTPRequest(method, path, endpoint, duration)
//		RecordOrderStatus("success", endpoint)
//		time.Sleep(500 * time.Millisecond) // 模拟请求间隔
//	}
//
//	select {} // 阻塞主 goroutine，保持程序运行
//}

package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

// 定义指标（常见类型示例）
// 计数器：统计请求总数（只增不减）
var httpReqCnt = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_cnt",
		Help: "Number of HTTP requests",
	},
	[]string{"method", "path", "endpoint"}) // 标签：请求方法、路径、端点

var orderNum = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "order_num",
		Help: "Number of orders",
	}, []string{"status", "endpoint"}) // 标签：订单状态、端点

var goruntineNum = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "goruntine_num",
		Help: "Number of goroutines",
	}) // 无标签，表示当前运行的 goroutine 数量

// 直方图：统计请求延迟分布
var httpReqLatency = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "http_request_latency_seconds",
		Help: "HTTP request latency distributions",
	}, []string{"method", "path", "endpoint"}) // 标签：请求方法、路径、端点

// 摘要：统计请求延迟的分位数
var httpReqLatencySummary = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name: "http_request_latency_summary_seconds",
		Help: "HTTP request latency summary",
	},
	[]string{"method", "path", "endpoint"}) // 标签：请求方法、路径、端点

// 初始化指标 - 不注册到默认注册表，而是通过PushGateway推送
func init() {
	// 这里不再使用MustRegister，而是将指标添加到推送器
}

func RecordHTTPRequest(method, path, endpoint string, duration time.Duration) {
	// 记录 HTTP 请求计数
	httpReqCnt.WithLabelValues(method, path, endpoint).Inc()
	// 记录 HTTP 请求延迟
	httpReqLatency.WithLabelValues(method, path, endpoint).Observe(duration.Seconds())
	// 记录 HTTP 请求延迟摘要
	httpReqLatencySummary.WithLabelValues(method, path, endpoint).Observe(duration.Seconds())
}

func RecordOrderStatus(status, endpoint string) {
	// 增加订单状态计数
	orderNum.WithLabelValues(status, endpoint).Inc()
}

func RecordOrderNum(status, endpoint string, num float64) {
	// 设置订单数量
	orderNum.WithLabelValues(status, endpoint).Set(num)
}

// PushMetrics 推送指标到PushGateway
func PushMetrics(pushGatewayURL, jobName string, labels map[string]string) error {
	// 创建推送器
	pusher := push.New(pushGatewayURL, jobName).
		Collector(httpReqCnt).
		Collector(orderNum).
		Collector(goruntineNum).
		Collector(httpReqLatency).
		Collector(httpReqLatencySummary)

	// 添加实例标签（可选）
	for k, v := range labels {
		pusher = pusher.Grouping(k, v)
	}

	// 执行推送
	if err := pusher.Push(); err != nil {
		return fmt.Errorf("推送指标失败: %v", err)
	}
	return nil
}

// StartMetricsPusher 启动定时推送指标的 goroutine
func StartMetricsPusher(pushGatewayURL, jobName string, interval time.Duration, labels map[string]string) {
	// 定时更新 goroutine 数量并推送指标
	go func() {
		for {
			// 更新goroutine数量
			goruntineNum.Set(float64(runtime.NumGoroutine()))

			// 推送指标
			if err := PushMetrics(pushGatewayURL, jobName, labels); err != nil {
				fmt.Printf("推送指标错误: %v\n", err)
			} else {
				fmt.Printf("成功推送指标到 %s\n", pushGatewayURL)
			}

			time.Sleep(interval)
		}
	}()
}

func main() {
	// 配置PushGateway地址和任务名称
	pushGatewayURL := "http://localhost:9990" // PushGateway默认地址
	jobName := "pushgateway"
	pushInterval := 10 * time.Second // 每10秒推送一次

	// 实例标签（用于区分不同实例）
	labels := map[string]string{
		"instance": "service-1",
		"env":      "test",
	}

	// 启动指标推送器
	StartMetricsPusher(pushGatewayURL, jobName, pushInterval, labels)

	// 模拟 HTTP 请求和订单状态更新
	for {
		method := "GET"
		i := rand.Intn(100) // 模拟随机资源 ID
		// 模拟请求路径和端点
		path := fmt.Sprintf("/api/v1/resource/%d", i)
		endpoint := "example_endpoint" + fmt.Sprintf("%d", i%5) // 模拟不同的端点

		// 模拟请求方法和延迟
		if rand.Intn(2) == 0 {
			method = "POST"
			path = fmt.Sprintf("/api/v1/resource/%d/order", i) // 模拟订单创建路径
		} else {
			method = "GET"
			path = fmt.Sprintf("/api/v1/resource/%d/status", i) // 模拟订单状态查询路径
		}
		duration := time.Duration(rand.Intn(1000)) * time.Millisecond

		RecordHTTPRequest(method, path, endpoint, duration)
		RecordOrderStatus("success", endpoint)

		time.Sleep(500 * time.Millisecond) // 模拟请求间隔
	}

	select {} // 阻塞主 goroutine，保持程序运行
}
