# code_exec_supported_languages Tool

## Overview

The `code_exec_supported_languages` MCP tool returns a list of programming languages supported by the Code Executor Module. This tool is useful for dynamically discovering which languages are available before attempting to execute code.

## Tool Information

- **Name**: `code_exec_supported_languages`
- **Description**: Get list of supported programming languages
- **Parameters**: None (empty object)
- **Returns**: JSON object with success status and list of supported languages

## Usage

### Request

The tool accepts an empty JSON object as input:

```json
{}
```

### Response

The tool returns a JSON object with the following structure:

```json
{
  "success": true,
  "languages": ["python", "javascript", "bash", "go"]
}
```

**Response Fields:**
- `success` (boolean): Always `true` for this tool
- `languages` (array of strings): List of supported programming language identifiers

## Supported Languages

The default configuration includes:

1. **python** - Python 3.x
2. **javascript** - JavaScript (Node.js)
3. **bash** - Bash/Shell scripts
4. **go** - Go programming language

The actual list of supported languages depends on the module configuration and may vary based on:
- Configuration settings
- Available interpreters/compilers on the system
- Module initialization parameters

## Example Usage

### Using the Tool Directly

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "AgentFramework/agent/aiosandbox/code_exec"
)

func main() {
    // Create module
    config := code_exec.CodeExecutorConfig{
        Timeout:            60000,
        MemoryLimit:        512,
        CPULimit:           2,
        SupportedLanguages: []string{"python", "javascript", "bash", "go"},
    }
    
    module, err := code_exec.NewCodeExecutorModule(config)
    if err != nil {
        panic(err)
    }
    defer module.Close()
    
    // Get tools
    ctx := context.Background()
    tools, err := module.GetTools(ctx)
    if err != nil {
        panic(err)
    }
    
    // Find and use the supported languages tool
    for _, tool := range tools {
        info, _ := tool.Info(ctx)
        if info.Name == "code_exec_supported_languages" {
            result, err := tool.InvokableRun(ctx, "{}")
            if err != nil {
                panic(err)
            }
            
            var response map[string]interface{}
            json.Unmarshal([]byte(result), &response)
            
            fmt.Printf("Supported Languages: %v\n", response["languages"])
            break
        }
    }
}
```

### Using via MCP Protocol

When using the tool through the MCP protocol:

**Request:**
```json
{
  "tool": "code_exec_supported_languages",
  "arguments": {}
}
```

**Response:**
```json
{
  "success": true,
  "languages": ["python", "javascript", "bash", "go"]
}
```

## Use Cases

### 1. Dynamic Language Detection

Before executing code, check which languages are available:

```go
// Get supported languages
result, _ := tool.InvokableRun(ctx, "{}")
var response map[string]interface{}
json.Unmarshal([]byte(result), &response)

languages := response["languages"].([]interface{})

// Check if Python is supported
pythonSupported := false
for _, lang := range languages {
    if lang.(string) == "python" {
        pythonSupported = true
        break
    }
}

if pythonSupported {
    // Execute Python code
    executeCode("python", "print('Hello')")
}
```

### 2. User Interface Generation

Generate a language selector in a UI:

```go
result, _ := tool.InvokableRun(ctx, "{}")
var response map[string]interface{}
json.Unmarshal([]byte(result), &response)

languages := response["languages"].([]interface{})

// Generate dropdown options
for _, lang := range languages {
    fmt.Printf("<option value='%s'>%s</option>\n", lang, lang)
}
```

### 3. Validation Before Execution

Validate user input before attempting code execution:

```go
func validateLanguage(userLang string) bool {
    result, _ := tool.InvokableRun(ctx, "{}")
    var response map[string]interface{}
    json.Unmarshal([]byte(result), &response)
    
    languages := response["languages"].([]interface{})
    for _, lang := range languages {
        if lang.(string) == userLang {
            return true
        }
    }
    return false
}

