// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the requirements of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

// docgen is a tool to generate API documentation from Go code comments
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// APIInfo contains information about an API
 type APIInfo struct {
	Type     string
	Name     string
	Comment  string
	Methods  []MethodInfo
	Fields   []FieldInfo
}

// MethodInfo contains information about a method
 type MethodInfo struct {
	Name     string
	Comment  string
	Signature string
}

// FieldInfo contains information about a struct field
 type FieldInfo struct {
	Name     string
	Type     string
	Comment  string
	Tags     string
}

// DocGenerator generates documentation from Go code
 type DocGenerator struct {
	outputDir string
	packages  []string
}

// NewDocGenerator creates a new DocGenerator
 func NewDocGenerator(outputDir string, packages []string) *DocGenerator {
	return &DocGenerator{
		outputDir: outputDir,
		packages:  packages,
	}
}

// Generate generates documentation for the specified packages
 func (d *DocGenerator) Generate() error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(d.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, pkgPath := range d.packages {
		if err := d.generatePackageDocs(pkgPath); err != nil {
			return fmt.Errorf("failed to generate docs for package %s: %w", pkgPath, err)
		}
	}

	return nil
}

// generatePackageDocs generates documentation for a single package
 func (d *DocGenerator) generatePackageDocs(pkgPath string) error {
	// Parse the package
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkgPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse package %s: %w", pkgPath, err)
	}

	for pkgName, pkg := range pkgs {
		// Extract API information
		apis := d.extractAPIs(pkg, fset)

		// Generate markdown documentation
		md := d.generateMarkdown(pkgName, apis)

		// Write to file
		outputFile := filepath.Join(d.outputDir, fmt.Sprintf("%s.md", pkgName))
		if err := os.WriteFile(outputFile, []byte(md), 0644); err != nil {
			return fmt.Errorf("failed to write documentation for package %s: %w", pkgName, err)
		}

		fmt.Printf("Generated documentation for package %s at %s\n", pkgName, outputFile)
	}

	return nil
}

// extractAPIs extracts API information from the package
 func (d *DocGenerator) extractAPIs(pkg *ast.Package, fset *token.FileSet) []APIInfo {
	var apis []APIInfo

	// Iterate through all files in the package
	for _, file := range pkg.Files {
		// Extract type declarations
		for _, decl := range file.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						// Check if it's an interface or struct
						var api APIInfo
						api.Name = typeSpec.Name.Name

						// Extract comment
						if genDecl.Doc != nil {
							api.Comment = genDecl.Doc.Text()
						} else if typeSpec.Comment != nil {
							api.Comment = typeSpec.Comment.Text()
						}

						// Check if it's an interface
						if ifaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
							api.Type = "interface"
							api.Methods = d.extractMethods(ifaceType, fset)
							apis = append(apis, api)
						} else if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							api.Type = "struct"
							api.Fields = d.extractFields(structType)
							// Check if it has methods
							api.Methods = d.extractStructMethods(file, api.Name)
							apis = append(apis, api)
						}
					}
				}
			}
		}
	}

	return apis
}

// extractMethods extracts methods from an interface
 func (d *DocGenerator) extractMethods(iface *ast.InterfaceType, fset *token.FileSet) []MethodInfo {
	var methods []MethodInfo

	if iface.Methods != nil {
		for _, field := range iface.Methods.List {
			if len(field.Names) > 0 {
				method := MethodInfo{
					Name: field.Names[0].Name,
				}

				// Extract comment
				if field.Doc != nil {
					method.Comment = field.Doc.Text()
				} else if field.Comment != nil {
					method.Comment = field.Comment.Text()
				}

				// Extract signature
				method.Signature = d.extractSignature(field.Type)

				methods = append(methods, method)
			}
		}
	}

	return methods
}

// extractStructMethods extracts methods from a struct in the file
 func (d *DocGenerator) extractStructMethods(file *ast.File, structName string) []MethodInfo {
	var methods []MethodInfo

	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
				recvType := funcDecl.Recv.List[0].Type
				if starExpr, ok := recvType.(*ast.StarExpr); ok {
					recvType = starExpr.X
				}
				if ident, ok := recvType.(*ast.Ident); ok && ident.Name == structName {
					method := MethodInfo{
						Name: funcDecl.Name.Name,
					}

					// Extract comment
					if funcDecl.Doc != nil {
						method.Comment = funcDecl.Doc.Text()
					}

					// Extract signature
					method.Signature = d.extractFunctionSignature(funcDecl)

					methods = append(methods, method)
				}
			}
		}
	}

	return methods
}

