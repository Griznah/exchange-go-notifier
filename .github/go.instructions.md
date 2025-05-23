# 🧠 Copilot Behavior Guide for Go Development

## 🎯 Project Intent
This project is written entirely in Go. Code suggestions must follow idiomatic Go best practices and prioritize clarity, correctness, and simplicity.

## ✅ Coding Guidelines

- Use **standard library** whenever possible.
- Follow the **Go Code Review Comments**: https://github.com/golang/go/wiki/CodeReviewComments
- Avoid excessive abstraction, interfaces, or overengineering.
- Write **idiomatic error handling** (explicit errors, no panics).
- Write code that is compatible with **Go modules** and supports `go test` and `go build`.
- Always comment each function and package with a clear purpose.

## 🧪 Testing Instructions

- Use **table-driven tests** with the standard `testing` package.
- Use `t.Run()` for subtests.
- Avoid third-party test frameworks unless necessary.

### Example:
```go
func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"simple", 1, 2, 3},
		{"zero", 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Add(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Add() = %d, want %d", got, tt.want)
			}
		})
	}
}
