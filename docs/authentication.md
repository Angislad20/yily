# Proposal: YILY Authentication

## 1. Objective

Implement an authentication system fully managed by YILY, without relying on an identity provider, SMTP, or any external service.

The first version focuses exclusively on the Web interface and the Super Admin.

The goal is to keep the solution lightweight while properly protecting access to the instance and sensitive operations.

## 2. Authentication

Access to YILY is handled using:

- an email address;
- a password.

The password is protected using Argon2id and is never stored in plaintext.

No external authentication mechanism is required.

## 3. Sessions

After successful authentication, YILY creates a server-side session.

The browser receives a session identifier through a secure cookie.

The cookie uses, among other protections:

- `HttpOnly` to prevent direct access from JavaScript;
- `Secure` when HTTPS is used;
- `SameSite` to reduce risks from cross-site requests.

Sessions must support:

- expiration;
- revocation;
- invalidation on logout;
- invalidation after a password change.

The session identifier must be renewed after successful authentication to prevent session fixation.

Server-side sessions allow YILY to keep control of authentication state and immediately revoke a compromised session.

## 4. CSRF Protection

Cookie-based authentication requires protection against CSRF attacks.

Requests that modify the application state must therefore be protected, including:

- creating or modifying a variable;
- deleting a variable;
- changing the password;
- modifying sensitive settings.

The exact CSRF protection mechanism will be defined during implementation.

## 5. Login Attempt Management

The number of consecutive authentication attempts will be limited to 5.

After 5 failed attempts, authentication is temporarily blocked.

The account must not be automatically deleted or permanently locked, as this could allow an attacker to intentionally cause a denial of service against the Super Admin.

Error messages should remain sufficiently generic to avoid unnecessarily revealing account information.

Common situations should also be handled properly, including:

- incorrect password;
- Caps Lock being enabled;
- expired session;
- revoked session;
- invalid credentials.

Rate limiting can be strengthened later if necessary.

## 6. Lost Password

YILY will not depend on any external service for password recovery.

We do not propose a recovery mechanism that bypasses authentication.

If the Super Admin permanently loses their credentials and has no available way to restore access, YILY must not provide a procedure that bypasses the instance's security mechanisms.

This is an intentional decision: a permanent recovery mechanism would itself introduce another attack surface.

An instance may be restored from a backup when the elements required for restoration are available.

## 7. Personal Access Token — Step 2

Personal Access Tokens (PATs) will be introduced in a second step for use cases requiring token-based authentication, particularly API access.

The token will be generated using cryptographically secure randomness with 256 bits of entropy.

It may use a YILY-specific format, for example:

`yily_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`

The complete token will not be stored in the database.

YILY will only store its SHA-256 hash.

When a PAT is presented, YILY hashes the provided value and compares it with the stored hash.

PATs must be revocable and may have an expiration time.

PATs are not used as the primary authentication mechanism for the Web interface.

## 8. Security Principles

The solution must ensure that:

- passwords are never stored in plaintext;
- session identifiers are generated using cryptographically secure randomness;
- expired or revoked sessions can no longer be used;
- a new session is created after authentication;
- existing sessions are invalidated after a password change;
- PATs are never stored in plaintext;
- sensitive secrets are not exposed in logs or error messages;
- sensitive operations require valid authentication;
- requests using session cookies are protected against CSRF;
- network communications use HTTPS when YILY is exposed over a network.

## 9. Scope

The first step includes:

- email and password authentication;
- password hashing with Argon2id;
- server-side sessions;
- secure cookies;
- session expiration and revocation;
- CSRF protection;
- a limit of 5 consecutive authentication attempts;
- explicit handling of authentication errors.

PAT support constitutes the second step.

The goal is to keep authentication minimal and self-contained, adapted to YILY's self-hosted model, without introducing additional mechanisms until they address a real requirement.