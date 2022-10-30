# Versionist
_A tool to manage multi module project import versions and go version_

Versionist is used to try reduce the amount of toil that would be needed to ensure module versions are correct.
However, `go mod tidy` is required to be run each project

## Versionist Config

The configuration file for the projec follows the following:

```yaml
go_version: 1.19
projects:
- package: github.com/awesome/package
  version: latest                       # Latest is a special keyword that is used to resolve the most recent value from GOPROXY settings
  match:                                # Match is not required but will exactly match the package name and include the additional regexp
  - regexp:^github.com/awesome/package/tools$
  - regexp:^github.com/awesome/package/components/(.*)$
```