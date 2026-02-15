// Agent Framework - Network Tools Module
// Copyright (C) 2025 Agent Framework Contributors
// SPDX-License-Identifier: Apache-2.0

package sys

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// NetworkToolsModule 网络工具模块
type NetworkToolsModule struct {
	config NetworkConfig
	mu     sync.RWMutex
	stats  *NetworkStats
	client *http.Client
}

// NetworkConfig 网络工具配置
type NetworkConfig struct {
	Timeout            int      `json:"timeout"`              // 请求超时（毫秒）
	MaxRedirects       int      `json:"max_redirects"`        // 最大重定向次数
	UserAgent          string   `json:"user_agent"`           // 用户代理
	AllowedHosts       []string `json:"allowed_hosts"`        // 允许访问的主机
	BlockedHosts      []string `json:"blocked_hosts"`       // 禁止访问的主机
	AllowedIPRanges    []string `json:"allowed_ip_ranges"`    // 允许的 IP 范围
	EnableDNSLookup    bool     `json:"enable_dns_lookup"`    // 启用 DNS 查询
	EnablePortScan     bool     `json:"enable_port_scan"`      // 启用端口扫描
	MaxPortScanPorts  int      `json:"max_port_scan_ports"`  // 端口扫描最大端口数
}

// NetworkStats 网络统计信息
type NetworkStats struct {
	TotalRequests     int64     `json:"total_requests"`
	SuccessRequests   int64     `json:"success_requests"`
	FailedRequests    int64     `json:"failed_requests"`
	TotalBytesSent    int64     `json:"total_bytes_sent"`
	TotalBytesRecv    int64     `json:"total_bytes_recv"`
	Pings             int64     `json:"pings"`
	PortScans         int64     `json:"port_scans"`
	mu                sync.RWMutex `json:"-"`
}

// NetworkRequestInfo 网络请求信息
type NetworkRequestInfo struct {
	URL          string            `json:"url"`
	Method       string            `json:"method"`
	StatusCode   int               `json:"status_code"`
	ContentType  string            `json:"content_type"`
	ContentLength int64            `json:"content_length"`
	ResponseTime int64             `json:"response_time_ms"`
	Headers      map[string]string `json:"headers"`
}

// PingResult Ping 结果
type PingResult struct {
	Host        string  `json:"host"`
	Success     bool    `json:"success"`
	TimeMs      int64   `json:"time_ms"`
	IP          string  `json:"ip"`
	Error       string  `json:"error,omitempty"`
}

// PortInfo 端口信息
type PortInfo struct {
	Port    int    `json:"port"`
	Status  string `json:"status"` // open, closed, filtered
	Service string `json:"service,omitempty"`
}

// NewNetworkToolsModule 创建网络工具模块实例
func NewNetworkToolsModule(config NetworkConfig) (*NetworkToolsModule, error) {
	if config.Timeout <= 0 {
		config.Timeout = 30000 // 默认30秒
	}
	if config.MaxRedirects <= 0 {
		config.MaxRedirects = 10 // 默认最多10次重定向
	}
	if config.UserAgent == "" {
		config.UserAgent = "AgentFramework/1.0 NetworkTools"
	}
	if config.MaxPortScanPorts <= 0 {
		config.MaxPortScanPorts = 1024 // 默认扫描前1024个端口
	}

	stats := &NetworkStats{}

	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Millisecond,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	return &NetworkToolsModule{
		config: config,
		stats:  stats,
		client: client,
	}, nil
}

// GetTools 返回网络工具模块的 MCP 工具列表
func (m *NetworkToolsModule) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	tools := []tool.BaseTool{
		// HTTP 请求工具
		&httpRequestTool{module: m},
		// DNS 查询工具
		&dnsLookupTool{module: m},
		// Ping 工具
		&pingTool{module: m},
		// 端口扫描工具
		&portScanTool{module: m},
		// 网络接口信息工具
		&networkInterfacesTool{module: m},
		// IP 地址查询工具
		&ipQueryTool{module: m},
	}

	return tools, nil
}

