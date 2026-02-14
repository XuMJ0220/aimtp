# Tencent Go Backend & Git Standards

## 1. Branch Naming Strategy
**Format**: `type/description-in-english`

- **Type**: Must use one of the "Commit Types" listed below.
- **Description**: Strict **English**, lowercase, kebab-case (hyphen-separated).
- **Examples**:
  - `feat/wechat-login`
  - `fix/db-connection-leak`
  - `refactor/auth-middleware`

## 2. Generate Commit Messages
When writing commit messages, strictly follow Conventional Commits (Hybrid Style):

**Format**: `<type>(<scope>): <subject>`

- **Type**: MUST be **English** (`feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`).
- **Scope**: MUST be **English** (module/package name, e.g., `auth`, `api`).
- **Subject**: MUST be **Chinese** (unless instructed otherwise).

**Example**:
`feat(user): 增加用户实名认证接口`
`fix(db): 修复连接池在高并发下的内存泄漏`