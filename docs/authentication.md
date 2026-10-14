# Proposal: Authentication and Security Architecture for YILY

## 1. Objective

YILY must provide authentication fully managed by the application, without relying on an external identity provider, SMTP, or any other third-party service.

The solution must cover YILY's three primary usage modes:

- Web Interface
- HTTP API
- CLI

The objective is to maintain a lightweight architecture while clearly separating authentication, authorization, business logic, persistence, and cryptography.

## 2. Proposed Architecture

```text
                 Web / API / CLI
                       |
                       v
               Authentication
                       |
                       v
                 Middleware
                       |
                       v
                Authorization
                       |
                       v
                    Service
                  /         \
                 v           v
           Repository      Crypto
                              |
                              v
                         Master Key
```

Each layer has a specific responsibility and must not take on the responsibilities of another layer.

## 3. Authentication

YILY manages user accounts directly.

Passwords are protected with **Argon2id**, paired with a cryptographically random salt. They are never stored in plaintext and are not encrypted as ordinary application data.

Authentication remains independent of the Master Key used to protect stored secrets.

### Why this choice?

Argon2id is designed for password hashing and significantly increases the cost of brute-force and dictionary attacks.

It also maintains a clear separation between:

- credentials used to access YILY;
- the key used to decrypt secrets stored by YILY.

## 4. Session Management

The Web interface uses **server-side sessions**.

After successful authentication, the server generates a random session identifier and sends it to the browser via a secure cookie.

The cookie must use:

- `HttpOnly`
- `Secure` when HTTPS is used
- `SameSite`

Sessions must be expirable and revocable.

A new session is generated after authentication to mitigate session fixation attacks.

### Why this choice?

Server-side sessions are straightforward to control and allow immediate revocation of compromised sessions.

They also prevent placing sensitive information or authorization claims directly in the browser.

## 5. Authorization

Authorization is decoupled from authentication.

YILY initially uses a simple role-based model:

- `User`
- `Admin`
- `Super Admin`

Permissions are verified before sensitive operations, in particular:

- reading a variable;
- creating or modifying a variable;
- deleting a variable;
- creating versions;
- retrieving a plaintext value;
- user management;
- modifying security settings.

The identity used for these checks must come exclusively from a validated authentication context.

### Why this choice?

This model remains sufficiently lightweight for YILY while supporting multiple users on a single instance.

It also avoids introducing an overly complex permission system prematurely when current requirements remain modest.

## 6. Middleware

The middleware is responsible for validating authentication on protected HTTP requests.

It retrieves and validates the session, then makes the authenticated identity available in the request context.

The middleware must not contain business logic.

Authorization decisions remain enforced at the appropriate layer of the application so that access control does not rely solely on the HTTP middleware.

## 7. Service

The Service layer remains responsible for orchestrating business operations.

It coordinates:

```text
Authorization
      |
      v
Service
 /       \
v         v
Repository Crypto
```

The Service determines when data must be created, modified, retrieved, or decrypted.

The Repository must not be responsible for encryption or business rules.

### Why this choice?

The same business operations can thus be reused across:

- the HTTP API;
- the Web interface;
- the CLI.

The CLI does not need to duplicate application business logic.

## 8. Secret Access

Secrets must not be decrypted unnecessarily.

An operation retrieving variable metadata must not automatically retrieve its plaintext value.

Retrieving a decrypted value must be performed through a dedicated operation and require corresponding authorization.

### Why this choice?

This minimizes unnecessary exposure of sensitive data and clearly separates:

- access to metadata;
- actual access to the secret value.

It also remains consistent with the cryptographic versioning already implemented.

## 9. HTTP API

The API enforces the same authentication and authorization rules as the Web interface.

Requests originating from the Web interface can reuse the session mechanism.

For external clients or the CLI communicating with a remote YILY instance, YILY can use **API tokens**.

Tokens must be:

- cryptographically random;
- associated with a user;
- scoped to required permissions when applicable;
- expirable;
- revocable.

Only the hash of the token should be stored when its plaintext value is not required for validation.

