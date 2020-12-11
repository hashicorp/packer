# Packer Plugin SDK

This SDK enables building Packer plugins. This allows Packer's users to use both the officially-supported builders, provisioners, and post-processors, and custom in-house solutions.

Packer itself is a tool for building identical machine images for multiple platforms from a single source configuration. You can find more about Packer on its [website](https://www.packer.io) and [its GitHub repository](https://github.com/hashicorp/packer).

## Packer CLI Compatibility

Packer v1.7.0 or later is needed for this SDK. Versions of Packer prior to that release are still compatible with third-party plugins, but the plugins should use the plugin tooling from inside earlier versions of Packer to ensure complete API compatibility.

## Go Compatibility

The Packer Plugin SDK is built in Go, and uses the [support policy](https://golang.org/doc/devel/release.html#policy) of Go as its support policy. The two latest major releases of Go are supported by the SDK.

Currently, that means Go **1.14** or later must be used when building a provider with the SDK.

## Getting Started

See the [Extending Packer](https://www.packer.io/docs/extending) docs for a guided tour of plugin development.

## Documentation

See the [Extending Packer](https://www.packer.io/docs/extending) section on the Packer website.

## Packer Scope (Plugins VS Core)

### Packer Core

 - acts as an RPC _client_
 - interacts with the user
 - parses (HCL/JSON) configuration
 - manages build as whole, asks **plugin(s)** to manage the image lifecycle and modify the image being built.
 - discovers **plugin(s)** and their versions per configuration
 - manages **plugin** lifecycles (i.e. spins up & tears down plugin process)
 - passes relevant parts of parsed (valid JSON/HCL) and interpolated configuration to **plugin(s)**

### Packer Provider (via this SDK)

 - acts as RPC _server_
 - executes any domain-specific logic based on received parsed configuration. For builders this includes managing the vm lifecycle on a give hypervisor or cloud; for provisioners this involves calling the operation on the remote instance.
 - tests domain-specific logic via provided acceptance test framework
 - provides **Core** with template validation, artifact information, and information about whether the plugin process succeeded or failed.

## Migrating to SDK from built-in SDK

Migrating to the standalone SDK v1 is covered on the [Plugin SDK section](https://www.packer.io/docs/extend/plugin-sdk.html) of the website.

## Versioning

The Packer Plugin SDK is a [Go module](https://github.com/golang/go/wiki/Modules) versioned using [semantic versioning](https://semver.org/).

## Contributing

See [`.github/CONTRIBUTING.md`](https://github.com/hashicorp/packer-plugin-sdk/blob/master/.github/CONTRIBUTING.md)

## License

[Mozilla Public License v2.0](https://github.com/hashicorp/Packer-plugin-sdk/blob/master/LICENSE)
