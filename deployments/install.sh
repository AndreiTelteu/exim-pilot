#!/bin/bash

# Exim Control Panel Installation Script
# This script installs Exim-Pilot on Ubuntu/Debian systems

set -e

# Configuration
INSTALL_DIR="/opt/exim-pilot"
SERVICE_USER="exim-pilot"
SERVICE_GROUP="exim-pilot"
BINARY_NAME="exim-pilot"
SERVICE_NAME="exim-pilot"
CONFIG_FILE="$INSTALL_DIR/config/config.yaml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        exit 1
    fi
}

# Check system requirements
check_requirements() {
    log_info "Checking system requirements..."
    
    # Check OS
    if ! grep -q "Ubuntu\|Debian" /etc/os-release; then
        log_error "This installer is designed for Ubuntu/Debian systems"
        exit 1
    fi
    
    # Check if Exim is installed
    if ! command -v exim4 &> /dev/null; then
        log_warning "Exim4 not found. Please install Exim4 first:"
        echo "  sudo apt update"
        echo "  sudo apt install exim4-daemon-heavy"
        exit 1
    fi
    
    # Check if systemd is available
    if ! command -v systemctl &> /dev/null; then
        log_error "systemd is required but not found"
        exit 1
    fi
    
    log_success "System requirements check passed"
}

# Create system user and group
create_user() {
    log_info "Creating system user and group..."
    
    if ! getent group "$SERVICE_GROUP" > /dev/null 2>&1; then
        groupadd --system "$SERVICE_GROUP"
        log_success "Created group: $SERVICE_GROUP"
    else
        log_info "Group $SERVICE_GROUP already exists"
    fi
    
    if ! getent passwd "$SERVICE_USER" > /dev/null 2>&1; then
        useradd --system --gid "$SERVICE_GROUP" --home-dir "$INSTALL_DIR" \
                --shell /bin/false --comment "Exim Control Panel" "$SERVICE_USER"
        log_success "Created user: $SERVICE_USER"
    else
        log_info "User $SERVICE_USER already exists"
    fi
}

# Create directory structure
create_directories() {
    log_info "Creating directory structure..."
    
    # Create main directories
    mkdir -p "$INSTALL_DIR"/{bin,config,data,logs,backups}
    
    # Set ownership and permissions
    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR"
    chmod 755 "$INSTALL_DIR"
    chmod 755 "$INSTALL_DIR"/{bin,config}
    chmod 750 "$INSTALL_DIR"/{data,logs,backups}
    
    log_success "Directory structure created"
}

# Install binary
install_binary() {
    log_info "Installing binary..."
    
    if [[ ! -f "$BINARY_NAME" ]]; then
        log_error "Binary file '$BINARY_NAME' not found in current directory"
        log_info "Please ensure you have built the binary first:"
        echo "  make build"
        exit 1
    fi
    
    # Copy binary
    cp "$BINARY_NAME" "$INSTALL_DIR/bin/"
    chown "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/bin/$BINARY_NAME"
    chmod 755 "$INSTALL_DIR/bin/$BINARY_NAME"
    
    log_success "Binary installed to $INSTALL_DIR/bin/$BINARY_NAME"
}

# Create default configuration
create_config() {
    log_info "Creating default configuration..."
    
    if [[ -f "$CONFIG_FILE" ]]; then
        log_warning "Configuration file already exists: $CONFIG_FILE"
        log_info "Backing up existing configuration..."
        cp "$CONFIG_FILE" "$CONFIG_FILE.backup.$(date +%Y%m%d_%H%M%S)"
    fi
    
    # Generate random session secret
    SESSION_SECRET=$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | xxd -p -c 32)
    
    # Create configuration file
    cat > "$CONFIG_FILE" << EOF
# Exim Control Panel Configuration
# Generated on $(date)

server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 15
  write_timeout: 15
  idle_timeout: 60
  allowed_origins: ["*"]
  log_requests: true
  tls_enabled: false
  # tls_cert_file: "/etc/ssl/certs/exim-pilot.crt"
  # tls_key_file: "/etc/ssl/private/exim-pilot.key"

database:
  path: "$INSTALL_DIR/data/exim-pilot.db"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5  # minutes
  backup_enabled: true
  backup_interval: 24   # hours
  backup_path: "$INSTALL_DIR/backups"

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
  level: "info"
  file: "$INSTALL_DIR/logs/exim-pilot.log"
  max_size: 100      # MB
  max_backups: 5
  max_age: 30        # days
  compress: true

retention:
  log_entries_days: 90
  audit_log_days: 365
  queue_snapshots_days: 30
  delivery_attempt_days: 180
  cleanup_interval: 24  # hours

security:
  session_timeout: 60      # minutes
  max_login_attempts: 5
  login_lockout_time: 15   # minutes
  csrf_protection: true
  secure_cookies: true
  content_redaction: true
  audit_all_actions: true
  trusted_proxies: []

auth:
  default_username: "admin"
  default_password: "admin123"  # CHANGE THIS AFTER INSTALLATION
  password_min_length: 8
  require_strong_password: true
  session_secret: "$SESSION_SECRET"
EOF
    
    # Set ownership and permissions
    chown "$SERVICE_USER:$SERVICE_GROUP" "$CONFIG_FILE"
    chmod 640 "$CONFIG_FILE"
    
    log_success "Configuration file created: $CONFIG_FILE"
    log_warning "IMPORTANT: Change the default admin password after installation!"
}

# Install systemd service
install_service() {
    log_info "Installing systemd service..."
    
    if [[ ! -f "deployments/systemd/exim-pilot.service" ]]; then
        log_error "Service file not found: deployments/systemd/exim-pilot.service"
        exit 1
    fi
    
    # Copy service file
    cp "deployments/systemd/exim-pilot.service" "/etc/systemd/system/"
    
    # Reload systemd
    systemctl daemon-reload
    
    log_success "Systemd service installed"
}