// HTTP 请求工具
type httpRequestTool struct {
	module *NetworkToolsModule
}

func (t *httpRequestTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "http_request",
		Desc: "Make an HTTP request with support for various methods and custom headers",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"url": {
				Type:     "string",
				Desc:     "Target URL",
				Required:  true,
			},
			"method": {
				Type:    "string",
				Desc:    "HTTP method (GET, POST, PUT, DELETE, PATCH, HEAD)",
			},
			"headers": {
				Type: "object",
				Desc: "Custom headers as key-value pairs",
			},
			"body": {
				Type: "string",
				Desc: "Request body for POST/PUT/PATCH requests",
			},
			"timeout": {
				Type: "integer",
				Desc: "Request timeout in milliseconds",
			},
		}),
	}, nil
}

func (t *httpRequestTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		URL     string            `json:"url"`
		Method  string            `json:"method"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
		Timeout int               `json:"timeout"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Method == "" {
		args.Method = "GET"
	}

	result, err := t.module.makeRequest(args.Method, args.URL, args.Headers, args.Body, args.Timeout)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// DNS 查询工具
type dnsLookupTool struct {
	module *NetworkToolsModule
}

func (t *dnsLookupTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "dns_lookup",
		Desc: "Perform DNS lookup for a hostname",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"hostname": {
				Type:     "string",
				Desc:     "Hostname to lookup",
				Required:  true,
			},
			"record_type": {
				Type: "string",
				Desc: "DNS record type (A, AAAA, MX, TXT, CNAME)",
			},
		}),
	}, nil
}

func (t *dnsLookupTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Hostname    string `json:"hostname"`
		RecordType  string `json:"record_type"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.RecordType == "" {
		args.RecordType = "A"
	}

	result, err := t.module.dnsLookup(args.Hostname, args.RecordType)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// Ping 工具
type pingTool struct {
	module *NetworkToolsModule
}

func (t *pingTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "ping",
		Desc: "Ping a host to check connectivity",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"host": {
				Type:     "string",
				Desc:     "Host to ping (hostname or IP)",
				Required:  true,
			},
			"count": {
				Type: "integer",
				Desc: "Number of ping packets (default: 4)",
			},
		}),
	}, nil
}

func (t *pingTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Host  string `json:"host"`
		Count int    `json:"count"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Count <= 0 {
		args.Count = 4
	}

	result, err := t.module.pingHost(args.Host, args.Count)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 端口扫描工具
type portScanTool struct {
	module *NetworkToolsModule
}

func (t *portScanTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "port_scan",
		Desc: "Scan common ports on a host to check which are open",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"host": {
				Type:     "string",
				Desc:     "Target host",
				Required:  true,
			},
			"ports": {
				Type: "array",
				Desc: "Specific ports to scan (default: common ports)",
			},
			"common_only": {
				Type: "boolean",
				Desc: "Only scan common ports (default: true)",
			},
		}),
	}, nil
}

