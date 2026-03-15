package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var Global = New()

type Collector struct {
	counters   sync.Map
	histograms sync.Map
}

type counter struct {
	values sync.Map
}

type histogram struct {
	mu      sync.Mutex
	buckets []float64
	values  map[string]*histData
}

type histData struct {
	count   int64
	sum     float64
	buckets []int64
}

func New() *Collector { return &Collector{} }

func (c *Collector) Inc(name string, labels map[string]string) {
	raw, _ := c.counters.LoadOrStore(name, &counter{})
	ctr := raw.(*counter)
	key := labelsKey(labels)
	raw2, _ := ctr.values.LoadOrStore(key, new(int64))
	atomic.AddInt64(raw2.(*int64), 1)
}

func (c *Collector) Observe(name string, labels map[string]string, value float64) {
	bkts := []float64{10, 50, 100, 250, 500, 1000, 2500, 5000}
	raw, _ := c.histograms.LoadOrStore(name, &histogram{buckets: bkts, values: make(map[string]*histData)})
	h := raw.(*histogram)
	key := labelsKey(labels)

	h.mu.Lock()
	defer h.mu.Unlock()
	d, ok := h.values[key]
	if !ok {
		d = &histData{buckets: make([]int64, len(h.buckets))}
		h.values[key] = d
	}
	d.count++
	d.sum += value
	for i, b := range h.buckets {
		if value <= b {
			d.buckets[i]++
		}
	}
}

func (c *Collector) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var sb strings.Builder

		c.counters.Range(func(name, raw interface{}) bool {
			ctr := raw.(*counter)
			ctr.values.Range(func(key, val interface{}) bool {
				v := atomic.LoadInt64(val.(*int64))
				if key.(string) == "" {
					fmt.Fprintf(&sb, "%s %d\n", name, v)
				} else {
					fmt.Fprintf(&sb, "%s{%s} %d\n", name, key, v)
				}
				return true
			})
			return true
		})

		c.histograms.Range(func(name, raw interface{}) bool {
			h := raw.(*histogram)
			h.mu.Lock()
			defer h.mu.Unlock()
			for key, d := range h.values {
				lbl := ""
				if key != "" {
					lbl = key + ","
				}
				for i, b := range h.buckets {
					fmt.Fprintf(&sb, "%s_bucket{%sle=\"%g\"} %d\n", name, lbl, b, d.buckets[i])
				}
				fmt.Fprintf(&sb, "%s_bucket{%sle=\"+Inf\"} %d\n", name, lbl, d.count)
				fmt.Fprintf(&sb, "%s_count{%s} %d\n", name, strings.TrimSuffix(lbl, ","), d.count)
				fmt.Fprintf(&sb, "%s_sum{%s} %g\n", name, strings.TrimSuffix(lbl, ","), d.sum)
			}
			return true
		})

		ctx.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(sb.String()))
	}
}

func RequestMetrics(c *Collector) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		dur := float64(time.Since(start).Milliseconds())
		labels := map[string]string{
			"method": ctx.Request.Method,
			"path":   ctx.FullPath(),
			"status": fmt.Sprintf("%d", ctx.Writer.Status()),
		}
		c.Inc("captcha_http_requests_total", labels)
		c.Observe("captcha_http_duration_ms", labels, dur)
	}
}

// Application metrics
func RecordPrecheck(action, scene string, score float64) {
	Global.Inc("captcha_precheck_total", map[string]string{"action": action, "scene": scene})
	Global.Observe("captcha_risk_score", map[string]string{"scene": scene}, score)
}

func RecordChallenge(challengeType, result string, durationMs float64) {
	Global.Inc("captcha_challenges_total", map[string]string{"type": challengeType, "result": result})
	Global.Observe("captcha_challenge_duration_ms", map[string]string{"type": challengeType}, durationMs)
}

func RecordBot(detectionMethod string) {
	Global.Inc("captcha_bots_detected_total", map[string]string{"method": detectionMethod})
}

func RecordRateLimit(ip string) {
	Global.Inc("captcha_rate_limits_total", nil)
}

func labelsKey(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, labels[k]))
	}
	return strings.Join(parts, ",")
}
