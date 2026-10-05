// Package metrics 提供进程内的接口请求指标采集（重启清零，无外部依赖）。
package metrics

import (
	"sort"
	"sync"
	"time"
)

// Metric 单个接口（METHOD + 路由模板）的统计结果。
type Metric struct {
	Method    string
	Path      string
	Total     int64
	Success   int64
	Error     int64
	ErrorRate float64 // 0-100
	AvgMs     float64
	MaxMs     float64
	LastMs    float64
}

type routeKey struct {
	method string
	path   string
}

type routeStat struct {
	total   int64
	success int64
	errs    int64
	sumMs   float64
	maxMs   float64
	lastMs  float64
}

// Collector 接口指标采集器（并发安全）。
type Collector struct {
	mu        sync.RWMutex
	startedAt time.Time
	routes    map[routeKey]*routeStat
}

// NewCollector 创建采集器
func NewCollector() *Collector {
	return &Collector{
		startedAt: time.Now(),
		routes:    make(map[routeKey]*routeStat),
	}
}

// StartedAt 采集起始时间（进程启动时间）
func (c *Collector) StartedAt() time.Time {
	return c.startedAt
}

// Record 记录一次请求。path 建议传路由模板（如 /admin/v1/admin/{id}）。
func (c *Collector) Record(method, path string, cost time.Duration, isError bool) {
	if method == "" || path == "" {
		return
	}
	ms := float64(cost.Microseconds()) / 1000
	key := routeKey{method: method, path: path}

	c.mu.Lock()
	defer c.mu.Unlock()
	st, ok := c.routes[key]
	if !ok {
		st = &routeStat{}
		c.routes[key] = st
	}
	st.total++
	if isError {
		st.errs++
	} else {
		st.success++
	}
	st.sumMs += ms
	st.lastMs = ms
	if ms > st.maxMs {
		st.maxMs = ms
	}
}

// Snapshot 返回当前指标快照（按请求量倒序）及汇总值。
func (c *Collector) Snapshot() (items []Metric, totalRequests, totalErrors int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	items = make([]Metric, 0, len(c.routes))
	for k, st := range c.routes {
		m := Metric{
			Method:  k.method,
			Path:    k.path,
			Total:   st.total,
			Success: st.success,
			Error:   st.errs,
			MaxMs:   round2(st.maxMs),
			LastMs:  round2(st.lastMs),
		}
		if st.total > 0 {
			m.AvgMs = round2(st.sumMs / float64(st.total))
			m.ErrorRate = round2(float64(st.errs) / float64(st.total) * 100)
		}
		items = append(items, m)
		totalRequests += st.total
		totalErrors += st.errs
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Total == items[j].Total {
			return items[i].Path < items[j].Path
		}
		return items[i].Total > items[j].Total
	})
	return items, totalRequests, totalErrors
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