func (t *portScanTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Host       string `json:"host"`
		Ports      []int  `json:"ports"`
		CommonOnly bool    `json:"common_only"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.scanPorts(args.Host, args.Ports, args.CommonOnly)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// 网络接口信息工具
type networkInterfacesTool struct {
	module *NetworkToolsModule
}

func (t *networkInterfacesTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "network_interfaces",
		Desc:        "Get list of network interfaces and their information",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (t *networkInterfacesTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	result, err := t.module.getNetworkInterfaces()
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// IP 地址查询工具
type ipQueryTool struct {
	module *NetworkToolsModule
}

func (t *ipQueryTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "ip_query",
		Desc: "Query IP address information",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"ip": {
				Type:     "string",
				Desc:     "IP address to query",
				Required:  true,
			},
		}),
	}, nil
}

func (t *ipQueryTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		IP string `json:"ip"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	result, err := t.module.queryIP(args.IP)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// Close 关闭网络工具模块
func (m *NetworkToolsModule) Close() error {
	m.client.CloseIdleConnections()
	return nil
}

// GetStats 获取统计信息
func (m *NetworkToolsModule) GetStats() map[string]int64 {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	return map[string]int64{
		"total_requests":    m.stats.TotalRequests,
		"success_requests":  m.stats.SuccessRequests,
		"failed_requests":   m.stats.FailedRequests,
		"total_bytes_sent":   m.stats.TotalBytesSent,
		"total_bytes_recv":   m.stats.TotalBytesRecv,
		"pings":             m.stats.Pings,
		"port_scans":        m.stats.PortScans,
	}
}

// ==================== 核心功能实现 ====================

// makeRequest 发送 HTTP 请求
func (m *NetworkToolsModule) makeRequest(method, url string, headers map[string]string, body string, timeout int) (map[string]any, error) {
	m.stats.mu.Lock()
	m.stats.TotalRequests++
	m.stats.mu.Unlock()

	// 验证 URL
	if !m.isURLAllowed(url) {
		return map[string]any{
			"success": false,
			"error":   "URL is not allowed by policy",
		}, nil
	}

	// 创建请求
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Failed to create request: %v", err),
		}, nil
	}

	// 设置请求头
	req.Header.Set("User-Agent", m.config.UserAgent)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 设置超时
	client := m.client
	if timeout > 0 {
		client = &http.Client{
			Timeout: time.Duration(timeout) * time.Millisecond,
		}
	}

	// 发送请求
	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		m.stats.mu.Lock()
		m.stats.FailedRequests++
		m.stats.mu.Unlock()

		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Request failed: %v", err),
			"time_ms":  duration.Milliseconds(),
		}, nil
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Failed to read response: %v", err),
		}, nil
	}

	// 更新统计
	m.stats.mu.Lock()
	m.stats.SuccessRequests++
	m.stats.TotalBytesSent += int64(req.ContentLength)
	m.stats.TotalBytesRecv += int64(len(respBody))
	m.stats.mu.Unlock()

	// 构建响应头
	respHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			respHeaders[key] = values[0]
		}
	}

	return map[string]any{
		"success":        true,
		"status_code":    resp.StatusCode,
		"status":         resp.Status,
		"headers":        respHeaders,
		"content_type":   resp.Header.Get("Content-Type"),
		"content_length": len(respBody),
		"body":          string(respBody),
		"time_ms":        duration.Milliseconds(),
	}
}

// dnsLookup 执行 DNS 查询
func (m *NetworkToolsModule) dnsLookup(hostname, recordType string) (map[string]any, error) {
	if !m.config.EnableDNSLookup {
		return map[string]any{
			"success": false,
			"error":   "DNS lookup is disabled",
		}, nil
	}

	// 简单的 DNS 查询实现
	var records []string

	switch strings.ToUpper(recordType) {
	case "A":
		ips, err := net.LookupIP(hostname)
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   fmt.Sprintf("DNS lookup failed: %v", err),
			}, nil
		}
		for _, ip := range ips {
			if ip.To4() != nil {
				records = append(records, ip.String())
			}
		}
	case "AAAA":
		ips, err := net.LookupIP(hostname)
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   fmt.Sprintf("DNS lookup failed: %v", err),
			}, nil
		}
		for _, ip := range ips {
			if ip.To4() == nil {
				records = append(records, ip.String())
			}
		}
	case "TXT":
		txtRecords, err := net.LookupTXT(hostname)
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   fmt.Sprintf("DNS lookup failed: %v", err),
			}, nil
		}
		records = txtRecords
	case "MX":
		mxRecords, err := net.LookupMX(hostname)
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   fmt.Sprintf("DNS lookup failed: %v", err),
			}, nil
		}
		for _, mx := range mxRecords {
			records = append(records, mx.Host)
		}
	case "CNAME":
		cname, err := net.LookupCNAME(hostname)
		if err != nil {
			return map[string]any{
				"success": false,
				"error":   fmt.Sprintf("DNS lookup failed: %v", err),
			}, nil
		}
		records = append(records, cname)
	default:
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Unsupported record type: %s", recordType),
		}, nil
	}

	return map[string]any{
		"success":      true,
		"hostname":     hostname,
		"record_type":  recordType,
		"records":      records,
		"count":        len(records),
	}, nil
}

