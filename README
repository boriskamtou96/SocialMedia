# 🚀 SocialMedia API

> A high-performance, scalable, and production-ready Social Media RESTful API built with **Go**, **PostgreSQL**, **Docker**, and **Kubernetes**.

---

## 🛠️ Tech Stack & Tools

| Technology | Role | Description |
| :--- | :--- | :--- |
| **[Go](https://golang.org)** | Language | High-performance, concurrent backend service |
| **[PostgreSQL](https://www.postgresql.org/)** | Database | Relational database for structured data storage |
| **[golang-migrate](https://github.com/golang-migrate/migrate)** | Database | Database migration management |
| **[Air](https://github.com/air-verse/air)** | Dev Tool | Live-reloading for Go applications during development |
| **[Docker](https://www.docker.com/)** | Containerization | Containerized development and deployment |
| **[Kubernetes](https://kubernetes.io/)** | Orchestration | Scalable container orchestration for production |

---

## ✨ Features

- 🔒 **Authentication & Authorization:** Secure user registration, login, and JWT-based session management.
- 👥 **User Management:** Profiles, followers, and social connections.
- 📝 **Posts & Feed:** Create, read, update, and delete posts with dynamic newsfeed aggregation.
- 💬 **Interactions:** Comments, likes, and engagement features.
- ⚡ **Developer Experience:** Instant hot-reload with Air and automated DB migrations.
- 🐳 **Cloud-Ready:** Multi-stage Dockerized builds and production-grade Kubernetes deployment manifests.

---

## 📁 Project Structure

```text
.
├── cmd/
│   └── api/             # Application entry point (main.go)
├── internal/
│   ├── config/          # Environment & app configuration
│   ├── database/        # DB connection & repository implementations
│   ├── handler/         # HTTP handlers / Controllers
│   ├── middleware/      # Auth, logging, CORS middlewares
│   └── service/         # Core business logic
├── migrations/          # SQL migration files (.sql)
├── k8s/                 # Kubernetes deployment & service manifests
├── .air.toml            # Air live-reload configuration
├── docker-compose.yml   # Local environment orchestrator
├── Dockerfile           # Multi-stage build image
└── Makefile             # Task automation shortcuts