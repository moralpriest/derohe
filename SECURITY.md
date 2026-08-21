# Security Policy

This repository is a fork of [DEROFDN/derohe](https://github.com/DEROFDN/derohe)
and is consensus-identical to its `community-dev` branch by design (enforced by
the consensus-guard workflow). Please report vulnerabilities that affect
**upstream DERO** to upstream first (private security advisory on
DEROFDN/derohe), and vulnerabilities specific to **this fork** (walletapi/,
rpc/, cmd/, CI/CD) here.

## Reporting a Vulnerability

**Do not open a public issue.** Report privately through one of:

- **GitHub private security advisory** (preferred):
  https://github.com/moralpriest/derohe/security/advisories/new
- Email the repository maintainers directly.

Please include:

1. Affected component and version (binary version, or commit hash)
2. Steps to reproduce (minimal; include network/config context if relevant)
3. Impact assessment — especially anything that could fork the chain,
   steal funds, or crash nodes (DoS)
4. Suggested fix, if you have one

## Scope

In scope:

- Consensus logic (block/transaction validation) — note that consensus
  fixes must land **upstream first** to avoid a fork of this node
- Wallet, smart-contract execution (DVM), and RPC layers
- P2P networking and node DoS resilience
- Build/release pipeline integrity (reproducible builds, signed artifacts,
  Docker images, CI/CD workflows)

Out of scope:

- Issues already fixed or tracked upstream in DEROFDN/derohe
- Cosmetic / non-security defects (use the issue tracker)
- Third-party dependencies (report to the dependency's own project, and
  reference it here so we can triage our vendored copy)

## Response

We aim to acknowledge reports within **3 business days** and to ship a fix
(and, where applicable, a patched release) as quickly as the issue warrants.
We ask reporters to allow a **90-day** coordinated-disclosure window before
public release of details.

## Security of releases

Every tagged release ships:

- Reproducible builds (`-trimpath -ldflags=-buildid=`)
- SHA-256 checksums (`SHA256SUMS.txt`)
- Sigstore Cosign signatures and SLSA provenance attestations
- An SPDX SBOM
- A container image on GHCR with provenance

Verify artifacts before use:

```bash
# checksums
sha256sum -c SHA256SUMS.txt

# cosign signature
cosign verify-blob <file> --bundle <file>.bundle \
  --certificate-identity=https://github.com/moralpriest/derohe/.github/workflows/release.yml@refs/tags/<tag> \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com

# GitHub attestation
gh attestation verify <file> --owner moralpriest
```

## Thanks

Thank you for helping keep the network safe.
