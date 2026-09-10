## Sparkpost Provider for Terraform

A provider to create and manage various resources in Sparkpost using Terraform / OpenTofu.

Not all areas are covered and it's not 100% at present but it works. There are various things to improve in the future.

If you have a resource you would like adding, please raise an issue.

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

See [docs/](docs/) for full schema reference, and [CHANGELOG.md](CHANGELOG.md) for release history.

#### Note

This Terraform provider is not affiliated with, endorsed by, or maintained by SparkPost. It is an independent project developed and maintained by volunteers.

Use it at your own discretion. For official support or features, please refer to SparkPost’s own tools and documentation.

This provider is provided “as-is” without any warranties.