package site

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// --- 流量统计（W-018）：基于 Nginx 访问日志解析 PV/UV ---

var (
	// combined log: IP - user [time] "METHOD uri PROTO" status bytes "referer" "ua"
	logLineRegex  = regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "([^"]*)" (\d{3}) (\d+|-)`)
	logTimeLayout = "02/Jan/2006:15:04:05 -0700"
)

// TrafficBucket 流量统计时间桶
type TrafficBucket struct {
	Time string `json:"time"` // 2026-09-13 14:00 或 2026-09-13
	PV   int64  `json:"pv"`
	UV   int64  `json:"uv"`
}

// TrafficSummary 流量统计汇总
type TrafficSummary struct {
	Range     string          `json:"range"`
	TotalPV   int64           `json:"total_pv"`
	TotalUV   int64           `json:"total_uv"`
	UniqueIPs int64           `json:"unique_ips"`
	TopURLs   []URLCount      `json:"top_urls"`
	TopIPs    []URLCount      `json:"top_ips"`
	Status2xx int64           `json:"status_2xx"`
	Status4xx int64           `json:"status_4xx"`
	Status5xx int64           `json:"status_5xx"`
	Buckets   []TrafficBucket `json:"buckets"`
}

// URLCount URL/IP 计数项
type URLCount struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// GetTraffic 解析站点访问日志，返回 PV/UV 趋势（24h 按小时，7d/30d 按天）
func (s *Service) GetTraffic(siteID uint, rangeStr string) (*TrafficSummary, error) {
	site, err := s.GetSite(siteID)
	if err != nil {
		return nil, err
	}

	dur := 24 * time.Hour
	bucketFormat := "2006-01-02 15:00"
	bucketDur := time.Hour
	switch rangeStr {
	case "7d":
		dur = 7 * 24 * time.Hour
		bucketFormat = "2006-01-02"
		bucketDur = 24 * time.Hour
	case "30d":
		dur = 30 * 24 * time.Hour
		bucketFormat = "2006-01-02"
		bucketDur = 24 * time.Hour
	default:
		rangeStr = "24h"
	}

	logPath := filepath.Join(siteRootDir(site.PrimaryDomain), "logs", "access.log")
	f, err := os.Open(logPath)
	if err != nil {
		// Windows 开发环境回退到 runtime 目录
		logPath = filepath.Join("./runtime/www", site.PrimaryDomain, "logs", "access.log")
		f, err = os.Open(logPath)
		if err != nil {
			return &TrafficSummary{Range: rangeStr, Buckets: []TrafficBucket{}, TopURLs: []URLCount{}, TopIPs: []URLCount{}}, nil
		}
	}
	defer f.Close()

	now := time.Now()
	cutoff := now.Add(-dur)
	bucketCounts := map[string]int64{}
	bucketUV := map[string]map[string]bool{}
	ipSet := map[string]bool{}
	urlCounts := map[string]int64{}
	ipCounts := map[string]int64{}
	var pv, s2, s4, s5 int64

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		m := logLineRegex.FindStringSubmatch(scanner.Text())
		if m == nil {
			continue
		}
		ts, err := time.Parse(logTimeLayout, m[2])
		if err != nil || ts.Before(cutoff) || ts.After(now.Add(time.Minute)) {
			continue
		}
		ip := m[1]
		status, _ := strconv.Atoi(m[4])
		uri := m[3]
		// 提取路径（去掉 method 与 query）
		if fields := strings.Fields(uri); len(fields) >= 2 {
			uri = fields[1]
			if idx := strings.Index(uri, "?"); idx > 0 {
				uri = uri[:idx]
			}
		} else {
			uri = "-"
		}

		pv++
		ipSet[ip] = true
		ipCounts[ip]++
		urlCounts[uri]++
		switch {
		case status >= 200 && status < 300:
			s2++
		case status >= 400 && status < 500:
			s4++
		case status >= 500:
			s5++
		}

		bkey := ts.Format(bucketFormat)
		bucketCounts[bkey]++
		if bucketUV[bkey] == nil {
			bucketUV[bkey] = map[string]bool{}
		}
		bucketUV[bkey][ip] = true
	}

	// 补齐空桶保证时间轴连续
	buckets := make([]TrafficBucket, 0, dur/bucketDur+1)
	for t := now.Truncate(bucketDur).Add(-dur.Truncate(bucketDur)); !t.After(now.Truncate(bucketDur)); t = t.Add(bucketDur) {
		key := t.Format(bucketFormat)
		b := TrafficBucket{Time: key, PV: bucketCounts[key], UV: int64(len(bucketUV[key]))}
		buckets = append(buckets, b)
	}

	summary := &TrafficSummary{
		Range:     rangeStr,
		TotalPV:   pv,
		TotalUV:   int64(len(ipSet)),
		UniqueIPs: int64(len(ipSet)),
		Status2xx: s2,
		Status4xx: s4,
		Status5xx: s5,
		Buckets:   buckets,
		TopURLs:   topN(urlCounts, 10),
		TopIPs:    topN(ipCounts, 10),
	}
	return summary, nil
}

// topN 取计数前 N
func topN(counts map[string]int64, n int) []URLCount {
	items := make([]URLCount, 0, len(counts))
	for k, v := range counts {
		items = append(items, URLCount{Key: k, Count: v})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Count > items[j].Count })
	if len(items) > n {
		items = items[:n]
	}
	return items
}