// pingHost Ping 主机
func (m *NetworkToolsModule) pingHost(host string, count int) (map[string]any, error) {
	m.stats.mu.Lock()
	m.stats.Pings++
	m.stats.mu.Unlock()

	results := make([]PingResult, 0, count)

	// 解析主机名
	ipAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		// 尝试 DNS 查询
		addrs, err := net.LookupHost(host)
		if err != nil || len(addrs) == 0 {
			return map[string]any{
				"success": false,
				"error":   fmt.Sprintf("Failed to resolve host: %v", err),
			}, nil
		}
		ipAddr = &net.IPAddr{IP: net.ParseIP(addrs[0])}
	}

	targetIP := ipAddr.IP.String()

	// 执行 ping 测试
	for i := 0; i < count; i++ {
		result := m.pingOnce(targetIP)
		result.Host = host
		results = append(results, result)
		time.Sleep(time.Second)
	}

	// 计算统计
	successCount := 0
	var totalTime int64
	for _, r := range results {
		if r.Success {
			successCount++
			totalTime += r.TimeMs
		}
	}

	avgTime := int64(0)
	if successCount > 0 {
		avgTime = totalTime / int64(successCount)
	}

	return map[string]any{
		"success":      true,
		"host":         host,
		"ip":           targetIP,
		"results":      results,
		"packets_sent": count,
		"packets_recv": successCount,
		"packet_loss":   float64(count-successCount) / float64(count) * 100,
		"avg_time_ms":  avgTime,
	}, nil
}

// pingOnce 执行单次 ping
func (m *NetworkToolsModule) pingOnce(ip string) PingResult {
	startTime := time.Now()

	// 尝试连接到常见端口（简化版 ping）
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, "80"), time.Second*2)
	if err != nil {
		return PingResult{
			Success: false,
			IP:      ip,
			Error:   err.Error(),
		}
	}
	defer conn.Close()

	duration := time.Since(startTime)

	return PingResult{
		Success: true,
		TimeMs:  duration.Milliseconds(),
		IP:      ip,
	}
}

// scanPorts 扫描端口
func (m *NetworkToolsModule) scanPorts(host string, ports []int, commonOnly bool) (map[string]any, error) {
	if !m.config.EnablePortScan {
		return map[string]any{
			"success": false,
			"error":   "Port scanning is disabled",
		}, nil
	}

	m.stats.mu.Lock()
	m.stats.PortScans++
	m.stats.mu.Unlock()

	// 确定要扫描的端口
	var scanPorts []int
	if len(ports) > 0 {
		scanPorts = ports
	} else if commonOnly {
		scanPorts = m.getCommonPorts()
	} else {
		scanPorts = m.getDefaultPorts()
	}

	// 限制端口数量
	if len(scanPorts) > m.config.MaxPortScanPorts {
		scanPorts = scanPorts[:m.config.MaxPortScanPorts]
	}

	results := make([]PortInfo, 0, len(scanPorts))

	// 扫描端口
	for _, port := range scanPorts {
		info := PortInfo{Port: port}

		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), time.Second*2)
		if err != nil {
			info.Status = "closed"
		} else {
			info.Status = "open"
			info.Service = m.getPortService(port)
			conn.Close()
		}

		results = append(results, info)
	}

	// 统计
	openCount := 0
	for _, r := range results {
		if r.Status == "open" {
			openCount++
		}
	}

	return map[string]any{
		"success":    true,
		"host":       host,
		"ports":      results,
		"total":      len(results),
		"open":       openCount,
		"closed":     len(results) - openCount,
	}, nil
}

