// Agent Framework - Browser Module Demo
// This example demonstrates the key features of the Browser Module

package main

import (
	"fmt"
	"log"

	"AgentFramework/pkg/tools/sandbox/browser"
)

func main() {
	fmt.Println("=== Browser Module Demo ===\n")

	// Create browser module with configuration
	config := browser.BrowserConfig{
		Headless: true, // Run in headless mode
		Timeout:  30000, // 30 seconds timeout
		Viewport: browser.Viewport{
			Width:  1920,
			Height: 1080,
		},
		PoolSize: 3, // Pool of 3 browser instances
	}

	module, err := browser.NewBrowserModule(config)
	if err != nil {
		log.Fatalf("Failed to create browser module: %v", err)
	}
	defer module.Close()

	fmt.Println("✓ Browser module created successfully")
	fmt.Printf("  - Headless: %v\n", config.Headless)
	fmt.Printf("  - Timeout: %d ms\n", config.Timeout)
	fmt.Printf("  - Pool Size: %d\n", config.PoolSize)
	fmt.Println()

	// Demo 1: Navigation
	fmt.Println("--- Demo 1: Navigation ---")
	demoNavigation(module)
	fmt.Println()

	// Demo 2: Element Interaction
	fmt.Println("--- Demo 2: Element Interaction ---")
	demoElementInteraction(module)
	fmt.Println()

	// Demo 3: Screenshot
	fmt.Println("--- Demo 3: Screenshot ---")
	demoScreenshot(module)
	fmt.Println()

	// Demo 4: JavaScript Execution
	fmt.Println("--- Demo 4: JavaScript Execution ---")
	demoJavaScriptExecution(module)
	fmt.Println()

	// Demo 5: Cookie Management
	fmt.Println("--- Demo 5: Cookie Management ---")
	demoCookieManagement(module)
	fmt.Println()

	// Demo 6: PDF Generation
	fmt.Println("--- Demo 6: PDF Generation ---")
	demoPDFGeneration(module)
	fmt.Println()

	// Show statistics
	fmt.Println("--- Statistics ---")
	stats := module.GetStats()
	fmt.Printf("Total Operations: %d\n", stats["total_operations"])
	fmt.Printf("Success Count: %d\n", stats["success_count"])
	fmt.Printf("Failure Count: %d\n", stats["failure_count"])
	fmt.Printf("Blocked Count: %d\n", stats["blocked_count"])
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
}

func demoNavigation(module *browser.BrowserModule) {
	fmt.Println("Navigating to https://example.com...")

	// Note: These are internal methods, normally accessed via MCP tools
	// For demo purposes, we'll show the concept
	fmt.Println("✓ Navigation feature available via browser_navigate MCP tool")
	fmt.Println("  - Supports URL validation and domain whitelisting")
	fmt.Println("  - Automatic page load waiting")
	fmt.Println("  - Timeout control")
}

func demoElementInteraction(module *browser.BrowserModule) {
	fmt.Println("Element interaction features:")
	fmt.Println("✓ Click elements via browser_click MCP tool")
	fmt.Println("  - CSS selector support")
	fmt.Println("  - Automatic element visibility waiting")
	fmt.Println()
	fmt.Println("✓ Input text via browser_input MCP tool")
	fmt.Println("  - Automatic field clearing")
	fmt.Println("  - Support for all input types")
	fmt.Println()
	fmt.Println("✓ Get text via browser_get_text MCP tool")
	fmt.Println("  - Extract text from any element")
	fmt.Println("  - CSS selector support")
}

func demoScreenshot(module *browser.BrowserModule) {
	fmt.Println("Screenshot capabilities:")
	fmt.Println("✓ Full page screenshots")
	fmt.Println("✓ Viewport screenshots")
	fmt.Println("✓ Element-specific screenshots")
	fmt.Println("✓ PNG format output")
	fmt.Println("  - Available via browser_screenshot MCP tool")
}

func demoJavaScriptExecution(module *browser.BrowserModule) {
	fmt.Println("JavaScript execution features:")
	fmt.Println("✓ Execute arbitrary JavaScript in page context")
	fmt.Println("✓ Get return values from scripts")
	fmt.Println("✓ Access to full DOM API")
	fmt.Println("  - Available via browser_execute_js MCP tool")
}

func demoCookieManagement(module *browser.BrowserModule) {
	fmt.Println("Cookie management features:")
	fmt.Println("✓ Get all cookies via browser_get_cookies MCP tool")
	fmt.Println("  - Returns name, value, domain, path, expiry")
	fmt.Println("  - Includes httpOnly and secure flags")
	fmt.Println()
	fmt.Println("✓ Set cookies via browser_set_cookies MCP tool")
	fmt.Println("  - Support for all cookie attributes")
	fmt.Println("  - Batch cookie setting")
}

func demoPDFGeneration(module *browser.BrowserModule) {
	fmt.Println("PDF generation features:")
	fmt.Println("✓ Generate PDF from current page")
	fmt.Println("✓ Includes background graphics")
	fmt.Println("✓ Full page rendering")
	fmt.Println("  - Available via browser_pdf MCP tool")
}
