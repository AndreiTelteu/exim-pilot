# Exim Control Panel Deployment

This directory contains deployment configurations and scripts for the Exim Control Panel.

## Contents

- `install.sh` - Automated installation script for Ubuntu/Debian systems
- `systemd/` - Systemd service configuration files
- `nginx/` - Nginx reverse proxy configuration examples
- `docker/` - Docker deployment configurations (if applicable)

## Quick Installation

### Prerequisites

- Ubuntu 18.04+ or Debian 9+
- Exim 4.90+ installed and configured
- Root access for installation

### Automated Installation

1. **Download the installer**:

   ```bash
   wget https://github.com/andreitelteu/exim-pilot/releases/latest/download/install.sh
   chmod +x install.sh
   ```

2. **Run the installer**:

   ```bash
   sudo ./install.sh
   ```

3. **Access the web interface**:
   - Open http://your-server-ip:8080
   - Login with username: `admin`, password: `admin123`
   - **IMPORTANT**: Change the default password immediately!

### Manual Installation

See [INSTALLATION.md](../docs/INSTALLATION.md) for detailed manual installation instructions.

## Configuration Management

### Configuration Tool

The `exim-pilot-config` tool helps manage configuration and database migrations:

```bash
# Generate default configuration
exim-pilot-config -generate -config /opt/exim-pilot/config/config.yaml

# Validate configuration
exim-pilot-config -validate -config /opt/exim-pilot/config/config.yaml

# Run database migrations
exim-pilot-config -migrate up -config /opt/exim-pilot/config/config.yaml

# Check migration status
exim-pilot-config -migrate status -config /opt/exim-pilot/config/config.yaml
```

### Configuration File

The main configuration file is located at `/opt/exim-pilot/config/config.yaml`. Key sections include:

- **Server**: HTTP server settings, TLS configuration
- **Database**: SQLite database settings and backup configuration
- **Exim**: Exim binary paths, log file locations, spool directory
- **Logging**: Application logging configuration
- **Security**: Authentication, session management, security policies
- **Retention**: Data retention policies for logs and audit trails

### Environment Variables

Configuration can be overridden using environment variables:

```bash
export EXIM_PILOT_PORT=8080
export EXIM_PILOT_ADMIN_PASSWORD="your-secure-password"
export EXIM_PILOT_TLS_ENABLED=true
export EXIM_PILOT_LOG_LEVEL=info
```

## Service Management

### Systemd Service

The application runs as a systemd service:

```bash
# Start service
sudo systemctl start exim-pilot

# Stop service
sudo systemctl stop exim-pilot

# Restart service
sudo systemctl restart exim-pilot

# Check status
sudo systemctl status exim-pilot

# View logs
sudo journalctl -u exim-pilot -f
```

### Service Configuration

The systemd service file is located at `/etc/systemd/system/exim-pilot.service` and includes:

- Security hardening settings
- Resource limits
- Automatic restart policies
- Environment variable loading

## Security Considerations

### File Permissions

- Configuration files: `640` (readable by exim-pilot user only)
- Database files: `640` (readable by exim-pilot user only)
- Log files: `640` (readable by exim-pilot user only)
- Binary files: `755` (executable by all, writable by root only)

### Network Security

- Default port: 8080 (configurable)
- TLS/HTTPS support with Let's Encrypt or custom certificates
- CORS configuration for web interface security
- Rate limiting and login attempt restrictions

### System Security

- Runs as dedicated `exim-pilot` user (not root)
- Minimal file system permissions
- systemd security hardening enabled
- Input validation and SQL injection prevention

## Backup and Recovery

### Automatic Backups

The application includes automatic database backup functionality:

- Configurable backup interval (default: 24 hours)
- Compressed backup files
- Configurable retention period
- Optional remote backup upload

### Manual Backup

