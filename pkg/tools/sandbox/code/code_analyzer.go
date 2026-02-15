// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// CodeAnalyzer 代码分析器
type CodeAnalyzer struct {
	dangerousPatterns  map[string][]*regexp.Regexp
	securityRules      map[string][]SecurityRule
	networkPatterns    map[string][]*regexp.Regexp
	fileSystemPatterns map[string][]*regexp.Regexp
	processPatterns    map[string][]*regexp.Regexp
	cryptoPatterns     map[string][]*regexp.Regexp
	databasePatterns   map[string][]*regexp.Regexp
	qualityRules       map[string][]QualityRule
	customRules        []CustomRule
}

// CustomRule 自定义规则
type CustomRule struct {
	Name        string
	Language    string
	Pattern     *regexp.Regexp
	Category    string
	Severity    string
	Description string
	Suggestion  string
}

// CustomRulesConfig 自定义规则配置
type CustomRulesConfig struct {
	CustomRules []CustomRuleConfig `yaml:"custom_rules"`
}

// CustomRuleConfig 自定义规则配置项
type CustomRuleConfig struct {
	Name        string `yaml:"name"`
	Language    string `yaml:"language"`
	Pattern     string `yaml:"pattern"`
	Category    string `yaml:"category"`
	Severity    string `yaml:"severity"`
	Description string `yaml:"description"`
	Suggestion  string `yaml:"suggestion"`
}

// SecurityRule 安全规则
type SecurityRule struct {
	Name        string
	Description string
	Pattern     *regexp.Regexp
	Severity    string // "high", "medium", "low"
}

// QualityRule 代码质量规则
type QualityRule struct {
	Name        string
	Description string
	Pattern     *regexp.Regexp
	Category    string // "naming", "style", "performance", "best_practice"
	Severity    string // "info", "low"
	Suggestion  string
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	Safe            bool                  `json:"safe"`
	Issues          []SecurityIssue       `json:"issues"`
	Complexity      int                   `json:"complexity"`
	LineCount       int                   `json:"line_count"`
	CharCount       int                   `json:"char_count"`
	HasDangerousOps bool                  `json:"has_dangerous_ops"`
	Language        string                `json:"language"`
	Suggestions     []string              `json:"suggestions"`
	NetworkOps      []NetworkOperation    `json:"network_ops"`
	FileSystemOps   []FileSystemOperation `json:"filesystem_ops"`
	ProcessOps      []ProcessOperation    `json:"process_ops"`
	CryptoIssues    []CryptoIssue         `json:"crypto_issues"`
	DatabaseOps     []DatabaseOperation   `json:"database_ops"`
	QualityIssues   []QualityIssue        `json:"quality_issues"`
	Score           int                   `json:"score"` // 代码质量评分 0-100
}

// NetworkOperation 网络操作
type NetworkOperation struct {
	Type        string `json:"type"` // "http", "https", "socket", "dns"
	Line        int    `json:"line"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

// FileSystemOperation 文件系统操作
type FileSystemOperation struct {
	Type        string `json:"type"` // "read", "write", "delete", "permission"
	Line        int    `json:"line"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Path        string `json:"path,omitempty"`
}

