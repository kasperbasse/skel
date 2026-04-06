---
applyTo: "internal/github/**"
---

# internal/GitHub - Review Instructions

## Security: SSRF via raw_url

`validateRawURL` is the critical security gate. The GitHub API response embeds a
`raw_url` this code fetches directly. A crafted response could point to an internal
address (e.g. `169.254.169.254`).

Any code path that fetches `raw_url` MUST call `validateRawURL` first.
Allowed: HTTPS only, exact host `gist.githubusercontent.com` (case-insensitive),
no explicit port, no userinfo.

## Token handling

`ResolveToken` checks `GITHUB_TOKEN` first, then `gh auth token`.
Never log or include the token in error messages. Errors must only indicate
whether the token is missing or invalid - not reveal its value.

## Size caps

- `FetchGist`: 60 MB cap on API response body.
- `FindProfileJSON`: caller-supplied `maxSize`; enforced on both `Size` metadata
  AND bytes actually read. Both checks are required - do not remove either.

## Profile shape constraint

`FindProfileJSON` requires exactly one `.json` file per gist.
Zero or multiple `.json` files are errors. Do not relax this.

## Testability pattern

`httpClient` and `apiBase` are package-level variables so tests can swap them via
`httptest.NewServer` without constructor injection. Do not refactor these into
struct fields or function parameters.