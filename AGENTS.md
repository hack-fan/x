# x

Go utility library monorepo: `github.com/hack-fan/x` (Go 1.27).
Each top-level dir is an independent module with its own `go.mod` (the root `go.mod` is an empty placeholder); there is no `go.work`.

## Packages

- **xerr** - Errors carrying HTTP status codes
- **xlog** - Zap logger wrapper, WeWork notification
- **xecho** - Echo v5 helpers (auth, error handler, logger, pagination)
- **xdb** - GORM/MySQL helpers
- **rdb** - Redis client helpers
- **xmp** - WeChat Mini Program helpers
- **xobj** - Object storage (Tencent COS)
- **xpay** - Payment integration
- **xtype** - JSON and value type helpers

## Commands

```bash
make lint     # golangci-lint run in every module (lint.sh)
make test     # go test ./... in every module (test.sh)
make up       # go get -u + go mod tidy in every module (up.sh)
./release.sh  # tag and push a patch release for each changed module
```

## Gotchas

- Module list is hard-coded in `lint.sh`, `test.sh`, `up.sh`; update all three when adding a module (`release.sh` auto-discovers).
- xecho and xobj import xerr by published tag (see their `go.mod`), not the local copy. An xerr change reaches them only after releasing xerr and running `go get github.com/hack-fan/x/xerr@<tag>` in the dependent module.
- Layering (bottom → top): `xerr` → `xlog` → other packages. A package may import only layers below it; `xerr` and `xlog` must not import other x packages (avoids cycles).

## Dev Rules

- English for all comments and docs
- Follow Go conventions (gofmt, godoc comments)
- Config structs use `default:"value"` tags
- Don't commit unless asked
- Run `make lint` and `make test` before commit

## Git Commit

Format: `<type>: <subject>` (lowercase subject, one line only)
Types: feat/fix/chore/docs/test/refactor/improve

## GitHub Release

1. `git pull` (use `--rebase` on conflicts)
2. Commit all changes and `git push` — `release.sh` refuses dirty trees, untracked files, or unpushed commits
3. Run `./release.sh`: tags `<module>/vX.Y.Z` (patch bump) for each module changed since its last tag