// ProcessOperation 进程操作
type ProcessOperation struct {
	Type        string `json:"type"` // "create", "kill", "signal", "ipc"
	Line        int    `json:"line"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

// CryptoIssue 加密问题
type CryptoIssue struct {
	Type        string `json:"type"` // "weak_algorithm", "hardcoded_key", "insecure_random"
	Line        int    `json:"line"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// DatabaseOperation 数据库操作
type DatabaseOperation struct {
	Type        string `json:"type"` // "sql_injection", "connection_leak", "unparameterized_query"
	Line        int    `json:"line"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// SecurityIssue 安全问题
type SecurityIssue struct {
	Rule        string `json:"rule"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Line        int    `json:"line"`
	Code        string `json:"code"`
}

// QualityIssue 代码质量问题
type QualityIssue struct {
	Rule        string   `json:"rule"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Severity    string   `json:"severity"`
	Line        int      `json:"line"`
	Column      int      `json:"column"`
	Code        string   `json:"code"`
	Context     []string `json:"context"`
	Suggestion  string   `json:"suggestion"`
}

// NewCodeAnalyzer 创建代码分析器
func NewCodeAnalyzer() *CodeAnalyzer {
	analyzer := &CodeAnalyzer{
		dangerousPatterns:  make(map[string][]*regexp.Regexp),
		securityRules:      make(map[string][]SecurityRule),
		networkPatterns:    make(map[string][]*regexp.Regexp),
		fileSystemPatterns: make(map[string][]*regexp.Regexp),
		processPatterns:    make(map[string][]*regexp.Regexp),
		cryptoPatterns:     make(map[string][]*regexp.Regexp),
		databasePatterns:   make(map[string][]*regexp.Regexp),
		qualityRules:       make(map[string][]QualityRule),
	}

	// 初始化危险模式
	analyzer.initDangerousPatterns()

	// 初始化安全规则
	analyzer.initSecurityRules()

	// 初始化网络操作模式
	analyzer.initNetworkPatterns()

	// 初始化文件系统操作模式
	analyzer.initFileSystemPatterns()

	// 初始化进程操作模式
	analyzer.initProcessPatterns()

	// 初始化加密操作模式
	analyzer.initCryptoPatterns()

	// 初始化数据库操作模式
	analyzer.initDatabasePatterns()

	// 初始化代码质量规则
	analyzer.initQualityRules()

	return analyzer
}

// initDangerousPatterns 初始化危险模式
func (a *CodeAnalyzer) initDangerousPatterns() {
	// Python 危险模式
	a.dangerousPatterns["python"] = []*regexp.Regexp{
		regexp.MustCompile(`\beval\s*\(`),
		regexp.MustCompile(`\bexec\s*\(`),
		regexp.MustCompile(`\b__import__\s*\(`),
		regexp.MustCompile(`\bcompile\s*\(`),
		regexp.MustCompile(`\bos\.system\s*\(`),
		regexp.MustCompile(`\bsubprocess\.(call|run|Popen)`),
		regexp.MustCompile(`\bopen\s*\([^)]*['"]w`),
		regexp.MustCompile(`\brm\s+-rf`),
		regexp.MustCompile(`\bshutil\.rmtree`),
	}

	// JavaScript 危险模式
	a.dangerousPatterns["javascript"] = []*regexp.Regexp{
		regexp.MustCompile(`\beval\s*\(`),
		regexp.MustCompile(`\bFunction\s*\(`),
		regexp.MustCompile(`\brequire\s*\(\s*['"]child_process`),
		regexp.MustCompile(`\bfs\.(unlink|rmdir|rm)`),
		regexp.MustCompile(`\bprocess\.exit`),
		regexp.MustCompile(`\b__dirname`),
		regexp.MustCompile(`\b__filename`),
	}

	// Bash 危险模式
	a.dangerousPatterns["bash"] = []*regexp.Regexp{
		regexp.MustCompile(`\brm\s+-rf`),
		regexp.MustCompile(`\b>\s*/dev/`),
		regexp.MustCompile(`\bdd\s+if=`),
		regexp.MustCompile(`\bmkfs`),
		regexp.MustCompile(`\bshutdown`),
		regexp.MustCompile(`\breboot`),
		regexp.MustCompile(`\bkill\s+-9`),
		regexp.MustCompile(`\bcurl.*\|\s*bash`),
		regexp.MustCompile(`\bwget.*\|\s*sh`),
	}

	// Go 危险模式
	a.dangerousPatterns["go"] = []*regexp.Regexp{
		regexp.MustCompile(`\bos\.Remove`),
		regexp.MustCompile(`\bos\.RemoveAll`),
		regexp.MustCompile(`\bexec\.Command`),
		regexp.MustCompile(`\bunsafe\.Pointer`),
		regexp.MustCompile(`\bsyscall\.`),
	}
}

// initNetworkPatterns 初始化网络操作模式
func (a *CodeAnalyzer) initNetworkPatterns() {
	// Python 网络操作模式
	a.networkPatterns["python"] = []*regexp.Regexp{
		// HTTP/HTTPS
		regexp.MustCompile(`\brequests\.(get|post|put|delete|patch|head|options)`),
		regexp.MustCompile(`\burllib\.(request|urlopen)`),
		regexp.MustCompile(`\bhttp\.client\.(HTTPConnection|HTTPSConnection)`),
		regexp.MustCompile(`\bhttplib\.(HTTPConnection|HTTPSConnection)`),
		// Socket
		regexp.MustCompile(`\bsocket\.(socket|create_connection|create_server)`),
		regexp.MustCompile(`\bsocket\.AF_INET`),
		regexp.MustCompile(`\bsocket\.SOCK_STREAM`),
		// DNS
		regexp.MustCompile(`\bsocket\.gethostbyname`),
		regexp.MustCompile(`\bsocket\.getaddrinfo`),
		regexp.MustCompile(`\bdns\.resolver`),
	}

	// JavaScript 网络操作模式
	a.networkPatterns["javascript"] = []*regexp.Regexp{
		// HTTP/HTTPS
		regexp.MustCompile(`\bfetch\s*\(`),
		regexp.MustCompile(`\baxios\.(get|post|put|delete|patch)`),
		regexp.MustCompile(`\bhttp\.(get|request)`),
		regexp.MustCompile(`\bhttps\.(get|request)`),
		regexp.MustCompile(`\bXMLHttpRequest`),
		// Socket
		regexp.MustCompile(`\bnet\.(createConnection|createServer|Socket)`),
		regexp.MustCompile(`\bWebSocket`),
		// DNS
		regexp.MustCompile(`\bdns\.(lookup|resolve)`),
	}

	// Go 网络操作模式
	a.networkPatterns["go"] = []*regexp.Regexp{
		// HTTP/HTTPS
		regexp.MustCompile(`\bhttp\.(Get|Post|Head|Do)`),
		regexp.MustCompile(`\bhttp\.Client`),
		regexp.MustCompile(`\bhttp\.NewRequest`),
		// Socket
		regexp.MustCompile(`\bnet\.(Dial|Listen|DialTCP|ListenTCP)`),
		regexp.MustCompile(`\bnet\.Conn`),
		// DNS
		regexp.MustCompile(`\bnet\.LookupHost`),
		regexp.MustCompile(`\bnet\.LookupIP`),
		regexp.MustCompile(`\bnet\.LookupAddr`),
	}

	// Bash 网络操作模式
	a.networkPatterns["bash"] = []*regexp.Regexp{
		regexp.MustCompile(`\bcurl\s+`),
		regexp.MustCompile(`\bwget\s+`),
		regexp.MustCompile(`\bnc\s+`),
		regexp.MustCompile(`\bnetcat\s+`),
		regexp.MustCompile(`\btelnet\s+`),
		regexp.MustCompile(`\bssh\s+`),
		regexp.MustCompile(`\bscp\s+`),
		regexp.MustCompile(`\bftp\s+`),
	}
}

// initFileSystemPatterns 初始化文件系统操作模式
func (a *CodeAnalyzer) initFileSystemPatterns() {
	// Python 文件系统操作模式
	a.fileSystemPatterns["python"] = []*regexp.Regexp{
		// 敏感目录访问
		regexp.MustCompile(`['"](/etc|/sys|/proc|/dev|/root|C:\\Windows|C:\\Program Files)`),
		// 文件操作
		regexp.MustCompile(`\bos\.(remove|unlink|rmdir|removedirs)`),
		regexp.MustCompile(`\bshutil\.(rmtree|move|copy)`),
		regexp.MustCompile(`\bopen\s*\([^)]*['"][wa]`),
		// 权限修改
		regexp.MustCompile(`\bos\.(chmod|chown)`),
		regexp.MustCompile(`\bos\.access`),
		// 符号链接
		regexp.MustCompile(`\bos\.(symlink|link)`),
	}

	// JavaScript 文件系统操作模式
	a.fileSystemPatterns["javascript"] = []*regexp.Regexp{
		// 敏感目录访问
		regexp.MustCompile(`['"](/etc|/sys|/proc|/dev|/root|C:\\\\Windows|C:\\\\Program Files)`),
		// 文件操作
		regexp.MustCompile(`\bfs\.(unlink|rm|rmdir|rmSync|rmdirSync)`),
		regexp.MustCompile(`\bfs\.(writeFile|writeFileSync|appendFile)`),
		// 权限修改
		regexp.MustCompile(`\bfs\.(chmod|chown)`),
		// 符号链接
		regexp.MustCompile(`\bfs\.(symlink|link)`),
	}

	// Go 文件系统操作模式
	a.fileSystemPatterns["go"] = []*regexp.Regexp{
		// 敏感目录访问
		regexp.MustCompile(`['"](/etc|/sys|/proc|/dev|/root|C:\\Windows|C:\\Program Files)`),
		// 文件操作
		regexp.MustCompile(`\bos\.(Remove|RemoveAll)`),
		regexp.MustCompile(`\bos\.(Create|OpenFile)`),
		regexp.MustCompile(`\bioutil\.WriteFile`),
		// 权限修改
		regexp.MustCompile(`\bos\.(Chmod|Chown)`),
		// 符号链接
		regexp.MustCompile(`\bos\.(Symlink|Link)`),
	}

	// Bash 文件系统操作模式
	a.fileSystemPatterns["bash"] = []*regexp.Regexp{
		// 敏感目录访问
		regexp.MustCompile(`(/etc|/sys|/proc|/dev|/root|C:\\Windows|C:\\Program Files)`),
		// 文件操作
		regexp.MustCompile(`\brm\s+`),
		regexp.MustCompile(`\brmdir\s+`),
		regexp.MustCompile(`\bmv\s+`),
		regexp.MustCompile(`\bcp\s+`),
		// 权限修改
		regexp.MustCompile(`\bchmod\s+`),
		regexp.MustCompile(`\bchown\s+`),
		// 符号链接
		regexp.MustCompile(`\bln\s+-s`),
	}
}

// initProcessPatterns 初始化进程操作模式
func (a *CodeAnalyzer) initProcessPatterns() {
	// Python 进程操作模式
	a.processPatterns["python"] = []*regexp.Regexp{
		// 进程创建
		regexp.MustCompile(`\bsubprocess\.(Popen|call|run|check_output)`),
		regexp.MustCompile(`\bos\.(system|spawn|exec)`),
		regexp.MustCompile(`\bmultiprocessing\.Process`),
		// 进程终止
		regexp.MustCompile(`\bos\.kill`),
		regexp.MustCompile(`\bsignal\.SIGKILL`),
		// 进程间通信
		regexp.MustCompile(`\bmultiprocessing\.(Queue|Pipe)`),
	}

	// JavaScript 进程操作模式
	a.processPatterns["javascript"] = []*regexp.Regexp{
		// 进程创建
		regexp.MustCompile(`\bchild_process\.(spawn|exec|execFile|fork)`),
		regexp.MustCompile(`\bspawn\s*\(`),
		regexp.MustCompile(`\bexec\s*\(`),
		regexp.MustCompile(`\bprocess\.exit`),
		// 进程终止
		regexp.MustCompile(`\bprocess\.kill`),
		// 进程间通信
		regexp.MustCompile(`\bprocess\.(send|on\(['"]message)`),
	}

	// Go 进程操作模式
	a.processPatterns["go"] = []*regexp.Regexp{
		// 进程创建
		regexp.MustCompile(`\bexec\.(Command|CommandContext)`),
		regexp.MustCompile(`\bos\.StartProcess`),
		// 进程终止
		regexp.MustCompile(`\bos\.Process\.Kill`),
		regexp.MustCompile(`\bsyscall\.Kill`),
		// 进程间通信
		regexp.MustCompile(`\bos\.Pipe`),
	}

	// Bash 进程操作模式
	a.processPatterns["bash"] = []*regexp.Regexp{
		// 进程创建
		regexp.MustCompile(`\b&\s*$`), // 后台进程
		regexp.MustCompile(`\bexec\s+`),
		// 进程终止
		regexp.MustCompile(`\bkill\s+`),
		regexp.MustCompile(`\bkillall\s+`),
		regexp.MustCompile(`\bpkill\s+`),
	}
}

// initCryptoPatterns 初始化加密操作模式
func (a *CodeAnalyzer) initCryptoPatterns() {
	// Python 加密操作模式
	a.cryptoPatterns["python"] = []*regexp.Regexp{
		// 弱加密算法
		regexp.MustCompile(`\bhashlib\.(md5|sha1)\(`),
		regexp.MustCompile(`\bCrypto\.Cipher\.(DES|RC4)`),
		// 硬编码密钥
		regexp.MustCompile(`(password|passwd|pwd|secret|key)\s*=\s*['"][^'"]+['"]`),
		// 不安全的随机数
		regexp.MustCompile(`\brandom\.(random|randint|choice)`),
	}

	// JavaScript 加密操作模式
	a.cryptoPatterns["javascript"] = []*regexp.Regexp{
		// 弱加密算法
		regexp.MustCompile(`\bcrypto\.createHash\(['"]md5['"]`),
		regexp.MustCompile(`\bcrypto\.createHash\(['"]sha1['"]`),
		// 硬编码密钥
		regexp.MustCompile(`(password|passwd|pwd|secret|key)\s*[:=]\s*['"][^'"]+['"]`),
		// 不安全的随机数
		regexp.MustCompile(`\bMath\.random\(`),
	}

	// Go 加密操作模式
	a.cryptoPatterns["go"] = []*regexp.Regexp{
		// 弱加密算法
		regexp.MustCompile(`\bmd5\.(New|Sum)`),
		regexp.MustCompile(`\bsha1\.(New|Sum)`),
		regexp.MustCompile(`\bdes\.(NewCipher|NewTripleDESCipher)`),
		// 硬编码密钥
		regexp.MustCompile(`(password|passwd|pwd|secret|key)\s*:?=\s*"[^"]+"`),
		// 不安全的随机数
		regexp.MustCompile(`\bmath/rand\.(Int|Intn|Float)`),
	}

	// Bash 加密操作模式
	a.cryptoPatterns["bash"] = []*regexp.Regexp{
		// 弱加密算法
		regexp.MustCompile(`\bmd5sum\s+`),
		regexp.MustCompile(`\bsha1sum\s+`),
		// 硬编码密钥
		regexp.MustCompile(`(PASSWORD|PASSWD|PWD|SECRET|KEY)=['"][^'"]+['"]`),
	}
}

// initDatabasePatterns 初始化数据库操作模式
func (a *CodeAnalyzer) initDatabasePatterns() {
	// Python 数据库操作模式
	a.databasePatterns["python"] = []*regexp.Regexp{
		// SQL 注入风险
		regexp.MustCompile(`execute\s*\([^)]*%s`),
		regexp.MustCompile(`execute\s*\([^)]*\+`),
		regexp.MustCompile(`execute\s*\([^)]*f['"]`),
		regexp.MustCompile(`execute\s*\([^)]*\.format\(`),
		// 未参数化查询
		regexp.MustCompile(`cursor\.execute\s*\(\s*['"].*%.*['"]`),
		// 数据库连接
		regexp.MustCompile(`\b(mysql|psycopg2|sqlite3|pymongo)\.connect`),
	}

	// JavaScript 数据库操作模式
	a.databasePatterns["javascript"] = []*regexp.Regexp{
		// SQL 注入风险
		regexp.MustCompile(`query\s*\([^)]*\+`),
		regexp.MustCompile("query\\s*\\(\\s*`.*\\$\\{"),
		regexp.MustCompile(`execute\s*\([^)]*\+`),
		// 未参数化查询
		regexp.MustCompile(`\.query\s*\(\s*['"].*\+`),
		// 数据库连接
		regexp.MustCompile(`\b(mysql|pg|mongodb|sqlite3)\.connect`),
		regexp.MustCompile(`\bcreateConnection\s*\(`),
	}

	// Go 数据库操作模式
	a.databasePatterns["go"] = []*regexp.Regexp{
		// SQL 注入风险
		regexp.MustCompile(`\bExec\s*\([^)]*\+`),
		regexp.MustCompile(`\bQuery\s*\([^)]*\+`),
		regexp.MustCompile(`fmt\.Sprintf\s*\([^)]*SELECT`),
		regexp.MustCompile(`db\.(Exec|Query)\s*\([^)]*\+`),
		// 未参数化查询
		regexp.MustCompile(`db\.(Exec|Query)\s*\(\s*".*%`),
		// 数据库连接
		regexp.MustCompile(`sql\.Open\s*\(`),
		regexp.MustCompile(`\bmongo\.Connect`),
	}

	// Bash 数据库操作模式
	a.databasePatterns["bash"] = []*regexp.Regexp{
		// SQL 注入风险
		regexp.MustCompile(`mysql.*-e.*\$`),
		regexp.MustCompile(`psql.*-c.*\$`),
		regexp.MustCompile(`sqlite3.*".*\$`),
	}
}

// initSecurityRules 初始化安全规则
func (a *CodeAnalyzer) initSecurityRules() {
	// Python 安全规则
	a.securityRules["python"] = []SecurityRule{
		{
			Name:        "eval_usage",
			Description: "使用 eval() 可能导致代码注入",
			Pattern:     regexp.MustCompile(`\beval\s*\(`),
			Severity:    "high",
		},
		{
			Name:        "exec_usage",
			Description: "使用 exec() 可能导致代码注入",
			Pattern:     regexp.MustCompile(`\bexec\s*\(`),
			Severity:    "high",
		},
		{
			Name:        "os_system",
			Description: "使用 os.system() 可能导致命令注入",
			Pattern:     regexp.MustCompile(`\bos\.system\s*\(`),
			Severity:    "high",
		},
		{
			Name:        "file_write",
			Description: "文件写入操作需要谨慎处理",
			Pattern:     regexp.MustCompile(`\bopen\s*\([^)]*['"]w`),
			Severity:    "medium",
		},
	}

	// JavaScript 安全规则
	a.securityRules["javascript"] = []SecurityRule{
		{
			Name:        "eval_usage",
			Description: "使用 eval() 可能导致代码注入",
			Pattern:     regexp.MustCompile(`\beval\s*\(`),
			Severity:    "high",
		},
		{
			Name:        "function_constructor",
			Description: "使用 Function 构造函数可能导致代码注入",
			Pattern:     regexp.MustCompile(`\bFunction\s*\(`),
			Severity:    "high",
		},
		{
			Name:        "child_process",
			Description: "使用 child_process 可能导致命令注入",
			Pattern:     regexp.MustCompile(`\brequire\s*\(\s*['"]child_process`),
			Severity:    "high",
		},
	}

	// Bash 安全规则
	a.securityRules["bash"] = []SecurityRule{
		{
			Name:        "rm_rf",
			Description: "使用 rm -rf 可能导致数据丢失",
			Pattern:     regexp.MustCompile(`\brm\s+-rf`),
			Severity:    "high",
		},
		{
			Name:        "pipe_to_shell",
			Description: "管道到 shell 可能导致代码注入",
			Pattern:     regexp.MustCompile(`\bcurl.*\|\s*bash|\bwget.*\|\s*sh`),
			Severity:    "high",
		},
	}

	// Go 安全规则
	a.securityRules["go"] = []SecurityRule{
		{
			Name:        "unsafe_pointer",
			Description: "使用 unsafe.Pointer 可能导致内存安全问题",
			Pattern:     regexp.MustCompile(`\bunsafe\.Pointer`),
			Severity:    "medium",
		},
		{
			Name:        "exec_command",
			Description: "使用 exec.Command 需要谨慎处理输入",
			Pattern:     regexp.MustCompile(`\bexec\.Command`),
			Severity:    "medium",
		},
	}
}

// Analyze 分析代码
func (a *CodeAnalyzer) Analyze(language, code string) *AnalysisResult {
	language = strings.ToLower(language)

	// 预先分割行，避免重复分割
	lines := strings.Split(code, "\n")
	lineCount := len(lines)

	// 对于大文件使用并行分析
	if lineCount > 100 {
		return a.analyzeParallel(language, code, lines)
	}

	// 对于小文件使用串行分析
	return a.analyzeSerial(language, code, lines)
}

// analyzeSerial 串行分析（小文件）
func (a *CodeAnalyzer) analyzeSerial(language, code string, lines []string) *AnalysisResult {
	result := &AnalysisResult{
		Safe:          true,
		Issues:        []SecurityIssue{},
		Language:      language,
		LineCount:     len(lines),
		CharCount:     len(code),
		NetworkOps:    []NetworkOperation{},
		FileSystemOps: []FileSystemOperation{},
		ProcessOps:    []ProcessOperation{},
		CryptoIssues:  []CryptoIssue{},
		DatabaseOps:   []DatabaseOperation{},
		QualityIssues: []QualityIssue{},
	}

	// 检查危险模式
	if patterns, exists := a.dangerousPatterns[language]; exists {
		for _, pattern := range patterns {
			if pattern.MatchString(code) {
				result.HasDangerousOps = true
				result.Safe = false
			}
		}
	}

	// 一次遍历检测所有模式
	a.detectAllPatterns(language, lines, result)

	// 计算复杂度和评分
	result.Complexity = a.calculateComplexity(code)
	result.Score = a.calculateQualityScore(result)
	result.Suggestions = a.generateSuggestions(result)

	return result
}

// analyzeParallel 并行分析（大文件）
func (a *CodeAnalyzer) analyzeParallel(language, code string, lines []string) *AnalysisResult {
	var wg sync.WaitGroup
	resultChan := make(chan *AnalysisResult, 6)

	// 并行执行各种检测
	wg.Add(6)

	go func() {
		defer wg.Done()
		result := &AnalysisResult{
			NetworkOps: []NetworkOperation{},
		}
		a.detectNetworkOpsOptimized(language, lines, result)
		resultChan <- result
	}()

	go func() {
		defer wg.Done()
		result := &AnalysisResult{
			FileSystemOps: []FileSystemOperation{},
		}
		a.detectFileSystemOpsOptimized(language, lines, result)
		resultChan <- result
	}()

	go func() {
		defer wg.Done()
		result := &AnalysisResult{
			ProcessOps: []ProcessOperation{},
		}
		a.detectProcessOpsOptimized(language, lines, result)
		resultChan <- result
	}()

	go func() {
		defer wg.Done()
		result := &AnalysisResult{
			CryptoIssues: []CryptoIssue{},
		}
		a.detectCryptoIssuesOptimized(language, lines, result)
		resultChan <- result
	}()

	go func() {
		defer wg.Done()
		result := &AnalysisResult{
			DatabaseOps: []DatabaseOperation{},
		}
		a.detectDatabaseOpsOptimized(language, lines, result)
		resultChan <- result
	}()

	go func() {
		defer wg.Done()
		result := &AnalysisResult{
			QualityIssues: []QualityIssue{},
		}
		a.checkQualityIssuesOptimized(language, lines, result)
		resultChan <- result
	}()

	// 等待所有检测完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 合并结果
	finalResult := &AnalysisResult{
		Safe:          true,
		Language:      language,
		LineCount:     len(lines),
		CharCount:     len(code),
		Issues:        []SecurityIssue{},
		NetworkOps:    []NetworkOperation{},
		FileSystemOps: []FileSystemOperation{},
		ProcessOps:    []ProcessOperation{},
		CryptoIssues:  []CryptoIssue{},
		DatabaseOps:   []DatabaseOperation{},
		QualityIssues: []QualityIssue{},
	}

	for result := range resultChan {
		finalResult.NetworkOps = append(finalResult.NetworkOps, result.NetworkOps...)
		finalResult.FileSystemOps = append(finalResult.FileSystemOps, result.FileSystemOps...)
		finalResult.ProcessOps = append(finalResult.ProcessOps, result.ProcessOps...)
		finalResult.CryptoIssues = append(finalResult.CryptoIssues, result.CryptoIssues...)
		finalResult.DatabaseOps = append(finalResult.DatabaseOps, result.DatabaseOps...)
		finalResult.QualityIssues = append(finalResult.QualityIssues, result.QualityIssues...)

		if !result.Safe {
			finalResult.Safe = false
		}
	}

	// 检查危险模式
	if patterns, exists := a.dangerousPatterns[language]; exists {
		for _, pattern := range patterns {
			if pattern.MatchString(code) {
				finalResult.HasDangerousOps = true
				finalResult.Safe = false
			}
		}
	}

	// 计算复杂度和评分
	finalResult.Complexity = a.calculateComplexity(code)
	finalResult.Score = a.calculateQualityScore(finalResult)
	finalResult.Suggestions = a.generateSuggestions(finalResult)

	return finalResult
}

// detectAllPatterns 一次遍历检测所有模式（优化版本）
func (a *CodeAnalyzer) detectAllPatterns(language string, lines []string, result *AnalysisResult) {
	// 获取所有模式
	networkPatterns := a.networkPatterns[language]
	fileSystemPatterns := a.fileSystemPatterns[language]
	processPatterns := a.processPatterns[language]
	cryptoPatterns := a.cryptoPatterns[language]
	databasePatterns := a.databasePatterns[language]

	// 一次遍历所有行
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		// 检测网络操作
		for _, pattern := range networkPatterns {
			if pattern.MatchString(line) {
				opType := a.detectNetworkOpType(line)
				result.NetworkOps = append(result.NetworkOps, NetworkOperation{
					Type:        opType,
					Line:        i + 1,
					Code:        trimmedLine,
					Description: fmt.Sprintf("Network operation detected: %s", opType),
				})
				break // 找到一个匹配就跳出
			}
		}

		// 检测文件系统操作
		for _, pattern := range fileSystemPatterns {
			if pattern.MatchString(line) {
				opType := a.detectFileSystemOpType(line)
				result.FileSystemOps = append(result.FileSystemOps, FileSystemOperation{
					Type:        opType,
					Line:        i + 1,
					Code:        trimmedLine,
					Description: fmt.Sprintf("File system operation detected: %s", opType),
				})
				break
			}
		}

		// 检测进程操作
		for _, pattern := range processPatterns {
			if pattern.MatchString(line) {
				opType := a.detectProcessOpType(line)
				result.ProcessOps = append(result.ProcessOps, ProcessOperation{
					Type:        opType,
					Line:        i + 1,
					Code:        trimmedLine,
					Description: fmt.Sprintf("Process operation detected: %s", opType),
				})
				break
			}
		}

		// 检测加密问题
		for _, pattern := range cryptoPatterns {
			if pattern.MatchString(line) {
				issue := a.detectCryptoIssue(line)
				result.CryptoIssues = append(result.CryptoIssues, CryptoIssue{
					Type:        issue.Type,
					Line:        i + 1,
					Code:        trimmedLine,
					Description: issue.Description,
					Severity:    issue.Severity,
				})
				if issue.Severity == "high" {
					result.Safe = false
				}
				break
			}
		}

		// 检测数据库操作
		for _, pattern := range databasePatterns {
			if pattern.MatchString(line) {
				issue := a.detectDatabaseIssue(line)
				result.DatabaseOps = append(result.DatabaseOps, DatabaseOperation{
					Type:        issue.Type,
					Line:        i + 1,
					Code:        trimmedLine,
					Description: issue.Description,
					Severity:    issue.Severity,
				})
				if issue.Severity == "high" {
					result.Safe = false
				}
				break
			}
		}
	}

	// 应用安全规则
	if rules, exists := a.securityRules[language]; exists {
		for _, rule := range rules {
			for i, line := range lines {
				if rule.Pattern.MatchString(line) {
					result.Issues = append(result.Issues, SecurityIssue{
						Rule:        rule.Name,
						Description: rule.Description,
						Severity:    rule.Severity,
						Line:        i + 1,
						Code:        strings.TrimSpace(line),
					})
					if rule.Severity == "high" {
						result.Safe = false
					}
				}
			}
		}
	}

	// 检查代码质量
	a.checkQualityIssues(language, lines, result)
}

// detectNetworkOpsOptimized 优化的网络操作检测
func (a *CodeAnalyzer) detectNetworkOpsOptimized(language string, lines []string, result *AnalysisResult) {
	patterns := a.networkPatterns[language]
	if patterns == nil {
		return
	}

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		for _, pattern := range patterns {
			if pattern.MatchString(line) {
				opType := a.detectNetworkOpType(line)
				result.NetworkOps = append(result.NetworkOps, NetworkOperation{
					Type:        opType,
					Line:        i + 1,
					Code:        trimmedLine,
					Description: fmt.Sprintf("Network operation detected: %s", opType),
				})
				break
			}
		}
	}
}

// detectFileSystemOpsOptimized 优化的文件系统操作检测
func (a *CodeAnalyzer) detectFileSystemOpsOptimized(language string, lines []string, result *AnalysisResult) {
	patterns := a.fileSystemPatterns[language]
	if patterns == nil {
		return
	}

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		for _, pattern := range patterns {
			if pattern.MatchString(line) {
				opType := a.detectFileSystemOpType(line)
				result.FileSystemOps = append(result.FileSystemOps, FileSystemOperation{
					Type:        opType,
					Line:        i + 1,
					Code:        trimmedLine,
					Description: fmt.Sprintf("File system operation detected: %s", opType),
				})
				break
			}
		}
	}
}

// detectProcessOpsOptimized 优化的进程操作检测
func (a *CodeAnalyzer) detectProcessOpsOptimized(language string, lines []string, result *AnalysisResult) {
	patterns := a.processPatterns[language]
	if patterns == nil {
		return
	}

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		for _, pattern := range patterns {
			if pattern.MatchString(line) {
				opType := a.detectProcessOpType(line)
				result.ProcessOps = append(result.ProcessOps, ProcessOperation{
					Type:        opType,
					Line:        i + 1,
					Code:        trimmedLine,
					Description: fmt.Sprintf("Process operation detected: %s", opType),
				})
				break
			}
		}
	}
}

// detectCryptoIssuesOptimized 优化的加密问题检测
func (a *CodeAnalyzer) detectCryptoIssuesOptimized(language string, lines []string, result *AnalysisResult) {
	patterns := a.cryptoPatterns[language]
	if patterns == nil {
		return
	}

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		for _, pattern := range patterns {
			if pattern.MatchString(line) {
				issue := a.detectCryptoIssue(line)
				result.CryptoIssues = append(result.CryptoIssues, CryptoIssue{
					Type:        issue.Type,
					Line:        i + 1,
					Code:        trimmedLine,
					Description: issue.Description,
					Severity:    issue.Severity,
				})
				if issue.Severity == "high" {
					result.Safe = false
				}
				break
			}
		}
	}
}

// detectDatabaseOpsOptimized 优化的数据库操作检测
func (a *CodeAnalyzer) detectDatabaseOpsOptimized(language string, lines []string, result *AnalysisResult) {
	patterns := a.databasePatterns[language]
	if patterns == nil {
		return
	}

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		for _, pattern := range patterns {
			if pattern.MatchString(line) {
				issue := a.detectDatabaseIssue(line)
				result.DatabaseOps = append(result.DatabaseOps, DatabaseOperation{
					Type:        issue.Type,
					Line:        i + 1,
					Code:        trimmedLine,
					Description: issue.Description,
					Severity:    issue.Severity,
				})
				if issue.Severity == "high" {
					result.Safe = false
				}
				break
			}
		}
	}
}

