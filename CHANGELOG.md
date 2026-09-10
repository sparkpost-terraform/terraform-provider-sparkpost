# Changelog

Notable changes to this provider, starting from the 1.0.0 release. Earlier
0.x versions predate this file - see the GitHub releases for those.

## 1.0.0

First stable release.

### Resources

- `sparkpost_subaccount`
- `sparkpost_domain`
- `sparkpost_domain_ownership_verification`
- `sparkpost_domain_bounce_verification`
- `sparkpost_tracking_domain`
- `sparkpost_tracking_domain_verification`
- `sparkpost_tracking_domain_association`
- `sparkpost_tracking_domain_https_configuration`

### Data sources

- `sparkpost_subaccounts`

### Fixed during the 1.0.0 betas

- `sparkpost_subaccounts` could return subaccounts in a different order on
  every read, producing plan diffs with no real change behind them.
- `sparkpost_domain` and `sparkpost_tracking_domain` could be forced into an
  unwanted replace when an unset optional attribute was refreshed from the
  API and diverged from its planned `null` value.
- Renaming a `sparkpost_subaccount` (or any in-place update to it) could
  cascade into destroying and recreating every domain and tracking domain
  under it, because computed attributes weren't preserved across the update.
- Errors from the SparkPost API now include its actual error message
  instead of just the HTTP status code.
- `sparkpost_tracking_domain`'s `https` attribute was removed in favour of
  `sparkpost_tracking_domain_https_configuration`, which enables a managed
  certificate (if requested) before turning HTTPS on, avoiding a chicken-and-
  egg failure where verifying a domain with `https = true` requires a
  certificate that doesn't exist until after the domain is verified.
