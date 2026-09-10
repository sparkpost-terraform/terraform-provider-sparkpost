# Changelog

Notable changes to this provider, starting from the 1.0.0 release. Earlier
0.x versions predate this file - see the GitHub releases for those.

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
