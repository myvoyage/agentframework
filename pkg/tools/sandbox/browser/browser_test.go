// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package browser

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewBrowserModule tests browser module creation
func TestNewBrowserModule(t *testing.T) {
	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{
			Width:  1920,
			Height: 1080,
		},
		PoolSize: 2,
	}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)
	require.NotNil(t, module)
	defer module.Close()

	assert.Equal(t, config.Headless, module.config.Headless)
	assert.Equal(t, config.Timeout, module.config.Timeout)
	assert.Equal(t, config.PoolSize, module.config.PoolSize)
}

// TestNewBrowserModuleDefaults tests default configuration
func TestNewBrowserModuleDefaults(t *testing.T) {
	config := BrowserConfig{}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)
	require.NotNil(t, module)
	defer module.Close()

	assert.Equal(t, 30000, module.config.Timeout)
	assert.Equal(t, int64(1920), module.config.Viewport.Width)
	assert.Equal(t, int64(1080), module.config.Viewport.Height)
	assert.Equal(t, 5, module.config.PoolSize)
}

// TestBrowserPool tests browser pool functionality
func TestBrowserPool(t *testing.T) {
	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{Width: 1920, Height: 1080},
		PoolSize: 2,
	}

	pool, err := NewBrowserPool(config)
	require.NoError(t, err)
	require.NotNil(t, pool)
	defer pool.Close()

	// Get first context
	ctx1, err := pool.Get()
	require.NoError(t, err)
	require.NotNil(t, ctx1)
	assert.True(t, ctx1.inUse)

	// Get second context
	ctx2, err := pool.Get()
	require.NoError(t, err)
	require.NotNil(t, ctx2)
	assert.True(t, ctx2.inUse)

	// Put contexts back
	pool.Put(ctx1)
	pool.Put(ctx2)

	assert.False(t, ctx1.inUse)
	assert.False(t, ctx2.inUse)
}

// TestBrowserPoolConcurrent tests concurrent access to browser pool
func TestBrowserPoolConcurrent(t *testing.T) {
	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{Width: 1920, Height: 1080},
		PoolSize: 3,
	}

	pool, err := NewBrowserPool(config)
	require.NoError(t, err)
	require.NotNil(t, pool)
	defer pool.Close()

	done := make(chan bool, 5)

	// Simulate 5 concurrent operations with pool size of 3
	for i := 0; i < 5; i++ {
		go func(id int) {
			ctx, err := pool.Get()
			if err != nil {
				t.Errorf("Failed to get context %d: %v", id, err)
				done <- false
				return
			}

			// Simulate work
			time.Sleep(100 * time.Millisecond)

			pool.Put(ctx)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		success := <-done
		assert.True(t, success)
	}
}

// TestNavigate tests navigation functionality
func TestNavigate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{Width: 1920, Height: 1080},
		PoolSize: 1,
	}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)
	defer module.Close()

	// Test navigation to example.com
	result, err := module.navigate("https://example.com", 0)
	require.NoError(t, err)
	assert.True(t, result["success"].(bool))
	assert.Contains(t, result, "title")
	assert.Contains(t, result, "duration")
}

// TestNavigateBlocked tests domain blocking
func TestNavigateBlocked(t *testing.T) {
	config := BrowserConfig{
		Headless:       true,
		Timeout:        30000,
		Viewport:       Viewport{Width: 1920, Height: 1080},
		PoolSize:       1,
		BlockedDomains: []string{"blocked.com"},
	}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)
	defer module.Close()

	// Test navigation to blocked domain
	result, err := module.navigate("https://blocked.com", 0)
	require.NoError(t, err)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["error"], "blocked")

	// Check stats
	stats := module.GetStats()
	assert.Equal(t, int64(1), stats["blocked_count"])
}

// TestNavigateAllowedDomains tests domain whitelist
func TestNavigateAllowedDomains(t *testing.T) {
	config := BrowserConfig{
		Headless:       true,
		Timeout:        30000,
		Viewport:       Viewport{Width: 1920, Height: 1080},
		PoolSize:       1,
		AllowedDomains: []string{"example.com"},
	}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)
	defer module.Close()

	// Test navigation to non-allowed domain
	result, err := module.navigate("https://google.com", 0)
	require.NoError(t, err)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["error"], "not in allowed list")
}

