<p align="center">
  <a href="https://www.pulumi.com?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=top-logo" title="Pulumi ESC: Open source secrets management solution">
    <img src="https://www.pulumi.com/images/logo/logo-on-white-box.svg?" width="350" alt="Pulumi logo">
   </a>

 [![License](https://img.shields.io/github/license/pulumi/pulumi)](LICENSE)
 [![Slack](http://www.pulumi.com/images/docs/badges/slack.svg)](https://slack.pulumi.com?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=slack-badge)
 [![X (formerly Twitter) Follow](https://img.shields.io/twitter/follow/PulumiCorp)](https://x.com/PulumiCorp)
</p>

# Pulumi ESC (Environments, Secrets, and Configuration)

**[Pulumi ESC](https://www.pulumi.com/product/esc/)** is an open-source secrets management platform that tames secrets and configuration complexity across cloud infrastructure and application environments.

With Pulumi ESC, teams can aggregate secrets and configurations from many sources, manage hierarchical collections of configurations and secrets as Environments, and consume them through various means, including CLI, SDK, REST API, Pulumi Cloud Web Console, and Pulumi-service provider.

Pulumi ESC not only enhances your applications and IaC, including Pulumi IaC, but also significantly bolsters your day-to-day developer workflow with its robust security features. For instance, the Pulumi ESC CLI (esc) empowers you to provide your developers with immediate, just-in-time authenticated, and short-lived access to cloud credentials across any cloud provider with a single command: `esc run aws-staging -- aws s3 ls`.

In this example, an ESC Environment named aws-staging has all the necessary staging Environment configuration and OIDC setup to connect to AWS. Running this command opens up a temporary Environment and executes the aws s3 ls command in that environment. The temporary AWS credentials are not stored anywhere, making them secure and allowing you to switch between different Environments dynamically.

![Pulumi's open source secrets management solution overview](./assets/esc.gif)

Pulumi ESC is offered as a managed service as part of [Pulumi Cloud](https://www.pulumi.com/product/pulumi-cloud/?utm_campaign=pulumi-esc-github-repo&utm_source=github.com), and this repo contains the implementation of the following key components of the ESC open source secrets and configuration management solution:

1. The `esc` CLI:  A CLI tool for managing and consuming Environments, secrets, and configuration using Pulumi ESC.
2. The Pulumi ESC evaluator:  The core specification and implementation of the document format for defining Environments and the syntax and semantics for evaluating Environments to produce a set of configurations and secrets.

<div>
<a href="https://www.pulumi.com/docs/esc/get-started/?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=get-started-button" title="Get Started">
    <img src="https://www.pulumi.com/images/get-started.svg?" align="center" width="120" alt="Click here to get started with Pulumi ESC">
</a>
</div>

## Table of contents

- :rocket: [Getting Started](#getting-started-with-pulumi-esc)
- :blue_book: [Documentation](https://pulumi.com/docs/pulumi-cloud/esc)
- :hammer_and_wrench: [How Pulumi ESC Works](#how-pulumi-esc-works)
- :white_check_mark: [Pulumi ESC Features](#pulumi-esc-features)
- :compass: [Plumi ESC Roadmap](#resources)
- :busts_in_silhouette: [Community](#resources)
- :computer: [Resources](#resources)

## Getting Started with Pulumi ESC

For a hands-on, self-paced tutorial, see our [Pulumi ESC Getting Started guide]((https://pulumi.com/docs/pulumi-cloud/esc/get-started?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=getting-started-install)) to get up and running quickly. This guide will walk you through the installation process, setting up your first Environment, and executing commands using the 'esc' CLI.

### Download and Install Pulumi ESC

1. **Install**:

 To install the latest Pulumi ESC release, run the following (see full
 [installation instructions](https://www.pulumi.com/docs/install/esc/?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=getting-started-install) for additional installation options):

 ```sh
curl -fsSL https://get.pulumi.com/ | sh
 ```

### Building the ESC CLI Locally

You can build the CLI locally for testing by cloning this repo. Then run:

```sh
make install
```

This will produce an `esc` binary in your `GOBIN` directory.

## How Pulumi ESC Works

![Pulumi ESC: Open source secrets management overview](./assets/overview.png)

1. Pulumi ESC enables you to define Environments, which are collections of secrets and configurations. Each Environment can be composed of multiple Environments.
2. Pulumi ESC supports a variety of configuration and secrets sources, and it has an extensible plugin model that allows third-party sources.
3. Pulumi ESC has a rich API that allows for easy integration. Values in an Environment can be accessed from any execution Environment.
4. Every Environment can be locked down with RBAC, versioned, and audited.

### Why Pulumi ESC?

Pulumi ESC was designed to address a set of challenges that many infrastructure and application development teams face in managing configuration and secrets across their various Environments:

- **Secrets and Configuration Sprawl**: Data in many systems is challenging to audit. There is a lot of application-specific logic to acquire and compose configuration from multiple sources and divergent solutions for Infrastructure and Application configuration.
- **Duplication and Copy/Paste**: Secrets are duplicated in many places. Frequently coupled to application/system-specific configuration stores.
- **Too Many Long-lived Static Secrets**: Long-lived static credentials are overused, exposing companies to significant security risks. Rotation is operationally challenging. Not all systems support direct integration with OIDC and other dynamic secret provisioning systems.

Pulumi ESC was born to address these problems head-on with the following features.

### Pulumi ESC Features

- **Hierarchical Environments**: Environments contain collections of secrets and configurations but can also import one or more other Environments. Values can be overridden, interpolated from other values, and arbitrarily nested. This allows for flexible composition and reuse and avoids copy-paste.
- **Dynamic + Static Secrets**: This service supports static values and dynamic values pulled from systems. Static values can be encrypted, and dynamic secrets plugins include AWS OIDC, HashiCorp Vault, AWS Secrets Manager, 1Password, and Pulumi StackReference.
- **Auditable**: Every Environment opening is recorded in audit logs, providing a concrete set of configurations derived from imported Environments and dynamic secrets.
- **Consume from Anywhere**: The `esc` CLI and the Pulumi ESC Rest API enable Environments to be accessed from any application, infrastructure provider, or automation system. At launch, first-class integrations are available with Pulumi IaC, local environment and .env files, GitHub Actions, and more.
- **Authentication and RBAC**: Pulumi ESC brokers access secrets and configurations in other systems, so authentication and granular RBAC are critical to ensuring robust access controls across your organization. Pulumi ESC leverages the same Pulumi Cloud identity, RBAC, Teams, SAML/SCIM, and scoped access tokens that are used for Pulumi IaC today, extending these to manage access to Environments as well as Stacks.
- **Configuration as Code**: Environments are defined as YAML documents that describe how to project and compose secrets and configuration, integrate dynamic configuration providers, and compute new configuration from other values (e.g., construing a URL from a DNS name or concatenating multiple configuration values into a derived value). The incredible flexibility of a code-based approach over traditional point-and-click interfaces allows Pulumi ESC to offer rich expressiveness for managing complex configurations.
- **Open Source + Managed**: Offers an open-source server with pluggable storage and authentication and a managed service in Pulumi Cloud and Pulumi Cloud Self-hosted options.
- **Version Control and Rollback**: Manage Environment changes with full audit and rollback capabilities.
- **Language SDKs**: Use ESC in Python, TypeScript/JavaScript, and Go applications.
- **Traceability and Auditing**: Environments must be “opened” to compute and see the set of values they provide. This action is recorded in audit logs, including a full record of how each value was sourced from within the hierarchy of Environments that contributed to it.
- **Composable Environments**: Combine multiple Environments for greater flexibility.
- **Dynamic Configuration Providers**: Support for dynamic configuration providers for more flexible management.
- **Fully Managed**: Pulumi ESC is offered as a fully managed cloud service in Pulumi Cloud (and will soon be available in the Pulumi Cloud self-hosted offering). The `pulumi/esc` project is open source, and contains the evaluation engine for Environments, the `esc` CLI, and in the future, the extensible plugins for source and target integrations.

## Pulumi ESC Roadmap

Review the planned work for the upcoming quarter and a selected backlog of issues that are on our mind but not yet scheduled on the [Pulumi Roadmap.](https://github.com/orgs/pulumi/projects/44)

## Community

- Join us in the [Pulumi Community Slack](https://slack.pulumi.com/?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=welcome-slack) to connect with our community and engineering team and ask questions. All conversations and questions are welcome.
- Send us a tweet via [@PulumiCorp](https://twitter.com/PulumiCorp)
- Watch videos and workshops on [Pulumi TV](https://www.youtube.com/pulumitv)

## Resources

- [Docs](https://pulumi.com/docs/pulumi-cloud/esc?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=esc-resources)
- [Slack](https://slack.pulumi.com/?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=welcome-slack)
- [Twitter](https://twitter.com/PulumiCorp)
- [YouTube](https://www.youtube.com/pulumitv)
- [Blog](https://pulumi.com/blog?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=esc-resources)
- [Roadmap](https://github.com/orgs/pulumi/projects/44)
