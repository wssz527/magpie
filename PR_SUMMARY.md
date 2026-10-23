# Kimi Code native protocol routing

Upstream issue: https://github.com/yetone/magpie/issues/3426

Implementation branch:
https://github.com/wssz527/magpie/tree/fix/kimi-code-native-protocols

Patch commit:
https://github.com/wssz527/magpie/commit/a005fc1b92026195cf7b0904b4762f74a589f7b1

Branch head with this summary:
https://github.com/wssz527/magpie/commit/38e57c7a8598bc94ebb9bc45bc818e5ecd45843b

One-commit upstream compare:
https://github.com/yetone/magpie/compare/main...wssz527:fix/kimi-code-native-protocols

Magpie-managed Kimi Code now writes a separate provider for each protocol
while retaining the same `magpie/...` model and group IDs:

- Kimi/Claude-only groups: Anthropic Messages on `magpie-anthropic`.
- GPT/Codex-only groups: OpenAI Responses on `magpie-responses`.
- Chat-only and mixed groups: Chat on `providers.magpie`.
- Legacy `kimi-cli`: Chat only.

Magpie still handles routing, accounts, failover and Usage. A normal sync
upgrades a previous Chat-only config by adding the two native providers.

Validation passed:

- `go test ./internal/agent -count=1`
- `go vet ./internal/agent`
- targeted Kimi and provider protocol tests

Upstream `yetone/magpie` has `pull_request_creation_policy` set to
`collaborators_only`, so GitHub refuses to create a cross-fork PR. The
ready-to-review compare and test summary are posted on issue #3426.
