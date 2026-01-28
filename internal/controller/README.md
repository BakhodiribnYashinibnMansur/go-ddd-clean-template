# Controller Layer (`internal/controller`)

The **Controller** layer (Transport Layer) allows the outside world to interact with the application.

## Responsibilities
- **Input Handling**: Parses incoming HTTP requests (JSON body, query params, headers).
- **Validation**: Performs initial structural validation of the input (using `validator` tags).
- **Delegation**: Calls the appropriate method in the `UseCase` layer.
- **Response Formatting**: Formats the result from the UseCase into a standard HTTP response (JSON) and handles errors.

## Structure
- **`restapi/`**: Contains HTTP REST handlers.
    - `router.go`: Defines the routing paths and links handlers to middlewares.
    - **`v1/`**: Versioned API handlers (e.g., `login`, `register`).
- **`middleware/`**: Interceptors for cross-cutting concerns (Auth, Logging, CORS).

## Framework
This project uses **Fiber** (or similar Go web frameworks) for HTTP routing and context management. However, the handlers try to decouple from the framework where possible to allow easier switching.
