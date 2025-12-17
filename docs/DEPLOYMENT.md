# Deployment Guide - Event Hub

> Production deployment guide for Event Hub API

## Table of Contents

- [Deployment Options](#deployment-options)
- [Docker Deployment](#docker-deployment)
- [Manual Deployment](#manual-deployment)
- [Cloud Platforms](#cloud-platforms)
- [Security Checklist](#security-checklist)
- [Monitoring](#monitoring)
- [Backup & Recovery](#backup--recovery)

## Deployment Options

Event Hub can be deployed using various methods:

1. **Docker Compose** - Simple, single-server deployment
2. **Docker Swarm** - Multi-node orchestration
3. **Kubernetes** - Enterprise-grade orchestration
4. **Manual** - Traditional server deployment
5. **Cloud Platforms** - AWS, GCP, Azure, DigitalOcean

## Docker Deployment

### Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- Domain name (optional, for production)
- SSL certificates (recommended)

### Production Docker Compose Setup

#### 1. Prepare Environment

```bash
# Create production directory
mkdir -p /opt/event-hub
cd /opt/event-hub

# Clone repository
git clone https://github.com/kimashii-dan/event-hub.git .

# Create production environment file
cp docker/.env.example docker/.env
```

#### 2. Configure Environment Variables

Edit `docker/.env` with production values:

```bash
# Database Configuration
DB_HOST=db
DB_USER=event_hub_user
DB_PASSWORD=<STRONG_RANDOM_PASSWORD>  # Use: openssl rand -base64 32
DB_PORT=5432
DB_NAME=event_hub_prod

# Server Configuration
SERVER_PORT=8000
ENV=production

# JWT Configuration
JWT_SECRET=<STRONG_RANDOM_SECRET>  # Use: openssl rand -base64 64
JWT_EXPIRATION_TIME=24h

# Database Connection Pool (optional)
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
```

#### 3. Production Docker Compose File

Create `docker/docker-compose.prod.yaml`:

```yaml
version: '3.8'

services:
  db:
    image: postgres:15-alpine
    container_name: event_hub_db
    restart: always
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - event_hub_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER} -d ${DB_NAME}"]
      interval: 10s
      timeout: 5s
      retries: 5

  backend:
    build:
      context: ..
      dockerfile: docker/Dockerfile
      args:
        - GO_VERSION=1.21
    container_name: event_hub_backend
    restart: always
    ports:
      - "8000:8000"
    env_file:
      - .env
    depends_on:
      db:
        condition: service_healthy
    networks:
      - event_hub_network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

volumes:
  postgres_data:
    driver: local

networks:
  event_hub_network:
    driver: bridge
```

#### 4. Production Dockerfile

Create optimized `docker/Dockerfile.prod`:

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/event-hub \
    cmd/app/main.go

# Runtime stage
FROM alpine:3.18

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/event-hub .
COPY --from=builder /app/migrations ./migrations

# Create non-root user
RUN adduser -D -u 1000 appuser && \
    chown -R appuser:appuser /root
USER appuser

EXPOSE 8000

CMD ["./event-hub"]
```

#### 5. Deploy

```bash
# Build and start services
docker compose -f docker/docker-compose.prod.yaml up -d --build

# Check status
docker compose -f docker/docker-compose.prod.yaml ps

# View logs
docker compose -f docker/docker-compose.prod.yaml logs -f backend

# Verify health
curl http://localhost:8000/
```

#### 6. Setup Reverse Proxy (Nginx)

```nginx
# /etc/nginx/sites-available/event-hub
server {
    listen 80;
    server_name api.yourdomain.com;

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    # SSL certificates
    ssl_certificate /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Logging
    access_log /var/log/nginx/event-hub-access.log;
    error_log /var/log/nginx/event-hub-error.log;

    # Proxy settings
    location / {
        proxy_pass http://localhost:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Rate limiting (optional)
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
    limit_req zone=api_limit burst=20 nodelay;
}
```

Enable and restart Nginx:
```bash
sudo ln -s /etc/nginx/sites-available/event-hub /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### SSL Certificate with Let's Encrypt

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx -d api.yourdomain.com

# Auto-renewal is configured automatically
# Test renewal
sudo certbot renew --dry-run
```

## Manual Deployment

### On Ubuntu/Debian Server

#### 1. Install Dependencies

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install PostgreSQL
sudo apt install postgresql postgresql-contrib -y

# Install Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### 2. Setup Database

```bash
# Switch to postgres user
sudo -u postgres psql

# Create database and user
CREATE DATABASE event_hub_prod;
CREATE USER event_hub_user WITH ENCRYPTED PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE event_hub_prod TO event_hub_user;
\q
```

#### 3. Deploy Application

```bash
# Create app directory
sudo mkdir -p /opt/event-hub
cd /opt/event-hub

# Clone repository
git clone https://github.com/kimashii-dan/event-hub.git .
cd backend

# Install dependencies
go mod download

# Build application
go build -o event-hub cmd/app/main.go

# Create environment file
cat > .env << EOF
DB_HOST=localhost
DB_USER=event_hub_user
DB_PASSWORD=secure_password
DB_PORT=5432
DB_NAME=event_hub_prod
SERVER_PORT=8000
JWT_SECRET=$(openssl rand -base64 64)
JWT_EXPIRATION_TIME=24h
ENV=production
EOF

# Test application
./event-hub
```

#### 4. Create Systemd Service

```bash
# Create service file
sudo nano /etc/systemd/system/event-hub.service
```

```ini
[Unit]
Description=Event Hub API Service
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/event-hub/backend
Environment="ENV=production"
ExecStart=/opt/event-hub/backend/event-hub
Restart=always
RestartSec=10
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=event-hub

[Install]
WantedBy=multi-user.target
```

```bash
# Set permissions
sudo chown -R www-data:www-data /opt/event-hub

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable event-hub
sudo systemctl start event-hub

# Check status
sudo systemctl status event-hub

# View logs
sudo journalctl -u event-hub -f
```

## Cloud Platforms

### AWS Deployment (EC2 + RDS)

#### 1. Setup RDS PostgreSQL

- Create RDS PostgreSQL instance
- Configure security group (allow port 5432)
- Note endpoint, username, password

#### 2. Setup EC2 Instance

```bash
# Launch Ubuntu 22.04 EC2 instance
# Configure security group:
#   - Allow SSH (port 22) from your IP
#   - Allow HTTP (port 80) from anywhere
#   - Allow HTTPS (port 443) from anywhere

# SSH into instance
ssh -i your-key.pem ubuntu@ec2-xx-xx-xx-xx.compute.amazonaws.com

# Follow manual deployment steps
# Update DB_HOST to RDS endpoint
```

#### 3. Use AWS Secrets Manager (Optional)

```go
// internal/config/config.go
import "github.com/aws/aws-sdk-go/service/secretsmanager"

func getSecret(secretName string) (string, error) {
    // Implement AWS Secrets Manager retrieval
}
```

### Google Cloud Platform (GCP)

#### Using Cloud Run + Cloud SQL

```bash
# Build container
gcloud builds submit --tag gcr.io/PROJECT_ID/event-hub

# Deploy to Cloud Run
gcloud run deploy event-hub \
  --image gcr.io/PROJECT_ID/event-hub \
  --platform managed \
  --region us-central1 \
  --add-cloudsql-instances PROJECT_ID:REGION:INSTANCE_NAME \
  --set-env-vars DB_HOST=/cloudsql/PROJECT_ID:REGION:INSTANCE_NAME \
  --set-env-vars DB_USER=event_hub_user \
  --set-secrets DB_PASSWORD=db-password:latest \
  --set-secrets JWT_SECRET=jwt-secret:latest
```

### DigitalOcean

#### Using App Platform

1. Connect GitHub repository
2. Configure environment variables in dashboard
3. Setup managed PostgreSQL database
4. Deploy automatically on push

## Security Checklist

### Before Production

- [ ] Change all default passwords
- [ ] Use strong JWT secret (64+ characters)
- [ ] Enable HTTPS/TLS
- [ ] Configure firewall (only necessary ports)
- [ ] Disable debug logging
- [ ] Set `ENV=production`
- [ ] Remove development endpoints
- [ ] Implement rate limiting
- [ ] Enable CORS with specific origins
- [ ] Regular security updates
- [ ] Database connection encryption
- [ ] Secure environment variables (not in code)
- [ ] Implement request validation
- [ ] Add API authentication
- [ ] Regular backups
- [ ] Monitor access logs

### Security Headers

Add to Nginx configuration:

```nginx
# Security headers
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "no-referrer-when-downgrade" always;
add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
```

## Monitoring

### Application Monitoring

#### Health Check Endpoint

```go
// Already implemented: GET /
// Returns 200 OK if service is healthy
```

#### Prometheus Metrics (Future)

```go
// pkg/metrics/prometheus.go
import "github.com/prometheus/client_golang/prometheus"

var (
    requestCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
)
```

### Database Monitoring

```sql
-- Check active connections
SELECT count(*) FROM pg_stat_activity;

-- Check slow queries
SELECT pid, now() - pg_stat_activity.query_start AS duration, query 
FROM pg_stat_activity 
WHERE state = 'active' 
ORDER BY duration DESC;

-- Check database size
SELECT pg_size_pretty(pg_database_size('event_hub_prod'));
```

### Log Aggregation

Use services like:
- **ELK Stack** (Elasticsearch, Logstash, Kibana)
- **Grafana Loki**
- **CloudWatch Logs** (AWS)
- **Stackdriver** (GCP)

### Alerting

Setup alerts for:
- High CPU/Memory usage
- Database connection pool exhaustion
- High error rate (5xx responses)
- Disk space running low
- SSL certificate expiration

## Backup & Recovery

### Database Backup

#### Automated Backups

```bash
# Create backup script
sudo nano /opt/scripts/backup-db.sh
```

```bash
#!/bin/bash

BACKUP_DIR="/opt/backups/postgres"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
DB_NAME="event_hub_prod"
DB_USER="event_hub_user"

mkdir -p $BACKUP_DIR

# Create backup
pg_dump -U $DB_USER -d $DB_NAME -F c -f $BACKUP_DIR/backup_$TIMESTAMP.dump

# Keep only last 7 days
find $BACKUP_DIR -type f -mtime +7 -delete

# Upload to S3 (optional)
# aws s3 cp $BACKUP_DIR/backup_$TIMESTAMP.dump s3://your-bucket/backups/
```

```bash
# Make executable
sudo chmod +x /opt/scripts/backup-db.sh

# Add to crontab (daily at 2 AM)
sudo crontab -e
0 2 * * * /opt/scripts/backup-db.sh
```

#### Manual Backup

```bash
# Backup
pg_dump -U event_hub_user -d event_hub_prod -F c -f backup.dump

# Restore
pg_restore -U event_hub_user -d event_hub_prod -c backup.dump
```

### Application State Backup

```bash
# Backup environment variables
cp .env .env.backup.$(date +%Y%m%d)

# Backup application binary
cp event-hub event-hub.backup.$(date +%Y%m%d)
```

## Scaling

### Horizontal Scaling

1. **Load Balancer**: Use Nginx, HAProxy, or cloud load balancer
2. **Multiple App Instances**: Run several backend containers
3. **Database Replication**: Master-slave PostgreSQL setup
4. **Caching Layer**: Add Redis for sessions/caching

### Vertical Scaling

- Increase server resources (CPU, RAM)
- Optimize database queries
- Add database indexes
- Connection pooling tuning

## Rollback Procedure

### Quick Rollback

```bash
# Stop current service
sudo systemctl stop event-hub

# Restore previous binary
cp event-hub.backup.YYYYMMDD event-hub

# Restore database (if needed)
pg_restore -U event_hub_user -d event_hub_prod -c backup.dump

# Start service
sudo systemctl start event-hub
```

### Docker Rollback

```bash
# Tag current version before update
docker tag event_hub_backend:latest event_hub_backend:previous

# If issues occur, rollback
docker compose -f docker/docker-compose.prod.yaml down
docker tag event_hub_backend:previous event_hub_backend:latest
docker compose -f docker/docker-compose.prod.yaml up -d
```

## Maintenance

### Zero-Downtime Deployment

1. Deploy new version alongside old version
2. Test new version health checks
3. Gradually shift traffic to new version
4. Monitor for errors
5. Complete switch or rollback if issues

### Database Migrations

```bash
# Test migration on staging first
# Backup database before migration
# Run migration during low-traffic period
# Monitor for errors

# Rollback if needed (run down migration)
```

---

**Production Deployment Checklist**: Always test on staging environment first!

**Need help?** Consult [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) or open an issue.
