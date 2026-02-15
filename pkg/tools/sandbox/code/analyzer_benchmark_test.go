// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package code

import (
	"strings"
	"testing"
)

// BenchmarkAnalyzer_SmallFile benchmarks analyzer performance on small files (< 100 lines)
func BenchmarkAnalyzer_SmallFile(b *testing.B) {
	analyzer := NewCodeAnalyzer()

	// Small Python code (50 lines)
	code := `
import requests
import hashlib
import os

def process_data(data):
    password = "secret123"
    response = requests.get("http://api.example.com")
    hash = hashlib.md5(response.content).hexdigest()
    
    result = ""
    for i in range(10):
        result += str(i)
    
    return result

def main():
    data = process_data("test")
    print(data)

if __name__ == "__main__":
    main()
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.Analyze("python", code)
	}
}

// BenchmarkAnalyzer_MediumFile benchmarks analyzer performance on medium files (100-500 lines)
func BenchmarkAnalyzer_MediumFile(b *testing.B) {
	analyzer := NewCodeAnalyzer()

	// Generate medium-sized code (200 lines)
	var sb strings.Builder
	sb.WriteString("import requests\nimport hashlib\nimport os\n\n")

	for i := 0; i < 40; i++ {
		sb.WriteString(`
def function_` + string(rune('a'+i%26)) + `(data):
    password = "secret123"
    response = requests.get("http://api.example.com")
    hash = hashlib.md5(response.content).hexdigest()
    return hash
`)
	}

	code := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.Analyze("python", code)
	}
}

// BenchmarkAnalyzer_LargeFile benchmarks analyzer performance on large files (> 500 lines)
func BenchmarkAnalyzer_LargeFile(b *testing.B) {
	analyzer := NewCodeAnalyzer()

	// Generate large code (1000 lines)
	var sb strings.Builder
	sb.WriteString("import requests\nimport hashlib\nimport os\nimport socket\n\n")

	for i := 0; i < 200; i++ {
		sb.WriteString(`
def function_` + string(rune('a'+i%26)) + `(data):
    password = "secret123"
    response = requests.get("http://api.example.com")
    hash = hashlib.md5(response.content).hexdigest()
    return hash
`)
	}

	code := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.Analyze("python", code)
	}
}

// BenchmarkAnalyzer_ParallelSmall benchmarks parallel analysis on small files
func BenchmarkAnalyzer_ParallelSmall(b *testing.B) {
	analyzer := NewCodeAnalyzer()

	code := `
import requests
import hashlib

def process_data(data):
    password = "secret123"
    response = requests.get("http://api.example.com")
    hash = hashlib.md5(response.content).hexdigest()
    return hash
`

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			analyzer.Analyze("python", code)
		}
	})
}

// BenchmarkAnalyzer_ParallelLarge benchmarks parallel analysis on large files
func BenchmarkAnalyzer_ParallelLarge(b *testing.B) {
	analyzer := NewCodeAnalyzer()

	// Generate large code (1000 lines)
	var sb strings.Builder
	sb.WriteString("import requests\nimport hashlib\nimport os\nimport socket\n\n")

	for i := 0; i < 200; i++ {
		sb.WriteString(`
def function_` + string(rune('a'+i%26)) + `(data):
    password = "secret123"
    response = requests.get("http://api.example.com")
    hash = hashlib.md5(response.content).hexdigest()
    return hash
`)
	}

	code := sb.String()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			analyzer.Analyze("python", code)
		}
	})
}

// BenchmarkAnalyzer_ComplexDetection benchmarks complex detection patterns
func BenchmarkAnalyzer_ComplexDetection(b *testing.B) {
	analyzer := NewCodeAnalyzer()

	code := `
import requests
import hashlib
import os
import socket
import subprocess

def complex_function(data):
    # Network operations
    response = requests.get("http://api.example.com")
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    
    # Crypto issues
    password = "secret123"
    hash = hashlib.md5(response.content).hexdigest()
    
    # File operations
    with open('/etc/passwd', 'r') as f:
        data = f.read()
    
    # Process operations
    subprocess.run(['ls', '-la'])
    os.system('rm -rf /tmp/data')
    
    # Database operations
    cursor.execute("SELECT * FROM users WHERE id = %s" % user_id)
    
    # Quality issues
    myVariable = 10
    result = ""
    for i in range(100):
        result += "text"
    
    try:
        pass
    except:
        pass
    
    return result
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.Analyze("python", code)
	}
}