// checkQualityIssuesOptimized 优化的代码质量检查
func (a *CodeAnalyzer) checkQualityIssuesOptimized(language string, lines []string, result *AnalysisResult) {
	a.checkQualityIssues(language, lines, result)
}

// detectAllPatterns helper - applies security rules
func (a *CodeAnalyzer) applySecurityRules(language string, lines []string, result *AnalysisResult) {
	// 应用安全规则
	if rules, exists := a.securityRules[language]; exists {
		for _, rule := range rules {
			for i, line := range lines {
				if rule.Pattern.MatchString(line) {
					result.Issues = append(result.Issues, SecurityIssue{
						Rule:        rule.Name,
						Description: rule.Description,
						Severity:    rule.Severity,
						Line:        i + 1,
						Code:        strings.TrimSpace(line),
					})
					if rule.Severity == "high" {
						result.Safe = false
					}
				}
			}
		}
	}
}

// calculateComplexity 计算代码复杂度
func (a *CodeAnalyzer) calculateComplexity(code string) int {
	complexity := 1 // 基础复杂度

	// 计算控制流语句
	controlFlowPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\bif\b`),
		regexp.MustCompile(`\belse\b`),
		regexp.MustCompile(`\belif\b`),
		regexp.MustCompile(`\bfor\b`),
		regexp.MustCompile(`\bwhile\b`),
		regexp.MustCompile(`\bswitch\b`),
		regexp.MustCompile(`\bcase\b`),
		regexp.MustCompile(`\btry\b`),
		regexp.MustCompile(`\bcatch\b`),
		regexp.MustCompile(`\b&&\b`),
		regexp.MustCompile(`\b\|\|\b`),
	}

	for _, pattern := range controlFlowPatterns {
		matches := pattern.FindAllString(code, -1)
		complexity += len(matches)
	}

	return complexity
}

// generateSuggestions 生成建议
func (a *CodeAnalyzer) generateSuggestions(result *AnalysisResult) []string {
	suggestions := []string{}

	if result.HasDangerousOps {
		suggestions = append(suggestions, "代码包含潜在危险操作，建议仔细审查")
	}

	if result.Complexity > 20 {
		suggestions = append(suggestions, "代码复杂度较高，建议简化逻辑")
	}

	if result.LineCount > 100 {
		suggestions = append(suggestions, "代码行数较多，建议拆分为多个函数")
	}

	highSeverityCount := 0
	for _, issue := range result.Issues {
		if issue.Severity == "high" {
			highSeverityCount++
		}
	}

	if highSeverityCount > 0 {
		suggestions = append(suggestions, fmt.Sprintf("发现 %d 个高危安全问题，强烈建议修复", highSeverityCount))
	}

	if len(result.Issues) == 0 && !result.HasDangerousOps {
		suggestions = append(suggestions, "代码看起来安全，未发现明显问题")
	}

	return suggestions
}

// IsSafe 检查代码是否安全
func (a *CodeAnalyzer) IsSafe(language, code string) bool {
	result := a.Analyze(language, code)
	return result.Safe
}

// GetIssues 获取代码问题
func (a *CodeAnalyzer) GetIssues(language, code string) []SecurityIssue {
	result := a.Analyze(language, code)
	return result.Issues
}

// ValidateCode 验证代码安全性
func (a *CodeAnalyzer) ValidateCode(language, code string) error {
	result := a.Analyze(language, code)

	if !result.Safe {
		return fmt.Errorf("code contains security issues: %d issues found", len(result.Issues))
	}

	return nil
}

// detectNetworkOpType 检测网络操作类型
func (a *CodeAnalyzer) detectNetworkOpType(line string) string {
	line = strings.ToLower(line)
	if strings.Contains(line, "http") || strings.Contains(line, "fetch") || strings.Contains(line, "axios") || strings.Contains(line, "requests") || strings.Contains(line, "curl") || strings.Contains(line, "wget") {
		if strings.Contains(line, "https") {
			return "https"
		}
		return "http"
	}
	if strings.Contains(line, "socket") || strings.Contains(line, "websocket") || strings.Contains(line, "nc") || strings.Contains(line, "netcat") {
		return "socket"
	}
	if strings.Contains(line, "dns") || strings.Contains(line, "gethostbyname") || strings.Contains(line, "lookup") {
		return "dns"
	}
	return "network"
}

// detectFileSystemOpType 检测文件系统操作类型
func (a *CodeAnalyzer) detectFileSystemOpType(line string) string {
	line = strings.ToLower(line)
	if strings.Contains(line, "remove") || strings.Contains(line, "unlink") || strings.Contains(line, "rmdir") || strings.Contains(line, "rm ") {
		return "delete"
	}
	if strings.Contains(line, "write") || strings.Contains(line, "create") || strings.Contains(line, "'w'") || strings.Contains(line, "\"w\"") {
		return "write"
	}
	if strings.Contains(line, "chmod") || strings.Contains(line, "chown") {
		return "permission"
	}
	if strings.Contains(line, "symlink") || strings.Contains(line, "link") {
		return "symlink"
	}
	if strings.Contains(line, "read") || strings.Contains(line, "open") {
		return "read"
	}
	return "filesystem"
}

// detectProcessOpType 检测进程操作类型
func (a *CodeAnalyzer) detectProcessOpType(line string) string {
	line = strings.ToLower(line)
	if strings.Contains(line, "kill") || strings.Contains(line, "sigkill") {
		return "kill"
	}
	if strings.Contains(line, "signal") || strings.Contains(line, "sig") {
		return "signal"
	}
	if strings.Contains(line, "pipe") || strings.Contains(line, "queue") || strings.Contains(line, "send") {
		return "ipc"
	}
	if strings.Contains(line, "spawn") || strings.Contains(line, "exec") || strings.Contains(line, "popen") || strings.Contains(line, "command") || strings.Contains(line, "fork") {
		return "create"
	}
	return "process"
}

// detectCryptoIssue 检测加密问题
func (a *CodeAnalyzer) detectCryptoIssue(line string) CryptoIssue {
	line = strings.ToLower(line)

	if strings.Contains(line, "md5") || strings.Contains(line, "sha1") || strings.Contains(line, "des") || strings.Contains(line, "rc4") {
		return CryptoIssue{
			Type:        "weak_algorithm",
			Description: "Weak cryptographic algorithm detected (MD5, SHA1, DES, or RC4)",
			Severity:    "high",
		}
	}

	if regexp.MustCompile(`(password|passwd|pwd|secret|key)\s*[:=]\s*['"][^'"]+['"]`).MatchString(line) {
		return CryptoIssue{
			Type:        "hardcoded_key",
			Description: "Hardcoded password or secret key detected",
			Severity:    "high",
		}
	}

	if strings.Contains(line, "math.random") || (strings.Contains(line, "random") && !strings.Contains(line, "secrets") && !strings.Contains(line, "crypto")) {
		return CryptoIssue{
			Type:        "insecure_random",
			Description: "Insecure random number generator detected",
			Severity:    "medium",
		}
	}

	return CryptoIssue{
		Type:        "crypto",
		Description: "Cryptographic operation detected",
		Severity:    "low",
	}
}

// detectDatabaseIssue 检测数据库问题
func (a *CodeAnalyzer) detectDatabaseIssue(line string) DatabaseOperation {
	lineLower := strings.ToLower(line)

	// SQL 注入风险检测 - 字符串拼接
	if regexp.MustCompile(`(execute|query|exec)\s*\([^)]*(\+|%s|\.format|f['"]|\$\{)`).MatchString(lineLower) {
		return DatabaseOperation{
			Type:        "sql_injection",
			Description: "Potential SQL injection vulnerability detected",
			Severity:    "high",
		}
	}

	// 未参数化查询检测
	if regexp.MustCompile(`(execute|query|exec)\s*\(\s*['"].*(%|SELECT|INSERT|UPDATE|DELETE)`).MatchString(lineLower) {
		return DatabaseOperation{
			Type:        "unparameterized_query",
			Description: "Unparameterized database query detected",
			Severity:    "high",
		}
	}

	// 数据库连接检测
	if strings.Contains(lineLower, "connect") || strings.Contains(lineLower, "open") {
		return DatabaseOperation{
			Type:        "connection_leak",
			Description: "Database connection detected - ensure proper closure",
			Severity:    "medium",
		}
	}

	return DatabaseOperation{
		Type:        "database",
		Description: "Database operation detected",
		Severity:    "low",
	}
}

// detectDatabaseOpType 检测数据库操作类型
func (a *CodeAnalyzer) detectDatabaseOpType(line string) string {
	issue := a.detectDatabaseIssue(line)
	return issue.Type
}

// initQualityRules 初始化代码质量规则
func (a *CodeAnalyzer) initQualityRules() {
	// Python 命名规范
	a.qualityRules["python"] = []QualityRule{
		// 变量命名 - snake_case
		{
			Name:        "python_var_naming",
			Description: "Python 变量应使用 snake_case 命名",
			Pattern:     regexp.MustCompile(`\b[a-z]+[A-Z][a-zA-Z]*\s*=`),
			Category:    "naming",
			Severity:    "info",
			Suggestion:  "使用 snake_case 命名变量，例如: my_variable",
		},
		// 函数命名 - snake_case
		{
			Name:        "python_func_naming",
			Description: "Python 函数应使用 snake_case 命名",
			Pattern:     regexp.MustCompile(`\bdef\s+[a-z]+[A-Z][a-zA-Z]*\s*\(`),
			Category:    "naming",
			Severity:    "info",
			Suggestion:  "使用 snake_case 命名函数，例如: my_function",
		},
		// 类命名 - PascalCase
		{
			Name:        "python_class_naming",
			Description: "Python 类应使用 PascalCase 命名",
			Pattern:     regexp.MustCompile(`\bclass\s+[a-z][a-zA-Z]*\s*[:(]`),
			Category:    "naming",
			Severity:    "info",
			Suggestion:  "使用 PascalCase 命名类，例如: MyClass",
		},
		// 最佳实践 - 异常处理
		{
			Name:        "python_bare_except",
			Description: "避免使用裸 except 语句",
			Pattern:     regexp.MustCompile(`\bexcept\s*:`),
			Category:    "best_practice",
			Severity:    "low",
			Suggestion:  "指定具体的异常类型，例如: except ValueError:",
		},
		// 最佳实践 - pass 语句
		{
			Name:        "python_empty_except",
			Description: "避免空的 except 块",
			Pattern:     regexp.MustCompile(`except[^:]*:\s+pass`),
			Category:    "best_practice",
			Severity:    "low",
			Suggestion:  "至少记录异常或重新抛出",
		},
		// 性能 - 字符串拼接
		{
			Name:        "python_string_concat",
			Description: "循环中避免使用 + 拼接字符串",
			Pattern:     regexp.MustCompile(`\+=\s*['"]`),
			Category:    "performance",
			Severity:    "info",
			Suggestion:  "使用 join() 或列表推导式",
		},
	}

	// JavaScript 命名规范
	a.qualityRules["javascript"] = []QualityRule{
		// 变量命名 - camelCase
		{
			Name:        "js_var_naming",
			Description: "JavaScript 变量应使用 camelCase 命名",
			Pattern:     regexp.MustCompile(`\b(var|let|const)\s+[A-Z][a-zA-Z]*\s*=`),
			Category:    "naming",
			Severity:    "info",
			Suggestion:  "使用 camelCase 命名变量，例如: myVariable",
		},
		// 函数命名 - camelCase
		{
			Name:        "js_func_naming",
			Description: "JavaScript 函数应使用 camelCase 命名",
			Pattern:     regexp.MustCompile(`\bfunction\s+[A-Z][a-zA-Z]*\s*\(`),
			Category:    "naming",
			Severity:    "info",
			Suggestion:  "使用 camelCase 命名函数，例如: myFunction",
		},
		// 类命名 - PascalCase
		{
			Name:        "js_class_naming",
			Description: "JavaScript 类应使用 PascalCase 命名",
			Pattern:     regexp.MustCompile(`\bclass\s+[a-z][a-zA-Z]*\s*[{]`),
			Category:    "naming",
			Severity:    "info",
			Suggestion:  "使用 PascalCase 命名类，例如: MyClass",
		},
		// 最佳实践 - var 使用
		{
			Name:        "js_var_usage",
			Description: "避免使用 var，使用 let 或 const",
			Pattern:     regexp.MustCompile(`\bvar\s+`),
			Category:    "best_practice",
			Severity:    "info",
			Suggestion:  "使用 let 或 const 代替 var",
		},
		// 最佳实践 - == 使用
		{
			Name:        "js_equality",
			Description: "使用 === 代替 ==",
			Pattern:     regexp.MustCompile(`[^=!]==[^=]`),
			Category:    "best_practice",
			Severity:    "info",
			Suggestion:  "使用严格相等 === 避免类型转换问题",
		},
		// 最佳实践 - 空 catch
		{
			Name:        "js_empty_catch",
			Description: "避免空的 catch 块",
			Pattern:     regexp.MustCompile(`catch\s*\([^)]*\)\s*\{\s*\}`),
			Category:    "best_practice",
			Severity:    "low",
			Suggestion:  "至少记录错误或重新抛出",
		},
		// 性能 - console.log
		{
			Name:        "js_console_log",
			Description: "生产代码中避免使用 console.log",
			Pattern:     regexp.MustCompile(`\bconsole\.log\s*\(`),
			Category:    "performance",
			Severity:    "info",
			Suggestion:  "使用适当的日志库或移除调试语句",
		},
	}

	// Go 命名规范
	a.qualityRules["go"] = []QualityRule{
		// 变量命名 - camelCase (检测 snake_case)
		{
			Name:        "go_var_naming",
			Description: "Go 变量应使用 camelCase 或 PascalCase 命名",
			Pattern:     regexp.MustCompile(`\b[a-z]+_[a-z_]+\s*:=`),
			Category:    "naming",
			Severity:    "info",
			Suggestion:  "Go 不推荐使用下划线，使用 camelCase，例如: myVariable",
		},
		// 常量命名 - PascalCase 或 camelCase (检测 snake_case)
		{
			Name:        "go_const_naming",
			Description: "Go 常量应使用 PascalCase 或 camelCase 命名",
			Pattern:     regexp.MustCompile(`\bconst\s+[a-z]+_[a-z_]+`),
			Category:    "naming",
			Severity:    "info",
			Suggestion:  "Go 不推荐使用下划线，使用 camelCase 或 PascalCase",
		},
		// 函数命名 - PascalCase 或 camelCase (检测 snake_case)
		{
			Name:        "go_func_naming",
			Description: "Go 函数应使用 PascalCase 或 camelCase 命名",
			Pattern:     regexp.MustCompile(`\bfunc\s+[a-z]+_[a-z_]+\s*\(`),
			Category:    "naming",
			Severity:    "info",
			Suggestion:  "Go 不推荐使用下划线，使用 camelCase 或 PascalCase",
		},
		// 最佳实践 - 错误检查
		{
			Name:        "go_error_check",
			Description: "应该检查错误返回值",
			Pattern:     regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*\s*,\s*_\s*:=.*\(`),
			Category:    "best_practice",
			Severity:    "low",
			Suggestion:  "检查并处理错误，不要忽略错误返回值",
		},
		// 性能 - 字符串拼接
		{
			Name:        "go_string_concat",
			Description: "循环中避免使用 + 拼接字符串",
			Pattern:     regexp.MustCompile(`\+=\s*"`),
			Category:    "performance",
			Severity:    "info",
			Suggestion:  "使用 strings.Builder 或 bytes.Buffer",
		},
	}
}

// checkQualityIssues 检查代码质量问题
func (a *CodeAnalyzer) checkQualityIssues(language string, lines []string, result *AnalysisResult) {
	issues := []QualityIssue{}

	// 检查命名规范、最佳实践和性能问题
	rules, exists := a.qualityRules[language]
	if exists {
		for _, rule := range rules {
			for i, line := range lines {
				if rule.Pattern.MatchString(line) {
					issue := QualityIssue{
						Rule:        rule.Name,
						Description: rule.Description,
						Category:    rule.Category,
						Severity:    rule.Severity,
						Line:        i + 1,
						Code:        strings.TrimSpace(line),
						Suggestion:  rule.Suggestion,
					}
					// 增强问题信息
					issue = a.enhanceQualityIssue(issue, lines, rule.Pattern)
					issues = append(issues, issue)
				}
			}
		}
	}

	// 检查代码风格问题
	styleIssues := a.checkCodeStyle(lines)
	issues = append(issues, styleIssues...)

	// 应用自定义规则
	customIssues := a.applyCustomRules(language, lines)
	issues = append(issues, customIssues...)

	result.QualityIssues = issues
}

// checkCodeStyle 检查代码风格问题
func (a *CodeAnalyzer) checkCodeStyle(lines []string) []QualityIssue {
	issues := []QualityIssue{}

	for i, line := range lines {
		lineNum := i + 1

		// 检查行长度（超过 120 字符）
		if len(line) > 120 {
			issue := QualityIssue{
				Rule:        "line_too_long",
				Description: fmt.Sprintf("行长度 %d 超过建议的 120 字符", len(line)),
				Category:    "style",
				Severity:    "info",
				Line:        lineNum,
				Code:        line[:50] + "...",
				Suggestion:  "将长行拆分为多行",
			}
			issue = a.enhanceQualityIssue(issue, lines, nil)
			issues = append(issues, issue)
		}

		// 检查尾随空格
		if len(line) > 0 && line[len(line)-1] == ' ' {
			issue := QualityIssue{
				Rule:        "trailing_whitespace",
				Description: "行尾有多余的空格",
				Category:    "style",
				Severity:    "info",
				Line:        lineNum,
				Code:        strings.TrimSpace(line),
				Suggestion:  "删除行尾空格",
			}
			issue = a.enhanceQualityIssue(issue, lines, nil)
			issues = append(issues, issue)
		}

		// 检查制表符（建议使用空格）
		if strings.Contains(line, "\t") {
			issue := QualityIssue{
				Rule:        "tab_character",
				Description: "使用了制表符",
				Category:    "style",
				Severity:    "info",
				Line:        lineNum,
				Code:        strings.TrimSpace(line),
				Suggestion:  "使用空格代替制表符",
			}
			issue = a.enhanceQualityIssue(issue, lines, nil)
			issues = append(issues, issue)
		}
	}

	// 检查连续空行（超过 2 行）
	emptyLineCount := 0
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			emptyLineCount++
			if emptyLineCount > 2 {
				issue := QualityIssue{
					Rule:        "too_many_blank_lines",
					Description: "连续空行过多",
					Category:    "style",
					Severity:    "info",
					Line:        i + 1,
					Code:        "",
					Suggestion:  "减少连续空行数量（最多 2 行）",
				}
				issue = a.enhanceQualityIssue(issue, lines, nil)
				issues = append(issues, issue)
			}
		} else {
			emptyLineCount = 0
		}
	}

	return issues
}

