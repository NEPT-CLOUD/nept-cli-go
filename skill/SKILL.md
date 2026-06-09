---
name: nept-platform-capabilities
description: Rules and capabilities on supported databases (PostgreSQL only) and frameworks available on the Nept Cloud platform.
---

# Nept Cloud Platform Capabilities & Constraints

This skill details the exact capabilities, supported databases, versions, and frameworks available on the Nept Cloud platform. AI coding agents should consult these guidelines when modifying deployment logic or responding to user inquiries about platform support.

---

## 1. Supported Databases & Constraints

> [!IMPORTANT]
> Although the Nept CLI has options/arguments for `mysql`, `mongodb`, and `redis` under its command definition (inherited from early specifications), **only PostgreSQL (PSQL)** is currently supported and provisioned on the Nept backend.

* **Database Engine**: PostgreSQL
* **Host Cluster**: Eun1-hel1-az1 (`eun1-hel1-az1.nept.space`)
* **Default Port**: 5432
* **Version Support**: 
  * Default and primary version: **PostgreSQL 16**
  * The database is deployed natively on a shared cluster, meaning the version is managed by the cluster environment. Explicit version overrides for PostgreSQL in CLI commands default to the cluster-supported version (v16).

---

## 2. Supported App Frameworks & Runtimes

Nept Cloud supports auto-detecting, building, and deploying the following frameworks and language runtimes:

### A. JavaScript & TypeScript (Node.js 22 Runtime)
* **Next.js**: Build (`npm install && npm run build`), Output (`.next`), Default Port `3000`
* **React**: Build (`npm install && npm run build`), Output (`build`), Default Port `3000`
* **Vue.js**: Build (`npm install && npm run build`), Output (`dist`), Default Port `8080`
* **Nuxt.js**: Build (`npm install && npm run build`), Output (`.output`), Default Port `3000`
* **Angular**: Build (`npm install && npm run build`), Output (`dist`), Default Port `4200`
* **SvelteKit**: Build (`npm install && npm run build`), Output (`build`), Default Port `3000` (executed via Node)
* **Svelte**: Build (`npm install && npm run build`), Output (`public`), Default Port `5000`
* **Vite**: Build (`npm install && npm run build`), Output (`dist`), Default Port `5173`
* **Astro**: Build (`npm install && npm run build`), Output (`dist`), Default Port `4321`
* **Gatsby**: Build (`npm install && npm run build`), Output (`public`), Default Port `8000`
* **Remix**: Build (`npm install && npm run build`), Output (`build`), Default Port `3000`
* **Express.js**: Build (`npm install`), Output (`.`), Default Port `3000`
* **Nest.js**: Build (`npm install && npm run build`), Output (`dist`), Default Port `3000` (starts via `npm run start:prod`)
* **Fastify**: Build (`npm install`), Output (`.`), Default Port `3000`

### B. Bun Runtime (Latest Bun Version)
* **Bun**: Build (`bun install`), Output (`.`), Default Port `3000` (starts via `bun run start`)

### C. Python (Python 3.12 Runtime)
* **FastAPI**: Build (`pip install -r requirements.txt`), Default Port `8000` (runs via Uvicorn)
* **Django**: Build (`pip install -r requirements.txt`), Default Port `8000` (runs via Gunicorn)
* **Flask**: Build (`pip install -r requirements.txt`), Default Port `5000` (runs via Gunicorn)
* **Generic Python**: Build (`pip install -r requirements.txt`), Default Port `8000` (runs via `python app.py`)

### D. System Languages
* **Go (Go 1.22)**: Build (`go build -o main .`), Default Port `8080`
* **Rust (Rust 1.x)**: Build (`cargo build --release`), Default Port `8080`

### E. Static Websites
* **html-css-js**: Starts via `serve` utility on port `80`

---

## 3. Agent Instructions for Deployment

1. **Database Selection**: If asked to deploy or configure a database, always default to `postgres`. Avoid selecting `mysql`, `mongodb`, or `redis` as they will bypass the native DatabaseDeployer module in the engine.
2. **Framework Override**: If auto-detection fails, use the `--framework` CLI flag specifying any of the names in the list above (case-sensitive) to force the preset configuration.