```bash
# Create manual backup
sudo -u exim-pilot cp /opt/exim-pilot/data/exim-pilot.db \
  /opt/exim-pilot/backups/manual-backup-$(date +%Y%m%d_%H%M%S).db

# Restore from backup
sudo systemctl stop exim-pilot
sudo -u exim-pilot cp /opt/exim-pilot/backups/backup-file.db \
  /opt/exim-pilot/data/exim-pilot.db
sudo systemctl start exim-pilot
```

## Monitoring and Maintenance

### Health Checks

The application provides health check endpoints:

- `GET /api/v1/health` - Basic health check
- `GET /api/v1/status` - Detailed system status

### Log Management

- Application logs: `/opt/exim-pilot/logs/exim-pilot.log`
- System logs: `journalctl -u exim-pilot`
- Log rotation configured automatically
- Configurable log levels and retention

### Performance Monitoring

Monitor key metrics:

- Database size and performance
- Memory and CPU usage
- Log processing throughput
- Queue processing performance

## Troubleshooting

### Common Issues

1. **Service won't start**:

   ```bash
   sudo systemctl status exim-pilot
   sudo journalctl -u exim-pilot -n 50
   ```

2. **Configuration errors**:

   ```bash
   exim-pilot-config -validate -config /opt/exim-pilot/config/config.yaml
   ```

3. **Database issues**:

   ```bash
   exim-pilot-config -migrate status -config /opt/exim-pilot/config/config.yaml
   ```

4. **Permission problems**:
   ```bash
   sudo chown -R exim-pilot:exim-pilot /opt/exim-pilot/
   sudo chmod 640 /opt/exim-pilot/config/config.yaml
   ```

### Debug Mode

Enable debug logging for troubleshooting:

```bash
# Set environment variable
export EXIM_PILOT_LOG_LEVEL=debug

# Or update configuration file
logging:
  level: "debug"
```

## Updates and Upgrades

### Update Process

1. **Stop the service**:

   ```bash
   sudo systemctl stop exim-pilot
   ```

2. **Backup current installation**:

   ```bash
   sudo cp -r /opt/exim-pilot /opt/exim-pilot.backup
   ```

3. **Download new version**:

   ```bash
   wget https://github.com/andreitelteu/exim-pilot/releases/latest/download/exim-pilot
   ```

4. **Replace binary**:

   ```bash
   sudo cp exim-pilot /opt/exim-pilot/bin/
   sudo chown exim-pilot:exim-pilot /opt/exim-pilot/bin/exim-pilot
   sudo chmod 755 /opt/exim-pilot/bin/exim-pilot
   ```

5. **Run migrations**:

   ```bash
   sudo -u exim-pilot exim-pilot-config -migrate up -config /opt/exim-pilot/config/config.yaml
   ```

6. **Start service**:
   ```bash
   sudo systemctl start exim-pilot
   ```

### Rollback Process

If an update fails:

1. **Stop service**:

   ```bash
   sudo systemctl stop exim-pilot
   ```

2. **Restore backup**:

   ```bash
   sudo rm -rf /opt/exim-pilot
   sudo mv /opt/exim-pilot.backup /opt/exim-pilot
   ```

3. **Start service**:
   ```bash
   sudo systemctl start exim-pilot
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

# Remove application files
sudo rm -rf /opt/exim-pilot

# Remove system user
sudo userdel exim-pilot
sudo groupdel exim-pilot

# Remove logrotate configuration
sudo rm /etc/logrotate.d/exim-pilot
```

## Support

For additional help:

- Check the [Installation Guide](../docs/INSTALLATION.md)
- Check the [Deployment Guide](../docs/DEPLOYMENT.md)
- Review application logs
- Check GitHub issues and documentation

## File Structure

```
/opt/exim-pilot/
├── bin/
│   ├── exim-pilot              # Main application binary
│   └── exim-pilot-config       # Configuration management tool
├── config/
│   ├── config.yaml            # Main configuration file
│   └── production.env         # Environment variables (optional)
├── data/
│   └── exim-pilot.db          # SQLite database
├── logs/
│   └── exim-pilot.log         # Application logs
├── backups/                   # Database backups
└── scripts/                   # Maintenance scripts
```