// calculateQualityScore 计算代码质量评分
func (a *CodeAnalyzer) calculateQualityScore(result *AnalysisResult) int {
	score := 100

	// 安全问题扣分
	for _, issue := range result.Issues {
		switch issue.Severity {
		case "high":
			score -= 20
		case "medium":
			score -= 10
		case "low":
			score -= 5
		}
	}

	// 加密问题扣分
	for _, issue := range result.CryptoIssues {
		switch issue.Severity {
		case "high":
			score -= 15
		case "medium":
			score -= 8
		case "low":
			score -= 3
		}
	}

	// 数据库问题扣分
	for _, issue := range result.DatabaseOps {
		switch issue.Severity {
		case "high":
			score -= 15
		case "medium":
			score -= 8
		case "low":
			score -= 3
		}
	}

	// 代码复杂度扣分
	if result.Complexity > 50 {
		score -= 15
	} else if result.Complexity > 30 {
		score -= 10
	} else if result.Complexity > 20 {
		score -= 5
	}

	// 代码行数扣分（过长）
	if result.LineCount > 500 {
		score -= 10
	} else if result.LineCount > 300 {
		score -= 5
	}

	// 质量问题扣分
	for _, issue := range result.QualityIssues {
		switch issue.Severity {
		case "info":
			score -= 1
		case "low":
			score -= 2
		}
	}

	// 确保分数在 0-100 范围内
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// LoadCustomRules 从 YAML 文件加载自定义规则
func (a *CodeAnalyzer) LoadCustomRules(rulesFile string) error {
	// 读取文件
	data, err := os.ReadFile(rulesFile)
	if err != nil {
		return fmt.Errorf("failed to read custom rules file: %w", err)
	}

	// 解析 YAML
	var config CustomRulesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse custom rules YAML: %w", err)
	}

	// 验证并加载规则
	for _, ruleConfig := range config.CustomRules {
		if err := a.validateCustomRule(ruleConfig); err != nil {
			return fmt.Errorf("invalid custom rule '%s': %w", ruleConfig.Name, err)
		}

		// 编译正则表达式
		pattern, err := regexp.Compile(ruleConfig.Pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern for rule '%s': %w", ruleConfig.Name, err)
		}

		// 添加自定义规则
		a.customRules = append(a.customRules, CustomRule{
			Name:        ruleConfig.Name,
			Language:    strings.ToLower(ruleConfig.Language),
			Pattern:     pattern,
			Category:    ruleConfig.Category,
			Severity:    ruleConfig.Severity,
			Description: ruleConfig.Description,
			Suggestion:  ruleConfig.Suggestion,
		})
	}

	return nil
}

