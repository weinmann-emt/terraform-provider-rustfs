# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | ✅ |
| < latest| ❌ |

Only the latest released version receives security updates.

## Reporting a Vulnerability

If you discover a security vulnerability, please **DO NOT** open a public issue.

**Preferred Method:** Use GitHub's [Private Vulnerability Reporting](https://github.com/weinmann-emt/terraform-provider-rustfs/security/advisories/new)

### What to Include

- **Vulnerability Type**: What type of vulnerability is it
- **Affected Versions**: Which provider versions are affected
- **Impact**: What is the impact (data exposure, privilege escalation, etc.)
- **Reproduction Steps**: Detailed steps to reproduce
- **Proof of Concept**: If possible, include a minimal PoC

### Response Timeline

- **Initial Response**: Within 48 hours
- **Detailed Assessment**: Within 7 days
- **Public Disclosure**: After a fix is released, typically within 14 days

## Security Best Practices for Users

1. **Credential Management**
   - Use environment variables (`RUSTFS_USER`, `RUSTFS_SECRET`) instead of hardcoded credentials
   - Rotate credentials regularly

2. **Network Security**
   - Use TLS (`ssl = true`) for production deployments
   - Consider private networks for sensitive data

3. **State Protection**
   - Encrypt Terraform state files
   - Use remote state backends with proper access controls

4. **Monitoring**
   - Enable RustFS audit logging
   - Review provider logs for accidental credential exposure
