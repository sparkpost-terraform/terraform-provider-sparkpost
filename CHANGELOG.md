# Changelog

Notable changes to this provider, starting from the 1.0.0 release. Earlier
0.x versions predate this file - see the GitHub releases for those.

## 1.2.0

### Added

- `sparkpost_default_bounce_domain` makes a bounce domain the account's
  default. SparkPost only accepts that once the bounce CNAME is verified, so
  it can be ordered after `sparkpost_domain_bounce_verification`.

### Fixed

- Changing `default_bounce_domain` on `sparkpost_domain` updates the domain
  in place instead of replacing it. Replacing it deleted the domain, then
  failed to recreate it, because SparkPost rejects `is_default_bounce_domain`
  on create. A new domain is now always created without it, then set as the
  default afterwards, with a warning if SparkPost refuses.

## 1.1.0

### Added

- `sparkpost_domain` exposes `dkim_selector` and `dkim_public_key`, so the
  DKIM TXT record that ownership verification needs can be created in the
  same run.

## 1.0.1

### Fixed

- `sparkpost_domain` and `sparkpost_tracking_domain` no longer lose track of a
  domain SparkPost registers while answering the create with an error, such
  as 400 "Restricted domain". The domain is saved to state with a warning,
  where every later run used to fail with a 409. A 409 from the create itself
  still fails, so a domain that already existed needs an explicit import.
- `sparkpost_subaccount` saves a newly created subaccount to state even when
  reading it back or setting its initial status fails, with a warning, so a
  retry no longer creates a duplicate.
- `sparkpost_tracking_domain_https_configuration` skips enabling a managed
  certificate that's already enabled, so a create that failed after that step
  can be retried.

## 1.0.0

First stable release. Changes below are relative to the last release, v0.3.3.

### Breaking changes

- `sparkpost_tracking_domain` no longer has an `https` attribute. Move it to
  the new `sparkpost_tracking_domain_https_configuration` resource, which
  also lets you enable a SparkPost-managed TLS certificate before turning
  HTTPS on.

### New resources

- `sparkpost_subaccount` - create, update and terminate SparkPost subaccounts
- `sparkpost_tracking_domain_https_configuration` - configure HTTPS on a
  tracking domain, optionally enabling a SparkPost-managed certificate first

### Improvements

- API errors now surface SparkPost's actual error message instead of just
  the HTTP status code
