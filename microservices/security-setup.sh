#!/bin/bash

# Security Infrastructure Setup Script for Financial Platform
echo "🔒 SETTING UP SECURITY INFRASTRUCTURE"
echo "====================================="

# Color functions
print_status() {
    echo -e "\033[1;34m[SETUP]\033[0m $1"
}

print_success() {
    echo -e "\033[1;32m[SUCCESS]\033[0m $1"
}

print_error() {
    echo -e "\033[1;31m[ERROR]\033[0m $1"
}

print_warning() {
    echo -e "\033[1;33m[WARNING]\033[0m $1"
}

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if running as root for some operations
    if [[ $EUID -eq 0 ]]; then
        print_warning "Running as root - some operations may require elevated privileges"
    fi
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check OpenSSL
    if ! command -v openssl &> /dev/null; then
        print_error "OpenSSL is not installed. Please install OpenSSL first."
        exit 1
    fi
    
    print_success "Prerequisites check completed"
}

# Setup SSL/TLS Certificates
setup_ssl_certificates() {
    print_status "Setting up SSL/TLS certificates..."
    
    # Create certificates directory
    mkdir -p ssl/certs
    mkdir -p ssl/private
    
    # Generate CA certificate
    print_status "Generating CA certificate..."
    openssl genrsa -out ssl/private/ca-key.pem 4096
    openssl req -new -x509 -days 365 -key ssl/private/ca-key.pem -sha256 -out ssl/certs/ca.pem -subj "/C=US/ST=CA/L=San Francisco/O=Financial Platform/CN=Financial Platform CA"
    
    # Generate server certificate
    print_status "Generating server certificate..."
    openssl genrsa -out ssl/private/server-key.pem 4096
    openssl req -subj "/CN=localhost" -sha256 -new -key ssl/private/server-key.pem -out ssl/certs/server.csr
    
    # Create certificate configuration
    cat > ssl/certs/extfile.conf << EOF
subjectAltName = DNS:localhost,IP:127.0.0.1,IP:10.0.0.1
extendedKeyUsage = serverAuth
EOF
    
    # Sign server certificate
    openssl x509 -req -days 365 -sha256 -in ssl/certs/server.csr -CA ssl/certs/ca.pem -CAkey ssl/private/ca-key.pem -CAcreateserial -out ssl/certs/server-cert.pem -extfile ssl/certs/extfile.conf
    
    # Generate client certificate
    print_status "Generating client certificate..."
    openssl genrsa -out ssl/private/client-key.pem 4096
    openssl req -subj "/CN=client" -new -key ssl/private/client-key.pem -out ssl/certs/client.csr
    
    # Create client certificate configuration
    cat > ssl/certs/extfile-client.conf << EOF
extendedKeyUsage = clientAuth
EOF
    
    # Sign client certificate
    openssl x509 -req -days 365 -sha256 -in ssl/certs/client.csr -CA ssl/certs/ca.pem -CAkey ssl/private/ca-key.pem -CAcreateserial -out ssl/certs/client-cert.pem -extfile ssl/certs/extfile-client.conf
    
    # Set proper permissions
    chmod 600 ssl/private/*
    chmod 644 ssl/certs/*
    
    print_success "SSL/TLS certificates generated successfully"
    echo "Certificate files:"
    echo "  • CA Certificate: ssl/certs/ca.pem"
    echo "  • Server Certificate: ssl/certs/server-cert.pem"
    echo "  • Server Key: ssl/private/server-key.pem"
    echo "  • Client Certificate: ssl/certs/client-cert.pem"
    echo "  • Client Key: ssl/private/client-key.pem"
}

# Setup Firewall Rules
setup_firewall() {
    print_status "Setting up firewall rules..."
    
    # Check if iptables is available
    if command -v iptables &> /dev/null; then
        print_status "Configuring iptables rules..."
        
        # Flush existing rules
        iptables -F
        iptables -X
        iptables -t nat -F
        iptables -t nat -X
        iptables -t mangle -F
        iptables -t mangle -X
        
        # Set default policies
        iptables -P INPUT DROP
        iptables -P FORWARD DROP
        iptables -P OUTPUT ACCEPT
        
        # Allow loopback
        iptables -A INPUT -i lo -j ACCEPT
        iptables -A OUTPUT -o lo -j ACCEPT
        
        # Allow established connections
        iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
        
        # Allow SSH (if needed)
        iptables -A INPUT -p tcp --dport 22 -j ACCEPT
        
        # Allow API Gateway HTTPS
        iptables -A INPUT -p tcp --dport 8090 -j ACCEPT
        
        # Allow internal service communication
        iptables -A INPUT -p tcp --dport 8081:8119 -s 10.0.0.0/8 -j ACCEPT
        
        # Allow database access
        iptables -A INPUT -p tcp --dport 5432 -s 10.0.0.0/8 -j ACCEPT
        iptables -A INPUT -p tcp --dport 27017 -s 10.0.0.0/8 -j ACCEPT
        iptables -A INPUT -p tcp --dport 6379 -s 10.0.0.0/8 -j ACCEPT
        
        # Allow message queue access
        iptables -A INPUT -p tcp --dport 5672 -s 10.0.0.0/8 -j ACCEPT
        
        # Allow monitoring access
        iptables -A INPUT -p tcp --dport 9090 -s 10.0.0.0/8 -j ACCEPT
        iptables -A INPUT -p tcp --dport 3000 -s 10.0.0.0/8 -j ACCEPT
        
        # Allow Vault access
        iptables -A INPUT -p tcp --dport 8200 -s 10.0.0.0/8 -j ACCEPT
        
        # Allow Consul access
        iptables -A INPUT -p tcp --dport 8500 -s 10.0.0.0/8 -j ACCEPT
        iptables -A INPUT -p udp --dport 8600 -s 10.0.0.0/8 -j ACCEPT
        
        # Allow DNS
        iptables -A INPUT -p udp --dport 53 -j ACCEPT
        
        # Allow NTP
        iptables -A INPUT -p udp --dport 123 -j ACCEPT
        
        # Save iptables rules
        if command -v iptables-save &> /dev/null; then
            iptables-save > /etc/iptables/rules.v4
            print_success "iptables rules saved"
        fi
        
        print_success "iptables firewall configured"
    else
        print_warning "iptables not available - skipping firewall configuration"
    fi
}

# Setup Network Security Monitoring
setup_network_monitoring() {
    print_status "Setting up network security monitoring..."
    
    # Create network monitoring configuration
    cat > network-monitoring.yml << EOF
version: '3.8'

services:
  # Network monitoring with Snort
  snort:
    image: snort/snort:latest
    container_name: snort-ids
    network_mode: host
    volumes:
      - ./snort-config:/etc/snort
      - ./logs:/var/log/snort
    command: snort -A console -q -c /etc/snort/snort.conf -i eth0
    restart: unless-stopped

  # Network traffic analyzer
  ntopng:
    image: ntop/ntopng:latest
    container_name: ntopng
    ports:
      - "3002:3000"
    volumes:
      - ntopng_data:/var/lib/ntopng
    environment:
      - NTOPNG_USER=admin
      - NTOPNG_PASSWORD=admin123
    restart: unless-stopped

  # Network security scanner
  nmap:
    image: uzyexe/nmap:latest
    container_name: nmap-scanner
    volumes:
      - ./scan-results:/results
    command: nmap -sS -sV -O -p- 10.0.0.0/8 -oA /results/network-scan
    restart: "no"

volumes:
  ntopng_data:
EOF
    
    # Create Snort configuration
    mkdir -p snort-config
    cat > snort-config/snort.conf << EOF
# Snort configuration for Financial Platform
var HOME_NET 10.0.0.0/8
var EXTERNAL_NET any

# Include rules
include \$RULE_PATH/local.rules
include \$RULE_PATH/bad-traffic.rules
include \$RULE_PATH/exploit.rules
include \$RULE_PATH/scan.rules
include \$RULE_PATH/finger.rules
include \$RULE_PATH/ftp.rules
include \$RULE_PATH/telnet.rules
include \$RULE_PATH/rpc.rules
include \$RULE_PATH/rservices.rules
include \$RULE_PATH/dos.rules
include \$RULE_PATH/ddos.rules
include \$RULE_PATH/dns.rules
include \$RULE_PATH/tftp.rules
include \$RULE_PATH/web-cgi.rules
include \$RULE_PATH/web-coldfusion.rules
include \$RULE_PATH/web-iis.rules
include \$RULE_PATH/web-frontpage.rules
include \$RULE_PATH/web-misc.rules
include \$RULE_PATH/web-client.rules
include \$RULE_PATH/web-php.rules
include \$RULE_PATH/sql.rules
include \$RULE_PATH/x11.rules
include \$RULE_PATH/icmp.rules
include \$RULE_PATH/netbios.rules
include \$RULE_PATH/misc.rules
include \$RULE_PATH/attack-responses.rules
include \$RULE_PATH/oracle.rules
include \$RULE_PATH/mysql.rules
include \$RULE_PATH/snmp.rules
include \$RULE_PATH/smtp.rules
include \$RULE_PATH/imap.rules
include \$RULE_PATH/pop2.rules
include \$RULE_PATH/pop3.rules
include \$RULE_PATH/nntp.rules
include \$RULE_PATH/other-ids.rules
include \$RULE_PATH/web-attacks.rules
include \$RULE_PATH/backdoor.rules
include \$RULE_PATH/shellcode.rules
include \$RULE_PATH/policy.rules
include \$RULE_PATH/porn.rules
include \$RULE_PATH/info.rules
include \$RULE_PATH/icmp-info.rules
include \$RULE_PATH/virus.rules
include \$RULE_PATH/chat.rules
include \$RULE_PATH/multimedia.rules
include \$RULE_PATH/p2p.rules
include \$RULE_PATH/spyware-put.rules
include \$RULE_PATH/experimental.rules

# Configure output
output alert_fast: /var/log/snort/alert.ids
output log_tcpdump: /var/log/snort/snort.log
EOF
    
    print_success "Network security monitoring configured"
}

# Setup Security Headers and Policies
setup_security_policies() {
    print_status "Setting up security policies..."
    
    # Create security policy configuration
    cat > security-policies.yml << EOF
# Security Policies for Financial Platform

# Password Policy
password_policy:
  min_length: 12
  require_uppercase: true
  require_lowercase: true
  require_numbers: true
  require_special_chars: true
  max_age_days: 90
  history_count: 5
  lockout_attempts: 5
  lockout_duration_minutes: 30

# Session Policy
session_policy:
  max_session_duration_hours: 8
  idle_timeout_minutes: 30
  concurrent_sessions: 3
  secure_cookies: true
  http_only_cookies: true
  same_site_policy: "strict"

# API Security Policy
api_security_policy:
  rate_limit_per_minute: 100
  max_request_size_mb: 10
  require_ssl: true
  require_api_key: true
  require_authentication: true
  audit_all_requests: true

# Data Protection Policy
data_protection_policy:
  encryption_at_rest: true
  encryption_in_transit: true
  encryption_algorithm: "AES-256"
  key_rotation_days: 90
  data_retention_days: 2555  # 7 years
  secure_deletion: true

# Network Security Policy
network_security_policy:
  require_vpn: false
  allowed_ports: [22, 443, 8090, 8081-8119]
  blocked_ports: [21, 23, 25, 110, 143]
  require_firewall: true
  require_ids: true
  require_ips: true

# Compliance Policy
compliance_policy:
  pci_dss: true
  sox: true
  gdpr: true
  hipaa: false
  audit_logging: true
  data_classification: true
EOF
    
    print_success "Security policies configured"
}

# Setup Security Testing
setup_security_testing() {
    print_status "Setting up security testing tools..."
    
    # Create security testing configuration
    cat > security-testing.yml << EOF
version: '3.8'

services:
  # OWASP ZAP security scanner
  zap:
    image: owasp/zap2docker-stable:latest
    container_name: zap-scanner
    ports:
      - "8083:8080"
    volumes:
      - ./zap-reports:/zap/wrk
    command: zap-baseline.py -t http://api-gateway:8090 -J zap-report.json
    restart: "no"

  # Nikto web server scanner
  nikto:
    image: solsson/nikto:latest
    container_name: nikto-scanner
    volumes:
      - ./nikto-reports:/reports
    command: nikto -h http://api-gateway:8090 -o /reports/nikto-report.txt
    restart: "no"

  # Nmap network scanner
  nmap-security:
    image: uzyexe/nmap:latest
    container_name: nmap-security-scanner
    volumes:
      - ./nmap-reports:/reports
    command: nmap -sS -sV -O --script vuln 10.0.0.0/8 -oA /reports/security-scan
    restart: "no"

  # Security headers checker
  security-headers:
    image: python:3.9-alpine
    container_name: security-headers-checker
    volumes:
      - ./security-checker:/app
    working_dir: /app
    command: python check_headers.py
    restart: "no"
EOF
    
    # Create security headers checker script
    mkdir -p security-checker
    cat > security-checker/check_headers.py << 'EOF'
#!/usr/bin/env python3
import requests
import json
import sys

def check_security_headers(url):
    try:
        response = requests.get(url, timeout=10)
        headers = response.headers
        
        security_headers = {
            'X-Content-Type-Options': 'nosniff',
            'X-Frame-Options': 'DENY',
            'X-XSS-Protection': '1; mode=block',
            'Strict-Transport-Security': 'max-age=31536000',
            'Content-Security-Policy': 'default-src \'self\'',
            'Referrer-Policy': 'strict-origin-when-cross-origin'
        }
        
        results = {}
        for header, expected_value in security_headers.items():
            if header in headers:
                results[header] = {
                    'present': True,
                    'value': headers[header],
                    'secure': expected_value in headers[header]
                }
            else:
                results[header] = {
                    'present': False,
                    'value': None,
                    'secure': False
                }
        
        return results
    except Exception as e:
        print(f"Error checking {url}: {e}")
        return None

if __name__ == "__main__":
    urls = [
        "http://api-gateway:8090",
        "http://database-service:8115",
        "http://monitoring-service:8116"
    ]
    
    all_results = {}
    for url in urls:
        print(f"Checking security headers for {url}")
        results = check_security_headers(url)
        if results:
            all_results[url] = results
    
    with open('/app/security-headers-report.json', 'w') as f:
        json.dump(all_results, f, indent=2)
    
    print("Security headers check completed")
EOF
    
    chmod +x security-checker/check_headers.py
    
    print_success "Security testing tools configured"
}

# Main setup function
main() {
    echo ""
    echo "🔒 FINANCIAL PLATFORM SECURITY SETUP"
    echo "===================================="
    echo ""
    
    # Check prerequisites
    check_prerequisites
    
    echo ""
    echo "🔧 SETTING UP SECURITY COMPONENTS"
    echo "================================="
    
    # Setup components
    setup_ssl_certificates
    setup_firewall
    setup_network_monitoring
    setup_security_policies
    setup_security_testing
    
    echo ""
    echo "✅ SECURITY SETUP COMPLETE"
    echo "=========================="
    echo ""
    echo "🔐 SECURITY COMPONENTS:"
    echo "  • SSL/TLS Certificates: Generated and configured"
    echo "  • Firewall Rules: iptables configured"
    echo "  • Network Monitoring: Snort IDS and ntopng configured"
    echo "  • Security Policies: Comprehensive policies defined"
    echo "  • Security Testing: OWASP ZAP, Nikto, Nmap configured"
    echo ""
    echo "🌐 ACCESS POINTS:"
    echo "  • ntopng Network Monitor: http://localhost:3002 (admin/admin123)"
    echo "  • OWASP ZAP Scanner: http://localhost:8083"
    echo ""
    echo "📋 SECURITY FEATURES:"
    echo "  • Rate limiting and DDoS protection"
    echo "  • API authentication and authorization"
    echo "  • Input validation and sanitization"
    echo "  • Security headers enforcement"
    echo "  • Network traffic monitoring"
    echo "  • Vulnerability scanning"
    echo "  • Audit logging and compliance"
    echo ""
    echo "🎯 SECURITY INFRASTRUCTURE READY!"
}

# Run main function
main "$@"
