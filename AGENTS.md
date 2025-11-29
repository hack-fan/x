# LLM Integration Guide for x

Go utilities library: github.com/hack-fan/x (Go 1.22+)

## Packages

- **xecho** - Echo framework utilities (auth, error handler, logger, pagination)
- **xerr** - Custom error handling with HTTP status codes
- **xdb** - Database utilities (GORM/MySQL)
- **xlog** - Logging utilities (Zap wrapper, WeWork notification)
- **rdb** - Redis client utilities
- **xmp** - WeChat Mini Program utilities
- **xobj** - Object storage (Tencent COS)
- **xpay** - Payment integration
- **xtype** - Type utilities (JSON, values)

## Dev Rules

- Use English for all comments and docs
- Follow Go conventions (gofmt, godoc comments)
- Each sub-package has its own go.mod
- Dependency order to avoid cycles: `xerr` → `xlog` → other packages
- Config structs use `default:"value"` tags
- Don't commit unless asked
- Run `make lint` and `make test` before commit

## Git Commit

Commit message format: `<type>: <subject>` (lowercase subject, one line only)
Types: feat/fix/chore/docs/test/refactor/improve

## GitHub Release

1. `git pull` first, if conflicts, try --rebase.
2. commit all changes.
3. use release.sh to publish package releases.
