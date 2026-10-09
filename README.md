# Shop

A small e-commerce app for the code challenge: product CRUD, CSV import, search and a checkout with a fake payment. Go backend, React frontend, SQLite, all shipped as one Docker container.

The sample CSV was downloaded on **October 9, 2026** and lives in [`testdata/products.csv`](testdata/products.csv).

## Requirements

To **run the app** and **run the tests** you only need Git and Docker. Go, Node and the database all live inside containers.

| | macOS | Windows 10/11 | Linux |
| --- | --- | --- | --- |
| Git | `xcode-select --install` or `brew install git` | `winget install Git.Git` or [git-scm.com](https://git-scm.com/download/win) | your package manager, e.g. `sudo apt install git` |
| Docker | [Docker Desktop](https://docs.docker.com/desktop/setup/install/mac-install/) or `brew install --cask docker-desktop` | [Docker Desktop](https://docs.docker.com/desktop/setup/install/windows-install/) or `winget install Docker.DockerDesktop` (it asks to enable WSL 2, say yes and reboot) | [Docker Engine](https://docs.docker.com/engine/install/) with the Compose plugin |
| `make` (optional) | already there with the Xcode command line tools | not needed, every `make` command below has a plain Docker version | usually there, or `sudo apt install make` |

Only if you want to **work on the code without Docker** (`make dev`):

| | macOS | Windows | Linux |
| --- | --- | --- | --- |
| Go 1.26+ | `brew install go` | `winget install GoLang.Go` | [go.dev/dl](https://go.dev/dl/) |
| Node 20.19+ (24 recommended) | `brew install node` | `winget install OpenJS.NodeJS.LTS` | [nodejs.org](https://nodejs.org/) or nvm |

A few things to know:

- The first build downloads the Node, Go and runtime images, so keep about 2 GB of free disk space.
- On Linux, if Docker commands only work with `sudo`, add yourself to the docker group with `sudo usermod -aG docker $USER`, then log out and back in.
- Apple Silicon and Intel Macs both work, the images are multi-architecture.

Check what you have:

```bash
git --version
docker --version
docker compose version
```

On Windows, use PowerShell for every command in this file. Commands are the same unless a Windows version is shown.

## Quick start

**1. Get the code**

```bash
git clone https://github.com/jeancarh/interview-csv-handler.git
cd interview-csv-handler
```

**2. Make sure Docker is running**

Open Docker Desktop (or start the Docker service on Linux) and check it answers:

```bash
docker info
```

If that prints an error like "Cannot connect to the Docker daemon", Docker isn't running yet. On Windows, wait until Docker Desktop says "Engine running".

**3. Start the app**

```bash
make up
```

No `make` (Windows, or you just don't have it)? Run the same thing directly:

```bash
docker compose up --build -d --wait
```

The first run builds the image and takes a few minutes. Later runs take seconds. The command returns once the app passes its health check, and prints the URL.

**4. Open it**

| What | URL |
| --- | --- |
| Shop UI | http://localhost:8080 |
| API docs (Swagger UI) | http://localhost:8080/docs |
| Health check | http://localhost:8080/healthz |

The catalog is already filled from the sample CSV. You should see 87 products.

**5. Try it (2 minutes)**

1. Type `speak` in the search box: three speakers show up.
2. Click **Add to cart** on a couple of products, open **Cart** and pay with `4242 4242 4242 4242`. You get an approved order and the stock goes down.
3. Add something again and pay with `4000 0000 0000 0002`. The payment is declined, nothing is charged and the stock goes back.
4. Open **Import** and upload `testdata/products.csv`. The report opens on the 13 rows that need attention and explains each one.
5. Click **New product**, leave the name empty and save, to see the validation.

**6. Stop it**

```bash
make down    # stop, keep the data
make reset   # stop and delete the database, the next start seeds again
make logs    # see the logs while it runs
```

Without `make`: `docker compose down`, `docker compose down -v` and `docker compose logs -f shop`.

**Port 8080 already in use?** Pick another one and use it in the URLs above:

```bash
make up HOST_PORT=9090
# or, macOS / Linux
HOST_PORT=9090 docker compose up --build -d --wait
```

```powershell
# Windows PowerShell
$env:HOST_PORT=9090; docker compose up --build -d --wait
```

## Running without Docker

This is for working on the code. You need Go 1.26+ and Node 20.19+ (Node 24 recommended). Stop the container first (`make down`), since both use port 8080.

```bash
make dev
```

That starts the API on http://localhost:8080 and the Vite dev server on http://localhost:5173. Open **http://localhost:5173**: it reloads on changes and forwards `/api` calls to the backend. The database file is `backend/data/shop.db`. Press `Ctrl+C` to stop both.

`make dev` uses a Unix shell, so on Windows run each side in its own terminal instead (this works everywhere):

```bash
# terminal 1, macOS / Linux
cd backend
SEED_CSV=../testdata/products.csv go run ./cmd/api
```

```powershell
# terminal 1, Windows PowerShell
cd backend
$env:SEED_CSV="../testdata/products.csv"; go run ./cmd/api
```

```bash
# terminal 2, any OS
cd frontend
npm install
npm run dev
```

Configuration is done with environment variables:

| Variable | Default | Meaning |
| --- | --- | --- |
| `PORT` | `8080` | HTTP port |
| `DB_PATH` | `data/shop.db` | SQLite file, created if missing |
| `SEED_CSV` | empty | CSV to import on start, only when the catalog is empty |
| `MAX_UPLOAD_MB` | `5` | Max CSV upload size |
| `MAX_IMPORT_ROWS` | `10000` | Max rows per CSV |

## Using the app

1. **Products**: search by name, SKU or description (prefix search, so `speak` finds "Bluetooth Speaker"), filter by category or stock. Create, edit and delete from the same page.
2. **Import**: upload a CSV. You get a report with one line per row saying if it was created, updated, unchanged, skipped or rejected, and why. It opens on the rows that need attention and is paginated, so a big file doesn't bury the few rows you have to fix.
3. **Cart**: add products, change quantities, then pay. The payment is fake:
   - `4242 4242 4242 4242` is approved
   - `4000 0000 0000 0002` is declined (the order is saved as failed and the stock goes back)

## Using the API docs

Open http://localhost:8080/docs with the app running. Every endpoint has a description and examples.

1. Click an endpoint, for example **GET /api/v1/products**.
2. Click **Try it out**, fill the parameters (try `q = speaker`) and click **Execute**.
3. The response and the equivalent `curl` command show up below.

To import a CSV from Swagger, open **POST /api/v1/products/import**, click **Try it out** and pick the file.

To place an order, open **POST /api/v1/orders**. It needs an `Idempotency-Key` header, any unique string of 8+ characters works. Send the same request twice with the same key and the second response comes back with `Idempotent-Replayed: true` and the same order, without charging again.

The same thing with curl:

```bash
curl -F file=@testdata/products.csv http://localhost:8080/api/v1/products/import

curl "http://localhost:8080/api/v1/products?q=speaker&category=Electronics"

curl -X POST http://localhost:8080/api/v1/orders \
  -H "Idempotency-Key: my-first-order-1" \
  -H "Content-Type: application/json" \
  -d '{"items":[{"productId":1,"quantity":2}],"cardNumber":"4242 4242 4242 4242"}'
```

Money is always integer cents (`8999` is $89.99) and weight is integer grams, so there is no floating point anywhere. Errors use the RFC 9457 `application/problem+json` format, and validation errors include an `errors` object keyed by field.

## Tests

**Without installing anything (Docker only)**

```bash
make test-docker
```

That runs the backend suite in the official Go image and the frontend suite in the official Node image, with nothing installed on your machine. The first run downloads both images, about a minute. Without `make`, from the repo root:

```bash
# macOS / Linux
docker run --rm -v "$PWD:/repo" -v shop-go-cache:/go/pkg/mod -w /repo/backend golang:1.26 sh -c "go vet ./... && go test -race ./..."
docker run --rm -v "$PWD/frontend:/app" -v /app/node_modules -w /app node:24-alpine sh -c "npm ci && npm run typecheck && npm run lint && npm test"
```

```powershell
# Windows PowerShell
docker run --rm -v "${PWD}:/repo" -v shop-go-cache:/go/pkg/mod -w /repo/backend golang:1.26 sh -c "go vet ./... && go test -race ./..."
docker run --rm -v "${PWD}/frontend:/app" -v /app/node_modules -w /app node:24-alpine sh -c "npm ci && npm run typecheck && npm run lint && npm test"
```

**With Go and Node installed**

```bash
make test
```

Or one side at a time:

```bash
cd backend && go vet ./... && go test -race ./...
cd frontend && npm install && npm run typecheck && npm run lint && npm test
```

Every package should print `ok` on the Go side, and Vitest should end with `43 passed`.

**What's covered**

- Backend: domain and use cases are tested with in-memory fakes. The SQLite adapter and the HTTP layer are tested against a real temporary database, including an import of the sample file, search, a concurrency test that tries to oversell, and the full checkout flow. `-race` turns on Go's race detector.
- Frontend: typecheck, ESLint (which also enforces the layer rules described below) and Vitest with Testing Library: pure domain functions, plus pages rendered with in-memory adapters (XSS rendering, cart limits, form errors, idempotent retry, import report).

## Project layout

```
backend/
  cmd/api/                 wiring: config, database, adapters, HTTP server
  internal/catalog/        products, CSV import rules, search (domain + use cases + ports)
  internal/checkout/       orders, idempotency, payment port
  internal/adapter/
    httpapi/               REST handlers, DTOs, problem+json, middleware
    sqlite/                repositories, transactions, migration runner
    csvimport/             CSV parsing, no business rules
    payment/               fake card gateway
  migrations/0001_init.sql the only migration
  api/openapi.yaml         API contract, served at /docs
frontend/src/
  app/                     router, providers, picks the HTTP adapters
  features/<name>/
    domain/                types and plain functions, no React
    application/           hooks (TanStack Query) that talk to ports
    infrastructure/        HTTP adapters that implement the ports
    ui/                    pages (containers) and presentational components
  shared/                  http client, money helpers, small UI pieces
```

## Approach

I started with the sample CSV instead of the code. Reading it row by row showed that the file is the real test: currency symbols, `free` as a price, negative stock, script tags, SQL in a name, empty names, the same SKU three times with different data, empty rows. So the first thing I did was write down a rule for each of those cases (they're in [ISSUES.md](ISSUES.md)) and the questions I'd normally take to the product owner.

From there:

1. **Domain first.** Product rules, import rules and checkout rules went into plain Go packages with tests against in-memory fakes, before any database or HTTP code existed. Most decisions were made here, where they're cheap to change.
2. **Adapters next.** SQLite, the CSV parser, the fake payment and the HTTP layer, each tested against the real thing: a temp database, the actual sample file, a real HTTP server.
3. **Contract, then UI.** The OpenAPI spec was written before the frontend, and the frontend types are generated from it. The UI follows the same split as the backend, so components never know about HTTP.
4. **Packaging and docs last,** then a full run against the built container: import, search, CRUD, paid and declined orders, retries with the same idempotency key, an oversell attempt and a restart to check that data persists.

I treated it as a small but real product: one command to run, safe defaults, and no feature I couldn't explain.

## Design decisions

**Hexagonal architecture on the backend.** `catalog` and `checkout` hold the rules and define the interfaces they need (`Repository`, `Products`, `Orders`, `PaymentGateway`, `Transactor`). SQLite, HTTP, CSV and the fake payment are adapters around them, and `cmd/api/main.go` is the only place that knows about all of them. The point is that the rules can be tested without a database, and swapping SQLite for Postgres or the fake gateway for Stripe touches one adapter.

**Standard library HTTP.** Go 1.22+ routing handles methods and path params, so a framework would mostly add a dependency.

**SQLite through `modernc.org/sqlite`.** The challenge asks for a local database and SQLite needs no server. The pure Go driver means no CGO, so the binary is static and runs on a distroless image. WAL mode, a busy timeout and `BEGIN IMMEDIATE` transactions avoid most locking surprises.

**One migration.** The schema is in `migrations/0001_init.sql`, embedded in the binary and applied at startup by a small runner that records versions in `schema_migrations`. New changes would go in `0002_...`, never by editing `0001`.

**Full-text search with FTS5.** An FTS5 table kept in sync by triggers gives prefix search with relevance ranking without adding Elasticsearch. User input is quoted term by term before it reaches `MATCH`, so FTS syntax in the search box can't break the query.

**Spec-first OpenAPI.** The spec in `backend/api/openapi.yaml` is written by hand and served at `/docs`. The frontend generates its TypeScript types from the same file (`npm run gen:api`), so both sides share one contract. I skipped swaggo because it builds the spec from code comments, and the challenge asks for code without comments.

**Integer money.** Prices are cents and weights are grams, parsed from strings without going through floats. `0.1 + 0.2` problems simply can't happen.

**CSV import rules.** These are the decisions that matter most. The full list is in [ISSUES.md](ISSUES.md).
- SKU is the business key. A known SKU updates the product, a new one creates it.
- Bad rows are rejected with a reason, and the rest of the file still goes in, in one transaction. A structurally broken file (wrong columns, bad quotes) is rejected as a whole.
- If a SKU appears twice in the file, the last row wins and the earlier ones show up as skipped, with the line that replaced them.
- Importing the same file again changes nothing and reports every row as `unchanged`.
- Stock from the CSV is used only when a product is created. Updates never touch stock, so re-importing an old file can't undo purchases.
- Names like `<script>...` or `'); DROP TABLE` are stored as plain text. Queries are parameterized and React escapes output, so they are harmless, and dropping them would hide data the owner might want to fix.

**Checkout that can't double charge or oversell.**
- Stock is reserved with `UPDATE ... SET stock = stock - ? WHERE stock >= ?`, so two buyers can never take the last unit. There's a test with 20 goroutines fighting over 5 units.
- The whole reservation is one transaction. If any item is short, nothing is reserved.
- The `Idempotency-Key` header plus a hash of the request makes retries safe. Same key and same body returns the original order. Same key with a different body is rejected.
- The payment call happens outside the database transaction. If it's declined, the stock is released and the order is marked failed. The frontend reuses the key when you retry after a network error and creates a new one when the cart changes.

**Soft delete.** Deleted products disappear from the catalog, but orders keep a snapshot of name, SKU and price, and the SKU becomes free again (partial unique index on active rows).

**One container.** A multi-stage Dockerfile builds the React app, embeds it into the Go binary with `go:embed`, and copies that binary into a distroless image running as non-root. The result is a single process serving the UI, the API and the docs, with a built-in health check.

**Frontend layers.** Each feature has `domain`, `application`, `infrastructure` and `ui` folders, mirroring the backend. Components never call `fetch`. They use hooks, the hooks use a repository interface, and `app/App.tsx` plugs in the HTTP implementation. ESLint (`no-restricted-imports` per folder) fails the build if a layer imports something it shouldn't. I didn't add use case classes on top of the hooks because they would only forward calls in a CRUD app.

**Patterns and principles actually used**
- Ports and adapters, Repository (`catalog.Repository`, `checkout.Orders`), Strategy for payments (`checkout.PaymentGateway`, with `payment.Fake` as one implementation).
- Unit of Work through `Transactor.WithinTx`, which puts the transaction in the context so repositories join it without knowing.
- Factory with invariants: `catalog.NewProduct` is the only way to build a valid product.
- Dependency injection by hand in `main.go` and in `App.tsx`, no container.
- Container and presentational components in React (`ProductListPage` vs `ProductTable`).
- SRP (the CSV adapter only parses, the importer only applies rules), OCP and DIP (use cases depend on interfaces they own, so new adapters don't change them), ISP (checkout asks for `Get`, `ReserveStock` and `ReleaseStock`, not the whole catalog repository).

## Assumptions and questions I'd ask the product owner

The challenge leaves these open, so I picked a reasonable default for each and would confirm them:

1. **Duplicate SKUs in one file.** Last row wins. Would you rather reject conflicting rows?
2. **Who owns stock?** I assumed the warehouse (the app). The CSV sets initial stock only. If the CSV is an ERP snapshot, it should overwrite stock instead.
3. **Products missing from a new CSV.** Kept. Should the import be a full sync that archives them?
4. **Partial imports.** Valid rows go in even if others fail. Some teams prefer all or nothing.
5. **`$29.99`.** Accepted, with the symbol stripped and a note in the report. `free` is rejected, not read as zero.
6. **Zero price** (Mystery Box) is allowed. **Missing weight** is allowed, since gift cards and digital goods have none. **Missing category** is rejected.
7. **Category: free text or fixed list?** Free text, because a fixed list would reject rows like "Gifts" or "Pets". The UI builds the filter from categories in use.
8. **Currency.** USD only.

## Alternatives considered

| Topic | Chosen | Also considered | Why not |
| --- | --- | --- | --- |
| Database | SQLite | Postgres in a second container | More moving parts for a local challenge. The repository port makes the switch cheap later. |
| SQLite driver | modernc (pure Go) | mattn/go-sqlite3 | Needs CGO and a libc image. |
| Migrations | Embedded SQL + tiny runner | golang-migrate, goose | Another dependency to run one file. |
| HTTP | net/http | Gin, Echo, Chi | The stdlib router covers it now. |
| API docs | Hand-written OpenAPI | swaggo annotations | Spec built from comments, which the challenge asks to avoid. |
| Search | FTS5 | `LIKE '%x%'` | No ranking, no prefix matching per word, full scans. |
| Import duplicates | Last wins | Reject conflicts | Blocks valid updates. Easy to flip if the business wants it. |
| Import idempotency | Upsert + `unchanged` detection | Store a SHA-256 per uploaded file | The upsert already makes repeats harmless. A file hash would only help auditing. |
| Deployment shape | One container | nginx + API containers | Two things to run and version for no gain at this size. |
| Frontend state | TanStack Query + a small cart reducer | Redux | Most state is server cache, and Query handles that better. |
| Layer rules | ESLint `no-restricted-imports` | eslint-plugin-boundaries | Its current release pulls dependencies with known advisories. The built-in rule does the job. |

## Deploying to AWS

The goal is to keep it cheap without making it fragile. Since SQLite lives on local disk, the app runs as a single instance.

- **ECR** stores the image. The existing GitHub Actions workflow would get one more job that builds and pushes it on merge to main.
- **One EC2 `t4g.small`** (Graviton, about $12/month) runs the container with Docker, and an **EBS gp3** volume holds `/data`.
- **Litestream** streams the SQLite WAL to **S3** every second, which gives point-in-time restore for a few cents a month. A new instance restores from S3 on boot.
- **CloudFront** in front for TLS and caching of the static assets, with an ACM certificate (free).
- **CloudWatch** agent for logs (the app already logs JSON) and an alarm on `/healthz`.

That's roughly $15 to $20 a month. When traffic or the team grows, the path is: Postgres on RDS, then ECS Fargate behind an ALB with several tasks. Only the `sqlite` adapter changes. CSV imports could move to an S3 upload that triggers a job through SQS, so large files don't tie up a request.