// validateCustomRule 验证自定义规则配置
func (a *CodeAnalyzer) validateCustomRule(rule CustomRuleConfig) error {
	// 验证必填字段
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if rule.Language == "" {
		return fmt.Errorf("language is required")
	}

	if rule.Pattern == "" {
		return fmt.Errorf("pattern is required")
	}

	if rule.Description == "" {
		return fmt.Errorf("description is required")
	}

	// 验证语言
	validLanguages := map[string]bool{
		"python":     true,
		"javascript": true,
		"go":         true,
		"bash":       true,
	}

	if !validLanguages[strings.ToLower(rule.Language)] {
		return fmt.Errorf("unsupported language '%s', must be one of: python, javascript, go, bash", rule.Language)
	}

	// 验证严重级别
	validSeverities := map[string]bool{
		"high":   true,
		"medium": true,
		"low":    true,
		"info":   true,
	}

	if rule.Severity != "" && !validSeverities[strings.ToLower(rule.Severity)] {
		return fmt.Errorf("invalid severity '%s', must be one of: high, medium, low, info", rule.Severity)
	}

	// 默认严重级别
	if rule.Severity == "" {
		rule.Severity = "info"
	}

	// 验证类别
	validCategories := map[string]bool{
		"naming":        true,
		"style":         true,
		"performance":   true,
		"best_practice": true,
		"security":      true,
		"custom":        true,
	}

	if rule.Category != "" && !validCategories[strings.ToLower(rule.Category)] {
		return fmt.Errorf("invalid category '%s', must be one of: naming, style, performance, best_practice, security, custom", rule.Category)
	}

	// 默认类别
	if rule.Category == "" {
		rule.Category = "custom"
	}

	// 验证正则表达式
	if _, err := regexp.Compile(rule.Pattern); err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	return nil
}

