---
id: com.example.hello
name: Hello World
version: "1.0.0"
description: "A simple hello world skill"
category: "example"
author: "Agent Framework Contributors"
license: "AGPL-3.0-or-later"

triggers:
  - "say hello"
  - "greet me"

prerequisites: []

workflow:
  - id: "step1"
    name: "Say Hello"
    description: "Print a hello world message"
    action: "execute"
    parameters:
      command: "echo \"Hello, World!\""

config:
  verbose: false

metadata:
  tags: ["example", "hello"]
  os: ["darwin", "linux", "windows"]
  disabled: false
---

# Hello World Skill

A simple skill that prints "Hello, World!" when triggered.

## When to use

Use this skill when you want to test the markdown skill system or simply want to see a hello world message.

## Prerequisites

No prerequisites are required.

## Usage Examples

### Example 1: Simple Greeting
User: "Say hello"
Agent: Uses this skill to print "Hello, World!"

### Example 2: Greeting with Name
User: "Greet me"
Agent: Uses this skill to print "Hello, World!"

## Notes

This is a very simple skill for demonstration purposes only.
