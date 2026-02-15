# Code Formatter Tool - Documentation

## Overview

The `code_exec_format` MCP tool provides code formatting capabilities for multiple programming languages. It automatically formats code according to language-specific style guidelines and best practices.

## Supported Languages

| Language   | Formatter Used | Availability |
|------------|----------------|--------------|
| Python     | black / autopep8 | Optional (graceful fallback) |
| JavaScript | prettier       | Optional (graceful fallback) |
| Bash       | shfmt          | Optional (graceful fallback) |
| Go         | gofmt          | Built-in with Go |

## Tool Specification

### Name
`code_exec_format`

### Description
Format code in a specified language using language-specific formatting tools.

### Parameters

| Parameter  | Type   | Required | Description |
|-----------|--------|----------|-------------|
| language  | string | Yes      | Programming language (python, javascript, bash, go) |
| code      | string | Yes      | Code to format |

### Return Value

```json
{
  "success": true/false,
  "language": "python",
  "formatted_code": "formatted code string",
  "error": "error message if failed"
}
```

## Usage Examples

### Example 1: Format Python Code

**Input:**
```json
{
  "language": "python",
  "code": "x=1+2\ny=x*3\nprint(y)"
}
```

**Output:**
```json
{
  "success": true,
  "language": "python",
  "formatted_code": "x = 1 + 2\ny = x * 3\nprint(y)\n"
}
```

### Example 2: Format JavaScript Code

**Input:**
```json
{
  "language": "javascript",
  "code": "const x=1+2;console.log(x);"
}
```

**Output:**
```json
{
  "success": true,
  "language": "javascript",
  "formatted_code": "const x = 1 + 2;\nconsole.log(x);\n"
}
```

### Example 3: Format Bash Script

**Input:**
```json
{
  "language": "bash",
  "code": "if [ -f file.txt ]; then echo 'exists'; fi"
}
```

**Output:**
```json
{
  "success": true,
  "language": "bash",
  "formatted_code": "if [ -f file.txt ]; then\n\techo 'exists'\nfi\n"
}
```

### Example 4: Format Go Code

**Input:**
```json
{
  "language": "go",
  "code": "package main\nfunc main(){println(\"hello\")}"
}
```

**Output:**
```json
{
  "success": true,
  "language": "go",
  "formatted_code": "package main\n\nfunc main() { println(\"hello\") }\n"
}
```

### Example 5: Unsupported Language

**Input:**
```json
{
  "language": "ruby",
  "code": "puts 'hello'"
}
```

**Output:**
```json
{
  "success": false,
  "error": "Language not supported"
}
```

### Example 6: Missing Parameters

**Input:**
```json
{
  "language": "python"
}
```

**Output:**
```json
{
  "success": true,
  "language": "python",
  "formatted_code": ""
}
```

## Formatter Details

### Python Formatting

The Python formatter tries the following tools in order:

1. **black** - The uncompromising Python code formatter
   - Command: `black --quiet <file>`
   - Style: PEP 8 compliant with opinionated defaults
   - Installation: `pip install black`

2. **autopep8** - Automatically formats Python code to conform to PEP 8
   - Command: `autopep8 --in-place <file>`
   - Style: PEP 8 compliant
   - Installation: `pip install autopep8`

3. **Fallback** - Returns original code if no formatter is available

### JavaScript Formatting

The JavaScript formatter uses:

1. **prettier** - Opinionated code formatter
   - Command: `prettier --parser babel --stdin-filepath code.js`
   - Style: Consistent, opinionated formatting
   - Installation: `npm install -g prettier`

2. **Fallback** - Returns original code if prettier is not available

### Bash Formatting

The Bash formatter uses:

1. **shfmt** - A shell parser, formatter, and interpreter
   - Command: `shfmt -`
   - Style: Consistent shell script formatting
   - Installation: `go install mvdan.cc/sh/v3/cmd/shfmt@latest`

2. **Fallback** - Returns original code if shfmt is not available

### Go Formatting

The Go formatter uses:

1. **gofmt** - Go's built-in formatter
   - Command: `gofmt <file>`
   - Style: Official Go formatting style
   - Installation: Included with Go installation

2. **Fallback** - Returns original code if gofmt fails

## Behavior and Guarantees

### Graceful Degradation

The formatter is designed to **never fail** due to missing formatting tools:

- If a formatter is not installed, the tool returns the original code unchanged
- If formatting fails (e.g., syntax error), the tool returns the original code
- The `success` field is always `true` unless the language is not supported

### Semantic Preservation

The formatter **preserves code semantics**:

- Only whitespace, indentation, and style are changed
- Logic and behavior remain identical
- Tests verify that formatted code produces the same output as original code

### Idempotency

Formatting is **idempotent**:

- Formatting already-formatted code produces the same result
- Multiple format operations converge to the same output

## Error Handling

### Language Not Supported

```json
{
  "success": false,
  "error": "Language not supported"
}
```

This occurs when:
- The language is not in the supported languages list
- The language parameter is missing or empty

### Formatter Not Available

```json
{
  "success": true,
  "language": "python",
  "formatted_code": "original code unchanged"
}
```

This occurs when:
- The formatting tool is not installed
- The formatting tool fails to execute
- The code has syntax errors that prevent formatting

## Integration Examples

### Go Integration