// Usage
if !validateLanguage("ruby") {
    fmt.Println("Ruby is not supported")
}
```

## Configuration

The supported languages list is determined by the module configuration:

```go
config := CodeExecutorConfig{
    SupportedLanguages: []string{"python", "javascript", "bash", "go"},
}
```

To customize the supported languages:

```go
// Only Python and JavaScript
config := CodeExecutorConfig{
    SupportedLanguages: []string{"python", "javascript"},
}

// All languages
config := CodeExecutorConfig{
    SupportedLanguages: []string{"python", "javascript", "bash", "go"},
}
```

## Error Handling

This tool always returns a successful response with the configured languages list. It does not perform validation of whether the interpreters/compilers are actually installed on the system.

To verify if a language is actually executable, use the `code_exec_run` tool with a simple test:

```go
// Test if Python is actually available
result, err := codeExecRunTool.InvokableRun(ctx, `{
    "language": "python",
    "code": "print('test')",
    "timeout": 5000
}`)

if err != nil {
    fmt.Println("Python is not available on this system")
}
```

## Integration with Other Tools

The `code_exec_supported_languages` tool works in conjunction with:

1. **code_exec_run** - Execute code in supported languages
2. **code_exec_format** - Format code in supported languages
3. **code_exec_validate** - Validate code syntax in supported languages

### Workflow Example

```go
// 1. Get supported languages
langResult, _ := supportedLangsTool.InvokableRun(ctx, "{}")

// 2. Choose a language
language := "python"

// 3. Execute code
execResult, _ := runTool.InvokableRun(ctx, `{
    "language": "python",
    "code": "print('Hello, World!')",
    "timeout": 5000
}`)

// 4. Format code
formatResult, _ := formatTool.InvokableRun(ctx, `{
    "language": "python",
    "code": "x=1+2\nprint(x)"
}`)
```

## Testing

The tool includes comprehensive tests:

```bash
# Run tests
go test -v -run TestCodeExecSupportedLanguagesTool ./agent/aiosandbox/code_exec/

# Run with coverage
go test -cover -run TestCodeExecSupportedLanguagesTool ./agent/aiosandbox/code_exec/
```

Test coverage includes:
- Tool info retrieval
- Basic invocation
- JSON response validation
- Multiple language configurations
- Empty configuration handling

## Performance

The `code_exec_supported_languages` tool is extremely lightweight:
- **Execution Time**: < 1ms
- **Memory Usage**: Negligible
- **No External Dependencies**: Pure in-memory operation

The tool simply returns the configured language list without performing any system checks or external calls.

## Security Considerations

1. **No System Access**: The tool only returns configuration data
2. **No Code Execution**: No code is executed when calling this tool
3. **No Side Effects**: The tool is read-only and has no side effects
4. **Safe for Public APIs**: Can be safely exposed in public-facing APIs

## Limitations

1. **Configuration-Based**: Returns configured languages, not necessarily installed ones
2. **No Runtime Validation**: Does not check if interpreters/compilers are actually available
3. **Static List**: The list is determined at module initialization and doesn't change dynamically

## Future Enhancements

Potential improvements for future versions:

1. **Runtime Detection**: Automatically detect installed interpreters/compilers
2. **Version Information**: Include language version information
3. **Capability Flags**: Indicate which features are available for each language
4. **Dynamic Updates**: Support for adding/removing languages at runtime

## Related Documentation

- [Code Executor Module Overview](./example_usage.md)
- [code_exec_run Tool](./example_usage.md#example-2-execute-python-code)
- [code_exec_format Tool](./example_usage.md#example-10-format-code)
- [Design Document](../../../.kiro/specs/aio-sandbox-completion/design.md)

## Support

For issues or questions:
1. Check the [example_usage.md](./example_usage.md) for common patterns
2. Review the [design document](../../../.kiro/specs/aio-sandbox-completion/design.md)
3. Run the test suite to verify functionality
4. Check the integration example in [integration_test_example.go](./integration_test_example.go)