### Why this choice?

Tokens are better suited for non-browser clients than session cookies.

They also allow revoking access for a CLI or API client without changing the user's password.

## 10. CLI

The CLI must be able to operate with a local YILY instance without relying on a browser or an external authentication service.

For local access, security relies primarily on:

- operating system permissions;
- local database and file protections;
- the Master Key management mechanism.

When the CLI communicates with a remote instance, it uses dedicated credentials, such as an API token.

The CLI and HTTP API reuse the same business services.

### Why this choice?

This keeps the CLI lightweight while avoiding the creation of a second authentication system.

The CLI will be distributed as precompiled binaries via native package managers using **GoReleaser**, so users do not have to compile YILY themselves.

## 11. Password Reset

Password reset must be fully managed by YILY.

No email, SMTP server, or external service is required.

The proposed mechanism is a reset procedure controlled by an **Admin or Super Admin**.

When a reset is triggered:

1. YILY generates a temporary, random secret.
2. This secret is associated with the target account.
3. It has a limited validity period.
4. It can only be used once (single-use).
5. Its redemption allows setting a new password.
6. Existing sessions can be revoked.
7. The reset mechanism is invalidated upon use or expiration.

The temporary secret must not be stored in plaintext when its hash is sufficient for validation.

### Why this choice?

The procedure remains completely self-contained and adheres to the constraint of not depending on SMTP or third-party services.

It also allows precise control over expiration, single-use, and revocation of the recovery mechanism.

## 12. Authentication Data Model

The model can remain intentionally minimal:

```text
USER
- ID
- username / email
- password_hash
- role
- status
- created_at
- updated_at

SESSION
- ID
- user_id
- token_hash
- expires_at
- revoked_at

API_TOKEN
- ID
- user_id
- token_hash
- permissions
- expires_at
- revoked_at
```

Business resources must be associated with their owner or access scope when multiple users share a single instance.

## 13. Master Key Protection

The Master Key remains completely independent of the authentication system.

It must not be derived from a user's password.

It must not be stored alongside encrypted secrets without appropriate protection.

The configuration layer should only provide the representation expected by the application. Any transformation required for cryptographic operations must be performed by the component that actually uses the key, rather than by the general configuration package.

### Why this choice?

Compromising a user account must not automatically compromise the Master Key.

This preserves a clear separation between:

```text
Authentication
      ≠
Secret Protection
```

## 14. Key Security Safeguards

The initial implementation should notably include:

- Argon2id for password hashing;
- Server-side sessions for the Web interface;
- `HttpOnly`, `Secure`, and `SameSite` cookies;
- Session expiration and revocation;
- Rate limiting on login attempts;
- Generic error messages during authentication;
- Authorization checks before sensitive operations;
- CSRF protection for cookie-based flows where necessary;
- Random, expirable, and revocable API tokens;
- HTTPS for network communications;
- Zero secrets in logs;
- Tests covering unauthorized access and authorization bypass attempts;
- Session revocation following relevant security operations.

## 15. Implementation Scope

Once the proposal is approved, implementation can be organized around dedicated Go components:

```text
Authentication
    ├── User management
    ├── Password hashing / verification
    └── Password reset

Session
    ├── Session creation
    ├── Session validation
    └── Session revocation

Middleware
    └── Authentication context

Authorization
    └── Role / permission checks

Service
    └── Business operations

Repository
    ├── Users
    ├── Sessions
    └── API tokens

Crypto
    └── Existing encryption / decryption
```

This section intentionally does not define the exact Go interfaces or method names yet. These can be defined once the architectural direction is agreed upon.

## Conclusion

This proposal provides YILY with a **local, lightweight, and self-contained** authentication system adapted to its different usage modes.

The Web interface uses server-side sessions, the API can use tokens, and the CLI reuses the same business services without introducing a second authentication system.

Authorization is decoupled from authentication and takes place before sensitive operations, particularly secret decryption.

Finally, the Master Key and authentication mechanisms remain independent so that compromising a user account does not automatically compromise the cryptographic protection of stored secrets.