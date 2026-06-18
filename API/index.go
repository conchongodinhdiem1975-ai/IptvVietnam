package handler

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Channel định nghĩa cấu trúc luồng IPTV
type Channel struct {
	Name string
	URL  string
}

// CacheItem lưu trữ cấu trúc m3u8 trong RAM
type CacheItem struct {
	Data      string
	Timestamp time.Time
}

var CONFIG = struct {
	UserAgent string
	Channels  []Channel
}{
	UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	Channels: []Channel{
		{Name: "IPTV Cat 01", URL: "https://list.iptvcat.com/my_list/s/bd7d295cbc106d36491e59f262f3efb5.m3u8"},
		{Name: "IPTV Cat 02", URL: "https://list.iptvcat.com/my_list/s/ac86e70fac4f2daa2e7858101e3693f4.m3u8"},
		{Name: "IPTV Cat 03", URL: "https://list.iptvcat.com/my_list/s/0f10f2e8554a7e2a038f86cb41b9c1f7.m3u8"},
		{Name: "IPTV Cat 04", URL: "https://list.iptvcat.com/my_list/s/d1d4f0d5a76ed49118af5f4531f66453.m3u8"},
		{Name: "IPTV Cat 05", URL: "https://list.iptvcat.com/my_list/s/0ffcbd60dcb4b4b6339ea09733bccc58.m3u8"},
		{Name: "IPTV Cat 06", URL: "https://list.iptvcat.com/my_list/s/300ca37df9856b8951b4f91c7f2f501a.m3u8"},
		{Name: "IPTV Cat 07", URL: "https://list.iptvcat.com/my_list/s/d425b809055929b7a3742bb7d22a573f.m3u8"},
		{Name: "IPTV Cat 08", URL: "https://list.iptvcat.com/my_list/s/774e2a3c9c6fa9940b23c541015ea27a.m3u8"},
		{Name: "IPTV Cat 09", URL: "https://list.iptvcat.com/my_list/s/0f10f2e8554a7e2a038f86cb41b9c1f7.m3u8"},
		{Name: "Cartoon Network US Premium", URL: "http://23.237.104.106:8080/USA_CARTOON_NETWORK/index.m3u8"},
	},
}

// Khởi tạo Client điều tốc kết nối nội bộ và bộ nhớ đệm luồng
var (
	m3uCache   = make(map[string]CacheItem)
	cacheMutex sync.RWMutex
	httpClient = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        50,
			IdleConnTimeout:     60 * time.Second,
			MaxIdleConnsPerHost: 10,
		},
	}
)

// Handler là cổng vào duy nhất được Vercel gọi khi có Request
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	switch r.URL.Path {
	case "/":
		handleDashboard(w, r)
	case "/playlist":
		handlePlaylist(w, r)
	case "/proxy":
		handleProxy(w, r)
	default:
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}
}