// applyCustomRules 应用自定义规则
func (a *CodeAnalyzer) applyCustomRules(language string, lines []string) []QualityIssue {
	issues := []QualityIssue{}

	for _, rule := range a.customRules {
		// 只应用匹配语言的规则
		if rule.Language != strings.ToLower(language) {
			continue
		}

		// 检查每一行
		for i, line := range lines {
			if rule.Pattern.MatchString(line) {
				issue := QualityIssue{
					Rule:        rule.Name,
					Description: rule.Description,
					Category:    rule.Category,
					Severity:    rule.Severity,
					Line:        i + 1,
					Code:        strings.TrimSpace(line),
					Suggestion:  rule.Suggestion,
				}
				// 增强问题信息
				issue = a.enhanceQualityIssue(issue, lines, rule.Pattern)
				issues = append(issues, issue)
			}
		}
	}

	return issues
}

// GetCustomRulesCount 获取已加载的自定义规则数量
func (a *CodeAnalyzer) GetCustomRulesCount() int {
	return len(a.customRules)
}

// ClearCustomRules 清除所有自定义规则
func (a *CodeAnalyzer) ClearCustomRules() {
	a.customRules = []CustomRule{}
}

// extractContext 提取代码上下文（前后几行）
func (a *CodeAnalyzer) extractContext(lines []string, lineNum int, contextLines int) []string {
	context := []string{}

	// 计算上下文范围
	start := lineNum - contextLines - 1
	if start < 0 {
		start = 0
	}

	end := lineNum + contextLines
	if end > len(lines) {
		end = len(lines)
	}

	// 提取上下文
	for i := start; i < end; i++ {
		prefix := "  "
		if i == lineNum-1 {
			prefix = "> " // 标记问题行
		}
		context = append(context, fmt.Sprintf("%s%d: %s", prefix, i+1, lines[i]))
	}

	return context
}