// extractFields extracts fields from a struct
 func (d *DocGenerator) extractFields(structType *ast.StructType) []FieldInfo {
	var fields []FieldInfo

	if structType.Fields != nil {
		for _, field := range structType.Fields.List {
			fieldInfo := FieldInfo{}

			// Extract field names
			if len(field.Names) > 0 {
				fieldInfo.Name = field.Names[0].Name
			} else {
				// Embedded field
				fieldInfo.Name = d.extractTypeName(field.Type)
			}

			// Extract field type
			fieldInfo.Type = d.extractTypeName(field.Type)

			// Extract comment
			if field.Doc != nil {
				fieldInfo.Comment = field.Doc.Text()
			} else if field.Comment != nil {
				fieldInfo.Comment = field.Comment.Text()
			}

			// Extract tags
			if field.Tag != nil {
				fieldInfo.Tags = field.Tag.Value
			}

			fields = append(fields, fieldInfo)
		}
	}

	return fields
}

// extractTypeName extracts the name of a type
 func (d *DocGenerator) extractTypeName(expr ast.Expr) string {
	var sb strings.Builder

	switch e := expr.(type) {
	case *ast.Ident:
		sb.WriteString(e.Name)
	case *ast.StarExpr:
		sb.WriteString("*")
		sb.WriteString(d.extractTypeName(e.X))
	case *ast.SelectorExpr:
		sb.WriteString(d.extractTypeName(e.X))
		sb.WriteString(".")
		sb.WriteString(e.Sel.Name)
	case *ast.ArrayType:
		sb.WriteString("[]")
		sb.WriteString(d.extractTypeName(e.Elt))
	case *ast.MapType:
		sb.WriteString("map[")
		sb.WriteString(d.extractTypeName(e.Key))
		sb.WriteString("]")
		sb.WriteString(d.extractTypeName(e.Value))
	case *ast.FuncType:
		sb.WriteString(d.extractSignature(e))
	}

	return sb.String()
}

// extractSignature extracts the signature of a function or method
 func (d *DocGenerator) extractSignature(expr ast.Expr) string {
	var sb strings.Builder

	if funcType, ok := expr.(*ast.FuncType); ok {
		sb.WriteString("func ")
		if funcType.Params != nil {
			sb.WriteString("(")
			for i, param := range funcType.Params.List {
				if i > 0 {
					sb.WriteString(", ")
				}
				var names []string
				for _, name := range param.Names {
					names = append(names, name.Name)
				}
				if len(names) == 0 {
					// If no names, just write the type
					sb.WriteString(d.extractTypeName(param.Type))
				} else {
					sb.WriteString(strings.Join(names, ", "))
					sb.WriteString(" ")
					sb.WriteString(d.extractTypeName(param.Type))
				}
			}
			sb.WriteString(")")
		}
		if funcType.Results != nil && len(funcType.Results.List) > 0 {
			if len(funcType.Results.List) == 1 && len(funcType.Results.List[0].Names) == 0 {
				// Single unnamed return value
				sb.WriteString(" ")
				sb.WriteString(d.extractTypeName(funcType.Results.List[0].Type))
			} else {
				// Multiple return values or named return values
				sb.WriteString(" (")
				for i, result := range funcType.Results.List {
					if i > 0 {
						sb.WriteString(", ")
					}
					var names []string
					for _, name := range result.Names {
						names = append(names, name.Name)
					}
					if len(names) > 0 {
						sb.WriteString(strings.Join(names, ", "))
						sb.WriteString(" ")
					}
					sb.WriteString(d.extractTypeName(result.Type))
				}
				sb.WriteString(")")
			}
		}
	}

	return sb.String()
}

