# Inbox Zero Automation Platform

This repository scaffolds the Inbox Zero automation experience described in the BRD/EDD. It combines a Go HTTP service with an embedded React + Tailwind interface that visualizes deterministic and LLM-assisted email flows.

## Prerequisites

- Go 1.22+
- Node.js 20+ and npm

## Getting Started

1. Install web dependencies and build the static assets:

   ```bash
   cd web
   npm install
   npm run build
   ```

2. Run the Go service (serves API stubs and the embedded UI):

   ```bash
   go run ./cmd/server
   ```

3. Visit <http://localhost:8080> for the control center UI.

### Local Front-End Development

During UI development you can run Vite with hot reloading while proxying API requests to the Go service:

```bash
# Terminal 1
npm run dev

# Terminal 2 (from repo root)
go run ./cmd/server
```

Vite proxies `/api` requests to `http://localhost:8080` to reuse the Go stubs.

## Project Structure

```
cmd/server         Go entrypoint
internal/api       HTTP handlers returning roadmap-aligned payloads
internal/server    Echo configuration + embedded static assets
web                React + Tailwind experience shell
```

The API layer currently returns illustrative data for dashboards, focus plans, and automation simulations so product and engineering teams can align on the end-to-end flow before wiring real integrations.

## Roadmap Hooks

- Replace the static API payloads with calls into real ingestion, classification, and automation pipelines.
- Instrument OpenTelemetry exporters and persist audit logs per the compliance requirements in the BRD.
- Expand the React routes into feature-complete modules (focus timers, automation builder, governance console) as features graduate from design.