```go
// Create module
config := CodeExecutorConfig{
    SupportedLanguages: []string{"python", "javascript", "bash", "go"},
}
module, err := NewCodeExecutorModule(config)
if err != nil {
    log.Fatal(err)
}
defer module.Close()

// Get tools
tools, err := module.GetTools(context.Background())
if err != nil {
    log.Fatal(err)
}

// Find format tool
var formatTool tool.BaseTool
for _, t := range tools {
    info, _ := t.Info(context.Background())
    if info.Name == "code_exec_format" {
        formatTool = t
        break
    }
}

// Use format tool
input := `{"language":"python","code":"x=1+2\nprint(x)"}`
result, err := formatTool.InvokableRun(context.Background(), input)
if err != nil {
    log.Fatal(err)
}

fmt.Println(result)
```

### Direct Runner Usage

```go
// Create runner
runner := NewPythonRunner(config, tempDir)

// Format code
code := "x=1+2\nprint(x)"
formatted, err := runner.Format(code)
if err != nil {
    log.Fatal(err)
}

fmt.Println(formatted)
```

## Performance Considerations

### Execution Time

Formatting times vary by language and formatter:

- **Go (gofmt)**: ~50-200ms (fast, built-in)
- **Python (black)**: ~100-500ms (moderate, external tool)
- **JavaScript (prettier)**: ~100-400ms (moderate, external tool)
- **Bash (shfmt)**: ~50-200ms (fast, external tool)

### Temporary Files

Some formatters require temporary files:

- **Python**: Creates temporary `.py` file
- **Go**: Creates temporary `.go` file
- **JavaScript**: Uses stdin (no file needed)
- **Bash**: Uses stdin (no file needed)

Temporary files are automatically cleaned up after formatting.

### Caching

Currently, formatting results are **not cached**. Each format request:

1. Creates necessary temporary files
2. Invokes the formatter
3. Reads the result
4. Cleans up temporary files

Future versions may add caching for improved performance.

## Best Practices

### 1. Check Language Support First

```json
// First, get supported languages
{}

// Then format code
{
  "language": "python",
  "code": "..."
}
```

### 2. Handle Missing Formatters Gracefully

```go
result, err := formatTool.InvokableRun(ctx, input)
if err != nil {
    // Handle error
}

// Check if formatting succeeded
var response map[string]interface{}
json.Unmarshal([]byte(result), &response)

if response["success"].(bool) {
    formatted := response["formatted_code"].(string)
    // Use formatted code
} else {
    // Use original code or show error
}
```

### 3. Validate Code Before Formatting

While formatters handle syntax errors gracefully, it's better to validate first:

```json
// Validate first (future feature)
{
  "language": "python",
  "code": "..."
}

// Then format
{
  "language": "python",
  "code": "..."
}
```

### 4. Use Appropriate Timeouts

Formatting is usually fast, but set reasonable timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := formatTool.InvokableRun(ctx, input)
```

## Testing

The format tool includes comprehensive tests:

### Unit Tests

- `TestPythonRunner_Format`: Tests Python formatting
- `TestLanguageRunners_Format`: Tests all language runners
- `TestCodeExecFormatTool_InvokableRun`: Tests MCP tool interface

### Integration Tests

- `TestFormatTool_AllLanguages`: Tests formatting for all languages
- `TestFormatTool_ErrorCases`: Tests error handling
- `TestFormatTool_PreservesCodeSemantics`: Verifies semantic preservation

### Running Tests

```bash
# Run all format tests
go test -v -run TestFormat ./agent/aiosandbox/code_exec/

# Run specific test
go test -v -run TestFormatTool_AllLanguages ./agent/aiosandbox/code_exec/

# Run with coverage
go test -cover ./agent/aiosandbox/code_exec/
```

## Troubleshooting

### Formatter Not Found

**Problem**: Formatting returns original code unchanged

**Solution**: Install the required formatter:

```bash
# Python
pip install black
# or
pip install autopep8

# JavaScript
npm install -g prettier

# Bash
go install mvdan.cc/sh/v3/cmd/shfmt@latest

# Go
# Included with Go installation
```

### Formatting Fails Silently

**Problem**: Code is not formatted but no error is shown

**Possible Causes**:
1. Formatter not installed (graceful fallback)
2. Code has syntax errors
3. Formatter doesn't support the code style

**Solution**: Check formatter availability and code syntax

### Unexpected Formatting

**Problem**: Formatted code looks different than expected

**Explanation**: Each formatter has its own style:
- **black**: Very opinionated, may reformat significantly
- **prettier**: Opinionated, consistent style
- **gofmt**: Official Go style
- **shfmt**: Standard shell script style

**Solution**: This is expected behavior. Formatters enforce consistent style.

## Future Enhancements

Planned improvements:

1. **Caching**: Cache formatting results for identical code
2. **Custom Styles**: Support custom formatting configurations
3. **More Languages**: Add support for Rust, Java, C++, etc.
4. **Validation**: Integrate syntax validation before formatting
5. **Diff Output**: Show differences between original and formatted code
6. **Batch Formatting**: Format multiple files at once

## Related Tools

- `code_exec_run`: Execute code in various languages
- `code_exec_supported_languages`: Get list of supported languages
- `code_exec_validate`: Validate code syntax (future)

## References

- [black documentation](https://black.readthedocs.io/)
- [prettier documentation](https://prettier.io/)
- [shfmt documentation](https://github.com/mvdan/sh)
- [gofmt documentation](https://golang.org/cmd/gofmt/)
- [Eino MCP Framework](https://github.com/cloudwego/eino)