// extractFunctionSignature extracts the signature of a function declaration
 func (d *DocGenerator) extractFunctionSignature(funcDecl *ast.FuncDecl) string {
	var sb strings.Builder

	sb.WriteString("func ")
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		sb.WriteString("(")
		for i, recv := range funcDecl.Recv.List {
			if i > 0 {
				sb.WriteString(", ")
			}
			var names []string
			for _, name := range recv.Names {
				names = append(names, name.Name)
			}
			if len(names) > 0 {
				sb.WriteString(strings.Join(names, ", "))
				sb.WriteString(" ")
			}
			sb.WriteString(d.extractTypeName(recv.Type))
		}
		sb.WriteString(") ")
	}
	sb.WriteString(funcDecl.Name.Name)
	if funcDecl.Type.Params != nil {
		sb.WriteString("(")
		for i, param := range funcDecl.Type.Params.List {
			if i > 0 {
				sb.WriteString(", ")
			}
			var names []string
			for _, name := range param.Names {
				names = append(names, name.Name)
			}
			if len(names) == 0 {
				// If no names, just write the type
				sb.WriteString(d.extractTypeName(param.Type))
			} else {
				sb.WriteString(strings.Join(names, ", "))
				sb.WriteString(" ")
				sb.WriteString(d.extractTypeName(param.Type))
			}
		}
		sb.WriteString(")")
	}
	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
		if len(funcDecl.Type.Results.List) == 1 && len(funcDecl.Type.Results.List[0].Names) == 0 {
			// Single unnamed return value
			sb.WriteString(" ")
			sb.WriteString(d.extractTypeName(funcDecl.Type.Results.List[0].Type))
		} else {
			// Multiple return values or named return values
			sb.WriteString(" (")
			for i, result := range funcDecl.Type.Results.List {
				if i > 0 {
					sb.WriteString(", ")
				}
				var names []string
				for _, name := range result.Names {
					names = append(names, name.Name)
				}
				if len(names) > 0 {
					sb.WriteString(strings.Join(names, ", "))
					sb.WriteString(" ")
				}
				sb.WriteString(d.extractTypeName(result.Type))
			}
			sb.WriteString(")")
		}
	}

	return sb.String()
}

// generateMarkdown generates markdown documentation for the APIs
 func (d *DocGenerator) generateMarkdown(pkgName string, apis []APIInfo) string {
	var sb strings.Builder

	// Add header
	sb.WriteString(fmt.Sprintf("# %s Package API Documentation\n\n", pkgName))
	sb.WriteString("This document contains API documentation for the package.\n\n")

	// Group APIs by type
	interfaces := []APIInfo{}
	structs := []APIInfo{}

	for _, api := range apis {
		if api.Type == "interface" {
			interfaces = append(interfaces, api)
		} else if api.Type == "struct" {
			structs = append(structs, api)
		}
	}

	// Add interfaces
	if len(interfaces) > 0 {
		sb.WriteString("## Interfaces\n\n")
		for _, api := range interfaces {
			sb.WriteString(fmt.Sprintf("### %s\n\n", api.Name))
			if api.Comment != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", strings.TrimSpace(api.Comment)))
			}
			if len(api.Methods) > 0 {
				sb.WriteString("#### Methods\n\n")
				for _, method := range api.Methods {
					sb.WriteString(fmt.Sprintf("##### %s\n\n", method.Name))
					if method.Comment != "" {
						sb.WriteString(fmt.Sprintf("%s\n\n", strings.TrimSpace(method.Comment)))
					}
					sb.WriteString(fmt.Sprintf("```go\n%s\n```\n\n", method.Signature))
				}
			}
		}
	}

	// Add structs
	if len(structs) > 0 {
		sb.WriteString("## Structs\n\n")
		for _, api := range structs {
			sb.WriteString(fmt.Sprintf("### %s\n\n", api.Name))
			if api.Comment != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", strings.TrimSpace(api.Comment)))
			}
			if len(api.Fields) > 0 {
				sb.WriteString("#### Fields\n\n")
				sb.WriteString("| Name | Type | Tags | Comment |\n")
				sb.WriteString("|------|------|------|---------|\n")
				for _, field := range api.Fields {
					tags := field.Tags
					if tags != "" {
						tags = strings.Trim(tags, "`")
					}
					sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
						field.Name,
						field.Type,
						tags,
						strings.TrimSpace(field.Comment)))
				}
				sb.WriteString("\n")
			}
			if len(api.Methods) > 0 {
				sb.WriteString("#### Methods\n\n")
				for _, method := range api.Methods {
					sb.WriteString(fmt.Sprintf("##### %s\n\n", method.Name))
					if method.Comment != "" {
						sb.WriteString(fmt.Sprintf("%s\n\n", strings.TrimSpace(method.Comment)))
					}
					sb.WriteString(fmt.Sprintf("```go\n%s\n```\n\n", method.Signature))
				}
			}
		}
	}

	return sb.String()
}

func main() {
	// Parse command line arguments
	outputDir := flag.String("output", "./docs/api", "Output directory for documentation")
	flag.Parse()

	// Get packages from remaining arguments
	packages := flag.Args()
	if len(packages) == 0 {
		fmt.Println("Please specify at least one package")
		flag.Usage()
		os.Exit(1)
	}

	// Create and run the generator
	generator := NewDocGenerator(*outputDir, packages)
	if err := generator.Generate(); err != nil {
		fmt.Printf("Error generating documentation: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Documentation generation completed successfully")
}
