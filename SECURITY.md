# Security policy

## Supported versions

We release security fixes for the latest minor release on the default branch (`main`). Older tags may not receive backports unless noted in a security advisory.

## Reporting a vulnerability

Please **do not** open a public issue for security vulnerabilities.

Instead, use **GitHub private vulnerability reporting**:

1. Open the repository’s **Security** tab.
2. Choose **Report a vulnerability** (or use [Report a vulnerability](https://github.com/alanzng/manifestor/security/advisories/new) if you are signed in and have access).

Include as much of the following as you can:

- A clear description of the issue and its impact
- Steps to reproduce, or a minimal proof of concept
- Affected component (library, CLI, or HTTP server) and version or commit, if known

We will acknowledge receipt when we can and work with you on a coordinated disclosure timeline before public discussion or release notes, where practical.

## Scope

This project includes parsing and network-facing surfaces (for example, manifest fetching and the optional HTTP proxy). Reports about unsafe handling of untrusted manifests, request smuggling, SSRF, path traversal, denial of service, or similar issues in maintained code are in scope. Third-party content linked from manifests (for example, arbitrary segment URLs) is generally the responsibility of operators and players unless the flaw is in how this library or server processes those references.

Thank you for helping keep users of **manifestor** safe.