// findColumnNumber 查找匹配模式在行中的列号
func (a *CodeAnalyzer) findColumnNumber(line string, pattern *regexp.Regexp) int {
	loc := pattern.FindStringIndex(line)
	if loc != nil && len(loc) > 0 {
		return loc[0] + 1 // 列号从 1 开始
	}
	return 0
}

// enhanceQualityIssue 增强质量问题信息（添加上下文和列号）
func (a *CodeAnalyzer) enhanceQualityIssue(issue QualityIssue, lines []string, pattern *regexp.Regexp) QualityIssue {
	// 添加上下文
	if issue.Line > 0 && issue.Line <= len(lines) {
		issue.Context = a.extractContext(lines, issue.Line, 2)

		// 查找列号
		if pattern != nil {
			issue.Column = a.findColumnNumber(lines[issue.Line-1], pattern)
		}
	}

	return issue
}

// FormatAnalysisResult 格式化分析结果为可视化文本
func (a *CodeAnalyzer) FormatAnalysisResult(result *AnalysisResult) string {
	var sb strings.Builder

	// 标题
	sb.WriteString("=== Code Analysis Report ===\n\n")

	// 基本信息
	sb.WriteString(fmt.Sprintf("Language: %s\n", result.Language))
	sb.WriteString(fmt.Sprintf("Safety Status: %s\n", map[bool]string{true: "✓ SAFE", false: "✗ UNSAFE"}[result.Safe]))
	sb.WriteString(fmt.Sprintf("Quality Score: %d/100\n", result.Score))
	sb.WriteString(fmt.Sprintf("Lines of Code: %d\n", result.LineCount))
	sb.WriteString(fmt.Sprintf("Complexity: %d\n", result.Complexity))
	sb.WriteString("\n")

	// 安全问题
	if len(result.Issues) > 0 {
		sb.WriteString("--- Security Issues ---\n")
		for i, issue := range result.Issues {
			sb.WriteString(fmt.Sprintf("%d. [%s] %s (Line %d)\n", i+1, strings.ToUpper(issue.Severity), issue.Description, issue.Line))
			sb.WriteString(fmt.Sprintf("   Code: %s\n", issue.Code))
		}
		sb.WriteString("\n")
	}

	// 加密问题
	if len(result.CryptoIssues) > 0 {
		sb.WriteString("--- Cryptographic Issues ---\n")
		for i, issue := range result.CryptoIssues {
			sb.WriteString(fmt.Sprintf("%d. [%s] %s (Line %d)\n", i+1, strings.ToUpper(issue.Severity), issue.Description, issue.Line))
			sb.WriteString(fmt.Sprintf("   Code: %s\n", issue.Code))
		}
		sb.WriteString("\n")
	}

	// 数据库问题
	if len(result.DatabaseOps) > 0 {
		sb.WriteString("--- Database Issues ---\n")
		for i, issue := range result.DatabaseOps {
			sb.WriteString(fmt.Sprintf("%d. [%s] %s (Line %d)\n", i+1, strings.ToUpper(issue.Severity), issue.Description, issue.Line))
			sb.WriteString(fmt.Sprintf("   Code: %s\n", issue.Code))
		}
		sb.WriteString("\n")
	}

	// 代码质量问题
	if len(result.QualityIssues) > 0 {
		sb.WriteString("--- Code Quality Issues ---\n")
		for i, issue := range result.QualityIssues {
			sb.WriteString(fmt.Sprintf("%d. [%s/%s] %s (Line %d", i+1, issue.Category, issue.Severity, issue.Description, issue.Line))
			if issue.Column > 0 {
				sb.WriteString(fmt.Sprintf(", Col %d", issue.Column))
			}
			sb.WriteString(")\n")
			sb.WriteString(fmt.Sprintf("   Code: %s\n", issue.Code))
			if issue.Suggestion != "" {
				sb.WriteString(fmt.Sprintf("   Suggestion: %s\n", issue.Suggestion))
			}
			if len(issue.Context) > 0 {
				sb.WriteString("   Context:\n")
				for _, ctx := range issue.Context {
					sb.WriteString(fmt.Sprintf("     %s\n", ctx))
				}
			}
		}
		sb.WriteString("\n")
	}

	// 网络操作
	if len(result.NetworkOps) > 0 {
		sb.WriteString(fmt.Sprintf("--- Network Operations (%d detected) ---\n", len(result.NetworkOps)))
		for i, op := range result.NetworkOps {
			sb.WriteString(fmt.Sprintf("%d. %s operation (Line %d)\n", i+1, strings.ToUpper(op.Type), op.Line))
		}
		sb.WriteString("\n")
	}

	// 文件系统操作
	if len(result.FileSystemOps) > 0 {
		sb.WriteString(fmt.Sprintf("--- File System Operations (%d detected) ---\n", len(result.FileSystemOps)))
		for i, op := range result.FileSystemOps {
			sb.WriteString(fmt.Sprintf("%d. %s operation (Line %d)\n", i+1, strings.ToUpper(op.Type), op.Line))
		}
		sb.WriteString("\n")
	}

	// 进程操作
	if len(result.ProcessOps) > 0 {
		sb.WriteString(fmt.Sprintf("--- Process Operations (%d detected) ---\n", len(result.ProcessOps)))
		for i, op := range result.ProcessOps {
			sb.WriteString(fmt.Sprintf("%d. %s operation (Line %d)\n", i+1, strings.ToUpper(op.Type), op.Line))
		}
		sb.WriteString("\n")
	}

	// 建议
	if len(result.Suggestions) > 0 {
		sb.WriteString("--- Suggestions ---\n")
		for i, suggestion := range result.Suggestions {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, suggestion))
		}
		sb.WriteString("\n")
	}

	// 总结
	sb.WriteString("=== Summary ===\n")
	sb.WriteString(fmt.Sprintf("Total Issues: %d\n", len(result.Issues)+len(result.CryptoIssues)+len(result.DatabaseOps)+len(result.QualityIssues)))
	sb.WriteString(fmt.Sprintf("  - Security: %d\n", len(result.Issues)))
	sb.WriteString(fmt.Sprintf("  - Cryptographic: %d\n", len(result.CryptoIssues)))
	sb.WriteString(fmt.Sprintf("  - Database: %d\n", len(result.DatabaseOps)))
	sb.WriteString(fmt.Sprintf("  - Quality: %d\n", len(result.QualityIssues)))
	sb.WriteString(fmt.Sprintf("Operations Detected: %d\n", len(result.NetworkOps)+len(result.FileSystemOps)+len(result.ProcessOps)))
	sb.WriteString(fmt.Sprintf("  - Network: %d\n", len(result.NetworkOps)))
	sb.WriteString(fmt.Sprintf("  - File System: %d\n", len(result.FileSystemOps)))
	sb.WriteString(fmt.Sprintf("  - Process: %d\n", len(result.ProcessOps)))

	return sb.String()
}
