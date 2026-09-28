package config

import (
	"os"
	"strconv"
)

// Config 保存服务运行所需的全部配置。
type Config struct {
	// Addr 是 HTTP 服务监听地址，例如 ":8080"。
	Addr string
	// DatabaseURL 是 PostgreSQL 连接串。
	DatabaseURL string
	// FreeFlowSpeed 是计算拥堵指数时使用的自由流速度（km/h）。
	FreeFlowSpeed float64
	// LowSpeedThreshold 是"连续低速"告警的速度阈值（km/h）。
	LowSpeedThreshold float64
	// FlowSpikeFactor 是"流量突增"告警的倍数阈值。
	FlowSpikeFactor float64
}

// Load 从环境变量读取配置，未设置时使用默认值。
func Load() *Config {
	return &Config{
		Addr:              getEnv("SERVER_ADDR", ":8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://traffic:traffic@localhost:5432/traffic?sslmode=disable"),
		FreeFlowSpeed:     getFloatEnv("FREE_FLOW_SPEED", 60.0),
		LowSpeedThreshold: getFloatEnv("LOW_SPEED_THRESHOLD", 20.0),
		FlowSpikeFactor:   getFloatEnv("FLOW_SPIKE_FACTOR", 1.5),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getFloatEnv(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

