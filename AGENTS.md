# Repository Guidelines

## Project Structure & Module Organization
Packages are scoped by capability at the repo root (for example `policy`, `ldap`, `randreader`, `sync`). Shared primitives live in `env`, `net`, `safe`, and `xtime`, while integration helpers such as certificates sit under `certs/`. Tests follow Go’s convention of co-locating `_test.go` files beside the code they exercise. No secondary Go modules exist; the single `go.mod` at the root governs builds.

## Build, Test, and Development Commands
- `make lint` – installs tooling via `make getdeps` when missing, then runs `golangci-lint` with the repo’s strict config.
- `make test` – executes `make lint` and `go test -race -tags kqueue ./...` to cover all packages.
- `make test-ldap` – focuses on the `ldap` package against the endpoint defined in `LDAP_TEST_SERVER`.
- `make clean` – removes cache binaries and editor swap artifacts.
Run these commands before opening a pull request; they mirror the CI stack.

## Coding Style & Naming Conventions
Always format Go sources with `gofmt`/`goimports`. Keep comments minimal, explaining **why** the code exists, never **what**, and do not leave “removed because” notes when deleting code. Use descriptive package names that mirror directory names and exported identifiers with GoDoc-ready sentences. Stick to tab-indented Go style and avoid introducing logging or HTTP helpers that bypass established patterns in the maintained SILO repositories without prior discussion.

## Testing Guidelines
Write focused unit tests in `_test.go` files and organize table-driven tests for edge cases (credentials, policy evaluation, network fallbacks). Include subtests for protocol-specific behavior. Always run `make test`; add targeted benchmarks (`go test -bench`) when optimizing hot paths and share before/after numbers in review.

## Commit & Pull Request Guidelines
Structure commits around single concerns with imperative subjects (e.g., `Add`, `Fix`, `Remove`). Reference GitHub issues where applicable. Pull requests must describe motivation, list validation commands, call out performance impact, and note any documentation updates. Follow `CONTRIBUTING.md`: submit to `pgsty/silo-pkg` under AGPL-3.0-or-later with DCO sign-off, retaining original authorship and notices. Keep feedback actionable and ensure CI is green before requesting review.

## Security & Configuration Tips
Treat secrets, certificates, and credentials as sensitive: never commit real values. Validate inputs touching filesystem or network boundaries, rely on existing policy enforcement helpers, and update `README.md` when new environment variables or configuration flags are introduced.

## irulast Compatibility Policy

This branch maintains the historical `github.com/minio/pkg/v3` import identity
for the owned MinIO server fork. Preserve it when backporting current fixes from
PGSTY. Keep provenance, authorship, licenses, and DCO sign-off. Submit changes to
irulast/minio-pkg; the upstream contribution destination above applies only when
sending changes back upstream. Run builds and tests on the cluster through
irulast/minio-build; do not use local Docker or change scanner dispositions.

## Git 内容与大文件规则

用户于 2026-10-10 明确的通用仓库规则：

- 仓库默认保存源代码、配置和必要的轻量文档、元数据。构建产物、可执行文件、软件包、容器/虚拟机镜像、下载压缩包、数据库导出和大体积原始数据默认不进入 Git；文本格式的数据大文件同样需要检查。
- 在工作树内下载或生成上述文件之前，先选择仓库外的制品/数据位置，或补上合理的 `.gitignore` 规则并验证实际路径已被忽略，防止暂存和代理恢复/差异快照将它们写入 Git 对象库。
- 暂存、提交前检查新增文件的大小、类型和 staged 文件列表，只加入本次需要的源文件；不得用强制添加绕过忽略规则。优先保留来源清单、校验值及可复现的下载/构建方法。
- 对二进制素材、数据集或生成文件是否应纳入版本控制拿不准时，先问用户，再暂存或提交；确需例外时也先取得用户明确指示。
- 本规则不授权删除已有文件、stash、恢复快照或改写 Git 历史。忽略规则不会清除已存在的 Git 对象，历史清理需要单独明确范围。
