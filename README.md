# OVH DNS Manager

A simple and lightweight tool to manage OVH DNS zones via YAML configuration files. 
Designed to run as a one-shot operation in Alpine containers with minimal memory footprint.

## Features

- **Export existing DNS zones** from OVH to YAML format
- **One-way synchronization** from YAML configuration to OVH DNS
- **Dry-run mode** to preview changes before applying
- **One-shot execution** - runs, applies changes, and exits

## Installation

### From Source
```bash
go build -o ovh-dns-manager ./main.go
```

### Docker
```bash
docker build -t ovh-dns-manager .
```

## Configuration

### OVH Credentials
Create `ovh-credentials.yaml`:
```yaml
endpoint: ovh-eu  # ovh-eu, ovh-ca, ovh-us
application_key: your_application_key
application_secret: your_application_secret
consumer_key: your_consumer_key
timeout: 30  # seconds
```

To obtain credentials:
1. Go to [OVH API Console](https://eu.api.ovh.com/createToken/)
2. Set rights for `/domain/zone/*` on GET, POST, PUT, DELETE
3. Generate your keys

### DNS Zone Configuration
```yaml
domain: example.com
records:
  - name: ""           # Root domain (@)
    type: A
    target: 1.2.3.4
    ttl: 3600
  - name: www
    type: CNAME
    target: example.com.
    ttl: 3600
  - name: mail
    type: MX
    target: mail.example.com.
    priority: 10
    ttl: 3600
  - name: _dmarc
    type: TXT
    target: "v=DMARC1; p=none;"
    ttl: 3600
```

## Usage

### Export existing DNS zone
```bash
ovh-dns-manager export --domain example.com --output config.yaml
```

### Apply configuration (dry run)
```bash
ovh-dns-manager apply --config config.yaml --dry-run
```

### Apply configuration
```bash
ovh-dns-manager apply --config config.yaml
```

### Using custom credentials file
```bash
ovh-dns-manager apply --config config.yaml --credentials /path/to/creds.yaml
```

## Docker Usage

### Export
```bash
podman run --rm -v $(pwd):/app/data ovh-dns-manager \
    export --domain example.com --output /app/data/config.yaml \
    --credentials /app/data/ovh-credentials.yaml
```

### Apply
```bash
podman run --rm -v $(pwd):/app/data ovh-dns-manager \
    apply --config /app/data/config.yaml \
    --credentials /app/data/ovh-credentials.yaml
```

## Supported DNS Record Types

- **A** - IPv4 address
- **AAAA** - IPv6 address  
- **CNAME** - Canonical name
- **MX** - Mail exchanger (requires priority)
- **TXT** - Text record
- **NS** - Name server
- **SRV** - Service record (requires priority)

## Memory Usage

Typical memory usage:
- **Binary size**: ~15MB (Alpine container)
- **Runtime memory**: <20MB RAM
- **Export operation**: ~5-10MB additional per 1000 records
- **Apply operation**: ~3-5MB additional per 1000 records

## Workflow

1. **One-time setup**: Export existing DNS configuration
   ```bash
   ovh-dns-manager export --domain example.com
   ```

2. **Edit YAML file** with desired DNS records

3. **Preview changes** with dry-run
   ```bash
   ovh-dns-manager apply --config example.com.yaml --dry-run
   ```

4. **Apply changes** 
   ```bash
   ovh-dns-manager apply --config example.com.yaml
   ```

## Error Handling

- Validates YAML syntax and DNS record formats
- Handles OVH API rate limits with retries
- Provides detailed error messages for troubleshooting
- Exits with non-zero code on errors

## Limitations

- **One-way sync only**: Changes are applied from YAML to OVH only
- **No backup**: Always export current state before major changes
- **One domain per file**: Each YAML file manages one DNS zone
- **No record merging**: Duplicate name+type combinations will conflict