# browser_navigate Tool

## Description

Navigate to a specified URL in the browser. This is the primary tool for loading web pages and is typically the first step in any browser automation workflow.

## Tool Information

- **Name**: `browser_navigate`
- **Category**: Browser Automation
- **Priority**: P0 (Core functionality)

## Parameters

| Parameter | Type    | Required | Description                                    | Default |
|-----------|---------|----------|------------------------------------------------|---------|
| url       | string  | Yes      | The URL to navigate to (must include protocol) | -       |
| timeout   | integer | No       | Navigation timeout in milliseconds             | 30000   |

## Response Format

### Success Response

```json
{
  "success": true,
  "url": "https://example.com",
  "title": "Example Domain",
  "duration": 1234,
  "message": "Navigation successful"
}
```

### Error Response

```json
{
  "success": false,
  "error": "domain blocked.com is blocked",
  "url": "https://blocked.com",
  "duration": 0
}
```

## Examples

### Basic Navigation

```json
{
  "tool": "browser_navigate",
  "arguments": {
    "url": "https://example.com"
  }
}
```

### Navigation with Custom Timeout

```json
{
  "tool": "browser_navigate",
  "arguments": {
    "url": "https://slow-website.com",
    "timeout": 60000
  }
}
```

## Security Features

### Domain Whitelisting

When `AllowedDomains` is configured, only URLs from those domains can be accessed:

```go
config := BrowserConfig{
    AllowedDomains: []string{"example.com", "trusted.com"},
}
```

Attempting to navigate to `https://other.com` will result in:

```json
{
  "success": false,
  "error": "domain other.com is not in allowed list"
}
```

### Domain Blacklisting

When `BlockedDomains` is configured, URLs from those domains are blocked:

```go
config := BrowserConfig{
    BlockedDomains: []string{"malicious.com", "spam.com"},
}
```

Attempting to navigate to `https://malicious.com` will result in:

```json
{
  "success": false,
  "error": "domain malicious.com is blocked"
}
```

## Implementation Details

### Navigation Process

1. **URL Validation**: Checks if the URL is allowed based on whitelist/blacklist
2. **Browser Context**: Acquires a browser instance from the connection pool
3. **Navigation**: Uses chromedp to navigate to the URL
4. **Wait for Ready**: Waits for the page body to be ready
5. **Title Extraction**: Retrieves the page title
6. **Statistics Update**: Updates success/failure/blocked counters

### Timeout Behavior

- If no timeout is specified, uses the module's default timeout (30000ms)
- If the page doesn't load within the timeout, returns an error
- The browser context is properly cleaned up even on timeout

### Connection Pool

The tool automatically manages browser instances through a connection pool:
- Reuses existing browser instances when available
- Creates new instances up to the configured pool size
- Blocks and waits if all instances are in use
- Returns instances to the pool after use

## Error Scenarios

| Error | Cause | Solution |
|-------|-------|----------|
| "invalid URL" | Malformed URL | Ensure URL includes protocol (http:// or https://) |
| "domain X is blocked" | URL in BlockedDomains | Remove from blocklist or use different domain |
| "domain X is not in allowed list" | URL not in AllowedDomains | Add domain to whitelist |
| "failed to get browser context" | Pool exhausted or browser error | Check pool size and browser installation |
| "context deadline exceeded" | Navigation timeout | Increase timeout or check network connectivity |

## Performance Considerations

- **First Navigation**: Takes longer as it initializes the browser instance
- **Subsequent Navigations**: Faster due to connection pool reuse
- **Pool Size**: Larger pool size allows more concurrent operations
- **Headless Mode**: Significantly faster than headed mode

## Best Practices

1. **Always include protocol**: Use `https://example.com` not `example.com`
2. **Set appropriate timeouts**: Adjust based on expected page load time
3. **Handle errors gracefully**: Check the `success` field before proceeding
4. **Use domain restrictions**: Configure whitelist/blacklist for security
5. **Monitor statistics**: Track blocked attempts and failure rates

## Related Tools

- `browser_click`: Click elements after navigation
- `browser_input`: Input text into forms
- `browser_screenshot`: Capture page after loading
- `browser_get_text`: Extract content from the page
- `browser_execute_js`: Run JavaScript on the loaded page

## Statistics Tracking

Each navigation attempt updates the module statistics:

- `total_operations`: Incremented for every call
- `success_count`: Incremented on successful navigation
- `failure_count`: Incremented on errors (timeout, invalid URL, etc.)
- `blocked_count`: Incremented when domain is blocked

Access statistics via:

```go
stats := module.GetStats()
```

## Testing

The tool includes comprehensive tests:

- `TestNavigate`: Basic navigation functionality
- `TestNavigateBlocked`: Domain blocking
- `TestNavigateAllowedDomains`: Domain whitelisting
- `TestURLValidation`: URL validation logic

Run tests with:

```bash
go test -v ./agent/aiosandbox/browser/...
```
