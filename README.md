# 🚀 SocialMedia API

> A RESTful social media API built with **Go**, **Chi** and **PostgreSQL**: user invitations by email, JWT authentication, role-based permissions, posts, followers and a personalized feed.

---

## 🛠️ Tech Stack

| Technology | Role |
| :--- | :--- |
| **[Go](https://go.dev)** 1.26 + **[Chi](https://github.com/go-chi/chi)** | HTTP server and routing |
| **[PostgreSQL](https://www.postgresql.org/)** | Main database |
| **[Redis](https://redis.io/)** | Optional user cache |
| **[golang-migrate](https://github.com/golang-migrate/migrate)** | SQL migrations |
| **[SendGrid](https://sendgrid.com/)** | Invitation emails |
| **[Swag](https://github.com/swaggo/swag)** | Swagger / OpenAPI documentation |
| **[Zap](https://github.com/uber-go/zap)** | Structured logging |
| **[Air](https://github.com/air-verse/air)** | Live reload during development |
| **[Docker](https://www.docker.com/)** | Local database and production image |
| **[React](https://react.dev/) + [Vite](https://vite.dev/)** | Account confirmation page (`web/`) |

---

## ✨ Features

- 📧 **Registration by invitation:** new users receive an email with an activation link that opens the confirmation page in `web/`.
- 🔒 **JWT authentication:** log in with email and password to get a token valid for 3 days.
- 🛡️ **Role-based permissions:** `user`, `moderator` and `admin`. Moderators can edit posts, admins can edit and delete them.
- 📝 **Posts:** create, read (with comments), update and delete, with optimistic locking to prevent concurrent overwrites.
- 👥 **Followers:** follow and unfollow other users.
- 📰 **Feed:** posts from followed users with pagination, sorting, tag filter, search in titles and content, and date range.
- ⚡ **Redis cache:** authenticated users are cached in Redis when it is enabled.
- 🚦 **Rate limiting:** fixed-window limiter per IP (20 requests per 5 seconds by default).
- 📚 **Swagger UI** at `/v1/swagger/index.html`.
- 📊 **Metrics** (version, goroutines, memory) at `/v1/metrics`, protected by basic auth.

---

## 📁 Project Structure

```text
.
├── cmd/
│   ├── api/               # HTTP server: entry point, routes, handlers, middlewares
│   └── migrate/
│       ├── migrations/    # SQL migrations (golang-migrate)
│       └── seed/          # Fills the local database with fake data
├── internal/
│   ├── auth/              # JWT authenticator
│   ├── db/                # PostgreSQL connection
│   ├── env/               # Environment variable helpers
│   ├── mailer/            # SendGrid mailer and email templates
│   ├── ratelimiter/       # Fixed-window rate limiter
│   └── store/             # Repositories (users, posts, comments, followers, roles) and Redis cache
├── docs/                  # Generated Swagger docs (make gen-docs)
├── web/                   # React + Vite frontend (account confirmation page)
├── .github/workflows/     # CI: audit, release-please, API version update
├── .air.toml              # Air live-reload configuration
├── docker-compose.yml     # Local PostgreSQL
├── Dockerfile             # Production image (multi-stage, scratch)
└── Makefile               # Migrations, seed and Swagger generation
```

---

## 🏁 Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.26+
- [Docker](https://www.docker.com/)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate): `brew install golang-migrate`
- [Air](https://github.com/air-verse/air): `go install github.com/air-verse/air@latest`
- [Node.js](https://nodejs.org/) 20+ for the frontend
- Optional: [direnv](https://direnv.net/) to load `.envrc` automatically

### 1. Configure the environment

Create a `.envrc` file at the root. It is git-ignored and the `Makefile` reads it.

```bash
export ADDR=":8080"
export DB_ADDR="postgres://postgres:postgres@localhost/socialmedia?sslmode=disable"
export FRONTEND_URL="http://localhost:5173"
export CORS_ALLOWED_ORIGIN="http://localhost:5173"
export FROM_EMAIL="you@example.com"
export SENDGRID_API_KEY="your-sendgrid-api-key"
export JWT_SECRET="change-me"
export REDIS_ENABLE=false
```

Then run `direnv allow`, or `source .envrc` in each new terminal. The Configuration section below lists every variable.

> **Note:** registration sends an invitation email. Without a valid SendGrid API key, user creation fails and is rolled back.

### 2. Start the database and run the migrations

```bash
docker compose up -d
make migrate-up
make seed   # optional: fake users, posts and comments
```

### 3. Run the API

```bash
air                 # with live reload
go run ./cmd/api    # or without
```

The API listens on `http://localhost:8080/v1` and the Swagger UI is at `http://localhost:8080/v1/swagger/index.html`.

### 4. Run the frontend

```bash
cd web
npm install
npm run dev
```

The frontend calls `http://localhost:8080/v1` by default. Set `VITE_API_URL` to point it to another API.

---

## 🧰 Makefile Commands

| Command | Description |
| :--- | :--- |
| `make migrate-up` | Apply all migrations |
| `make migrate-down [n]` | Roll back the last `n` migrations (all if omitted) |
| `make migration <name>` | Create a new pair of migration files |
| `make seed` | Seed the local database (`localhost:5432/socialmedia`) |
| `make gen-docs` | Regenerate the Swagger docs ([swag CLI](https://github.com/swaggo/swag) required) |

---

## 📡 API Endpoints

All routes are prefixed with `/v1`. The Swagger UI documents the request and response bodies.

| Method | Endpoint | Auth | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/health` | — | Status, environment and version |
| `GET` | `/metrics` | Basic | Runtime metrics (expvar) |
| `POST` | `/authentication/user` | — | Register and receive an invitation email |
| `PUT` | `/users/activate/{token}` | — | Activate an account |
| `POST` | `/authentication/token` | — | Log in and get a JWT |
| `GET` | `/users` | — | List users |
| `GET` | `/users/{userID}` | JWT | Get a user |
| `PUT` | `/users/{userID}/follow` | JWT | Follow a user |
| `PUT` | `/users/{userID}/unfollow` | JWT | Unfollow a user |
| `GET` | `/users/feed` | JWT | Feed (`limit`, `offset`, `sort`, `tags`, `search`, `since`, `until`) |
| `POST` | `/posts` | JWT | Create a post |
| `GET` | `/posts/{postID}` | JWT | Get a post with its comments |
| `PATCH` | `/posts/{postID}` | JWT (moderator or admin) | Update a post |
| `DELETE` | `/posts/{postID}` | JWT (admin) | Delete a post |

Send the JWT in the `Authorization: Bearer <token>` header.

---

## ⚙️ Configuration

| Variable | Default | Description |
| :--- | :--- | :--- |
| `ADDR` | `:8080` | Address the server listens on |
| `DB_ADDR` | `postgres://postgres:postgres@localhost/socialmedia?sslmode=disable` | PostgreSQL connection string |
| `DB_MAX_OPEN_CONNS` | `25` | Max open connections |
| `DB_MAX_IDLE_CONNS` | `25` | Max idle connections |
| `BD_MAX_IDLE_TIME` | `15m` | Max idle time of a connection |
| `REDIS_ENABLE` | `true` | Enable the Redis user cache |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `REDIS_PASSWORD` | _(empty)_ | Redis password |
| `REDIS_DB` | `0` | Redis database |
| `FRONTEND_URL` | `http://localhost:5173` | Frontend URL used in activation links |
| `CORS_ALLOWED_ORIGIN` | `http://localhost:5174` | Origin allowed by CORS (set it to the frontend URL) |
| `FROM_EMAIL` | — | Sender address of the emails |
| `SENDGRID_API_KEY` | _(empty)_ | SendGrid API key |
| `JWT_SECRET` | `your-secret-key` | JWT signing secret (**change it in production**) |
| `JWT_AUDIENCE` | `GopherSocialMedia` | JWT audience |
| `JWT_ISSUER` | `GopherSocialMedia` | JWT issuer |
| `BASIC_AUTH_USER` | `admin` | Username for `/metrics` (**change it in production**) |
| `BASIC_AUTH_PASSWORD` | `admin` | Password for `/metrics` (**change it in production**) |
| `RATE_LIMITER_ENABLED` | `true` | Enable rate limiting |
| `RATELIMITER_REQUESTS_COUNT` | `20` | Requests allowed per IP every 5 seconds |

---

## 🐳 Docker

The `Dockerfile` builds a static binary into a `scratch` image (about 10 MB) that runs as an unprivileged user.

```bash
docker build -t socialmedia-api .
docker run -p 8080:8080 \
  -e DB_ADDR="postgres://postgres:postgres@host.docker.internal/socialmedia?sslmode=disable" \
  -e REDIS_ENABLE=false \
  socialmedia-api
```

On Apple Silicon, add `--platform linux/amd64` to `docker build` to target most cloud servers.

---

## ☁️ Free Deployment (Render + Neon)

This setup costs nothing: the API runs on [Render](https://render.com/)'s free plan from the `Dockerfile`, and PostgreSQL on [Neon](https://neon.com/)'s free plan.

### 1. Database on Neon

1. Create a free project on [neon.com](https://neon.com/).
2. Copy the **direct** connection string (turn off *Connection pooling*) and keep only `?sslmode=require` at the end, for example `postgres://USER:PASSWORD@HOST/neondb?sslmode=require`.
3. Run the migrations from your machine:

   ```bash
   migrate -path=./cmd/migrate/migrations -database="postgres://USER:PASSWORD@HOST/neondb?sslmode=require" up
   ```

### 2. API on Render

1. On Render, create a **Web Service** from this GitHub repository. Render detects the `Dockerfile` (language **Docker**).
2. Choose the **Free** instance type and set the **Health Check Path** to `/v1/health` (in *Advanced*).
3. Add the environment variables:

   | Variable | Value |
   | :--- | :--- |
   | `ADDR` | `:10000` (Render's default port) |
   | `DB_ADDR` | The Neon connection string |
   | `REDIS_ENABLE` | `false` |
   | `JWT_SECRET` | A long random string (`openssl rand -hex 32`) |
   | `BASIC_AUTH_USER` / `BASIC_AUTH_PASSWORD` | Your own credentials |
   | `FROM_EMAIL` / `SENDGRID_API_KEY` | Your SendGrid sender and key |
   | `FRONTEND_URL` / `CORS_ALLOWED_ORIGIN` | The frontend URL (step 3) |

Render redeploys automatically on every push to `main`.

### 3. Frontend on Render (optional)

1. Create a **Static Site** from the same repository with root directory `web`, build command `npm ci && npm run build` and publish directory `dist`.
2. Add the environment variable `VITE_API_URL=https://<your-api>.onrender.com/v1`.
3. In *Redirects/Rewrites*, add a rule with source `/*`, destination `/index.html` and action **Rewrite**, so that `/confirm/<token>` links work.

### Free plan limits

- **Render:** the API goes to sleep after 15 minutes without traffic, and the next request takes about a minute to wake it up.
- **Neon:** 1 GB of storage per project. The database pauses after 5 minutes of inactivity and resumes on the next query.
- **SendGrid:** no longer has a free plan. New accounts get a 60-day trial limited to 100 emails per day. After that, registration emails require a paid plan or another provider with a free tier (for example [Resend](https://resend.com/), which requires changing `internal/mailer`).
