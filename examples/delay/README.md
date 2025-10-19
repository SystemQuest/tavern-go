# Delay Example

This example demonstrates the use of `delay_before` and `delay_after` in test stages.

## Features

- **delay_before**: Pauses execution before sending the request
- **delay_after**: Pauses execution after validating the response

## Use Cases

### 1. Wait for Async Operations
```yaml
- name: Trigger async job
  request:
    url: "{base_url}/api/jobs"
    method: POST
  response:
    status_code: 202
  delay_after: 2.0  # Wait for job to complete

- name: Check job status
  request:
    url: "{base_url}/api/jobs/{job_id}"
    method: GET
  response:
    status_code: 200
    json:
      status: "completed"
```

### 2. Rate Limiting
```yaml
- name: API call 1
  request:
    url: "{base_url}/api/data"
  response:
    status_code: 200
  delay_after: 1.0  # Respect rate limit

- name: API call 2
  request:
    url: "{base_url}/api/data"
  response:
    status_code: 200
```

### 3. System Stabilization
```yaml
- name: Deploy configuration
  delay_before: 2.0  # Wait for system to be ready
  request:
    url: "{base_url}/api/config"
    method: PUT
  response:
    status_code: 200
  delay_after: 3.0  # Wait for config to propagate
```

## Running the Example

```bash
# From the project root
./bin/tavern examples/delay/delay.tavern.yaml

# With debug logging to see delay messages
./bin/tavern --log-level debug examples/delay/delay.tavern.yaml
```

## Notes

- Delays are specified in seconds (supports decimals, e.g., 0.5 for 500ms)
- Delays are optional - omit them if not needed
- `delay_before` executes before the HTTP request
- `delay_after` executes after response validation
- Use debug logging (`--log-level debug`) to see delay messages