# Set up log rotation
setup_logrotate() {
    log_info "Setting up log rotation..."
    
    cat > "/etc/logrotate.d/exim-pilot" << EOF
$INSTALL_DIR/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 640 $SERVICE_USER $SERVICE_GROUP
    postrotate
        systemctl reload $SERVICE_NAME > /dev/null 2>&1 || true
    endscript
}
EOF
    
    log_success "Log rotation configured"
}

# Configure file permissions for Exim log access
configure_permissions() {
    log_info "Configuring file permissions..."
    
    # Add exim-pilot user to adm group for log access
    usermod -a -G adm "$SERVICE_USER"
    
    # Ensure log directories are readable
    if [[ -d "/var/log/exim4" ]]; then
        chmod 755 /var/log/exim4
        log_success "Exim log directory permissions configured"
    else
        log_warning "Exim log directory not found: /var/log/exim4"
    fi
    
    # Ensure spool directory is readable
    if [[ -d "/var/spool/exim4" ]]; then
        chmod 755 /var/spool/exim4
        log_success "Exim spool directory permissions configured"
    else
        log_warning "Exim spool directory not found: /var/spool/exim4"
    fi
}

# Initialize database
initialize_database() {
    log_info "Initializing database..."
    
    # Run database migrations as the service user
    sudo -u "$SERVICE_USER" "$INSTALL_DIR/bin/$BINARY_NAME" --migrate-up --config "$CONFIG_FILE" || {
        log_error "Database initialization failed"
        exit 1
    }
    
    log_success "Database initialized"
}

# Start and enable service
start_service() {
    log_info "Starting and enabling service..."
    
    # Enable service
    systemctl enable "$SERVICE_NAME"
    
    # Start service
    systemctl start "$SERVICE_NAME"
    
    # Check status
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        log_success "Service started successfully"
    else
        log_error "Service failed to start"
        log_info "Check service status with: systemctl status $SERVICE_NAME"
        log_info "Check logs with: journalctl -u $SERVICE_NAME -f"
        exit 1
    fi
}

# Display post-installation information
show_info() {
    log_success "Installation completed successfully!"
    echo
    echo "=== Exim Control Panel Installation Summary ==="
    echo "Installation directory: $INSTALL_DIR"
    echo "Configuration file: $CONFIG_FILE"
    echo "Service name: $SERVICE_NAME"
    echo "Service user: $SERVICE_USER"
    echo
    echo "=== Next Steps ==="
    echo "1. Change the default admin password:"
    echo "   - Open http://localhost:8080 in your browser"
    echo "   - Login with username: admin, password: admin123"
    echo "   - Change the password in the user settings"
    echo
    echo "2. Configure TLS (recommended for production):"
    echo "   - Obtain SSL certificates"
    echo "   - Update $CONFIG_FILE with certificate paths"
    echo "   - Set tls_enabled: true"
    echo "   - Restart service: systemctl restart $SERVICE_NAME"
    echo
    echo "3. Adjust firewall settings if needed:"
    echo "   - Allow port 8080: ufw allow 8080"
    echo
    echo "=== Service Management ==="
    echo "Start service:   systemctl start $SERVICE_NAME"
    echo "Stop service:    systemctl stop $SERVICE_NAME"
    echo "Restart service: systemctl restart $SERVICE_NAME"
    echo "Service status:  systemctl status $SERVICE_NAME"
    echo "View logs:       journalctl -u $SERVICE_NAME -f"
    echo
    echo "=== Configuration ==="
    echo "Edit config:     $CONFIG_FILE"
    echo "After editing:   systemctl restart $SERVICE_NAME"
    echo
    log_warning "SECURITY: Remember to change the default admin password!"
}

# Uninstall function
uninstall() {
    log_info "Uninstalling Exim Control Panel..."
    
    # Stop and disable service
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        systemctl stop "$SERVICE_NAME"
    fi
    
    if systemctl is-enabled --quiet "$SERVICE_NAME"; then
        systemctl disable "$SERVICE_NAME"
    fi
    
    # Remove service file
    rm -f "/etc/systemd/system/$SERVICE_NAME.service"
    systemctl daemon-reload
    
    # Remove logrotate configuration
    rm -f "/etc/logrotate.d/exim-pilot"
    
    # Ask about data removal
    echo
    read -p "Remove all data and configuration files? [y/N]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$INSTALL_DIR"
        log_success "All files removed"
    else
        log_info "Data and configuration files preserved in $INSTALL_DIR"
    fi
    
    # Ask about user removal
    read -p "Remove system user $SERVICE_USER? [y/N]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        userdel "$SERVICE_USER" 2>/dev/null || true
        groupdel "$SERVICE_GROUP" 2>/dev/null || true
        log_success "System user removed"
    fi
    
    log_success "Uninstallation completed"
}

# Main installation function
main() {
    echo "=== Exim Control Panel Installer ==="
    echo
    
    # Parse command line arguments
    case "${1:-}" in
        --uninstall)
            check_root
            uninstall
            exit 0
            ;;
        --help|-h)
            echo "Usage: $0 [--uninstall] [--help]"
            echo
            echo "Options:"
            echo "  --uninstall  Remove Exim Control Panel"
            echo "  --help       Show this help message"
            exit 0
            ;;
    esac
    
    # Run installation steps
    check_root
    check_requirements
    create_user
    create_directories
    install_binary
    create_config
    install_service
    setup_logrotate
    configure_permissions
    initialize_database
    start_service
    show_info
}

# Run main function
main "$@"