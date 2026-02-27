package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP 请求总数
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "grape",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP 请求耗时
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "grape",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// 包下载次数
	PackageDownloadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "grape",
			Name:      "package_downloads_total",
			Help:      "Total number of package tarball downloads",
		},
		[]string{"package"},
	)

	// 包发布次数
	PackagePublishTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "grape",
			Name:      "package_publish_total",
			Help:      "Total number of package publish operations",
		},
		[]string{"package", "status"},
	)

	// 上游代理请求次数
	ProxyRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "grape",
			Name:      "proxy_requests_total",
			Help:      "Total number of upstream proxy requests",
		},
		[]string{"upstream", "status"},
	)

	// 存储中已缓存的包数量（Gauge）
	StoredPackagesTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "grape",
			Name:      "stored_packages_total",
			Help:      "Total number of packages stored locally",
		},
	)

	// 活跃用户数（Gauge）
	RegisteredUsersTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "grape",
			Name:      "registered_users_total",
			Help:      "Total number of registered users",
		},
	)

	// 登录尝试次数
	LoginAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "grape",
			Name:      "login_attempts_total",
			Help:      "Total number of login attempts",
		},
		[]string{"status"},
	)
)
