# Code Executor Module - Example Usage

## Overview

The `code_exec_run` MCP tool allows executing code in different programming languages within a sandboxed environment.

## Supported Languages

- Python
- JavaScript (Node.js)
- Bash/Shell
- Go

## Available Tools

### 1. code_exec_run

Execute code in a specified programming language.

**Parameters:**
- `language` (string, required): Programming language (python, javascript, bash, go)
- `code` (string, required): Code to execute
- `input` (string, optional): Optional input to pass to the code (stdin)
- `timeout` (integer, optional): Execution timeout in milliseconds (defaults to config value)

**Returns:**
```json
{
  "success": true/false,
  "language": "python",
  "output": "stdout output",
  "error": "stderr output or error message",
  "exit_code": 0,
  "duration": 123,
  "memory_mb": 45
}
```

### 2. code_exec_supported_languages

Get the list of supported programming languages.

**Parameters:**
- None (empty object `{}`)

**Returns:**
```json
{
  "success": true,
  "languages": ["python", "javascript", "bash", "go"]
}
```

### 3. code_exec_format

Format code in a specified programming language.

**Parameters:**
- `language` (string, required): Programming language
- `code` (string, required): Code to format

**Returns:**
```json
{
  "success": true/false,
  "language": "python",
  "formatted_code": "formatted code string",
  "error": "error message if failed"
}
```

## Usage Examples

### Example 1: Get Supported Languages

```json
{}
```

**Response:**
```json
{
  "success": true,
  "languages": ["python", "javascript", "bash", "go"]
}
```

**Use Case:** Before executing code, check which languages are available in the current configuration.

### Example 2: Execute Python Code

```json
{
  "language": "python",
  "code": "print('Hello, World!')",
  "timeout": 5000
}
```

**Response:**
```json
{
  "success": true,
  "language": "python",
  "output": "Hello, World!\n",
  "error": "",
  "exit_code": 0,
  "duration": 234,
  "memory_mb": 0
}
```

### Example 3: Execute JavaScript Code

```json
{
  "language": "javascript",
  "code": "console.log('Hello from Node.js'); console.log(2 + 2);",
  "timeout": 5000
}
```

**Response:**
```json
{
  "success": true,
  "language": "javascript",
  "output": "Hello from Node.js\n4\n",
  "error": "",
  "exit_code": 0,
  "duration": 156,
  "memory_mb": 0
}
```

### Example 4: Execute Bash Command

```json
{
  "language": "bash",
  "code": "echo 'Current directory:'; pwd",
  "timeout": 5000
}
```

**Response:**
```json
{
  "success": true,
  "language": "bash",
  "output": "Current directory:\n/path/to/dir\n",
  "error": "",
  "exit_code": 0,
  "duration": 45,
  "memory_mb": 0
}
```

### Example 5: Execute Go Code

```json
{
  "language": "go",
  "code": "package main\nimport \"fmt\"\nfunc main() {\n    fmt.Println(\"Hello from Go!\")\n}",
  "timeout": 10000
}
```

**Response:**
```json
{
  "success": true,
  "language": "go",
  "output": "Hello from Go!\n",
  "error": "",
  "exit_code": 0,
  "duration": 1234,
  "memory_mb": 0
}
```

### Example 6: Handle Syntax Error

```json
{
  "language": "python",
  "code": "print('unclosed string",
  "timeout": 5000
}
```

**Response:**
```json
{
  "success": false,
  "language": "python",
  "output": "",
  "error": "SyntaxError: unterminated string literal...",
  "exit_code": 1,
  "duration": 123,
  "memory_mb": 0
}
```

### Example 7: Handle Timeout

```json
{
  "language": "python",
  "code": "import time; time.sleep(10)",
  "timeout": 1000
}
```

**Response:**
```json
{
  "success": false,
  "language": "python",
  "output": "",
  "error": "signal: killed",
  "exit_code": -1,
  "duration": 1001,
  "memory_mb": 0
}
```

### Example 8: Unsupported Language

```json
{
  "language": "ruby",
  "code": "puts 'Hello'",
  "timeout": 5000
}
```

**Response:**
```json
{
  "success": false,
  "error": "Language not supported"
}
```

### Example 9: With Input Parameter

```json
{
  "language": "python",
  "code": "name = 'World'\nprint(f'Hello, {name}!')",
  "input": "",
  "timeout": 5000
}
```

**Response:**
```json
{
  "success": true,
  "language": "python",
  "output": "Hello, World!\n",
  "error": "",
  "exit_code": 0,
  "duration": 234,
  "memory_mb": 0
}
```

### Example 10: Format Code

```json
{
  "language": "python",
  "code": "x=1+2\nprint(x)"
}
```

**Response:**
```json
{
  "success": true,
  "language": "python",
  "formatted_code": "x = 1 + 2\nprint(x)\n"
}
```

## Security Features

1. **Sandboxed Execution**: Code runs in isolated environment
2. **Resource Limits**: CPU, memory, and time limits enforced
3. **Timeout Control**: Automatic termination on timeout
4. **Language Restrictions**: Only whitelisted languages allowed
5. **Error Handling**: Safe error capture and reporting

## Configuration

The module can be configured with:

```go
config := CodeExecutorConfig{
    Timeout:            60000,  // 60 seconds default
    MemoryLimit:        512,    // 512MB
    CPULimit:           2,      // 2 cores
    SupportedLanguages: []string{"python", "javascript", "bash", "go"},
}
```

## Integration with Eino MCP Framework

The tool follows the Eino MCP tool framework pattern:

```go
// Get tools from module
tools, err := module.GetTools(ctx)

// Use the tool
tool := tools[0] // code_exec_run
result, err := tool.InvokableRun(ctx, jsonInput)
```

## Notes

- The `input` parameter is available for future stdin support
- Memory usage tracking may not be fully implemented for all languages
- Go code execution requires Go compiler to be installed
- Python execution uses `-I` flag for isolation
- JavaScript execution requires Node.js to be installed