// Tuyến 0: Concurrency Dashboard - Quét đa luồng kiểm tra trạng thái kênh cực nhanh
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	type ChanStatus struct {
		Name   string
		Status string
		Color  string
	}

	statusChan := make(chan ChanStatus, len(CONFIG.Channels))
	var wg sync.WaitGroup

	// Kỹ thuật Goroutine: Kích hoạt quét đồng thời tất cả các kênh cùng một lúc (Zero-waiting)
	for _, ch := range CONFIG.Channels {
		wg.Add(1)
		go func(c Channel) {
			defer wg.Done()
			req, _ := http.NewRequest("HEAD", c.URL, nil)
			req.Header.Set("User-Agent", CONFIG.UserAgent)
			
			client := &http.Client{Timeout: 1500 * time.Millisecond} // Giới hạn phản hồi 1.5s
			resp, err := client.Do(req)
			
			if err == nil && resp.StatusCode == http.StatusOK {
				statusChan <- ChanStatus{Name: c.Name, Status: "ONLINE", Color: "#2ecc71"}
				resp.Body.Close()
			} else {
				statusChan <- ChanStatus{Name: c.Name, Status: "OFFLINE", Color: "#e74c3c"}
			}
		}(ch)
	}

	wg.Wait()
	close(statusChan)

	var rows strings.Builder
	for ch := range statusChan {
		rows.WriteString(fmt.Sprintf(`
			<div style="display: flex; justify-content: space-between; background: #2c3e50; padding: 12px; margin: 8px 0; border-radius: 6px;">
				<span style="font-weight: bold; color: #ecf0f1;">📺 %s</span>
				<span style="background: %s; color: white; padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: bold;">%s</span>
			</div>`, ch.Name, ch.Color, ch.Status))
	}

	baseUrl := fmt.Sprintf("http://%s", r.Host)
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		baseUrl = fmt.Sprintf("https://%s", r.Host)
	}

	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Go IPTV VIP Gateway</title>
	</head>
	<body style="background: #1a252f; font-family: system-ui, sans-serif; color: #95a5a6; padding: 20px; max-width: 600px; margin: 0 auto;">
		<h2 style="color: #00add8; border-bottom: 2px solid #00add8; padding-bottom: 10px;">🐹 Go IPTV VIP Core (Vercel)</h2>
		<p>Hệ thống lõi Engine: <span style="color: #2ecc71; font-weight: bold;">GOLANG NATIVE REALTIME</span></p>
		<p>Nạp Link này vào App Tivi của bạn: <br>
			<input type="text" value="%s/playlist" readonly style="width: 100%; background: #2c3e50; border: none; padding: 10px; color: #2ecc71; font-weight: bold; border-radius: 4px; box-sizing: border-box; margin-top: 5px;">
		</p>
		<h3 style="color: #ecf0f1; margin-top: 25px;">📊 Giám sát Luồng Kênh:</h3>
		<div>%s</div>
	</body>
	</html>`, baseUrl, rows.String())

	w.Write([]byte(html))
}

// Tuyến 1: Phát xuất Playlist cấu trúc M3U
func handlePlaylist(w http.ResponseWriter, r *http.Request) {
	baseUrl := fmt.Sprintf("http://%s", r.Host)
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		baseUrl = fmt.Sprintf("https://%s", r.Host)
	}

	var m3u strings.Builder
	m3u.WriteString("#EXTM3U\n")
	for _, ch := range CONFIG.Channels {
		m3u.WriteString(fmt.Sprintf("#EXTINF:-1, %s\n%s/proxy?url=%s\n", ch.Name, baseUrl, url.QueryEscape(ch.URL)))
	}

	w.Header().Set("Content-Type", "application/x-mpegURL")
	w.Write([]byte(m3u.String()))
}

// Tuyến 2: Siêu máy chủ Proxy xử lý dữ liệu Nhị phân Video (.ts) và bẻ khóa liên kết (.m3u8)
func handleProxy(w http.ResponseWriter, r *http.Request) {
	targetURLStr := r.URL.Query().Get("url")
	if targetURLStr == "" {
		http.Error(w, "Missing target URL", http.StatusBadRequest)
		return
	}

	targetURL, err := url.Parse(targetURLStr)
	if err != nil {
		http.Error(w, "Invalid target URL", http.StatusBadRequest)
		return
	}

	baseUrl := fmt.Sprintf("http://%s", r.Host)
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		baseUrl = fmt.Sprintf("https://%s", r.Host)
	}

	// 1. XỬ LÝ CẤU TRÚC KÊNH M3U8 (Có áp dụng Thread-Safe Cache bằng Mutex)
	if strings.Contains(targetURLStr, ".m3u8") {
		cacheMutex.RLock()
		cached, exists := m3uCache[targetURLStr]
		cacheMutex.RUnlock()

		if exists && time.Since(cached.Timestamp) < 10*time.Second {
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			w.Header().Set("X-Go-Cache", "HIT")
			w.Write([]byte(cached.Data))
			return
		}

		// Tạo Request kéo luồng có Context để tự động hủy khi Tivi ngắt kết nối
		req, _ := http.NewRequestWithContext(r.Context(), "GET", targetURLStr, nil)
		req.Header.Set("User-Agent", CONFIG.UserAgent)
		req.Header.Set("Referer", targetURL.Scheme+"://"+targetURL.Host+"/")

		resp, err := httpClient.Do(req)
		if err != nil {
			http.Error(w, "Fetch Origin Error", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		var rewrittenM3u8 strings.Builder
		scanner := bufio.NewScanner(resp.Body)
		
		for scanner.Scan() {
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)

			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				rewrittenM3u8.WriteString(line + "\n")
				continue
			}

			resolvedURL, err := targetURL.Parse(trimmed)
			if err != nil {
				rewrittenM3u8.WriteString(line + "\n")
				continue
			}

			rewrittenM3u8.WriteString(fmt.Sprintf("%s/proxy?url=%s\n", baseUrl, url.QueryEscape(resolvedURL.String())))
		}

		finalData := rewrittenM3u8.String()

		// Ghi đệm dữ liệu mới vào RAM
		cacheMutex.Lock()
		m3uCache[targetURLStr] = CacheItem{Data: finalData, Timestamp: time.Now()}
		cacheMutex.Unlock()

		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Header().Set("X-Go-Cache", "MISS")
		w.Write([]byte(finalData))
		return
	}

	// 2. CHUYỂN TIẾP FILE VIDEO NHỊ PHÂN .TS (Kỹ thuật Zero-Allocation qua io.Copy)
	req, _ := http.NewRequestWithContext(r.Context(), "GET", targetURLStr, nil)
	req.Header.Set("User-Agent", CONFIG.UserAgent)
	req.Header.Set("Referer", targetURL.Scheme+"://"+targetURL.Host+"/")

	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, "Stream Connection Error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", "video/mp2t")
	}

	w.WriteHeader(resp.StatusCode)
	
	// Bơm trực tiếp luồng nhị phân từ Card mạng về Tivi, RAM tiêu thụ cố định ở mức siêu thấp
	_, _ = io.Copy(w, resp.Body)
}
