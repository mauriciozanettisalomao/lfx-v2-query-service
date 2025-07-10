# Request ID Middleware

This middleware adds a unique request ID to each HTTP request for distributed tracing and logging correlation.

## Features

- **Automatic Request ID Generation**: Generates a UUID for each request if not provided
- **Header Propagation**: Accepts and returns request IDs via `X-Request-ID` header
- **Context Integration**: Stores request ID in context for use throughout the request lifecycle
- **Structured Logging**: Provides logging functions that automatically include request ID

## Usage

### Basic Setup

```go
import "github.com/linuxfoundation/lfx-v2-query-service/internal/middleware"

// Add to HTTP handler chain
handler = middleware.RequestIDMiddleware()(handler)
```

### Logging with Request ID

Instead of using `slog` directly, use the middleware's logging functions:

```go
// These functions automatically include the request ID from context
middleware.InfoContext(ctx, "processing request", "user_id", userID)
middleware.DebugContext(ctx, "query executed", "query", query)
middleware.ErrorContext(ctx, "error occurred", "error", err)
middleware.WarnContext(ctx, "warning message", "details", details)
```

### Extracting Request ID

```go
requestID := middleware.RequestIDFromContext(ctx)
if requestID != "" {
    // Use request ID for manual correlation
}
```

## Request Flow

1. **Incoming Request**: Middleware checks for `X-Request-ID` header
2. **ID Generation**: If no header exists, generates a new UUID
3. **Context Storage**: Stores request ID in request context
4. **Response Header**: Adds `X-Request-ID` to response headers
5. **Logging**: All logging functions include request ID automatically

## Log Output Example

```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "performing resource search",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "committee",
  "type": "committee"
}
```

## Benefits

- **Tracing**: Track requests across distributed systems
- **Debugging**: Correlate logs for specific requests
- **Monitoring**: Identify performance issues by request
- **Compliance**: Audit trails with request correlation

## Testing

Run the middleware tests:

```bash
go test ./internal/middleware/...
```

The tests cover:
- Request ID generation
- Header propagation
- Context integration
- Logging functionality 