// getCommonPorts 获取常用端口列表
func (m *NetworkToolsModule) getCommonPorts() []int {
	return []int{
		21,   // FTP
		22,   // SSH
		23,   // Telnet
		25,   // SMTP
		53,   // DNS
		80,   // HTTP
		110,  // POP3
		143,  // IMAP
		443,  // HTTPS
		445,  // SMB
		993,  // IMAPS
		995,  // POP3S
		3306, // MySQL
		3389, // RDP
		5432, // PostgreSQL
		6379, // Redis
		8080, // HTTP Alt
	}
}

// getDefaultPorts 获取默认端口列表
func (m *NetworkToolsModule) getDefaultPorts() []int {
	ports := make([]int, 1024)
	for i := range ports {
		ports[i] = i + 1
	}
	return ports
}

// getPortService 获取端口对应的服务
func (m *NetworkToolsModule) getPortService(port int) string {
	services := map[int]string{
		21:   "ftp",
		22:   "ssh",
		23:   "telnet",
		25:   "smtp",
		53:   "dns",
		80:   "http",
		110:  "pop3",
		143:  "imap",
		443:  "https",
		445:  "smb",
		993:  "imaps",
		995:  "pop3s",
		3306: "mysql",
		3389: "rdp",
		5432: "postgresql",
		6379: "redis",
		8080: "http-alt",
	}

	if service, ok := services[port]; ok {
		return service
	}
	return "unknown"
}

// getNetworkInterfaces 获取网络接口信息
func (m *NetworkToolsModule) getNetworkInterfaces() (map[string]any, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Failed to get interfaces: %v", err),
		}, nil
	}

	result := make([]map[string]any, 0, len(interfaces))

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var ips []string
		for _, addr := range addrs {
			ips = append(ips, addr.String())
		}

		result = append(result, map[string]any{
			"name":      iface.Name,
			"index":     iface.Index,
			"mtu":       iface.MTU,
			"hardware":   iface.HardwareAddr.String(),
			"flags":     iface.Flags.String(),
			"addresses": ips,
		})
	}

	return map[string]any{
		"success":    true,
		"interfaces": result,
		"count":      len(result),
	}, nil
}

// queryIP 查询 IP 地址信息
func (m *NetworkToolsModule) queryIP(ip string) (map[string]any, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return map[string]any{
			"success": false,
			"error":   "Invalid IP address",
		}, nil
	}

	info := map[string]any{
		"success": true,
		"ip":      ip,
		"version": 4,
	}

	if parsedIP.To4() == nil {
		info["version"] = 6
	}

	// 判断 IP 类型
	info["type"] = m.getIPType(parsedIP)
	info["private"] = m.isPrivateIP(parsedIP)

	// 本地回环检测
	info["loopback"] = parsedIP.IsLoopback()

	return info, nil
}

// getIPType 获取 IP 类型
func (m *NetworkToolsModule) getIPType(ip net.IP) string {
	if ip.IsLoopback() {
		return "loopback"
	}
	if ip.IsPrivate() {
		return "private"
	}
	if ip.IsGlobalUnicast() {
		return "public"
	}
	if ip.IsLinkLocalUnicast() {
		return "link-local"
	}
	return "unknown"
}

// isPrivateIP 判断是否为私有 IP
func (m *NetworkToolsModule) isPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// isURLAllowed 检查 URL 是否允许访问
func (m *NetworkToolsModule) isURLAllowed(url string) bool {
	// 解析 URL 获取主机
	parsedURL, err := regexp.MustCompile(`^[a-zA-Z]+://([^/]+)`).FindStringMatch(url)
	if err != nil || len(parsedURL) < 2 {
		return false
	}

	host := parsedURL[1]

	// 检查阻止列表
	for _, blockedHost := range m.config.BlockedHosts {
		if strings.Contains(strings.ToLower(host), strings.ToLower(blockedHost)) {
			return false
		}
	}

	// 如果允许列表为空，默认允许
	if len(m.config.AllowedHosts) == 0 {
		return true
	}

	// 检查允许列表
	for _, allowedHost := range m.config.AllowedHosts {
		if strings.Contains(strings.ToLower(host), strings.ToLower(allowedHost)) {
			return true
		}
	}

	return false
}
