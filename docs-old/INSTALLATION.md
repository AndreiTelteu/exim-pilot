# Exim Control Panel Installation Guide

This guide provides comprehensive instructions for installing and configuring the Exim Control Panel (Exim-Pilot) on Ubuntu/Debian systems.

## Table of Contents

- [System Requirements](#system-requirements)
- [Pre-Installation](#pre-installation)
- [Installation Methods](#installation-methods)
- [Configuration](#configuration)
- [Post-Installation](#post-installation)
- [Security Hardening](#security-hardening)
- [Troubleshooting](#troubleshooting)
- [Uninstallation](#uninstallation)

## System Requirements

### Operating System
- Ubuntu 18.04 LTS or later
- Debian 9 (Stretch) or later

### Software Dependencies
- Exim 4.90 or later
- systemd (for service management)
- SQLite 3 (included with most distributions)

### Hardware Requirements
- **Minimum**: 512MB RAM, 1GB disk space
- **Recommended**: 1GB RAM, 5GB disk space
- **CPU**: Any modern x86_64 processor

### Network Requirements
- Port 8080 (default, configurable)
- HTTPS port 443 (if using TLS)

## Pre-Installation

### 1. Install Exim

If Exim is not already installed:

```bash
sudo apt update
sudo apt install exim4-daemon-heavy
```

Configure Exim according to your needs:

```bash
sudo dpkg-reconfigure exim4-config
```

### 2. Verify Exim Installation

Check that Exim is running and accessible:

```bash
# Check Exim status
sudo systemctl status exim4

# Verify Exim binary location
which exim4

# Check log files exist
ls -la /var/log/exim4/

# Check spool directory
ls -la /var/spool/exim4/
```

### 3. Install Build Dependencies (if building from source)

```bash
sudo apt install build-essential git golang-go nodejs npm
```

## Installation Methods

### Method 1: Automated Installation (Recommended)

1. **Download the installer**:
   ```bash
   wget https://github.com/andreitelteu/exim-pilot/releases/latest/download/install.sh
   chmod +x install.sh
   ```

2. **Run the installer**:
   ```bash
   sudo ./install.sh
   ```

The installer will:
- Create system user and directories
- Install the binary and configuration
- Set up systemd service
- Configure log rotation
- Initialize the database
- Start the service

### Method 2: Manual Installation

#### Step 1: Download Binary

```bash
# Download latest release
wget https://github.com/andreitelteu/exim-pilot/releases/latest/download/exim-pilot
chmod +x exim-pilot
```

#### Step 2: Create System User

```bash
sudo groupadd --system exim-pilot
sudo useradd --system --gid exim-pilot --home-dir /opt/exim-pilot \
             --shell /bin/false --comment "Exim Control Panel" exim-pilot
```

#### Step 3: Create Directory Structure

```bash
sudo mkdir -p /opt/exim-pilot/{bin,config,data,logs,backups}
sudo chown -R exim-pilot:exim-pilot /opt/exim-pilot
sudo chmod 755 /opt/exim-pilot
sudo chmod 755 /opt/exim-pilot/{bin,config}
sudo chmod 750 /opt/exim-pilot/{data,logs,backups}
```

#### Step 4: Install Binary

```bash
sudo cp exim-pilot /opt/exim-pilot/bin/
sudo chown exim-pilot:exim-pilot /opt/exim-pilot/bin/exim-pilot
sudo chmod 755 /opt/exim-pilot/bin/exim-pilot
```

#### Step 5: Create Configuration

```bash
sudo /opt/exim-pilot/bin/exim-pilot-config -generate -config /opt/exim-pilot/config/config.yaml
```

#### Step 6: Install Systemd Service

Create `/etc/systemd/system/exim-pilot.service`:

```ini
[Unit]
Description=Exim Control Panel (Exim-Pilot)
Documentation=https://github.com/andreitelteu/exim-pilot
After=network.target
Wants=network.target

[Service]
Type=simple
User=exim-pilot
Group=exim-pilot
WorkingDirectory=/opt/exim-pilot
ExecStart=/opt/exim-pilot/bin/exim-pilot
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

Environment=EXIM_PILOT_CONFIG=/opt/exim-pilot/config/config.yaml

[Install]
WantedBy=multi-user.target
```

#### Step 7: Configure Permissions

```bash
# Add exim-pilot user to adm group for log access
sudo usermod -a -G adm exim-pilot

# Ensure log directories are accessible
sudo chmod 755 /var/log/exim4
sudo chmod 755 /var/spool/exim4
```

#### Step 8: Initialize Database

```bash
sudo -u exim-pilot /opt/exim-pilot/bin/exim-pilot-config -migrate up -config /opt/exim-pilot/config/config.yaml
```

#### Step 9: Start Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable exim-pilot
sudo systemctl start exim-pilot
```

### Method 3: Build from Source

#### Step 1: Clone Repository

```bash
git clone https://github.com/andreitelteu/exim-pilot.git
cd exim-pilot
```

#### Step 2: Build Application

```bash
# Install Go dependencies
go mod download

# Build frontend
cd web
npm install
npm run build
cd ..

# Build backend with embedded frontend
make build
```

#### Step 3: Follow Manual Installation

Continue with steps 2-9 from the manual installation method.

## Configuration

### Configuration File Location

The main configuration file is located at:
```
/opt/exim-pilot/config/config.yaml
```

### Configuration Sections

#### Server Configuration

```yaml
server:
  port: 8080                    # HTTP port
  host: "0.0.0.0"              # Bind address
  read_timeout: 15             # Request read timeout (seconds)
  write_timeout: 15            # Response write timeout (seconds)
  idle_timeout: 60             # Connection idle timeout (seconds)
  allowed_origins: ["*"]       # CORS allowed origins
  log_requests: true           # Log HTTP requests
  tls_enabled: false           # Enable HTTPS
  tls_cert_file: ""           # TLS certificate file
  tls_key_file: ""            # TLS private key file
```

#### Database Configuration

```yaml
database:
  path: "/opt/exim-pilot/data/exim-pilot.db"
  max_open_conns: 25           # Maximum open connections
  max_idle_conns: 5            # Maximum idle connections
  conn_max_lifetime: 5         # Connection lifetime (minutes)
  backup_enabled: true         # Enable automatic backups
  backup_interval: 24          # Backup interval (hours)
  backup_path: "/opt/exim-pilot/backups"
```

#### Exim Configuration

```yaml
exim:
  log_paths:                   # Exim log file paths
    - "/var/log/exim4/mainlog"
    - "/var/log/exim4/rejectlog"
    - "/var/log/exim4/paniclog"
  spool_dir: "/var/spool/exim4"
  binary_path: "/usr/sbin/exim4"
  config_file: "/etc/exim4/exim4.conf"
  queue_run_user: "Debian-exim"
```

#### Security Configuration

```yaml
security:
  session_timeout: 60          # Session timeout (minutes)
  max_login_attempts: 5        # Max failed login attempts
  login_lockout_time: 15       # Lockout duration (minutes)
  csrf_protection: true        # Enable CSRF protection
  secure_cookies: true         # Use secure cookies
  content_redaction: true      # Redact sensitive content
  audit_all_actions: true     # Audit all administrative actions
```

### Environment Variables

Configuration can be overridden using environment variables:

```bash
# Server configuration
export EXIM_PILOT_PORT=8080
export EXIM_PILOT_HOST="0.0.0.0"
export EXIM_PILOT_TLS_ENABLED=false

# Database configuration
export EXIM_PILOT_DB_PATH="/opt/exim-pilot/data/exim-pilot.db"

# Exim configuration
export EXIM_PILOT_BINARY_PATH="/usr/sbin/exim4"
export EXIM_PILOT_SPOOL_DIR="/var/spool/exim4"

# Authentication
export EXIM_PILOT_ADMIN_USER="admin"
export EXIM_PILOT_ADMIN_PASSWORD="your-secure-password"
export EXIM_PILOT_SESSION_SECRET="your-session-secret"

# Logging
export EXIM_PILOT_LOG_LEVEL="info"
export EXIM_PILOT_LOG_FILE="/opt/exim-pilot/logs/exim-pilot.log"
```

### Configuration Validation

Validate your configuration:

```bash
sudo /opt/exim-pilot/bin/exim-pilot-config -validate -config /opt/exim-pilot/config/config.yaml
```

## Post-Installation

### 1. Access the Web Interface

Open your browser and navigate to:
```
http://your-server-ip:8080
```

Default credentials:
- **Username**: admin
- **Password**: admin123

### 2. Change Default Password

**IMPORTANT**: Change the default password immediately after installation.

1. Log in with default credentials
2. Navigate to User Settings
3. Change password to a strong, unique password

### 3. Configure TLS (Recommended)

For production deployments, enable TLS:

#### Option A: Let's Encrypt (Recommended)

```bash
# Install certbot
sudo apt install certbot

# Obtain certificate
sudo certbot certonly --standalone -d your-domain.com

# Update configuration
sudo nano /opt/exim-pilot/config/config.yaml
```

Update the configuration:
```yaml
server:
  tls_enabled: true
  tls_cert_file: "/etc/letsencrypt/live/your-domain.com/fullchain.pem"
  tls_key_file: "/etc/letsencrypt/live/your-domain.com/privkey.pem"
```

#### Option B: Self-Signed Certificate

```bash
# Generate self-signed certificate
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /opt/exim-pilot/config/server.key \
  -out /opt/exim-pilot/config/server.crt

# Set permissions
sudo chown exim-pilot:exim-pilot /opt/exim-pilot/config/server.*
sudo chmod 600 /opt/exim-pilot/config/server.key
sudo chmod 644 /opt/exim-pilot/config/server.crt
```

Update configuration:
```yaml
server:
  tls_enabled: true
  tls_cert_file: "/opt/exim-pilot/config/server.crt"
  tls_key_file: "/opt/exim-pilot/config/server.key"
```

### 4. Configure Firewall

```bash
# Allow HTTP (if not using TLS)
sudo ufw allow 8080

# Allow HTTPS (if using TLS)
sudo ufw allow 443

# Enable firewall if not already enabled
sudo ufw enable
```

### 5. Set Up Reverse Proxy (Optional)

#### Nginx Configuration

Create `/etc/nginx/sites-available/exim-pilot`:

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

Enable the site:
```bash
sudo ln -s /etc/nginx/sites-available/exim-pilot /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## Security Hardening

### 1. File Permissions

Ensure proper file permissions:

```bash
# Configuration file should be readable only by exim-pilot user
sudo chmod 640 /opt/exim-pilot/config/config.yaml
sudo chown exim-pilot:exim-pilot /opt/exim-pilot/config/config.yaml

# Data directory should be private
sudo chmod 750 /opt/exim-pilot/data
sudo chown exim-pilot:exim-pilot /opt/exim-pilot/data

# Log directory should be private
sudo chmod 750 /opt/exim-pilot/logs
sudo chown exim-pilot:exim-pilot /opt/exim-pilot/logs
```

### 2. Network Security

```bash
# Restrict access to specific IPs (example)
sudo ufw allow from 192.168.1.0/24 to any port 8080

# Or use iptables
sudo iptables -A INPUT -p tcp --dport 8080 -s 192.168.1.0/24 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 8080 -j DROP
```

### 3. System Security

```bash
# Keep system updated
sudo apt update && sudo apt upgrade

# Install fail2ban for additional protection
sudo apt install fail2ban

# Configure fail2ban for exim-pilot (optional)
sudo nano /etc/fail2ban/jail.local
```

### 4. Application Security

Update configuration for enhanced security:

```yaml
security:
  session_timeout: 30          # Shorter session timeout
  max_login_attempts: 3        # Fewer login attempts
  login_lockout_time: 30       # Longer lockout time
  secure_cookies: true         # Always use secure cookies
  trusted_proxies:             # Define trusted proxy IPs
    - "127.0.0.1"
    - "192.168.1.1"
```

## Troubleshooting

### Service Issues

#### Service Won't Start

```bash
# Check service status
sudo systemctl status exim-pilot

# Check logs
sudo journalctl -u exim-pilot -f

# Check configuration
sudo /opt/exim-pilot/bin/exim-pilot-config -validate
```

#### Permission Errors

```bash
# Check file ownership
ls -la /opt/exim-pilot/

# Fix ownership if needed
sudo chown -R exim-pilot:exim-pilot /opt/exim-pilot/

# Check Exim log access
sudo -u exim-pilot ls -la /var/log/exim4/
```

### Database Issues

#### Database Connection Errors

```bash
# Check database file permissions
ls -la /opt/exim-pilot/data/

# Test database connection
sudo -u exim-pilot sqlite3 /opt/exim-pilot/data/exim-pilot.db ".tables"

# Run database migrations
sudo /opt/exim-pilot/bin/exim-pilot-config -migrate status
```

#### Migration Failures

```bash
# Check migration status
sudo /opt/exim-pilot/bin/exim-pilot-config -migrate status

# Manually run migrations
sudo /opt/exim-pilot/bin/exim-pilot-config -migrate up
```

### Web Interface Issues

#### Can't Access Web Interface

```bash
# Check if service is listening
sudo netstat -tlnp | grep 8080

# Check firewall
sudo ufw status

# Check logs for errors
sudo journalctl -u exim-pilot -n 50
```

#### Login Issues

```bash
# Reset admin password (if needed)
sudo sqlite3 /opt/exim-pilot/data/exim-pilot.db "UPDATE users SET password_hash = 'new-hash' WHERE username = 'admin';"

# Check session configuration
grep -A 10 "auth:" /opt/exim-pilot/config/config.yaml
```

### Log Processing Issues

#### Logs Not Appearing

```bash
# Check log file permissions
ls -la /var/log/exim4/

# Check if exim-pilot user can read logs
sudo -u exim-pilot cat /var/log/exim4/mainlog | head -5

# Check log paths in configuration
grep -A 5 "log_paths:" /opt/exim-pilot/config/config.yaml
```

### Performance Issues

#### High Memory Usage

```bash
# Check memory usage
ps aux | grep exim-pilot

# Adjust database connection limits
nano /opt/exim-pilot/config/config.yaml
# Reduce max_open_conns and max_idle_conns
```

#### Slow Response Times

```bash
# Check database size
du -h /opt/exim-pilot/data/exim-pilot.db

# Run database cleanup
sudo /opt/exim-pilot/bin/exim-pilot-config -migrate status

# Check log retention settings
grep -A 5 "retention:" /opt/exim-pilot/config/config.yaml
```

## Maintenance

### Regular Tasks

#### Database Backups

Backups are automatic if enabled in configuration. Manual backup:

```bash
sudo -u exim-pilot cp /opt/exim-pilot/data/exim-pilot.db \
  /opt/exim-pilot/backups/exim-pilot-$(date +%Y%m%d_%H%M%S).db
```

#### Log Rotation

Log rotation is configured automatically. Check configuration:

```bash
cat /etc/logrotate.d/exim-pilot
```

#### Updates

```bash
# Stop service
sudo systemctl stop exim-pilot

# Backup current installation
sudo cp -r /opt/exim-pilot /opt/exim-pilot.backup

# Download new version
wget https://github.com/andreitelteu/exim-pilot/releases/latest/download/exim-pilot

# Replace binary
sudo cp exim-pilot /opt/exim-pilot/bin/
sudo chown exim-pilot:exim-pilot /opt/exim-pilot/bin/exim-pilot
sudo chmod 755 /opt/exim-pilot/bin/exim-pilot

# Run migrations if needed
sudo /opt/exim-pilot/bin/exim-pilot-config -migrate up

# Start service
sudo systemctl start exim-pilot
```

### Monitoring

#### Health Checks

```bash
# Service status
sudo systemctl is-active exim-pilot

# HTTP health check
curl -f http://localhost:8080/api/v1/health || echo "Service unhealthy"

# Database check
sudo -u exim-pilot sqlite3 /opt/exim-pilot/data/exim-pilot.db "SELECT COUNT(*) FROM schema_migrations;"
```

#### Log Monitoring

```bash
# Monitor application logs
sudo journalctl -u exim-pilot -f

# Monitor application log file
sudo tail -f /opt/exim-pilot/logs/exim-pilot.log

# Check for errors
sudo journalctl -u exim-pilot --since "1 hour ago" | grep -i error
```

## Uninstallation

### Automated Uninstallation

```bash
sudo ./install.sh --uninstall
```

### Manual Uninstallation

```bash
# Stop and disable service
sudo systemctl stop exim-pilot
sudo systemctl disable exim-pilot

# Remove service file
sudo rm /etc/systemd/system/exim-pilot.service
sudo systemctl daemon-reload

# Remove logrotate configuration
sudo rm /etc/logrotate.d/exim-pilot

# Remove application files (optional - preserves data)
sudo rm -rf /opt/exim-pilot

# Remove system user (optional)
sudo userdel exim-pilot
sudo groupdel exim-pilot
```

## Support

### Getting Help

- **Documentation**: Check this guide and other documentation in the `docs/` directory
- **Issues**: Report bugs and issues on GitHub
- **Logs**: Always include relevant log output when reporting issues

### Useful Commands

```bash
# Service management
sudo systemctl {start|stop|restart|status} exim-pilot

# Configuration management
sudo /opt/exim-pilot/bin/exim-pilot-config -help

# Database management
sudo /opt/exim-pilot/bin/exim-pilot-config -migrate status

# Log viewing
sudo journalctl -u exim-pilot -f
sudo tail -f /opt/exim-pilot/logs/exim-pilot.log

# Configuration validation
sudo /opt/exim-pilot/bin/exim-pilot-config -validate
```

This completes the comprehensive installation guide for the Exim Control Panel.