# Event Hub

Event Management API for organizing and managing events.

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Git

### Setup

```bash
# Clone and setup
git clone https://github.com/kimashii-dan/event-hub.git
cd event-hub/backend
cp .env.example .env
# Edit .env with your database credentials

# Run application
cd docker
docker compose up --build
```

Application runs at: `http://localhost:8000`, you can change port to 8080 if you want

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
DB_HOST=db
DB_USER=postgres
DB_PASSWORD=your_password
DB_PORT=5432
DB_NAME=event_hub
```

## Git Workflow

### Development Process

1. **Create feature branch**: `git checkout -b feature/your-initials-feature-name`
2. **Make changes and commit**: `git commit -m "feat: description"`
3. **Push and create PR**: `git push origin feature/your-initials-feature-name`
4. **Team lead reviews and merges**

### Branch Naming

- Features: `feature/jd-user-auth`
- Bug fixes: `bugfix/jd-fix-login`
- Hotfixes: `hotfix/jd-security-patch`

### Daily Workflow

```bash
# Start work
git checkout main && git pull origin main
git checkout -b feature/initials-feature-name

# Finish work
git add . && git commit -m "feat: description"
git push origin feature/initials-feature-name
# Create PR on GitHub
```

## Database

- **PostgreSQL 15** with golang/migrate
- **ORM**: GORM

## Troubleshooting

**Port conflicts**: Change ports in `.env` or `docker-compose.yaml`
**Database issues**: Check `.env` values and run `docker compose logs db`
**Build failures**: Run `docker compose down && docker compose up --build`
