# Common Packages (`pkg`)

This directory contains shared libraries and utilities that are **not specific** to the application's domain logic. Code in this directory can theoretically be reused in other projects.

## Contents
- **`logger/`**: Structured logging wrapper (e.g., using `zap` or `zerolog`). Provides a standard interface for logging across the app.
- **`postgres/`**: Database connection initialization and configuration helpers.
- **`httpserver/`**: Wrapper around the HTTP server for graceful shutdown and configuration.
- **`jwt/`**: Utilities for parsing and generating JSON Web Tokens.
- **`validator/`**: Custom validation logic and error translation helpers (using `go-playground/validator`).
- **`errors/`**: Common error types and error handling utilities.

## Rule of Thumb
If a piece of code references `internal/domain`, it belongs in `internal`. If it is a generic tool (like a logger or a string helper) that could be published as a separate library, it belongs in `pkg`.
