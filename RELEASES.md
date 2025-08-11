# Release Management

## How to Create a Release

1. **Create and push a semantic version tag:**
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```

2. **GitHub Actions automatically:**
   - Builds cross-platform binaries (Linux, macOS, Windows - AMD64/ARM64)
   - Creates GitHub release with changelog and SHA256 checksums
   - Builds and pushes multi-arch container images to GitHub Container Registry
   - Runs security scans on container images

## Release Artifacts

### Binaries
- **Platforms**: Linux, macOS, Windows (AMD64/ARM64)
- **Location**: GitHub Releases page
- **Verification**: SHA256 checksums included

### Container Images
- **Registry**: `ghcr.io/epheo/ovh-dns-manager`
- **Tags**: 
  - `v1.2.3` (version-specific)
  - `1.2` (minor version)
  - `latest` (latest release)
- **Architectures**: AMD64, ARM64
- **Security**: Automated Trivy scanning

## Version Format

Use **semantic versioning** with `v` prefix:
- `v1.0.0` - Major release
- `v1.1.0` - Minor release  
- `v1.1.1` - Patch release

## Container Usage

```bash
# Version-specific
podman pull ghcr.io/epheo/ovh-dns-manager:1.2.3

# Latest
podman pull ghcr.io/epheo/ovh-dns-manager:latest
```