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

Pulumi ESC integrates with Pulumi Cloud identity and RBAC to provide rich control over access to secret configuration within an organization. Pulumi ESC supports multiple configuration providers, enabling static key/value configuration as well as dynamically retrieved configuration and secrets via OIDC and additional providers like 1Password and Vault.  Pulumi ESC is available via the new esc CLI, Pulumi Cloud, the Pulumi Cloud REST API, and Pulumi IaC stack configuration.

![Pulumi's open source secrets management solution overview](./assets/esc.gif)

<div>
<a href="https://www.pulumi.com/docs/esc/get-started/?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=get-started-button" title="Get Started">
    <img src="https://www.pulumi.com/images/get-started.svg?" align="center" width="120" alt="Click here to get started with Pulumi ESC">
</a>
</div>

This repo contains the implementation of the following key components:

- **The `esc` CLI** - A CLI tool to fully manage Pulumi ESC Environments.
- **The Pulumi ESC evaluator** - The core specification and implementation of the document format for defining Environments, and the syntax and semantics for evaluating Environments to produce a set of configurations and secrets.

## Table of contents

- :rocket: [Getting Started](#getting-started-with-pulumi-esc)
- :blue_book: [Documentation](https://pulumi.com/docs/pulumi-cloud/esc)
- :hammer_and_wrench: [How Pulumi ESC Works](#how-pulumi-esc-works)
- :white_check_mark: [Pulumi ESC Features](#pulumi-esc-features)
- :compass: [Plumi ESC Roadmap](#resources)
- :busts_in_silhouette: [Community](#resources)
- :computer: [Resources](#resources)

## Getting Started with Pulumi ESC

For a hands-on, self-paced tutorial, see our [Pulumi ESC Getting Started guide]((https://pulumi.com/docs/pulumi-cloud/esc/get-started?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=getting-started-install)) to get up and running quickly. This guide will walk you through the installation process, setting up your first Environment, and executing commands using the `esc` CLI.

### Install the Pulumi ESC CLI

```sh
curl -fsSL https://get.pulumi.com/ | sh
```

See the [installation instructions](https://www.pulumi.com/docs/install/esc/?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=getting-started-install) for additional options.

### Build Locally

You can build the CLI locally for testing.

1. Clone this repo.
2. Run

```sh
make install
```

This will produce an `esc` binary in your `GOBIN` directory.

## How Pulumi ESC Works

![Pulumi ESC: Open source secrets management overview](./assets/overview.png)

Pulumi ESC is offered as a managed service as part of [Pulumi Cloud](https://www.pulumi.com/product/pulumi-cloud/?utm_campaign=pulumi-esc-github-repo&utm_source=github.com).

1. Pulumi ESC enables you to define Environments, which are collections of secrets and configurations. Each Environment can be composed of multiple Environments.
2. Pulumi ESC supports a variety of configuration and secrets sources, and it has an extensible plugin model that allows third-party sources.
3. Pulumi ESC has a rich API that allows for easy integration. Values in an Environment can be accessed from any execution environment.
4. Every Environment can be locked down with RBAC, versioned, and audited.

Pulumi ESC encrypts all data at rest using AWS S3 encryption. All API routes serving the Pulumi ESC API are HTTPS, and authenticated via Pulumi Access Token.

### Why Pulumi ESC?

Pulumi ESC was designed to address a set of challenges that many infrastructure and application development teams face in managing configuration and secrets across their various environments:

- **Secrets and Configuration Sprawl**: Data in many systems is challenging to audit. There is a lot of application-specific logic to acquire and compose configuration from multiple sources and divergent solutions for Infrastructure and Application configuration.
- **Duplication and Copy/Paste**: Secrets are duplicated in many places. Frequently coupled to application/system-specific configuration stores.
- **Too Many Long-lived Static Secrets**: Long-lived static credentials are overused, exposing companies to significant security risks. Rotation is operationally challenging. Not all systems support direct integration with OIDC and other dynamic secret provisioning systems.

Pulumi ESC not only enhances your applications and IaC, including Pulumi IaC, but also significantly bolsters your day-to-day developer workflow with its robust security features. For instance, the Pulumi ESC CLI (esc) empowers you to provide your developers with immediate, just-in-time authenticated, and short-lived access to cloud credentials across any cloud provider with a single command: `esc run aws-staging -- aws s3 ls`.

### Pulumi ESC Features

- **Hierarchical Environments**: Environments contain collections of secrets and configurations but can also import one or more other Environments. Values can be overridden, interpolated from other values, and arbitrarily nested. This allows for flexible composition and reuse and avoids copy-paste.
- **Dynamic + Static Secrets**: This service supports static values and dynamic values pulled from systems. Static values can be encrypted, and dynamic secrets plugins include AWS OIDC, HashiCorp Vault, AWS Secrets Manager, 1Password, and Pulumi StackReference.
- **Auditable**: Every Environment interaction is recorded in audit logs, providing a concrete set of configurations derived from imported Environments and dynamic secrets.
- **Consume from Anywhere**: The `esc` CLI and the Pulumi ESC REST API enable Environments to be accessed from any application, infrastructure provider, or automation system. At launch, first-class integrations are available with Pulumi IaC, local environment and .env files, GitHub Actions, and more.
- **Authentication and RBAC**: Pulumi ESC brokers access secrets and configurations in other systems, so authentication and granular RBAC are critical to ensuring robust access controls across your organization. Pulumi ESC leverages the same Pulumi Cloud identity, RBAC, Teams, SAML/SCIM, and scoped access tokens that are used for Pulumi IaC today, extending these to manage access to Environments as well as Stacks.
- **Configuration as Code**: Environments are defined as YAML documents that describe how to project and compose secrets and configuration, integrate dynamic configuration providers, and compute new configuration from other values (e.g., construing a URL from a DNS name or concatenating multiple configuration values into a derived value). The incredible flexibility of a code-based approach over traditional point-and-click interfaces allows Pulumi ESC to offer rich expressiveness for managing complex configurations.
- **Open Source + Managed**: Offers an open-source server with pluggable storage and authentication and a managed service in Pulumi Cloud and Pulumi Cloud Self-hosted options.
- **Version Control and Rollback**: Manage Environment changes with full versioning and rollback capabilities.
- **Language SDKs**: Use ESC in Python, TypeScript/JavaScript, and Go applications.
- **Traceability and Auditing**: Environments must be “opened” to compute and see the set of values they provide. This action is recorded in audit logs, including a full record of how each value was sourced from within the hierarchy of Environments that contributed to it.
- **Composable Environments**: Combine multiple Environments for greater flexibility.
- **Dynamic Configuration Providers**: Support for dynamic configuration providers for more flexible management.
- **Fully Managed**: Pulumi ESC is offered as a fully managed cloud service in Pulumi Cloud (and will soon be available in the Pulumi Cloud self-hosted offering). The `pulumi/esc` project is open source, and contains the evaluation engine for Environments, the `esc` CLI, and in the future, the extensible plugins for source and target integrations.

## Community

Engage with our community to elevate your developer experience:

- **Join our online [Pulumi Community on Slack](https://slack.pulumi.com/)** - Interact with over 13K Pulumi developers for collaborative problem-solving and knowledge-sharing.
- **Join a [local Pulumi User Group (PUG)](https://www.meetup.com/pro/pugs/)** - Attend tech-packed meetups and hands-on virtual or in-person workshops.
- **Follow [@PulumiCorp](https://twitter.com/PulumiCorp) on X (Twitter)** - Get real-time updates, technical insights, and sneak peeks into the latest features.
- **Subscribe to our YouTube Channel, [PulumiTV](https://www.youtube.com/@PulumiTV)** - Learn about AI / ML essentials, launches, workshops, demos, and more.
- **Follow our [LinkedIn](https://www.linkedin.com/company/pulumi/)** - Uncover company news, achievements, and behind-the-scenes glimpses.

## Resources

- [Get Started with Pulumi ESC](https://www.pulumi.com/docs/esc/) - Complete a tutorial to learn how to manage ESC Environments.
- [Pulumi Blog](https://www.pulumi.com/blog/?utm_source=GitHub&utm_medium=referral&utm_campaign=workshops) - Stay in the loop with our latest tech announcements, insightful articles, and updates.
