// ============================================================================
// go.mod — Go Module Definition
// ============================================================================
// Module chứa backend Go của Cash Flow Minimizer.
// Dependencies chính:
//   - Fiber v2: HTTP framework hiệu năng cao
//   - lib/pq: PostgreSQL driver
//   - google/uuid: Sinh UUID

module github.com/yourusername/cash-flow-minimizer

go 1.22

require (
	github.com/gofiber/fiber/v2 v2.52.0 // HTTP framework
	github.com/lib/pq v1.10.9           // PostgreSQL driver
	github.com/google/uuid v1.6.0       // UUID generator
)