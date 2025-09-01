# Exim Control Panel Troubleshooting Guide

## Table of Contents

1. [Common Issues](#common-issues)
2. [Installation Problems](#installation-problems)
3. [Authentication Issues](#authentication-issues)
4. [Queue Management Problems](#queue-management-problems)
5. [Log Processing Issues](#log-processing-issues)
6. [Performance Problems](#performance-problems)
7. [Database Issues](#database-issues)
8. [Network and Connectivity](#network-and-connectivity)
9. [Exim Integration Problems](#exim-integration-problems)
10. [Diagnostic Tools](#diagnostic-tools)
11. [Log Analysis](#log-analysis)
12. [Getting Help](#getting-help)

## Common Issues

### Application Won't Start

**Symptoms:**
- Service fails to start
- "Connection refused" errors
- Process exits immediately

**Diagnostic Steps:**
1. Check service status:
   ```bash
   systemctl status exim-pilot
   ```

2. Check application logs:
   ```bash
   journalctl -u exim-pilot -f
   ```

3. Verify configuration:
   ```bash
   /opt/exim-pilot/bin/exim-pilot --config-check
   ```

**Common Causes and Solutions:**

**Port Already in Use:**
```bash
# Check what's using port 8080
sudo netstat -tlnp | grep :8080
sudo lsof -i :8080

# Solution: Change port in config.yaml or stop conflicting service
```

**Permission Issues:**
```bash
# Check file permissions
ls -la /opt/exim-pilot/
ls -la /var/log/exim4/

# Fix permissions
sudo chown -R exim-pilot:exim-pilot /opt/exim-pilot/
sudo chmod 755 /opt/exim-pilot/bin/exim-pilot
```

**Missing Dependencies:**
```bash
# Check for missing libraries
ldd /opt/exim-pilot/bin/exim-pilot

# Install missing packages
sudo apt update
sudo apt install libc6 libsqlite3-0
```

### Web Interface Not Loading

**Symptoms:**
- Blank page or loading spinner
- 404 errors for static assets
- JavaScript errors in browser console

**Diagnostic Steps:**
1. Check browser console for errors (F12)
2. Verify API connectivity:
   ```bash
   curl -u admin:password http://localhost:8080/api/v1/dashboard
   ```
3. Check if static assets are embedded:
   ```bash
   strings /opt/exim-pilot/bin/exim-pilot | grep -i "web/dist"
   ```

**Solutions:**

**Static Assets Not Embedded:**
```bash
# Rebuild with embedded assets
cd /opt/exim-pilot/source
make build-frontend
make build
sudo systemctl restart exim-pilot
```

**CORS Issues:**
- Check browser console for CORS errors
- Verify API base URL in frontend configuration
- Ensure proper CORS headers in server configuration

**JavaScript Errors:**
- Clear browser cache and cookies
- Try different browser or incognito mode
- Check for browser compatibility issues

## Installation Problems

### Database Initialization Fails

**Symptoms:**
- "Database connection failed" errors
- Missing tables errors
- Migration failures

**Solutions:**

**SQLite File Permissions:**
```bash
# Check database file and directory permissions
ls -la /opt/exim-pilot/data/
sudo chown exim-pilot:exim-pilot /opt/exim-pilot/data/exim-pilot.db
sudo chmod 644 /opt/exim-pilot/data/exim-pilot.db
sudo chmod 755 /opt/exim-pilot/data/
```

**Corrupted Database:**
```bash
# Check database integrity
sqlite3 /opt/exim-pilot/data/exim-pilot.db "PRAGMA integrity_check;"

# If corrupted, restore from backup or reinitialize
sudo systemctl stop exim-pilot
sudo mv /opt/exim-pilot/data/exim-pilot.db /opt/exim-pilot/data/exim-pilot.db.backup
sudo systemctl start exim-pilot
```

**Disk Space Issues:**
```bash
# Check available disk space
df -h /opt/exim-pilot/data/

# Clean up if needed
sudo find /opt/exim-pilot/data/ -name "*.log" -mtime +30 -delete
```

### Configuration File Problems

**Symptoms:**
- "Configuration file not found" errors
- Invalid configuration errors
- Default values not working

**Solutions:**

**Missing Configuration File:**
```bash
# Copy example configuration
sudo cp /opt/exim-pilot/config/config.example.yaml /opt/exim-pilot/config/config.yaml
sudo chown exim-pilot:exim-pilot /opt/exim-pilot/config/config.yaml
```

**Invalid YAML Syntax:**
```bash
# Validate YAML syntax
python3 -c "import yaml; yaml.safe_load(open('/opt/exim-pilot/config/config.yaml'))"

# Or use online YAML validator
```

**Path Configuration Issues:**
```bash
# Verify Exim paths exist
ls -la /var/log/exim4/
ls -la /var/spool/exim4/

# Update config.yaml with correct paths
sudo nano /opt/exim-pilot/config/config.yaml
```

## Authentication Issues

### Cannot Log In

**Symptoms:**
- "Invalid credentials" errors
- Login form doesn't respond
- Session expires immediately

**Diagnostic Steps:**
1. Check authentication configuration in config.yaml
2. Verify user credentials in database or configuration
3. Check session configuration

**Solutions:**

**Default Credentials:**
```bash
# Check if using default credentials
# Default: admin/admin (change immediately!)
```

**Session Cookie Issues:**
```bash
# Clear browser cookies for the site
# Check if HTTPS is required but using HTTP
# Verify session timeout configuration
```

**Database Authentication Issues:**
```bash
# Check user table in database
sqlite3 /opt/exim-pilot/data/exim-pilot.db "SELECT * FROM users;"

# Reset admin password (if supported)
/opt/exim-pilot/bin/exim-pilot --reset-password admin
```

### Session Expires Too Quickly

**Solutions:**
1. Increase session timeout in config.yaml:
   ```yaml
   auth:
     session_timeout: 3600  # 1 hour
   ```

2. Check system clock synchronization:
   ```bash
   timedatectl status
   sudo systemctl restart systemd-timesyncd
   ```

## Queue Management Problems

### Queue Not Loading

**Symptoms:**
- Empty queue display when messages exist
- "Failed to load queue" errors
- Timeout errors

**Diagnostic Steps:**
1. Check if Exim is running:
   ```bash
   systemctl status exim4
   ```

2. Test Exim queue commands manually:
   ```bash
   sudo exim -bp
   sudo exim -bpu
   ```

3. Check permissions:
   ```bash
   ls -la /var/spool/exim4/
   ```

**Solutions:**

**Exim Not Running:**
```bash
sudo systemctl start exim4
sudo systemctl enable exim4
```

**Permission Issues:**
```bash
# Add exim-pilot user to exim group
sudo usermod -a -G Debian-exim exim-pilot

# Or adjust permissions
sudo chmod g+r /var/spool/exim4/input/
```

**Queue Command Failures:**
```bash
# Check Exim configuration
sudo exim -bV
sudo exim -bt test@example.com

# Fix Exim configuration if needed
sudo nano /etc/exim4/update-exim4.conf.conf
sudo update-exim4.conf
sudo systemctl restart exim4
```

### Message Operations Fail

**Symptoms:**
- "Operation failed" errors
- Messages don't get delivered/frozen/deleted
- Permission denied errors

**Solutions:**

**Insufficient Privileges:**
```bash
# Ensure exim-pilot can execute Exim commands
sudo visudo
# Add: exim-pilot ALL=(root) NOPASSWD: /usr/sbin/exim
```

**Message Not Found:**
- Message may have been processed by Exim between listing and operation
- Refresh queue list and try again

**Exim Command Errors:**
```bash
# Test commands manually
sudo exim -M 1hKj4x-0008Oi-3r  # Deliver
sudo exim -Mf 1hKj4x-0008Oi-3r # Freeze
sudo exim -Mt 1hKj4x-0008Oi-3r # Thaw
sudo exim -Mrm 1hKj4x-0008Oi-3r # Remove
```

## Log Processing Issues

### Logs Not Updating

**Symptoms:**
- Log viewer shows old entries
- Real-time tail not working
- Missing log entries

**Diagnostic Steps:**
1. Check log file permissions:
   ```bash
   ls -la /var/log/exim4/
   ```

2. Verify log files are being written:
   ```bash
   tail -f /var/log/exim4/mainlog
   ```

3. Check file watcher status in application logs

**Solutions:**

**Permission Issues:**
```bash
# Add exim-pilot user to adm group (for log access)
sudo usermod -a -G adm exim-pilot

# Or adjust log file permissions
sudo chmod g+r /var/log/exim4/*.log
```

**Log Rotation Issues:**
```bash
# Check logrotate configuration
cat /etc/logrotate.d/exim4

# Manually rotate logs to test
sudo logrotate -f /etc/logrotate.d/exim4
```

**File Watcher Problems:**
```bash
# Check inotify limits
cat /proc/sys/fs/inotify/max_user_watches

# Increase if needed
echo 'fs.inotify.max_user_watches=524288' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

### Log Parsing Errors

**Symptoms:**
- Malformed log entries in database
- Missing information in log viewer
- Parser error messages in logs

**Solutions:**

**Unknown Log Format:**
- Check Exim version and log format
- Update parser configuration for custom log formats
- Report parsing issues with sample log lines

**Character Encoding Issues:**
```bash
# Check log file encoding
file /var/log/exim4/mainlog

# Convert if needed
iconv -f ISO-8859-1 -t UTF-8 /var/log/exim4/mainlog > /tmp/mainlog.utf8
```

## Performance Problems

### Slow Page Loading

**Symptoms:**
- Pages take long time to load
- Timeouts on large result sets
- High CPU/memory usage

**Diagnostic Steps:**
1. Check system resources:
   ```bash
   top
   free -h
   df -h
   ```

2. Check database performance:
   ```bash
   sqlite3 /opt/exim-pilot/data/exim-pilot.db ".timer on" "SELECT COUNT(*) FROM log_entries;"
   ```

3. Monitor application logs for slow queries

**Solutions:**

**Database Optimization:**
```bash
# Analyze and optimize database
sqlite3 /opt/exim-pilot/data/exim-pilot.db "ANALYZE;"
sqlite3 /opt/exim-pilot/data/exim-pilot.db "VACUUM;"

# Check index usage
sqlite3 /opt/exim-pilot/data/exim-pilot.db ".schema"
```

**Memory Issues:**
```bash
# Increase system memory if possible
# Reduce data retention periods in config.yaml
# Use more specific search filters
```

**Large Result Sets:**
- Use pagination with smaller page sizes
- Apply more specific filters
- Consider data archiving for old records

### High CPU Usage

**Causes and Solutions:**

**Log Processing Overload:**
```bash
# Check log processing rate
tail -f /opt/exim-pilot/logs/exim-pilot.log | grep "processed"

# Reduce processing frequency in config.yaml
# Implement log processing throttling
```

**Database Lock Contention:**
```bash
# Check for long-running queries
# Optimize database schema and queries
# Consider connection pooling adjustments
```

## Database Issues

### Database Corruption

**Symptoms:**
- "Database is locked" errors
- Inconsistent data display
- Application crashes

**Diagnostic Steps:**
```bash
# Check database integrity
sqlite3 /opt/exim-pilot/data/exim-pilot.db "PRAGMA integrity_check;"

# Check for locks
lsof /opt/exim-pilot/data/exim-pilot.db
```

**Solutions:**

**Database Recovery:**
```bash
# Stop application
sudo systemctl stop exim-pilot

# Backup current database
cp /opt/exim-pilot/data/exim-pilot.db /opt/exim-pilot/data/exim-pilot.db.backup

# Attempt repair
sqlite3 /opt/exim-pilot/data/exim-pilot.db ".recover" ".quit"

# If repair fails, restore from backup
cp /opt/exim-pilot/backups/latest.db /opt/exim-pilot/data/exim-pilot.db

# Restart application
sudo systemctl start exim-pilot
```

### Database Growth Issues

**Symptoms:**
- Disk space warnings
- Slow database operations
- Large database file size

**Solutions:**

**Data Cleanup:**
```bash
# Check database size
ls -lh /opt/exim-pilot/data/exim-pilot.db

# Clean old data (adjust dates as needed)
sqlite3 /opt/exim-pilot/data/exim-pilot.db "DELETE FROM log_entries WHERE timestamp < datetime('now', '-90 days');"
sqlite3 /opt/exim-pilot/data/exim-pilot.db "VACUUM;"
```

**Retention Policies:**
```yaml
# Configure in config.yaml
retention:
  log_entries_days: 30
  audit_log_days: 365
  queue_snapshots_days: 7
```

## Network and Connectivity

### API Connection Issues

**Symptoms:**
- "Connection refused" errors
- Timeouts on API requests
- Intermittent connectivity

**Diagnostic Steps:**
```bash
# Check if service is listening
sudo netstat -tlnp | grep :8080

# Test local connectivity
curl -v http://localhost:8080/api/v1/dashboard

# Check firewall rules
sudo ufw status
sudo iptables -L
```

**Solutions:**

**Firewall Configuration:**
```bash
# Allow port 8080
sudo ufw allow 8080/tcp

# Or for specific IP ranges
sudo ufw allow from 192.168.1.0/24 to any port 8080
```

**Network Interface Binding:**
```yaml
# In config.yaml, ensure correct binding
server:
  host: "0.0.0.0"  # Listen on all interfaces
  port: 8080
```

### WebSocket Connection Issues

**Symptoms:**
- Real-time updates not working
- WebSocket connection failures
- "Connection closed" errors

**Solutions:**

**Proxy Configuration:**
```nginx
# Nginx proxy configuration
location /api/v1/logs/tail {
    proxy_pass http://localhost:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
}
```

**Browser Issues:**
- Check browser WebSocket support
- Disable browser extensions that might block WebSockets
- Try different browser or incognito mode

## Exim Integration Problems

### Exim Commands Not Working

**Symptoms:**
- Queue operations fail
- "Command not found" errors
- Permission denied errors

**Diagnostic Steps:**
```bash
# Check Exim installation
which exim
exim -bV

# Test commands as exim-pilot user
sudo -u exim-pilot exim -bp
```

**Solutions:**

**Exim Not Installed:**
```bash
sudo apt update
sudo apt install exim4-daemon-light
```

**Path Issues:**
```bash
# Add Exim to PATH in service file
sudo systemctl edit exim-pilot
# Add:
# [Service]
# Environment="PATH=/usr/sbin:/usr/bin:/sbin:/bin"
```

**Permission Configuration:**
```bash
# Configure sudo access for Exim commands
sudo visudo
# Add:
# exim-pilot ALL=(root) NOPASSWD: /usr/sbin/exim
```

### Log File Access Issues

**Symptoms:**
- Cannot read Exim logs
- "Permission denied" errors
- Empty log display

**Solutions:**

**Group Membership:**
```bash
# Add user to appropriate groups
sudo usermod -a -G adm,Debian-exim exim-pilot

# Verify group membership
groups exim-pilot
```

**File Permissions:**
```bash
# Check and fix log permissions
sudo chmod g+r /var/log/exim4/*.log
sudo chgrp adm /var/log/exim4/*.log
```

## Diagnostic Tools

### Health Check Script

Create a comprehensive health check script:

```bash
#!/bin/bash
# /opt/exim-pilot/bin/health-check.sh

echo "=== Exim Control Panel Health Check ==="
echo

# Check service status
echo "1. Service Status:"
systemctl is-active exim-pilot
systemctl is-enabled exim-pilot
echo

# Check process
echo "2. Process Status:"
pgrep -f exim-pilot || echo "Process not running"
echo

# Check ports
echo "3. Port Status:"
netstat -tlnp | grep :8080 || echo "Port 8080 not listening"
echo

# Check database
echo "4. Database Status:"
if [ -f /opt/exim-pilot/data/exim-pilot.db ]; then
    sqlite3 /opt/exim-pilot/data/exim-pilot.db "SELECT COUNT(*) FROM log_entries;" 2>/dev/null || echo "Database error"
else
    echo "Database file not found"
fi
echo

# Check Exim integration
echo "5. Exim Integration:"
sudo -u exim-pilot exim -bp >/dev/null 2>&1 && echo "OK" || echo "Failed"
echo

# Check log files
echo "6. Log File Access:"
sudo -u exim-pilot test -r /var/log/exim4/mainlog && echo "OK" || echo "Failed"
echo

# Check disk space
echo "7. Disk Space:"
df -h /opt/exim-pilot/data/
echo

# Check memory usage
echo "8. Memory Usage:"
free -h
echo

echo "=== Health Check Complete ==="
```

### Log Analysis Commands

```bash
# Check recent errors
journalctl -u exim-pilot --since "1 hour ago" | grep -i error

# Monitor real-time logs
journalctl -u exim-pilot -f

# Check database statistics
sqlite3 /opt/exim-pilot/data/exim-pilot.db "
SELECT 
    'Log Entries' as table_name, COUNT(*) as count 
FROM log_entries 
UNION ALL 
SELECT 
    'Messages' as table_name, COUNT(*) as count 
FROM messages;
"

# Check log processing rate
tail -f /opt/exim-pilot/logs/exim-pilot.log | grep "processed"
```

### Performance Monitoring

```bash
# Monitor system resources
watch -n 5 'ps aux | grep exim-pilot; echo; free -h; echo; df -h'

# Database performance
sqlite3 /opt/exim-pilot/data/exim-pilot.db "
.timer on
EXPLAIN QUERY PLAN SELECT * FROM log_entries WHERE timestamp > datetime('now', '-1 hour');
"

# Network monitoring
netstat -i
ss -tuln | grep :8080
```

## Log Analysis

### Application Log Patterns

**Normal Operation:**
```
INFO: Starting Exim Control Panel
INFO: Database connection established
INFO: Log monitor started
INFO: Web server listening on :8080
INFO: Processed 150 log entries in 2.3s
```

**Warning Signs:**
```
WARN: High log processing backlog: 5000 entries
WARN: Database query took 5.2s
WARN: WebSocket connection dropped
ERROR: Failed to execute Exim command: permission denied
ERROR: Database connection lost
FATAL: Unable to start web server
```

### System Log Analysis

```bash
# Check for memory issues
journalctl -u exim-pilot | grep -i "out of memory\|killed"

# Check for permission issues
journalctl -u exim-pilot | grep -i "permission denied\|access denied"

# Check for network issues
journalctl -u exim-pilot | grep -i "connection refused\|timeout"
```

### Database Query Analysis

```sql
-- Check for slow queries
.timer on
SELECT COUNT(*) FROM log_entries WHERE timestamp > datetime('now', '-1 day');

-- Check index usage
EXPLAIN QUERY PLAN SELECT * FROM log_entries WHERE message_id = '1hKj4x-0008Oi-3r';

-- Check database size
SELECT 
    name,
    COUNT(*) as rows,
    (COUNT(*) * 1000) as estimated_size_bytes
FROM sqlite_master 
WHERE type='table' 
GROUP BY name;
```

## Getting Help

### Before Contacting Support

1. **Gather Information:**
   - Application version: `/opt/exim-pilot/bin/exim-pilot --version`
   - System information: `uname -a`
   - Error messages from logs
   - Steps to reproduce the issue

2. **Run Diagnostics:**
   ```bash
   # Run health check
   /opt/exim-pilot/bin/health-check.sh

   # Collect logs
   journalctl -u exim-pilot --since "1 hour ago" > /tmp/exim-pilot.log
   tail -n 100 /opt/exim-pilot/logs/exim-pilot.log > /tmp/app.log
   ```

3. **Try Basic Troubleshooting:**
   - Restart the service: `sudo systemctl restart exim-pilot`
   - Check configuration: `/opt/exim-pilot/bin/exim-pilot --config-check`
   - Verify permissions and file access

### Support Information

**Log Locations:**
- Application logs: `/opt/exim-pilot/logs/exim-pilot.log`
- System logs: `journalctl -u exim-pilot`
- Exim logs: `/var/log/exim4/`

**Configuration Files:**
- Main config: `/opt/exim-pilot/config/config.yaml`
- Service config: `/etc/systemd/system/exim-pilot.service`

**Important Directories:**
- Installation: `/opt/exim-pilot/`
- Data: `/opt/exim-pilot/data/`
- Backups: `/opt/exim-pilot/backups/`

### Emergency Recovery

**Service Won't Start:**
```bash
# Reset to defaults
sudo systemctl stop exim-pilot
sudo cp /opt/exim-pilot/config/config.example.yaml /opt/exim-pilot/config/config.yaml
sudo systemctl start exim-pilot
```

**Database Corruption:**
```bash
# Restore from backup
sudo systemctl stop exim-pilot
sudo cp /opt/exim-pilot/backups/latest.db /opt/exim-pilot/data/exim-pilot.db
sudo chown exim-pilot:exim-pilot /opt/exim-pilot/data/exim-pilot.db
sudo systemctl start exim-pilot
```

**Complete Reinstall:**
```bash
# Backup data
sudo cp -r /opt/exim-pilot/data /tmp/exim-pilot-backup

# Reinstall application
sudo systemctl stop exim-pilot
sudo rm -rf /opt/exim-pilot
# Reinstall from package/source

# Restore data
sudo cp -r /tmp/exim-pilot-backup/* /opt/exim-pilot/data/
sudo chown -R exim-pilot:exim-pilot /opt/exim-pilot/data/
sudo systemctl start exim-pilot
```

---

This troubleshooting guide covers the most common issues encountered with Exim Control Panel. For issues not covered here, please check the application logs and system logs for specific error messages, and consult your system administrator or the project documentation.