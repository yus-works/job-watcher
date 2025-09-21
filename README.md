# Job Watcher

A Go backend service that aggregates software engineering job postings from multiple remote job boards and RSS feeds into a unified interface.

## Overview

Job Watcher continuously fetches job listings from various APIs and RSS feeds, deduplicates them, and presents them through a clean web interface with filtering and search capabilities. Built with Go's standard library and SQLite, it prioritizes simplicity, performance, and reliability.

## Problem Statement

Monitoring multiple job boards is time-consuming and prone to missing opportunities. Job Watcher automates this process by:
- Aggregating listings from Remotive, RemoteOK, Jobicy, Himalayas, WeWorkRemotely, and other sources
- Deduplicating posts that appear across multiple platforms
- Providing a single interface with search and filtering capabilities

## Technical Architecture

### Data Pipeline
- **Multi-source ingestion**: Handles JSON APIs and RSS feeds with different schemas
- **Normalization**: Converts heterogeneous data into a consistent internal format
- **Deduplication**: Uses content-based hashing to identify and eliminate duplicate listings
- **Storage**: SQLite database with properly normalized schema including tagging system

### Concurrency & Performance
- Parallel feed fetching using goroutines and WaitGroups
- Custom HTTP client with connection pooling and optimized timeouts
- Context-based cancellation for handling slow or stuck sources
- Rate limiting with configurable cooldown periods to respect API limits

### Web Interface
- Server-side rendering with Go's `html/template`
- Dynamic updates via HTMX for modern UX without SPA complexity
- RESTful endpoints for job streaming and filtering
- Token-protected refresh endpoint with rate limiting

### Observability
- Structured logging with detailed network timing metrics
- Performance monitoring using `httptrace` hooks
- Request-level timing for DNS, TLS handshake, and time-to-first-byte
- Error tracking with contextual information for debugging

## Implementation Highlights

**Database Design**: The schema uses CHECK constraints to ensure data integrity, with separate tables for tags and job relationships. The INSERT OR IGNORE pattern efficiently handles deduplication at the database level.

**Error Handling**: Failed feed fetches are isolated and logged without affecting other sources. HTTP endpoints return appropriate status codes (200, 429, 500) with meaningful error messages.

**Deployment Ready**: Includes multi-stage Dockerfile producing minimal images, Fly.io configuration for cloud deployment, and environment-based configuration following 12-factor principles.

**Development Environment**: Nix flake provides reproducible development setup with all necessary tooling.

## Code Structure

```
├── main.go           # Application entry point and server setup
├── feeds/            # Feed parsing and normalization logic
├── storage/          # Database interface and operations
├── handlers/         # HTTP handlers and routing
├── templates/        # HTML templates
└── static/           # CSS and client-side assets
```

## Running the Project

```bash
# Local development
go run main.go

# Docker
docker build -t job-watcher .
docker run -p 8080:8080 job-watcher

# With Nix
nix develop
go run main.go
```

## Configuration

Environment variables:
- `DB_PATH`: SQLite database location
- `LOG_PATH`: Log file path
- `REFRESH_TOKEN`: Secret token for protected endpoints
- `PORT`: HTTP server port

## Design Decisions

**Standard Library Focus**: Uses Go's `net/http` and `html/template` instead of frameworks, reducing dependencies and improving maintainability.

**SQLite Over PostgreSQL**: For this use case, SQLite provides sufficient performance with simpler deployment and no external dependencies.

**HTMX Over React/Vue**: Delivers dynamic functionality without JavaScript build complexity, keeping the frontend lightweight and maintainable.

**Content Hashing for Deduplication**: More reliable than URL-based deduplication since the same job often appears with different URLs across platforms.

## Future Enhancements

- User accounts for saved searches and job tracking
- Email notifications for matching jobs
- Additional job board integrations
- Advanced filtering by salary range and tech stack
- API endpoints for programmatic access

## Technical Stack

- **Language**: Go 1.21+
- **Database**: SQLite
- **Frontend**: HTML templates + HTMX
- **Logging**: Logrus for structured logging
- **Deployment**: Docker, Fly.io
- **Development**: Nix for reproducible environments

---

This project demonstrates practical Go backend development with attention to production concerns: data integrity, error handling, observability, and deployment. The codebase emphasizes clarity and maintainability while solving a real-world problem.
