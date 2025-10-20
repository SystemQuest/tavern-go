# Example: Multiple Global Configuration Files
# Aligned with tavern-py commit 76569fd

This example demonstrates how to use multiple global configuration files with tavern-go.
Files are loaded in order and deep merged, with later files overwriting earlier ones.

## Usage

```bash
# Single global config (backward compatible)
tavern -c config/base.yaml test.tavern.yaml

# Multiple global configs (new feature)
tavern -c config/base.yaml -c config/env.yaml -c config/override.yaml test.tavern.yaml
```

## Configuration Files

### config/base.yaml
Base configuration with common settings:
- Base URL
- Default timeout
- Common headers

### config/staging.yaml
Staging environment overrides:
- Staging-specific base URL
- Staging API key

### config/production.yaml
Production environment overrides:
- Production base URL
- Production API key
- Stricter timeout

## Merge Behavior

Given these configs:

**base.yaml:**
```yaml
variables:
  base_url: http://localhost
  timeout: 30
  api_key: dev_key
```

**staging.yaml:**
```yaml
variables:
  base_url: https://staging.api.com
  api_key: staging_key
```

The final merged config will be:
```yaml
variables:
  base_url: https://staging.api.com  # overridden by staging.yaml
  timeout: 30                         # from base.yaml
  api_key: staging_key                # overridden by staging.yaml
```

## Benefits

1. **Environment Management**: Separate configs for dev/staging/prod
2. **Reusability**: Share common config across environments
3. **Flexibility**: Override specific values without duplicating entire configs
4. **Team Collaboration**: Different team members can maintain different config layers
