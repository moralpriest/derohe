# Go 1.26 Consensus-Critical Fix Patches

This directory contains **7 reference patches** that must be applied before
the derohe mainnet consensus node can be safely upgraded to Go 1.26.0.

These patches are **documentation only** — they are NOT auto-applied.
A follow-up PR must land the actual code changes.

## Why These Patches Are Required

The Go 1.17 → 1.26 upgrade changes the map implementation (Swiss Tables,
active since Go 1.24). Map iteration order in Go 1.26 differs from Go 1.17.
Where this order is then committed to disk or hashed, the resulting
state root diverges between nodes running different Go versions — a
direct hard-fork risk.

Additionally, the `gob` encoder used in chunk sync is Go-version-sensitive
and may produce incompatible byte streams across runtime versions.

## Patch Index

| # | File | Purpose | Audit Site |
|---|------|---------|-----------|
| 01 | `01-blockchain-blockchain-sc_change_cache-sort.diff` | Sort SC change cache by SCID | §5 site #1 |
| 02 | `02-dvm-sc-RawKeys-sort.diff` | Sort SC store keys | §5 site #2 |
| 03 | `03-dvm-simulator-Entries-sort.diff` | Sort all `Entries` iterations (3 sites) | §5 sites #3, #4, #5 |
| 04 | `04-dvm-simulator-total_per_asset-sort.diff` | Sort asset totals | §5 site #6 |
| 05 | `05-dvm-simulator-incoming_values-sort.diff` | Sort incoming values (2 sites) | §5 site #7 |
| 06 | `06-hardcoded_contracts-Entries-sort.diff` | Sort genesis/HF Entries | §5 site #8 |
| 07 | `07-p2p-chunk_server-replace-gob.diff` | Replace gob with custom encoding | §7 wire risk |

## How to Apply

Each patch is a `.diff` file with explanatory context. To apply:

```bash
# From the derohe repo root
cd /path/to/derohe

# Apply one patch
git apply docs/go-1.26-fix-patches/01-blockchain-blockchain-sc_change_cache-sort.diff

# Apply all in order
for p in docs/go-1.26-fix-patches/*.diff; do
    echo "Applying $p"
    git apply "$p" || { echo "FAILED: $p"; exit 1; }
done

# Verify
go mod tidy
go mod vendor
go build -mod=vendor ./...
go test -mod=vendor ./walletapi/xswd/...
```

## Verification (After Applying All Patches)

1. **Build:** `go build -mod=vendor ./...` must succeed
2. **Existing tests:** `go test -mod=vendor ./...` must not regress
3. **New tests (recommended):** add per-patch determinism regression tests
   (each patch file documents its specific test suggestion)
4. **Differential testnet:** run a shadow sync with one Go 1.17 node and
   one Go 1.26 node, verify state root matches at every block height
   (see `docs/go-1.26-differential-test-spec.md`)

## Risk If Skipped

| Site | Risk if not patched |
|------|---------------------|
| 1 (sc_change_cache) | Hard fork at first block with SC activity |
| 2 (RawKeys) | Hard fork at first SC write |
| 3-4 (Entries) | Hard fork at first SC finalization |
| 5 (ErrorDeposit Entries) | Hard fork on first SC error path |
| 6 (total_per_asset) | Hard fork at first multi-asset SC transfer |
| 7 (incoming_values) | Hard fork on first error-rollback |
| 8 (hardcoded_contracts) | Hard fork at block 1 (genesis SC) |
| gob (chunk_server) | Network segmentation, chunk sync failure |

**All 8 sites are consensus-critical. None may be skipped.**

## Related Documents

- `docs/go-1.26-upgrade-audit.md` — full audit, 17 sections
- `docs/go-1.26-differential-test-spec.md` — shadow testnet spec
- PR #19 — currently a draft, blocked on these patches
