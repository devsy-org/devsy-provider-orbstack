# OrbStack Provider for Devsy

Runs Devsy workspaces inside [OrbStack](https://orbstack.dev) Linux machines on
macOS. Each workspace gets an OrbStack machine; the provider installs Docker in
the machine and the devcontainer is deployed by Devsy's Docker driver, so
existing `devcontainer.json` configurations work unchanged.

## Getting started

Install [OrbStack](https://orbstack.dev) first (`orbctl` must be on `PATH`),
then add the provider:

```sh
devsy provider add devsy-org/devsy-provider-orbstack
devsy provider use orbstack
```

OrbStack is macOS-only.

### Creating your first devsy env

```sh
devsy workspace up .
```

The first run creates the machine and installs Docker in it, so it takes a
couple of minutes.

## Isolated machines

By default the provider creates OrbStack machines with `--isolated`, which
disables host file sharing and integration. This gives a clean, reproducible
workspace boundary. Set `ORBSTACK_ISOLATED=false` to share files and
integrations with the host.

## Multiple workspaces on one machine

By default each workspace gets its own machine. To run multiple devcontainers on
a single shared machine, enable single-machine mode:

```sh
devsy provider add devsy-org/devsy-provider-orbstack --single-machine
```

Devsy then reuses one machine for all workspaces on this provider and deletes it
only when the last workspace is removed.

## Options

| NAME | REQUIRED | DESCRIPTION | DEFAULT |
| ---- | -------- | ----------- | ------- |
| ORBSTACK_DISTRO | false | Linux distribution for the machine. | ubuntu |
| ORBSTACK_VERSION | false | Distro version, e.g. 24.04. Empty uses the latest. | |
| ORBSTACK_ARCH | false | CPU architecture (arm64 or amd64). Empty uses the host. | |
| ORBSTACK_CPUS | false | CPU core limit. Empty uses the OrbStack default. | |
| ORBSTACK_MEMORY | false | Memory limit, e.g. 4G. Empty uses the OrbStack default. | |
| ORBSTACK_DISK | false | Disk limit, e.g. 64G. Empty uses the OrbStack default. | |
| ORBSTACK_ISOLATED | false | Create an isolated machine (no host file sharing/integration). | true |
| ORBSTACK_PATH | false | Path to the orbctl binary. | orbctl |
| AGENT_PATH | false | Path where the Devsy agent is injected in the machine. | /tmp/devsy |
| INACTIVITY_TIMEOUT | false | Stops the machine after the given idle period, e.g. 10m or 1h. | 10m |
| INJECT_GIT_CREDENTIALS | false | Inject git credentials into the machine. | true |
| INJECT_DOCKER_CREDENTIALS | false | Inject docker credentials into the machine. | true |

Set options at add time or later:

```sh
devsy provider add devsy-org/devsy-provider-orbstack -o ORBSTACK_MEMORY=8G -o ORBSTACK_DISTRO=debian
devsy provider set orbstack --option ORBSTACK_ISOLATED=false
```

## Development

```sh
task build:cli            # cross-compile binaries into ./dist
task build:provider:dev   # generate ./dist/provider.yaml pointing at ./dist
task test
task lint
```

## License

MPL-2.0
