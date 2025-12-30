# HTTPS/TLS Configuration Guide

## Overview

This guide helps you set up HTTPS/TLS encryption for the Employee Management System.

## Quick Start (Development)

### Generate Self-Signed Certificate (for testing only)

```bash
# Using OpenSSL
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes \
  -subj "/C=IN/ST=State/L=City/O=Company/CN=localhost"

# This creates:
# - server.crt (certificate file)
# - server.key (private key file)
```

### Enable in .env

```dotenv
TLS_ENABLED=true
TLS_CERT_FILE=./server.crt
TLS_KEY_FILE=./server.key
TLS_MIN_VERSION=TLS12
```

### Start Server

```bash
ENVIRONMENT=development TLS_ENABLED=true go run main.go
```

### Test HTTPS

```bash
# With curl (skip certificate verification for self-signed)
curl -k https://localhost:8080/health

# Using insecure flag
curl --insecure https://localhost:8080/health
```

---

## Production Setup

### 1. Get a Real Certificate

#### Option A: Let's Encrypt (Recommended)

```bash
# Install certbot
# On Ubuntu/Debian: sudo apt-get install certbot

# Generate certificate for domain
sudo certbot certonly --standalone -d yourdomain.com

# Certificates created in:
# /etc/letsencrypt/live/yourdomain.com/fullchain.pem
# /etc/letsencrypt/live/yourdomain.com/privkey.pem
```

#### Option B: Self-signed (Organization)

```bash
# Generate private key (4096-bit RSA)
openssl genrsa -out server.key 4096

# Create certificate signing request
openssl req -new -key server.key -out server.csr \
  -subj "/C=IN/ST=Karnataka/L=Bangalore/O=YourCompany/CN=yourdomain.com"

# Self-sign the certificate (valid for 365 days)
openssl x509 -req -days 365 -in server.csr \
  -signkey server.key -out server.crt
```

### 2. Docker/Container Setup

```dockerfile
# Copy certificates
COPY server.crt /app/certs/
COPY server.key /app/certs/

# Set environment
ENV TLS_ENABLED=true
ENV TLS_CERT_FILE=/app/certs/server.crt
ENV TLS_KEY_FILE=/app/certs/server.key
ENV TLS_MIN_VERSION=TLS13
```

### 3. Kubernetes Setup

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: tls-certs
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-cert>
  tls.key: <base64-encoded-key>

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  TLS_ENABLED: "true"
  TLS_CERT_FILE: "/etc/tls/certs/tls.crt"
  TLS_KEY_FILE: "/etc/tls/certs/tls.key"
  TLS_MIN_VERSION: "TLS13"

---
apiVersion: v1
kind: Pod
spec:
  containers:
    - name: app
      volumeMounts:
        - name: tls
          mountPath: /etc/tls/certs
  volumes:
    - name: tls
      secret:
        secretName: tls-certs
```

### 4. Environment-Specific Config

#### Development

```bash
# Use self-signed
TLS_ENABLED=false  # For development without HTTPS
```

#### Staging

```bash
TLS_ENABLED=true
TLS_CERT_FILE=/etc/certs/staging.crt
TLS_KEY_FILE=/etc/certs/staging.key
TLS_MIN_VERSION=TLS12
```

#### Production

```bash
TLS_ENABLED=true
TLS_CERT_FILE=/etc/letsencrypt/live/yourdomain.com/fullchain.pem
TLS_KEY_FILE=/etc/letsencrypt/live/yourdomain.com/privkey.pem
TLS_MIN_VERSION=TLS13  # Require modern TLS
```

---

## Certificate Renewal (Let's Encrypt)

```bash
# Manual renewal
sudo certbot renew --quiet

# Automatic renewal (crontab)
# Add to crontab: 0 2 * * * certbot renew --quiet && systemctl restart your-service
```

---

## Verification & Testing

### 1. Check Certificate Validity

```bash
# View certificate details
openssl x509 -in server.crt -text -noout

# Check certificate expiration
openssl x509 -in server.crt -noout -dates

# Verify certificate chain
openssl verify -CAfile server.crt server.crt
```

### 2. Test HTTPS Connection

```bash
# Check TLS version
openssl s_client -connect localhost:8080 -tls1_2

# Check supported ciphers
openssl s_client -connect localhost:8080 -tls1_3

# Get certificate from server
openssl s_client -connect localhost:8080 < /dev/null | openssl x509 -text
```

### 3. Curl Tests

```bash
# Test with specific TLS version
curl --tlsv1.2 https://localhost:8080/health

# Test with certificate pinning
curl --cacert server.crt https://localhost:8080/health

# Verbose output
curl -v --cacert server.crt https://localhost:8080/health
```

---

## Troubleshooting

### Issue: "Certificate doesn't match"

- Check that CN (Common Name) matches the domain
- For localhost: CN=localhost

### Issue: "Bad certificate"

- Ensure TLS_CERT_FILE and TLS_KEY_FILE paths are correct
- Check file permissions (should be readable by server process)

### Issue: "TLS handshake failure"

- Verify both certificate and key files are present
- Check TLS_MIN_VERSION compatibility with client
- Ensure certificate is not expired

### Solution: Debug TLS

```bash
# Enable verbose logging
LOG_LEVEL=debug TLS_ENABLED=true go run main.go

# Check logs for TLS errors
tail -f app.log | grep -i tls
```

---

## Security Best Practices

✅ **Do:**

- Use TLS 1.3 in production
- Renew certificates 30 days before expiry
- Use strong key sizes (2048-bit minimum, 4096-bit recommended)
- Restrict certificate file permissions (mode 600 for keys)
- Monitor certificate expiration dates

❌ **Don't:**

- Use HTTP in production
- Share private key files
- Use self-signed certs in production for public APIs
- Disable certificate validation in clients
- Ignore TLS warnings

---

## Configuration Options

```go
type TLSConfig struct {
    Enabled    bool   // Enable/disable TLS
    CertFile   string // Path to certificate
    KeyFile    string // Path to private key
    MinVersion string // TLS12 or TLS13
}
```

### Supported Cipher Suites

**TLS 1.3:**

- TLS_AES_256_GCM_SHA384
- TLS_CHACHA20_POLY1305_SHA256
- TLS_AES_128_GCM_SHA256

**TLS 1.2:**

- TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
- TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305
- TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256

---

## Rate Limiting & HTTPS

Rate limiting works with both HTTP and HTTPS:

```bash
# HTTPS (if enabled)
for i in {1..101}; do curl -k https://localhost:8080/health; done

# Response after 100 requests/minute
# HTTP 429 Too Many Requests
```

---

## Monitoring

### Check if HTTPS is Working

```bash
# Health check over HTTPS
curl -k https://localhost:8080/health

# Should return:
{
  "status": "healthy",
  "checks": {
    "database": "ok",
    "email_queue": "ok"
  }
}
```

### Kubernetes Health Probes

```yaml
livenessProbe:
  httpGet:
    scheme: HTTPS
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    scheme: HTTPS
    path: /readiness
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
```

---

## Summary

| Environment | TLS      | Version     | Cert Source       |
| ----------- | -------- | ----------- | ----------------- |
| Development | Optional | TLS 1.2     | Self-signed       |
| Staging     | Yes      | TLS 1.2     | Self-signed or LE |
| Production  | **Yes**  | **TLS 1.3** | Let's Encrypt     |

---

For more information, see:

- [Mozilla TLS Configuration](https://wiki.mozilla.org/Security/Server_Side_TLS)
- [Let's Encrypt Documentation](https://letsencrypt.org/docs/)
- [Go TLS Documentation](https://golang.org/pkg/crypto/tls/)
