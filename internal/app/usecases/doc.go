// Package usecases contains the application's business use cases.
//
// Use cases orchestrate the flow of data between the domain layer and external layers,
// coordinating domain entities, services, and repositories to fulfill specific
// application requirements. They represent the "application-specific business rules"
// as defined in Clean Architecture.
//
// Key responsibilities:
//   - Orchestrate domain entities and services to accomplish user goals
//   - Coordinate repository calls and handle transactions
//   - Transform between DTOs and domain entities
//   - Enforce application-level business rules and workflows
//   - Handle cross-cutting concerns (logging, validation, authorization context)
//
// Use cases should remain independent of delivery mechanisms (HTTP, gRPC, CLI)
// and infrastructure concerns (databases, external APIs). They depend only on
// domain abstractions (ports/interfaces) rather than concrete implementations.
package usecases
