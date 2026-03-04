# CF Targets Plugin — Developer Notes

## Project Structure

- `cf_targets.go` — main plugin logic
- `cf_targets_test.go` — Ginkgo tests with FakeOS abstraction
- `internal/diff/` — vendored diff library (see below)

## Vendored Diff Code (`internal/diff/`)

The unified diff display uses a local copy of Go's internal diff package,
vendored from `golang.org/x/tools/internal/diff`.

This code is under the Go Authors' BSD license (see `internal/diff/LICENSE`).

### Upstream references

- **pkg.go.dev**: https://pkg.go.dev/golang.org/x/tools/internal/diff
- **GitHub (mirror)**: https://github.com/golang/tools/tree/master/internal/diff
- **Canonical source**: https://go.googlesource.com/tools/+/refs/heads/master/internal/diff/

### Current upstream sync point

Recorded in `internal/diff/SYNC_VERSION` (currently **v0.42.0**, synced March 2026).

### Key files

| File | Purpose |
|------|---------|
| `internal/diff/diff.go` | `Edit` type, `Apply`, `SortEdits` |
| `internal/diff/ndiff.go` | `Strings`, `Bytes`, `Lines` (line-level LCS diff) |
| `internal/diff/unified.go` | `Unified`, `ToUnified` (unified diff formatting) |
| `internal/diff/merge.go` | `Merge` (combine two edit lists) |
| `internal/diff/lcs/` | LCS algorithm engine (forward, backward, two-sided) |

The plugin uses `diff.Lines()` → `diff.ToUnified()` to produce line-level
unified diffs of JSON config files.

### How to check for updates

Run the automated check:
```
make check-diff-upstream
```

This clones upstream, reads the sync tag from `internal/diff/SYNC_VERSION`,
and reports whether any tracked files have changed. Exit code 0 means in sync,
1 means changes detected.

**Manual fallback** (if the script is unavailable):

1. Check the latest tag at https://github.com/golang/tools/tags
2. Compare against the sync tag in `internal/diff/SYNC_VERSION`
3. Review the upstream changelog for `internal/diff/`:
   ```
   git log --oneline <sync-tag>..HEAD -- internal/diff/
   ```
   (in a clone of golang.org/x/tools)
4. If there are meaningful changes, manually port them to `internal/diff/`.
   The code is vendored because upstream uses `internal/` which prevents
   direct import.
5. Update `internal/diff/SYNC_VERSION` to the new tag.
6. Run `go test ./internal/diff/...` to verify after any sync.

### Why not use a third-party diff package?

The following alternatives were evaluated:

| Package | Status | Notes |
|---------|--------|-------|
| [hexops/gotextdiff](https://github.com/hexops/gotextdiff) | **Archived** (Feb 2024) | Was the most commonly recommended Go diff library. Itself was a re-export of the same `x/tools/internal/diff` code. README directs contributions upstream. |
| [aymanbagabas/go-udiff](https://github.com/aymanbagabas/go-udiff) | Active | Also based on `x/tools/internal/diff`. Lightweight wrapper, Myers-based. A viable alternative but adds an external dependency for code we already vendor. |
| [sourcegraph/go-diff](https://github.com/sourcegraph/go-diff) | Active | Parses and formats unified diffs but does **not compute** diffs — different use case. |

Since the most recommended package (`gotextdiff`) was archived and itself was
just a re-export of the same upstream code, maintaining a local vendored copy
from `x/tools/internal/diff` directly remains the most appropriate approach.
This avoids a dependency on an archived project and keeps the code aligned
with the canonical upstream source.

## Build and Test

```
make build          # compile
make test           # run tests with -race
make check          # fmt + vet + lint
make gosec          # security scan (G304 is expected, see Makefile GOSEC_EXCLUDE)
make govulncheck    # dependency vulnerability check
make verify         # all of the above
```

## OS Abstraction

The `OS` interface in `cf_targets.go` wraps filesystem operations for
testability. `FakeOS` in the test file provides a test double. When adding
new OS calls, update both the interface and `FakeOS`.
