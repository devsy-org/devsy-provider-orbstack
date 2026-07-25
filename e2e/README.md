## E2E tests

The suite has two label groups (both require macOS, since the provider binary is
darwin-only):

- `integration` builds the provider and exercises the no-machine paths (manifest
  generation, option validation, provider-binary error handling). It does not
  require OrbStack to be installed or running.
- `vm` provisions a real OrbStack machine and runs a full workspace lifecycle.
  It requires OrbStack (`orbctl` on `PATH`, OrbStack running).

### Run

Build the provider binaries first, then run a label group from this directory:

```sh
task build:cli
go test -v -ginkgo.v -timeout 3600s --ginkgo.label-filter=integration
go test -v -ginkgo.v -timeout 3600s --ginkgo.label-filter=vm
```
