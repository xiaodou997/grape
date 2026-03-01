# 🍇 Grape

> **Lightweight Enterprise Private npm Registry**  
> One binary, zero debt.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![npm compatible](https://img.shields.io/badge/npm-compatible-brightgreen)](https://www.npmjs.com)
[![Docker](https://img.shields.io/badge/Docker-available-blue?logo=docker)](https://ghcr.io/xiaodou997/grape)

[**中文文档**](README_CN.md)

Grape is a **lightweight, high-performance** private npm registry written in Go, fully compatible with npm/yarn/pnpm/bun clients. Compared to Verdaccio, it offers **lower resource usage**, **stronger permission control**, and a **modern Web UI**.

## ✨ Features

| Feature | Description |
|---------|-------------|
| 🚀 **Single Binary** | No Node.js required, download and run |
| 📦 **npm Compatible** | Full support for npm/yarn/pnpm/bun |
| 🔀 **Multi-Upstream Routing** | Route by scope to different upstreams |
| 🔐 **User Authentication** | JWT auth with SQLite persistence |
| 💾 **Data Persistence** | SQLite storage, survives restarts |
| 🗄️ **Smart Caching** | Auto-cache public packages |
| 🌐 **Modern Web UI** | Vue 3 + Element Plus admin interface |
| 🪶 **Lightweight** | Memory usage < 10MB |
| 🔔 **Webhook Notifications** | Package publish/delete events |
| 📊 **Prometheus Metrics** | Built-in monitoring support |
| 🎫 **CI/CD Tokens** | Dedicated tokens for automation |
| 💾 **Backup & Restore** | Web UI and CLI support |
| 🔒 **Package-level ACL** | Fine-grained access control |

## 🚀 Quick Start

### Option 1: Download Pre-built Binary

```bash
# Linux (x86_64)
curl -sL https://github.com/xiaodou997/grape/releases/latest/download/grape-linux-amd64.tar.gz | tar xz
./grape-linux-amd64

# Linux (ARM64)
curl -sL https://github.com/xiaodou997/grape/releases/latest/download/grape-linux-arm64.tar.gz | tar xz
./grape-linux-arm64

# Access Web UI
open http://localhost:4873
```

### Option 2: Docker

```bash
# Pull and run
docker pull ghcr.io/xiaodou997/grape:latest
docker run -d --name grape -p 4873:4873 -p 4874:4874 ghcr.io/xiaodou997/grape:latest

# View logs
docker logs -f grape
```

### Option 3: Build from Source

```bash
git clone https://github.com/xiaodou997/grape.git
cd grape
make build
./bin/grape
```

## 📖 Usage

### Configure npm

```bash
# Set registry globally
npm set registry http://localhost:4874

# Or for specific scope (recommended)
npm set @mycompany:registry http://localhost:4874
```

### Install Packages

```bash
# Public packages (auto-cached from upstream)
npm install lodash
npm install express

# Scoped packages
npm install @mycompany/utils
```

### Publish Private Packages

```bash
# Login (default: admin / admin)
npm login --registry http://localhost:4874

# Publish
npm publish --registry http://localhost:4874
```

## 🏗️ Architecture

Grape uses a **dual-port architecture**:

| Port | Purpose |
|------|---------|
| **4873** | Web UI + Admin API |
| **4874** | npm Registry API |

This separation ensures:
- Web UI access on standard port
- Clean npm registry API compatibility
- Independent scaling if needed

## 🔧 Configuration

```yaml
# config.yaml
server:
  host: "0.0.0.0"
  web_port: 4873      # Web UI
  api_port: 4874      # npm API

registry:
  upstreams:
    - name: "npmjs"
      url: "https://registry.npmjs.org"
      scope: ""
      timeout: 30s
      enabled: true

auth:
  jwt_secret: "your-secret-key-change-in-production"
  jwt_expiry: 24h

storage:
  type: "local"
  path: "./data"

database:
  type: "sqlite"
  dsn: "./data/grape.db"
```

## 🌐 Web UI Features

- 📦 **Package Browser** - View cached packages and details
- 👤 **User Management** - Create/delete users, assign roles
- 🎫 **Token Management** - Create CI/CD tokens with permissions
- 💾 **Backup & Restore** - Export/import data
- 🗑️ **Garbage Collection** - Clean up old packages
- 📊 **System Monitoring** - Stats and status

**Default credentials:** `admin` / `admin`

> ⚠️ Change the default password immediately in production!

## 📦 Releases

| Tag Pattern | Example | Build Artifacts |
|-------------|---------|-----------------|
| `v*` | `v1.0.0` | Linux binaries + Docker image |
| `beta-v*` | `beta-v1.0.0` | Linux binaries only (pre-release) |

```bash
# Stable release
git tag v1.0.0 && git push origin v1.0.0

# Beta release
git tag beta-v1.0.0 && git push origin beta-v1.0.0
```

## 💡 Why Grape

| Feature | Description |
|---------|-------------|
| 🚀 **Single Binary** | Written in Go, no runtime dependencies |
| 🪶 **Lightweight** | Memory usage < 10MB, minimal footprint |
| 🔀 **Multi-Upstream** | Route packages by scope to different registries |
| 💾 **Data Persistence** | SQLite database, survives restarts |
| 🔐 **Modern Auth** | JWT authentication with database-backed users |
| 🎫 **CI/CD Tokens** | Dedicated tokens for automated publishing |
| 🌐 **Modern Web UI** | Vue 3 + Element Plus admin interface |
| 📊 **Prometheus Ready** | Built-in metrics for monitoring |
| 💾 **Backup & Restore** | Export/import data via Web UI or CLI |
| 🔒 **Package ACL** | Fine-grained access control per package |
| 🗑️ **Garbage Collection** | Clean up old versions automatically |

## 🗂️ Project Structure

```
grape/
├── cmd/grape/              # Application entry
├── internal/
│   ├── auth/               # Authentication (JWT + SQLite)
│   ├── config/             # Configuration (Viper)
│   ├── db/                 # Database models (GORM)
│   ├── logger/             # Logging (Zap)
│   ├── metrics/            # Prometheus metrics
│   ├── registry/           # npm registry core
│   ├── server/             # HTTP server (Gin)
│   ├── storage/            # Storage abstraction
│   ├── webhook/            # Webhook notifications
│   └── web/                # Embedded frontend
├── web/                    # Frontend source (Vue 3)
├── configs/                # Configuration examples
└── docs/                   # Documentation
```

## 🛠️ Development

```bash
# Development mode with hot reload (frontend + backend)
# Frontend: http://localhost:3000 (Vite)
# Backend:  http://localhost:4873 (Air)
make dev

# Build backend only
make build-only

# Build frontend
make build-frontend

# Full build
make build

# Run tests
make test
```

**Development mode features:**
- 🔄 Frontend hot reload (Vite) - instant updates on Vue/TS changes
- 🔄 Backend hot reload (Air) - auto rebuild on Go changes
- 🌐 Access frontend at `http://localhost:3000` for development
- 📡 API requests proxy to backend automatically

## 📚 Documentation

- [**Usage Guide**](docs/USAGE.md) - Package manager configuration
- [**API Reference**](docs/API.md) - Complete API documentation
- [**Deployment**](docs/DEPLOYMENT.md) - Production deployment guide
- [**Development**](docs/DEVELOPMENT.md) - Developer guide

## 🔒 Security Recommendations

1. **Change JWT secret** - Set a strong `auth.jwt_secret`
2. **Change default password** - Modify admin password immediately
3. **Use HTTPS** - Configure reverse proxy (nginx/caddy)
4. **Restrict network access** - Only allow trusted networks
5. **Disable self-registration** - Set `auth.allow_registration: false`

## 🗺️ Roadmap

### v0.2.0 (Planned)
- [ ] RBAC permission model
- [ ] PostgreSQL support
- [ ] Audit logs
- [ ] Package scope permissions

### v0.3.0 (Planned)
- [ ] Redis caching
- [ ] S3/MinIO storage
- [ ] Performance optimization

### v1.0.0 (Planned)
- [ ] LDAP/OIDC integration
- [ ] High availability cluster
- [ ] Helm Chart

## 🤝 Contributing

Contributions are welcome!

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

[Apache License 2.0](LICENSE)

## 🙏 Acknowledgments

- [npm](https://www.npmjs.com) - JavaScript package manager
- [Verdaccio](https://verdaccio.org) - Inspiration source
- [Gin](https://gin-gonic.com) - Go web framework
- [Vue 3](https://vuejs.org) - Frontend framework
- [Element Plus](https://element-plus.org) - UI component library

---

<p align="center">
  Made with ❤️ by the Grape Team<br>
  🍇 Light as air, powerful as a mountain
</p>