package ssmonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Chart struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	NameEN     string `json:"nameEN"`
	Unit       string `json:"unit"`
	Expression string `json:"-"`
}
type Point struct {
	Time  float64 `json:"time"`
	Value float64 `json:"value"`
}
type Series struct {
	Chart
	Points    []Point `json:"points"`
	Error     string  `json:"error,omitempty"`
	Available bool    `json:"available"`
}

func Catalog(kind string) []Chart {
	if kind == "logservice" {
		return []Chart{
			{"ready", "Ready 副本数", "Ready replicas", "nodes", `sum(ss_ls_ready_replicas{FILTER})`},
			{"desired", "期望副本数", "Desired replicas", "nodes", `sum(ss_ls_desired_replicas{FILTER})`},
			{"running", "LS 运行状态", "LS running state", "0 / 1", `min(ss_ls_running{FILTER})`},
			{"active", "注册表 Active LN", "Registry Active LNs", "nodes", `sum(ss_ls_active_nodes{FILTER})`},
			{"restarts", "10 分钟容器重启", "Container restarts in 10 minutes", "restarts", `sum(increase(ss_ls_container_restarts_total{FILTER}[10m]))`},
			{"registry", "LS 注册表采集状态", "LS registry collection health", "0 / 1", `min(ss_ls_registry_success{FILTER})`},
			{"collection", "Kubernetes 采集状态", "Kubernetes collection health", "0 / 1", `min(ss_ls_collection_success{FILTER})`},
		}
	}
	return []Chart{
		{"collection", "SS 数据库采集状态", "SS database collection health", "0 / 1", `min(ss_ob_collection_success{FILTER})`},
		{"read", "对象存储远端读取", "Remote object storage reads", "ops/s", `sum(rate(ss_ob_stat_total{FILTER,stat_id=~"60101|60107"}[2m]))`},
		{"write", "对象存储远端写入", "Remote object storage writes", "ops/s", `sum(rate(ss_ob_stat_total{FILTER,stat_id=~"60098|60104"}[2m]))`},
		{"read_bytes", "对象存储读取带宽", "Object storage read bandwidth", "bytes/s", `sum(rate(ss_ob_stat_total{FILTER,stat_id=~"60102|60108"}[2m]))`},
		{"write_bytes", "对象存储写入带宽", "Object storage write bandwidth", "bytes/s", `sum(rate(ss_ob_stat_total{FILTER,stat_id=~"60099|60105"}[2m]))`},
		{"head", "对象存储 HEAD 请求", "Object storage HEAD requests", "ops/s", `sum(rate(ss_ob_stat_total{FILTER,stat_id="240022"}[2m]))`},
		{"head_failures", "对象存储 HEAD 失败", "Object storage HEAD failures", "errors/s", `sum(rate(ss_ob_stat_total{FILTER,stat_id="240023"}[2m]))`},
		{"cache_hit_ratio", "SS 微缓存命中率（有请求时）", "SS micro-cache hit ratio (with traffic)", "%", `100 * sum(rate(ss_ob_stat_total{FILTER,stat_id="240001"}[2m])) / sum(rate(ss_ob_stat_total{FILTER,stat_id=~"240001|240002"}[2m]))`},
		{"cache_hits", "SS 微缓存命中", "SS micro-cache hits", "ops/s", `sum(rate(ss_ob_stat_total{FILTER,stat_id="240001"}[2m]))`},
		{"cache_misses", "SS 微缓存未命中", "SS micro-cache misses", "ops/s", `sum(rate(ss_ob_stat_total{FILTER,stat_id="240002"}[2m]))`},
		{"cache_bytes", "SS 微缓存持有数据量", "SS micro-cache held bytes", "bytes", `sum(ss_ob_stat{FILTER,stat_id="240010"})`},
		{"cache_disk", "SS 微缓存磁盘占用", "SS micro-cache disk usage", "bytes", `sum(ss_ob_stat{FILTER,stat_id="240019"})`},
		{"cache_errors", "SS 微缓存读写失败", "SS micro-cache get/add failures", "errors/s", `sum(rate(ss_ob_stat_total{FILTER,stat_id=~"240003|240004"}[2m]))`},
	}
}

type queryResponse struct {
	Status string `json:"status"`
	Data   struct {
		Result []struct {
			Values [][2]json.RawMessage `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// Public callers select a fixed catalog, not arbitrary PromQL. Labels are
// quoted to keep requests scoped to the RBAC-authorized Kubernetes identity.
func Query(ctx context.Context, baseURL, kind, namespace, name string, minutes int) ([]Series, error) {
	if kind != "logservice" && kind != "obcluster" {
		return nil, fmt.Errorf("unsupported monitoring scope")
	}
	if minutes != 15 && minutes != 60 && minutes != 360 && minutes != 1440 {
		return nil, fmt.Errorf("range must be 15, 60, 360 or 1440 minutes")
	}
	filter := "namespace=" + strconv.Quote(namespace) + "," + kind + "=" + strconv.Quote(name)
	catalog := Catalog(kind)
	out := make([]Series, len(catalog))
	end := time.Now().Unix()
	step := int64(minutes * 60 / 180)
	if step < 30 {
		step = 30
	}
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)
	for i, chart := range catalog {
		wg.Add(1)
		go func(i int, chart Chart) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				out[i] = Series{Chart: chart, Points: []Point{}, Error: "Monitoring request timed out"}
				return
			}
			series := Series{Chart: chart, Points: []Point{}}
			args := url.Values{"query": {strings.ReplaceAll(chart.Expression, "FILTER", filter)}, "start": {strconv.FormatInt(end-int64(minutes*60), 10)}, "end": {strconv.FormatInt(end, 10)}, "step": {strconv.FormatInt(step, 10)}}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v1/query_range?"+args.Encode(), nil)
			if err != nil {
				series.Error = "Invalid monitoring endpoint"
				out[i] = series
				return
			}
			resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
			if err != nil {
				series.Error = "Prometheus is unavailable"
				out[i] = series
				return
			}
			defer resp.Body.Close()
			var result queryResponse
			if resp.StatusCode != 200 || json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&result) != nil || result.Status != "success" {
				series.Error = "Prometheus query failed"
				out[i] = series
				return
			}
			for _, r := range result.Data.Result {
				for _, v := range r.Values {
					var ts float64
					var raw string
					if json.Unmarshal(v[0], &ts) != nil || json.Unmarshal(v[1], &raw) != nil {
						continue
					}
					n, err := strconv.ParseFloat(raw, 64)
					if err != nil || math.IsInf(n, 0) || math.IsNaN(n) {
						continue
					}
					series.Points = append(series.Points, Point{Time: ts, Value: n})
				}
			}
			series.Available = len(series.Points) > 0 && series.Points[len(series.Points)-1].Time >= float64(end-step*2-60)
			out[i] = series
		}(i, chart)
	}
	wg.Wait()
	return out, nil
}