// TestScreenshot tests screenshot functionality
func TestScreenshot(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{Width: 1920, Height: 1080},
		PoolSize: 1,
	}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)
	defer module.Close()

	// Navigate first
	_, err = module.navigate("https://example.com", 0)
	require.NoError(t, err)

	// Create temp directory for screenshots
	tempDir := t.TempDir()
	screenshotPath := filepath.Join(tempDir, "screenshot.png")

	// Take screenshot
	result, err := module.screenshot(screenshotPath, "", false)
	require.NoError(t, err)
	assert.True(t, result["success"].(bool))

	// Verify file exists
	_, err = os.Stat(screenshotPath)
	assert.NoError(t, err)
}

// TestScreenshotFullPage tests full page screenshot
func TestScreenshotFullPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{Width: 1920, Height: 1080},
		PoolSize: 1,
	}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)
	defer module.Close()

	// Navigate first
	_, err = module.navigate("https://example.com", 0)
	require.NoError(t, err)

	// Create temp directory for screenshots
	tempDir := t.TempDir()
	screenshotPath := filepath.Join(tempDir, "fullpage.png")

	// Take full page screenshot
	result, err := module.screenshot(screenshotPath, "", true)
	require.NoError(t, err)
	assert.True(t, result["success"].(bool))
	assert.True(t, result["full_page"].(bool))

	// Verify file exists
	_, err = os.Stat(screenshotPath)
	assert.NoError(t, err)
}

// TestGeneratePDF tests PDF generation
func TestGeneratePDF(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{Width: 1920, Height: 1080},
		PoolSize: 1,
	}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)
	defer module.Close()

	// Navigate first
	_, err = module.navigate("https://example.com", 0)
	require.NoError(t, err)

	// Create temp directory for PDF
	tempDir := t.TempDir()
	pdfPath := filepath.Join(tempDir, "page.pdf")

	// Generate PDF
	result, err := module.generatePDF(pdfPath)
	require.NoError(t, err)
	assert.True(t, result["success"].(bool))

	// Verify file exists
	_, err = os.Stat(pdfPath)
	assert.NoError(t, err)
}

// TestGetText tests text retrieval
func TestGetText(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{Width: 1920, Height: 1080},
		PoolSize: 1,
	}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)
	defer module.Close()

	// Navigate first
	_, err = module.navigate("https://example.com", 0)
	require.NoError(t, err)

	// Get text from h1 element
	result, err := module.getText("h1")
	require.NoError(t, err)
	assert.True(t, result["success"].(bool))
	assert.NotEmpty(t, result["text"])
}

// TestExecuteJS tests JavaScript execution
func TestExecuteJS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{Width: 1920, Height: 1080},
		PoolSize: 1,
	}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)
	defer module.Close()

	// Navigate first
	_, err = module.navigate("https://example.com", 0)
	require.NoError(t, err)

	// Execute JavaScript
	result, err := module.executeJS("document.title")
	require.NoError(t, err)
	assert.True(t, result["success"].(bool))
	assert.NotNil(t, result["result"])
}

// TestGetCookies tests cookie retrieval
func TestGetCookies(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{Width: 1920, Height: 1080},
		PoolSize: 1,
	}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)
	defer module.Close()

	// Navigate first
	_, err = module.navigate("https://example.com", 0)
	require.NoError(t, err)

	// Get cookies
	result, err := module.getCookies()
	require.NoError(t, err)
	assert.True(t, result["success"].(bool))
	assert.Contains(t, result, "cookies")
	assert.Contains(t, result, "count")
}

