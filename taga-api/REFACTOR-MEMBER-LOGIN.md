You are an autonomous Go coding agent.

GOAL:
Refactor the login feature to:

* Issue JWT on successful authentication
* Include user roles in JWT claims
* Add JWT validation for subsequent requests

CONTEXT:

* File(s): handler/memberlogin.go
* Relevant code: <only required snippet>

REQUIREMENTS:

* Use `github.com/golang-jwt/jwt/v5`
* Use `bcrypt` for password verification if not already used
* JWT must include:

  * `sub` (user identifier)
  * `roles` ([]string) (derive from existing user model or default to ["USER"])
  * `exp`, `iat`
* Read JWT secret from environment variable: `JWT_SECRET`
* Token response:
  {
  "access_token": "...",
  "token_type": "Bearer"
  }

MIDDLEWARE:

* Implement minimal JWT validation middleware
* Extract token from `Authorization: Bearer <token>`
* Validate signature and expiry
* Attach claims to request context
* Place middleware in same file unless clearly inappropriate

GIT:

* Create and use branch: `feature/jwt-auth`
* Stage and commit only modified/added files

CONSTRAINTS:

* Modify only what is required for the goal
* Do not touch unrelated code
* Follow existing style and patterns
* Keep changes minimal and production-ready
* Avoid introducing new dependencies beyond JWT and bcrypt
* Do not refactor unrelated logic
* Max total code changes: 120 lines
* No explanations, no comments unless necessary

PLANNING STEP (MANDATORY):

* Output a plan BEFORE making any code changes

Rules:

* Max 5 steps
* Each step max 12 words
* Only include actions required for JWT + middleware
* Do NOT introduce new architecture or refactors
* Prefer modifying existing file over creating new ones

EXECUTION:

* After the plan, immediately implement changes
* Implementation MUST strictly follow the plan
* Do not add extra steps during execution

OUTPUT ORDER (STRICT):

<plan>

```text
<step 1>
<step 2>
...
```

<file_path>

```diff
<git diff format OR full updated file>
```

<git_commands>

```bash
git checkout -b feature/jwt-auth
git add <files>
git commit -m "add jwt auth with roles"
```

<summary>

```text
- short bullet points only
```

VALIDATION CHECKS:

* `build.sh` runs successfully
* No unused imports or variables
* No breaking API changes unless required
* JWT is correctly generated and validated
* Roles are present in token and accessible via context
