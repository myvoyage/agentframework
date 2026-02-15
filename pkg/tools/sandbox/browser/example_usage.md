# Browser Module - Example Usage

This document provides examples of how to use the Browser Module MCP tools for browser automation.

## Table of Contents

1. [Basic Navigation](#basic-navigation)
2. [Element Interaction](#element-interaction)
3. [Page Capture](#page-capture)
4. [JavaScript Execution](#javascript-execution)
5. [Cookie Management](#cookie-management)
6. [Advanced Usage](#advanced-usage)

---

## Basic Navigation

### Navigate to a URL

```json
{
  "tool": "browser_navigate",
  "arguments": {
    "url": "https://example.com",
    "timeout": 30000
  }
}
```

**Response:**
```json
{
  "success": true,
  "url": "https://example.com",
  "title": "Example Domain",
  "duration": 1234,
  "message": "Navigation successful"
}
```

---

## Element Interaction

### Click an Element

```json
{
  "tool": "browser_click",
  "arguments": {
    "selector": "button.submit"
  }
}
```

**Response:**
```json
{
  "success": true,
  "selector": "button.submit",
  "message": "Element clicked successfully"
}
```

### Input Text

```json
{
  "tool": "browser_input",
  "arguments": {
    "selector": "input[name='username']",
    "value": "testuser"
  }
}
```

**Response:**
```json
{
  "success": true,
  "selector": "input[name='username']",
  "value": "testuser",
  "message": "Text input successful"
}
```

### Get Element Text

```json
{
  "tool": "browser_get_text",
  "arguments": {
    "selector": "h1.title"
  }
}
```

**Response:**
```json
{
  "success": true,
  "selector": "h1.title",
  "text": "Welcome to Example",
  "message": "Text retrieved successfully"
}
```

---

## Page Capture

### Take a Screenshot (Viewport)

```json
{
  "tool": "browser_screenshot",
  "arguments": {
    "path": "/tmp/screenshot.png"
  }
}
```

**Response:**
```json
{
  "success": true,
  "path": "/tmp/screenshot.png",
  "selector": "",
  "full_page": false,
  "size": 123456,
  "message": "Screenshot saved successfully"
}
```

### Take a Full Page Screenshot

```json
{
  "tool": "browser_screenshot",
  "arguments": {
    "path": "/tmp/fullpage.png",
    "full_page": true
  }
}
```

### Screenshot a Specific Element

```json
{
  "tool": "browser_screenshot",
  "arguments": {
    "path": "/tmp/element.png",
    "selector": "div.content"
  }
}
```

### Generate PDF

```json
{
  "tool": "browser_pdf",
  "arguments": {
    "path": "/tmp/page.pdf"
  }
}
```

**Response:**
```json
{
  "success": true,
  "path": "/tmp/page.pdf",
  "size": 234567,
  "message": "PDF generated successfully"
}
```

---

## JavaScript Execution

### Execute JavaScript

```json
{
  "tool": "browser_execute_js",
  "arguments": {
    "script": "document.title"
  }
}
```

**Response:**
```json
{
  "success": true,
  "script": "document.title",
  "result": "Example Domain",
  "message": "JavaScript executed successfully"
}
```

### Complex JavaScript Example

```json
{
  "tool": "browser_execute_js",
  "arguments": {
    "script": "Array.from(document.querySelectorAll('a')).map(a => a.href)"
  }
}
```

**Response:**
```json
{
  "success": true,
  "script": "Array.from(document.querySelectorAll('a')).map(a => a.href)",
  "result": ["https://example.com/page1", "https://example.com/page2"],
  "message": "JavaScript executed successfully"
}
```

---

## Cookie Management

### Get Cookies

```json
{
  "tool": "browser_get_cookies",
  "arguments": {}
}
```

**Response:**
```json
{
  "success": true,
  "cookies": [
    {
      "name": "session_id",
      "value": "abc123",
      "domain": "example.com",
      "path": "/",
      "expires": 1735689600,
      "httpOnly": true,
      "secure": true,
      "sameSite": "Lax"
    }
  ],
  "count": 1,
  "message": "Cookies retrieved successfully"
}
```

### Set Cookies

```json
{
  "tool": "browser_set_cookies",
  "arguments": {
    "cookies": [
      {
        "name": "user_pref",
        "value": "dark_mode",
        "domain": "example.com",
        "path": "/",
        "expires": 1735689600,
        "httpOnly": false,
        "secure": true,
        "sameSite": "Lax"
      }
    ]
  }
}
```

**Response:**
```json
{
  "success": true,
  "count": 1,
  "message": "Cookies set successfully"
}
```

---

## Advanced Usage

### Complete Form Submission Workflow

```javascript
// 1. Navigate to the page
await browser_navigate({
  url: "https://example.com/login"
});

// 2. Fill in the form
await browser_input({
  selector: "input[name='username']",
  value: "testuser"
});

await browser_input({
  selector: "input[name='password']",
  value: "password123"
});

// 3. Submit the form
await browser_click({
  selector: "button[type='submit']"
});

// 4. Wait and verify
await browser_get_text({
  selector: "div.welcome-message"
});

// 5. Take a screenshot
await browser_screenshot({
  path: "/tmp/logged-in.png"
});
```

### Web Scraping Example

```javascript
// 1. Navigate to the page
await browser_navigate({
  url: "https://example.com/products"
});

// 2. Extract product information using JavaScript
const products = await browser_execute_js({
  script: `
    Array.from(document.querySelectorAll('.product')).map(p => ({
      name: p.querySelector('.name').textContent,
      price: p.querySelector('.price').textContent,
      image: p.querySelector('img').src
    }))
  `
});

// 3. Take a screenshot for reference
await browser_screenshot({
  path: "/tmp/products.png",
  full_page: true
});

// 4. Generate PDF report
await browser_pdf({
  path: "/tmp/products-report.pdf"
});
```

### Testing with Domain Restrictions

When creating a BrowserModule with domain restrictions:

```go
config := BrowserConfig{
    Headless: true,
    Timeout:  30000,
    Viewport: Viewport{Width: 1920, Height: 1080},
    PoolSize: 5,
    AllowedDomains: []string{"example.com", "test.com"},
    BlockedDomains: []string{"malicious.com"},
}

module, err := NewBrowserModule(config)
```

Attempting to navigate to a blocked or non-allowed domain will result in:

```json
{
  "success": false,
  "error": "domain malicious.com is blocked",
  "url": "https://malicious.com"
}
```

---

## Error Handling

All tools return a consistent error format:

```json
{
  "success": false,
  "error": "Element not found: button.submit",
  "selector": "button.submit"
}
```

Common error scenarios:
- **Element not found**: The CSS selector doesn't match any element
- **Timeout**: Operation took longer than the configured timeout
- **Domain blocked**: URL is in the blocked domains list
- **Domain not allowed**: URL is not in the allowed domains list (when whitelist is configured)
- **Browser context error**: Failed to get a browser instance from the pool

---

## Performance Tips

1. **Use Browser Pool**: The module automatically manages a pool of browser instances for better performance
2. **Adjust Timeout**: Set appropriate timeout values based on your use case
3. **Headless Mode**: Use headless mode (default) for better performance
4. **Viewport Size**: Adjust viewport size based on your needs (default: 1920x1080)
5. **Reuse Connections**: The pool automatically reuses browser instances

---

## Security Considerations

1. **Domain Whitelisting**: Use `AllowedDomains` to restrict navigation to trusted domains
2. **Domain Blacklisting**: Use `BlockedDomains` to prevent navigation to known malicious domains
3. **JavaScript Execution**: Be cautious when executing user-provided JavaScript
4. **Cookie Security**: Always use `secure` and `httpOnly` flags for sensitive cookies
5. **Resource Limits**: Configure appropriate timeout and pool size limits

---

## Configuration Options

```go
type BrowserConfig struct {
    Headless       bool     // Run browser in headless mode (default: true)
    Timeout        int      // Operation timeout in milliseconds (default: 30000)
    UserAgent      string   // Custom user agent string
    Viewport       Viewport // Browser viewport size (default: 1920x1080)
    PoolSize       int      // Browser instance pool size (default: 5)
    AllowedDomains []string // Domain whitelist (empty = allow all)
    BlockedDomains []string // Domain blacklist
}
```

---

## Statistics and Monitoring

Get operation statistics:

```go
stats := module.GetStats()
// Returns:
// {
//   "total_operations": 100,
//   "success_count": 95,
//   "failure_count": 3,
//   "blocked_count": 2
// }
```

This helps monitor:
- Total number of operations performed
- Success rate
- Failure rate
- Number of blocked domain access attempts