// TestSetCookies tests cookie setting
func TestSetCookies(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{Width: 1920, Height: 1080},
		PoolSize: 1,
	}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)
	defer module.Close()

	// Navigate first
	_, err = module.navigate("https://example.com", 0)
	require.NoError(t, err)

	// Set cookies
	cookies := []map[string]interface{}{
		{
			"name":   "test_cookie",
			"value":  "test_value",
			"domain": "example.com",
			"path":   "/",
		},
	}

	result, err := module.setCookies(cookies)
	require.NoError(t, err)
	assert.True(t, result["success"].(bool))
	assert.Equal(t, 1, int(result["count"].(int)))
}

// TestGetTools tests MCP tools retrieval
func TestGetTools(t *testing.T) {
	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{Width: 1920, Height: 1080},
		PoolSize: 1,
	}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)
	defer module.Close()

	tools, err := module.GetTools(context.Background())
	require.NoError(t, err)
	assert.Len(t, tools, 9) // Should have 9 tools

	// Verify tool names
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		info, err := tool.Info(context.Background())
		require.NoError(t, err)
		toolNames[info.Name] = true
	}

	expectedTools := []string{
		"browser_navigate",
		"browser_click",
		"browser_input",
		"browser_screenshot",
		"browser_pdf",
		"browser_get_text",
		"browser_execute_js",
		"browser_get_cookies",
		"browser_set_cookies",
	}

	for _, name := range expectedTools {
		assert.True(t, toolNames[name], "Tool %s should exist", name)
	}
}

// TestGetStats tests statistics tracking
func TestGetStats(t *testing.T) {
	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{Width: 1920, Height: 1080},
		PoolSize: 1,
	}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)
	defer module.Close()

	// Initial stats should be zero
	stats := module.GetStats()
	assert.Equal(t, int64(0), stats["total_operations"])
	assert.Equal(t, int64(0), stats["success_count"])
	assert.Equal(t, int64(0), stats["failure_count"])
	assert.Equal(t, int64(0), stats["blocked_count"])
}

// TestURLValidation tests URL validation
func TestURLValidation(t *testing.T) {
	config := BrowserConfig{
		Headless:       true,
		AllowedDomains: []string{"example.com"},
		BlockedDomains: []string{"blocked.com"},
	}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)
	defer module.Close()

	// Test allowed domain
	err = module.isURLAllowed("https://example.com/page")
	assert.NoError(t, err)

	// Test blocked domain
	err = module.isURLAllowed("https://blocked.com/page")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blocked")

	// Test non-allowed domain
	err = module.isURLAllowed("https://google.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in allowed list")

	// Test invalid URL
	err = module.isURLAllowed("not-a-url")
	assert.Error(t, err)
}

// TestResourceCleanup tests proper resource cleanup
func TestResourceCleanup(t *testing.T) {
	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{Width: 1920, Height: 1080},
		PoolSize: 2,
	}

	module, err := NewBrowserModule(config)
	require.NoError(t, err)

	// Get some contexts
	ctx1, err := module.pool.Get()
	require.NoError(t, err)
	ctx2, err := module.pool.Get()
	require.NoError(t, err)

	// Put them back
	module.pool.Put(ctx1)
	module.pool.Put(ctx2)

	// Close module
	err = module.Close()
	assert.NoError(t, err)
}

// BenchmarkNavigate benchmarks navigation performance
func BenchmarkNavigate(b *testing.B) {
	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{Width: 1920, Height: 1080},
		PoolSize: 3,
	}

	module, err := NewBrowserModule(config)
	require.NoError(b, err)
	defer module.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = module.navigate("https://example.com", 0)
	}
}

// BenchmarkScreenshot benchmarks screenshot performance
func BenchmarkScreenshot(b *testing.B) {
	config := BrowserConfig{
		Headless: true,
		Timeout:  30000,
		Viewport: Viewport{Width: 1920, Height: 1080},
		PoolSize: 3,
	}

	module, err := NewBrowserModule(config)
	require.NoError(b, err)
	defer module.Close()

	// Navigate once
	_, err = module.navigate("https://example.com", 0)
	require.NoError(b, err)

	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := filepath.Join(tempDir, "screenshot.png")
		_, _ = module.screenshot(path, "", false)
	}
}
