# CI Configuration

tavern-go uses GitHub Actions for continuous integration.

## What it does

On every push to `main` or pull request:

1. **Test** - Run all tests
2. **Lint** - Check code quality
3. **Build** - Verify it compiles

## Configuration

See `.github/workflows/ci.yml`

Simple and straightforward - no complex matrices or deployments.

## Local Testing

Before pushing, run:

```bash
make test
make lint
make build
```

These are the same commands CI runs.
