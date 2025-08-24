# Exim Control Panel Deployment Guide

This guide covers deployment strategies, configuration management, and operational considerations for the Exim Control Panel in production environments.

## Table of Contents

- [Deployment Architecture](#deployment-architecture)
- [Environment Configuration](#environment-configuration)
- [Database Management](#database-management)
- [Security Configuration](#security-configuration)
- [Monitoring and Logging](#monitoring-and-logging)
- [Backup and Recovery](#backup-and-recovery)
- [Performance Tuning](#performance-tuning)
- [High Availability](#high-availability)
- [Troubleshooting](#troubleshooting)

## Deployment Architecture

### Single Server Deployment

```
┌─────────────────────────────────────┐
│           Server                    │
│  ┌─────────────────────────────────┐│
│  │        Exim-Pilot               ││
│  │  ┌─────────────┐ ┌─────────────┐││
│  │  │   Web UI    │ │   REST API  │││
│  │  └─────────────┘ └─────────────┘││
│  │  ┌─────────────┐ ┌─────────────┐││
│  │  │Log Processor│ │Queue Monitor│││
│  │  └─────────────┘ └─────────────┘││
│  │  ┌─────────────────────────────┐││
│  │  │        SQLite DB            │││
│  │  └─────────────────────────────┘││
│  └─────────────────────────────────┘│
│  ┌─────────────────────────────────┐│
│  │            Exim                 ││
│  │  ┌─────────┐ ┌─────────────────┐││
│  │  │ Logs    │ │     Queue       │││
│  │  └─────────┘ └─────────────────┘││
│  └─────────────────────────────────┘│
└─────────────────────────────────────┘
```

### Multi-Server Deployment with Load Balancer

```
                    ┌─────────────────┐
                    │  Load Balancer  │
                    │   (Nginx/HAProxy)│
                    └─────────┬───────┘
                              │
              ┌───────────────┼───────────────┐
              │               │               │
    ┌─────────▼─────┐ ┌───────▼─────┐ ┌───────▼─────┐
    │ Exim-Pilot 1  │ │ Exim-Pilot 2│ │ Exim-Pilot 3│
    │ (Read-Only)   │ │ (Read-Only) │ │ (Primary)   │
    └─────────┬─────┘ └───────┬─────┘ └───────┬─────┘
              │               │               │
              └───────────────┼───────────────┘
                              │
                    ┌─────────▼─────────┐
                    │   Shared Storage  │
                    │   (NFS/GlusterFS) │
                    └───────────────────┘
```

## Environment Configuration

### Production Configuration Template

Create `/opt/exim-pilot/config/production.yaml`:

```yaml
# Production Configuration for Exim Control Panel
server:
  port: 8080
  host: "127.0.0.1"  # Bind to localhost when using reverse proxy
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 120
  allowed_origins: 
    - "https://mail.yourdomain.com"
    - "https://exim.yourdomain.com"
  log_requests: true
  tls_enabled: true
  tls_cert_file: "/etc/ssl/certs/exim-pilot.crt"
  tls_key_file: "/etc/ssl/private/exim-pilot.key"

database:
  path: "/opt/exim-pilot/data/exim-pilot.db"
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: 10  # minutes
  backup_enabled: true
  backup_interval: 6     # hours
  backup_path: "/opt/exim-pilot/backups"

exim:
  log_paths:
    - "/var/log/exim4/mainlog"
    - "/var/log/exim4/rejectlog"
    - "/var/log/exim4/paniclog"
  spool_dir: "/var/spool/exim4"
  binary_path: "/usr/sbin/exim4"
  config_file: "/etc/exim4/exim4.conf"
  queue_run_user: "Debian-exim"
  log_rotation_dir: "/var/log/exim4"

logging:
  level: "warn"  # Reduced logging for production
  file: "/opt/exim-pilot/logs/exim-pilot.log"
  max_size: 100
  max_backups: 10
  max_age: 60
  compress: true

retention:
  log_entries_days: 180      # Longer retention for production
  audit_log_days: 730       # 2 years for compliance
  queue_snapshots_days: 90
  delivery_attempt_days: 365
  cleanup_interval: 12       # More frequent cleanup

security:
  session_timeout: 30        # Shorter timeout for security
  max_login_attempts: 3
  login_lockout_time: 30
  csrf_protection: true
  secure_cookies: true
  content_redaction: true
  audit_all_actions: true
  trusted_proxies:
    - "127.0.0.1"
    - "10.0.0.0/8"
    - "172.16.0.0/12"
    - "192.168.0.0/16"

auth:
  default_username: "admin"
  default_password: ""       # Set via environment variable
  password_min_length: 12    # Stronger passwords
  require_strong_password: true
  session_secret: ""         # Set via environment variable
```

### Environment Variables for Production

Create `/opt/exim-pilot/config/production.env`:

```bash
# Production Environment Variables
EXIM_PILOT_CONFIG=/opt/exim-pilot/config/production.yaml
EXIM_PILOT_LOG_LEVEL=warn

# Security
EXIM_PILOT_ADMIN_PASSWORD=your-very-secure-password-here
EXIM_PILOT_SESSION_SECRET=your-cryptographically-secure-session-secret-here

# Database
EXIM_PILOT_DB_PATH=/opt/exim-pilot/data/exim-pilot.db

# TLS
EXIM_PILOT_TLS_ENABLED=true
EXIM_PILOT_TLS_CERT=/etc/ssl/certs/exim-pilot.crt
EXIM_PILOT_TLS_KEY=/etc/ssl/private/exim-pilot.key

# Performance
EXIM_PILOT_DB_MAX_CONNS=50
```

### Systemd Service for Production

Update `/etc/systemd/system/exim-pilot.service`:

```ini
[Unit]
Description=Exim Control Panel (Exim-Pilot)
Documentation=https://github.com/andreitelteu/exim-pilot
After=network.target exim4.service
Wants=network.target
Requires=exim4.service

[Service]
Type=simple
User=exim-pilot
Group=exim-pilot
WorkingDirectory=/opt/exim-pilot

# Load environment variables
EnvironmentFile=/opt/exim-pilot/config/production.env

ExecStart=/opt/exim-pilot/bin/exim-pilot
ExecReload=/bin/kill -HUP $MAINPID

# Restart policy
Restart=always
RestartSec=10
StartLimitInterval=60
StartLimitBurst=3

# Output
StandardOutput=journal
StandardError=journal
SyslogIdentifier=exim-pilot

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/exim-pilot/data /opt/exim-pilot/logs /opt/exim-pilot/backups
ReadOnlyPaths=/var/log/exim4 /var/spool/exim4 /etc/ssl

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096
LimitMEMLOCK=64M

# Additional hardening
CapabilityBoundingSet=
AmbientCapabilities=
SystemCallFilter=@system-service
SystemCallErrorNumber=EPERM
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictRealtime=true
RestrictNamespaces=true
LockPersonality=true
MemoryDenyWriteExecute=true
RestrictSUIDSGID=true
RemoveIPC=true

[Install]
WantedBy=multi-user.target
```

## Database Management

### Migration Strategy

#### Pre-Deployment Migration Check

```bash
#!/bin/bash
# pre-deploy-check.sh

CONFIG_FILE="/opt/exim-pilot/config/production.yaml"
BACKUP_DIR="/opt/exim-pilot/backups"

echo "Pre-deployment migration check..."

# Backup current database
echo "Creating pre-deployment backup..."
sudo -u exim-pilot cp /opt/exim-pilot/data/exim-pilot.db \
  "$BACKUP_DIR/pre-deploy-$(date +%Y%m%d_%H%M%S).db"

# Check pending migrations
echo "Checking for pending migrations..."
sudo -u exim-pilot /opt/exim-pilot/bin/exim-pilot-config \
  -migrate status -config "$CONFIG_FILE"

# Validate configuration
echo "Validating configuration..."
sudo -u exim-pilot /opt/exim-pilot/bin/exim-pilot-config \
  -validate -config "$CONFIG_FILE"

echo "Pre-deployment check completed."
```

#### Deployment Migration Script

```bash
#!/bin/bash
# deploy-migrate.sh

set -e

CONFIG_FILE="/opt/exim-pilot/config/production.yaml"
SERVICE_NAME="exim-pilot"

echo "Starting deployment migration..."

# Stop service
echo "Stopping service..."
sudo systemctl stop "$SERVICE_NAME"

# Run migrations
echo "Running database migrations..."
sudo -u exim-pilot /opt/exim-pilot/bin/exim-pilot-config \
  -migrate up -config "$CONFIG_FILE"

# Start service
echo "Starting service..."
sudo systemctl start "$SERVICE_NAME"

# Verify service is running
sleep 5
if sudo systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "Service started successfully"
else
    echo "Service failed to start"
    sudo systemctl status "$SERVICE_NAME"
    exit 1
fi

echo "Deployment migration completed."
```

### Database Backup Strategy

#### Automated Backup Script

Create `/opt/exim-pilot/scripts/backup-database.sh`:

```bash
#!/bin/bash

set -e

# Configuration
DB_PATH="/opt/exim-pilot/data/exim-pilot.db"
BACKUP_DIR="/opt/exim-pilot/backups"
RETENTION_DAYS=30
REMOTE_BACKUP_HOST="backup.yourdomain.com"
REMOTE_BACKUP_PATH="/backups/exim-pilot"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Generate backup filename
BACKUP_FILE="exim-pilot-$(date +%Y%m%d_%H%M%S).db"
BACKUP_PATH="$BACKUP_DIR/$BACKUP_FILE"

echo "Creating database backup: $BACKUP_FILE"

# Create backup using SQLite backup command
sqlite3 "$DB_PATH" ".backup '$BACKUP_PATH'"

# Compress backup
gzip "$BACKUP_PATH"
BACKUP_PATH="$BACKUP_PATH.gz"

echo "Backup created: $BACKUP_PATH"

# Upload to remote backup server (optional)
if [ -n "$REMOTE_BACKUP_HOST" ]; then
    echo "Uploading backup to remote server..."
    scp "$BACKUP_PATH" "$REMOTE_BACKUP_HOST:$REMOTE_BACKUP_PATH/"
    echo "Remote backup completed"
fi

# Clean up old backups
echo "Cleaning up old backups (older than $RETENTION_DAYS days)..."
find "$BACKUP_DIR" -name "exim-pilot-*.db.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup process completed"
```

#### Cron Job for Automated Backups

Add to crontab for exim-pilot user:

```bash
# Edit crontab
sudo -u exim-pilot crontab -e

# Add backup job (every 6 hours)
0 */6 * * * /opt/exim-pilot/scripts/backup-database.sh >> /opt/exim-pilot/logs/backup.log 2>&1
```

### Database Recovery

#### Recovery Script

Create `/opt/exim-pilot/scripts/restore-database.sh`:

```bash
#!/bin/bash

set -e

if [ $# -ne 1 ]; then
    echo "Usage: $0 <backup-file>"
    echo "Example: $0 /opt/exim-pilot/backups/exim-pilot-20231201_120000.db.gz"
    exit 1
fi

BACKUP_FILE="$1"
DB_PATH="/opt/exim-pilot/data/exim-pilot.db"
SERVICE_NAME="exim-pilot"

# Verify backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
    echo "Backup file not found: $BACKUP_FILE"
    exit 1
fi

echo "Restoring database from: $BACKUP_FILE"

# Stop service
echo "Stopping service..."
sudo systemctl stop "$SERVICE_NAME"

# Backup current database
echo "Backing up current database..."
cp "$DB_PATH" "$DB_PATH.pre-restore-$(date +%Y%m%d_%H%M%S)"

# Restore from backup
echo "Restoring database..."
if [[ "$BACKUP_FILE" == *.gz ]]; then
    gunzip -c "$BACKUP_FILE" > "$DB_PATH"
else
    cp "$BACKUP_FILE" "$DB_PATH"
fi

# Set proper ownership
chown exim-pilot:exim-pilot "$DB_PATH"
chmod 640 "$DB_PATH"

# Start service
echo "Starting service..."
sudo systemctl start "$SERVICE_NAME"

# Verify service is running
sleep 5
if sudo systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "Database restore completed successfully"
else
    echo "Service failed to start after restore"
    sudo systemctl status "$SERVICE_NAME"
    exit 1
fi
```

## Security Configuration

### TLS/SSL Configuration

#### Let's Encrypt Setup

```bash
#!/bin/bash
# setup-letsencrypt.sh

DOMAIN="exim.yourdomain.com"
EMAIL="admin@yourdomain.com"

# Install certbot
sudo apt update
sudo apt install certbot

# Stop exim-pilot temporarily
sudo systemctl stop exim-pilot

# Obtain certificate
sudo certbot certonly --standalone \
  --email "$EMAIL" \
  --agree-tos \
  --no-eff-email \
  -d "$DOMAIN"

# Set up certificate renewal
sudo crontab -l | { cat; echo "0 12 * * * /usr/bin/certbot renew --quiet --post-hook 'systemctl reload exim-pilot'"; } | sudo crontab -

# Update configuration
sudo sed -i 's/tls_enabled: false/tls_enabled: true/' /opt/exim-pilot/config/production.yaml
sudo sed -i "s|tls_cert_file: \"\"|tls_cert_file: \"/etc/letsencrypt/live/$DOMAIN/fullchain.pem\"|" /opt/exim-pilot/config/production.yaml
sudo sed -i "s|tls_key_file: \"\"|tls_key_file: \"/etc/letsencrypt/live/$DOMAIN/privkey.pem\"|" /opt/exim-pilot/config/production.yaml

# Start service
sudo systemctl start exim-pilot

echo "TLS configuration completed for $DOMAIN"
```

### Firewall Configuration

#### UFW Rules

```bash
#!/bin/bash
# setup-firewall.sh

# Reset UFW
sudo ufw --force reset

# Default policies
sudo ufw default deny incoming
sudo ufw default allow outgoing

# SSH access
sudo ufw allow ssh

# HTTP/HTTPS for reverse proxy
sudo ufw allow 80
sudo ufw allow 443

# Exim-Pilot (if direct access needed)
# sudo ufw allow 8080

# Exim SMTP
sudo ufw allow 25
sudo ufw allow 587
sudo ufw allow 465

# Enable firewall
sudo ufw --force enable

echo "Firewall configuration completed"
```

#### iptables Rules

```bash
#!/bin/bash
# setup-iptables.sh

# Flush existing rules
iptables -F
iptables -X
iptables -t nat -F
iptables -t nat -X

# Default policies
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# Allow loopback
iptables -A INPUT -i lo -j ACCEPT

# Allow established connections
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# SSH
iptables -A INPUT -p tcp --dport 22 -j ACCEPT

# HTTP/HTTPS
iptables -A INPUT -p tcp --dport 80 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -j ACCEPT

# Exim SMTP
iptables -A INPUT -p tcp --dport 25 -j ACCEPT
iptables -A INPUT -p tcp --dport 587 -j ACCEPT
iptables -A INPUT -p tcp --dport 465 -j ACCEPT

# Exim-Pilot (restrict to specific networks)
iptables -A INPUT -p tcp --dport 8080 -s 192.168.0.0/16 -j ACCEPT
iptables -A INPUT -p tcp --dport 8080 -s 10.0.0.0/8 -j ACCEPT

# Save rules
iptables-save > /etc/iptables/rules.v4

echo "iptables configuration completed"
```

### Reverse Proxy Configuration

#### Nginx Configuration

Create `/etc/nginx/sites-available/exim-pilot`:

```nginx
# Rate limiting
limit_req_zone $binary_remote_addr zone=login:10m rate=5r/m;
limit_req_zone $binary_remote_addr zone=api:10m rate=30r/m;

# Upstream backend
upstream exim_pilot_backend {
    server 127.0.0.1:8080 max_fails=3 fail_timeout=30s;
    keepalive 32;
}

# HTTP to HTTPS redirect
server {
    listen 80;
    server_name exim.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    server_name exim.yourdomain.com;

    # SSL configuration
    ssl_certificate /etc/letsencrypt/live/exim.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/exim.yourdomain.com/privkey.pem;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_session_tickets off;

    # Modern configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # HSTS
    add_header Strict-Transport-Security "max-age=63072000" always;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Referrer-Policy "strict-origin-when-cross-origin";
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self' wss:";

    # Logging
    access_log /var/log/nginx/exim-pilot.access.log;
    error_log /var/log/nginx/exim-pilot.error.log;

    # Client settings
    client_max_body_size 10M;
    client_body_timeout 60s;
    client_header_timeout 60s;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;

    # Rate limiting for login
    location /api/v1/auth/login {
        limit_req zone=login burst=3 nodelay;
        proxy_pass http://exim_pilot_backend;
        include /etc/nginx/proxy_params;
    }

    # Rate limiting for API
    location /api/ {
        limit_req zone=api burst=10 nodelay;
        proxy_pass http://exim_pilot_backend;
        include /etc/nginx/proxy_params;
        
        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # Static files with caching
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        proxy_pass http://exim_pilot_backend;
        include /etc/nginx/proxy_params;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Main application
    location / {
        proxy_pass http://exim_pilot_backend;
        include /etc/nginx/proxy_params;
        
        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # Health check endpoint
    location /health {
        proxy_pass http://exim_pilot_backend/api/v1/health;
        include /etc/nginx/proxy_params;
        access_log off;
    }
}
```

Create `/etc/nginx/proxy_params`:

```nginx
proxy_set_header Host $http_host;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Proto $scheme;
proxy_set_header X-Forwarded-Host $host;
proxy_set_header X-Forwarded-Port $server_port;

proxy_connect_timeout 30s;
proxy_send_timeout 30s;
proxy_read_timeout 30s;

proxy_buffering on;
proxy_buffer_size 4k;
proxy_buffers 8 4k;
proxy_busy_buffers_size 8k;
```

## Monitoring and Logging

### Application Monitoring

#### Health Check Script

Create `/opt/exim-pilot/scripts/health-check.sh`:

```bash
#!/bin/bash

set -e

# Configuration
SERVICE_NAME="exim-pilot"
HEALTH_URL="http://localhost:8080/api/v1/health"
LOG_FILE="/opt/exim-pilot/logs/health-check.log"
ALERT_EMAIL="admin@yourdomain.com"

# Function to log messages
log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Function to send alert
send_alert() {
    local subject="$1"
    local message="$2"
    echo "$message" | mail -s "$subject" "$ALERT_EMAIL"
}

# Check if service is running
if ! systemctl is-active --quiet "$SERVICE_NAME"; then
    log_message "ERROR: Service $SERVICE_NAME is not running"
    send_alert "Exim-Pilot Service Down" "The Exim-Pilot service is not running on $(hostname)"
    exit 1
fi

# Check HTTP health endpoint
if ! curl -f -s "$HEALTH_URL" > /dev/null; then
    log_message "ERROR: Health check endpoint failed"
    send_alert "Exim-Pilot Health Check Failed" "The Exim-Pilot health check endpoint is not responding on $(hostname)"
    exit 1
fi

# Check database connectivity
if ! sudo -u exim-pilot sqlite3 /opt/exim-pilot/data/exim-pilot.db "SELECT 1;" > /dev/null 2>&1; then
    log_message "ERROR: Database connectivity check failed"
    send_alert "Exim-Pilot Database Error" "Cannot connect to Exim-Pilot database on $(hostname)"
    exit 1
fi

# Check disk space
DISK_USAGE=$(df /opt/exim-pilot | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -gt 90 ]; then
    log_message "WARNING: Disk usage is ${DISK_USAGE}%"
    send_alert "Exim-Pilot Disk Space Warning" "Disk usage is ${DISK_USAGE}% on $(hostname)"
fi

# Check log file size
LOG_SIZE=$(du -m /opt/exim-pilot/logs/exim-pilot.log 2>/dev/null | cut -f1 || echo 0)
if [ "$LOG_SIZE" -gt 100 ]; then
    log_message "WARNING: Log file size is ${LOG_SIZE}MB"
fi

log_message "INFO: Health check passed"
```

#### Cron Job for Health Checks

```bash
# Add to crontab
sudo crontab -e

# Health check every 5 minutes
*/5 * * * * /opt/exim-pilot/scripts/health-check.sh
```

### Log Management

#### Centralized Logging with rsyslog

Create `/etc/rsyslog.d/30-exim-pilot.conf`:

```
# Exim-Pilot logging configuration
if $programname == 'exim-pilot' then {
    /var/log/exim-pilot/exim-pilot.log
    stop
}

# Forward to central log server (optional)
if $programname == 'exim-pilot' then {
    @@logserver.yourdomain.com:514
}
```

#### Log Rotation Configuration

Update `/etc/logrotate.d/exim-pilot`:

```
/opt/exim-pilot/logs/*.log {
    daily
    missingok
    rotate 60
    compress
    delaycompress
    notifempty
    create 640 exim-pilot exim-pilot
    sharedscripts
    postrotate
        systemctl reload exim-pilot > /dev/null 2>&1 || true
    endscript
}

/var/log/exim-pilot/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 640 syslog adm
    sharedscripts
    postrotate
        systemctl reload rsyslog > /dev/null 2>&1 || true
    endscript
}
```

## Performance Tuning

### Database Optimization

#### SQLite Performance Settings

Create `/opt/exim-pilot/scripts/optimize-database.sh`:

```bash
#!/bin/bash

DB_PATH="/opt/exim-pilot/data/exim-pilot.db"

echo "Optimizing database performance..."

# Run VACUUM to reclaim space and defragment
sudo -u exim-pilot sqlite3 "$DB_PATH" "VACUUM;"

# Update statistics
sudo -u exim-pilot sqlite3 "$DB_PATH" "ANALYZE;"

# Set performance pragmas (these are set at runtime by the application)
echo "Database optimization completed"
```

#### Database Maintenance Cron Job

```bash
# Add to exim-pilot user crontab
sudo -u exim-pilot crontab -e

# Run database optimization weekly
0 2 * * 0 /opt/exim-pilot/scripts/optimize-database.sh >> /opt/exim-pilot/logs/maintenance.log 2>&1
```

### System Performance

#### Kernel Parameters

Add to `/etc/sysctl.d/99-exim-pilot.conf`:

```
# Network performance
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216

# File system performance
fs.file-max = 65536
vm.swappiness = 10

# SQLite performance
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5
```

Apply settings:
```bash
sudo sysctl -p /etc/sysctl.d/99-exim-pilot.conf
```

#### File Limits

Add to `/etc/security/limits.d/exim-pilot.conf`:

```
exim-pilot soft nofile 65536
exim-pilot hard nofile 65536
exim-pilot soft nproc 4096
exim-pilot hard nproc 4096
```

## High Availability

### Load Balancer Configuration

#### HAProxy Configuration

Create `/etc/haproxy/haproxy.cfg`:

```
global
    daemon
    chroot /var/lib/haproxy
    stats socket /run/haproxy/admin.sock mode 660 level admin
    stats timeout 30s
    user haproxy
    group haproxy

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    option httplog
    option dontlognull

frontend exim_pilot_frontend
    bind *:80
    bind *:443 ssl crt /etc/ssl/certs/exim-pilot.pem
    redirect scheme https if !{ ssl_fc }
    
    # Health check
    monitor-uri /health
    
    default_backend exim_pilot_backend

backend exim_pilot_backend
    balance roundrobin
    option httpchk GET /api/v1/health
    
    server pilot1 10.0.1.10:8080 check inter 5s fall 3 rise 2
    server pilot2 10.0.1.11:8080 check inter 5s fall 3 rise 2
    server pilot3 10.0.1.12:8080 check inter 5s fall 3 rise 2

listen stats
    bind *:8404
    stats enable
    stats uri /stats
    stats refresh 30s
    stats admin if TRUE
```

### Shared Storage Setup

#### NFS Configuration

On NFS server:

```bash
# Install NFS server
sudo apt install nfs-kernel-server

# Create shared directory
sudo mkdir -p /srv/nfs/exim-pilot
sudo chown exim-pilot:exim-pilot /srv/nfs/exim-pilot

# Configure exports
echo "/srv/nfs/exim-pilot 10.0.1.0/24(rw,sync,no_subtree_check,no_root_squash)" | sudo tee -a /etc/exports

# Export shares
sudo exportfs -a
sudo systemctl restart nfs-kernel-server
```

On client servers:

```bash
# Install NFS client
sudo apt install nfs-common

# Mount shared storage
sudo mkdir -p /opt/exim-pilot/shared
sudo mount -t nfs nfs-server:/srv/nfs/exim-pilot /opt/exim-pilot/shared

# Add to fstab for persistent mount
echo "nfs-server:/srv/nfs/exim-pilot /opt/exim-pilot/shared nfs defaults 0 0" | sudo tee -a /etc/fstab
```

## Troubleshooting

### Common Issues

#### Service Won't Start

```bash
# Check service status
sudo systemctl status exim-pilot

# Check configuration
sudo /opt/exim-pilot/bin/exim-pilot-config -validate

# Check logs
sudo journalctl -u exim-pilot -n 50

# Check file permissions
ls -la /opt/exim-pilot/
```

#### Database Issues

```bash
# Check database integrity
sudo -u exim-pilot sqlite3 /opt/exim-pilot/data/exim-pilot.db "PRAGMA integrity_check;"

# Check migration status
sudo /opt/exim-pilot/bin/exim-pilot-config -migrate status

# Repair database if needed
sudo -u exim-pilot sqlite3 /opt/exim-pilot/data/exim-pilot.db ".recover"
```

#### Performance Issues

```bash
# Check system resources
top
htop
iotop

# Check database size
du -h /opt/exim-pilot/data/

# Check log file sizes
du -h /opt/exim-pilot/logs/

# Run database optimization
/opt/exim-pilot/scripts/optimize-database.sh
```

### Debugging Tools

#### Debug Configuration

Create debug configuration file `/opt/exim-pilot/config/debug.yaml`:

```yaml
# Debug configuration - DO NOT USE IN PRODUCTION
logging:
  level: "debug"
  file: "/opt/exim-pilot/logs/debug.log"

server:
  log_requests: true

# Enable all debugging features
debug:
  enable_profiling: true
  enable_metrics: true
  log_sql_queries: true
```

#### Performance Profiling

```bash
# Enable Go profiling (if built with profiling support)
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof

# Analyze with go tool pprof
go tool pprof cpu.prof
```

This deployment guide provides comprehensive coverage of production deployment scenarios, security considerations, and operational procedures for the Exim Control Panel.