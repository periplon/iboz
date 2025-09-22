# Repository Guidelines

## Architectural Principles
- Treat SOLID principles as the baseline for both Go and TypeScript code.
- Preserve the hexagonal architecture already in place:
  - `internal/<domain>` packages hold the core domain logic and should be framework-agnostic.
  - `internal/api` contains inbound HTTP adapters (Echo handlers) that translate requests into domain invocations and map domain responses back into DTOs.
  - Outbound adapters (for persistence, external services, etc.) should live in their own `internal/<context>/adapter` packages and depend on domain-facing ports.
  - The entrypoint in `cmd/server` must only compose dependencies and start the application.
- Keep side-effectful code (I/O, logging, HTTP, persistence) at the edges. Domain services should expose interfaces so they can be unit tested in isolation.
- Prefer dependency injection via constructors or explicit setters over globals. When package-level state is unavoidable, guard it carefully with synchronization primitives.

## Go Service Conventions
- Target Go 1.24 features only when they do not break compatibility with earlier minor releases.
- Use `context.Context` as the first parameter for operations that perform I/O, might block, or can be cancelled.
- Keep request/response structs in the adapter layer; domain types should not import Echo or HTTP packages.
- Organize tests as table-driven cases in `_test.go` files colocated with the code under test. Use fakes that satisfy domain interfaces rather than spinning up servers in unit tests.
- Format Go code with `gofmt` (or `go fmt ./...`) and run `go test ./...` before committing. Add new dependencies via `go get` so `go.mod` and `go.sum` stay in sync.
- Embedded frontend assets under `internal/server/static` are generated—do not edit them manually. Rebuild them from the `web` project when UI changes ship.

## Frontend (React + Vite)
- Use React 19 with TypeScript and keep components typed explicitly. Shared, reusable UI lives in `web/src/components`; route-level shells belong in `web/src/pages`.
- Data fetching hooks should extend the patterns in `web/src/api/hooks.ts` (wrap `useApi` and encapsulate fetch semantics). Avoid calling `fetch` directly from components.
- Favor functional components, React hooks, and Tailwind utility classes. Place any custom Tailwind configuration in `tailwind.config.js` and shared styles in `web/src/index.css`.
- When UI changes require updated embedded assets, run `npm install` (if dependencies changed) followed by `npm run build` from the `web` directory. Commit the resulting changes under `internal/server/static`.
- Keep ESLint happy (`npm run lint`) and ensure the TypeScript project builds cleanly (`npm run build`) before submitting changes.

## Documentation
- Align feature work with the product context captured in `docs/BRD.md` and `docs/EDD.md`. Update these documents when changes affect scope, requirements, or design assumptions.
- Maintain the README to reflect the current developer workflow whenever commands or prerequisites change.

## Required Checks
- From the repository root run `go test ./...`.
- From the `web` directory run `npm run lint` and `npm run build`.
