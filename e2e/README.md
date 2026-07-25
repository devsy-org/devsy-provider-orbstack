## E2E tests

The suite has two label groups (both require macOS, since the provider binary is
darwin-only):

- `integration` builds the provider and exercises the no-machine paths (manifest
  generation, option validation, provider-binary error handling). It does not
  require OrbStack to be installed or running.
- `vm` provisions a real OrbStack machine and runs a full workspace lifecycle.
  It requires OrbStack (`orbctl` on `PATH`, OrbStack running). This label does
  not run in CI: OrbStack cannot start on GitHub-hosted macOS runners (they lack
  the Apple Virtualization capability it needs). Run it on a real Mac or a
  self-hosted macOS runner.

### Run

Build the provider binaries first, then run a label group from this directory:

```sh
task build:cli
go test -v -ginkgo.v -timeout 3600s --ginkgo.label-filter=integration
go test -v -ginkgo.v -timeout 3600s --ginkgo.label-filter=vm
```
