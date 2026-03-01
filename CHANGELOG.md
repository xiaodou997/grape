# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Internationalization (i18n) support with Chinese and English
- Language switch component with GitHub repository link
- Hot reload development mode with Air
- Change password functionality in user menu
- Frontend asset validation in build process
- Handler unit tests for authentication

### Fixed
- CSP policy to allow external HTTPS images in package README
- CSP policy to allow vue-i18n compatibility (unsafe-eval)

## [0.1.0] - 2024-03-01

### Added
- Initial release of Grape
- Single binary deployment with embedded Vue 3 frontend
- npm registry API compatibility
- Multi-upstream routing by scope
- JWT authentication with SQLite persistence
- Web UI for package management
- User management with role-based access
- CI/CD token management
- Backup and restore functionality
- Garbage collection for old packages
- Package-level ACL (owners)
- Webhook notifications
- Prometheus metrics support
- Docker image support

[Unreleased]: https://github.com/xiaodou997/grape/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/xiaodou997/grape/releases/tag/v0.1.0
