<p align="center">
  <a href="https://www.pulumi.com?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=top-logo" title="Pulumi ESC: Open source secrets management solution">
    <img src="https://www.pulumi.com/images/logo/logo-on-white-box.svg?" width="350">
   </a>
</p>

# Secrets Management for Multi-Cloud Environments

**[Pulumi ESC](https://www.pulumi.com/product/esc/?utm_source=github.com&utm_medium=referral&utm_campaign=pulumi+esc+github+repo&utm_content=intro)** is a centralized secrets management & orchestration service that makes it easy to tame secrets sprawl and configuration complexity securely across all your cloud infrastructure and applications. You can pull and sync secrets with any secrets store – including HashiCorp Vault, AWS Secrets Manager, Azure Key Vault, GCP Secret Manager, 1Password, and more – and consume secrets in any application, tool, or CI/CD platform.

Pulumi ESC simplifies the adoption of dynamic, on-demand secrets as a best practice. It leverages Pulumi Cloud identity, RBAC, Teams, SAML/SCIM, OIDC, and scoped access tokens used for Pulumi IaC to ensure secrets management complies with enterprise security policies. Every time secrets or configuration values are accessed or changed with Pulumi ESC, the action is fully logged for auditing. So you can trust (and prove) your secrets are secure. Pulumi ESC makes it easy to eliminate the need for developers to copy and paste secrets and store them in plaintext on their computers. Developers can easily access secrets via CLI, API, Kubernetes operator, the Pulumi Cloud UI, and in-code with Typescript/Javascript, Python, and Go SDKs.

Be sure to check out the **[Pulumi ESC explainer video](https://www.youtube.com/watch?v=JY3Cm1UUIYE)**.

## Table of contents

- :clapper: [Demo](#pulumi-esc-demo)
- :rocket: [Getting Started](#getting-started-with-pulumi-esc)
- :blue_book: [Documentation](https://pulumi.com/docs/pulumi-cloud/esc)
- :hammer_and_wrench: [How It Works](#how-pulumi-esc-works)
- :white_check_mark: [Features](#pulumi-esc-features)
- :compass:	[Roadmap](#resources)
- :busts_in_silhouette: [Community](#resources)
- :computer: [Resources](#resources)

## Pulumi ESC Demo

Pulumi ESC not only works great for your applications and IaC, including Pulumi IaC, but it also makes your day-to-day developer workflow much more secure and streamlined. For example, the Pulumi ESC CLI (esc) allows you to give your developers immediate, just-in-time authenticated, and short-lived access to cloud credentials across any cloud provider with just a single command: `esc run aws-staging -- aws s3 ls`.

In this example, an ESC environment named aws-staging has all the necessary staging environment configuration and OIDC setup to connect to AWS. Running this command opens up a temporary environment and executes the aws s3 ls command in that environment. The temporary AWS credentials are not stored anywhere, making them secure and also allowing you to switch between different environments dynamically.

![Pulumi's open source secrets management solution overview](./assets/esc.gif)

Pulumi ESC is also offered as a managed service as part of [Pulumi Cloud,](https://www.pulumi.com/product/pulumi-cloud/?utm_campaign=pulumi-esc-github-repo&utm_source=github.com) and this repo contains the implementation of the following key components of the ESC open source secrets and configuration management solution:

1. The `esc` CLI:  A CLI tool for managing and consuming environments, secrets and configuration using Pulumi ESC.
2. The Pulumi ESC evaluator:  The core specification and implementation of the document format for defining environments, and the syntax and semantics for evaluating environments to produce a set of configuration and secrets.

<div>
<a href="https://www.pulumi.com/docs/esc/get-started/?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=get-started-button" title="Get Started">
    <img src="https://www.pulumi.com/images/get-started.svg?" align="center" width="120" alt="Click here to get started with Pulumi's open source secrets manager ESC">
</a>
</div>

## Getting Started with Pulumi ESC

For a hands-on, self-paced tutorial see our Pulumi ESC [Getting Started](https://pulumi.com/docs/pulumi-cloud/esc/get-started?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=getting-started-install) to quickly get up and running.

### Download and Install Pulumi ESC

1. **Install**:

    To install the latest Pulumi ESC release, run the following (see full
    [installation instructions](https://www.pulumi.com/docs/install/esc/?utm_campaign=pulumi-esc-github-repo&utm_source=github.com&utm_medium=getting-started-install) for additional installation options):

    ```bash
    $ curl -fsSL https://get.pulumi.com/esc/install.sh | sh
    ```

### Building the ESC CLI Locally

You can build the CLI locally for testing by cloning this repo and running:

```shell
$ make install
```

This will produce an `esc` binary in your `GOBIN` directory.

## How Pulumi ESC Works

![Pulumi ESC: Open source secrets management overview](./assets/overview.png)

1. Pulumi ESC enables you to define environments, which are collections of secrets and configuration. Each environment can be composed from multiple environments.
2. Pulumi ESC supports a variety of configuration and secrets sources, and it has an extensible plugin model that allows third-party sources.
3. Pulumi ESC has a rich API that allows for easy integration.  Every value in an environment can be accessed from any execution environment.
4. Every environment can be locked down with RBAC, versioned, and audited.

### Why Pulumi ESC?

Pulumi ESC was designed to address a set of challenges that many infrastructure and application development teams face in managing configuration and secrets across their various environments:

* __Stop secret sprawl__: Pull and sync secrets and configuration with any secrets store – HashiCorp Vault, AWS Secrets Manager, Azure Key Vault, GCP Secret Manager, 1Password, and more – and consume in any application, tool, or CI/CD platform.
* __Trust (and prove) your secrets are secure__: Adopt dynamic, short-lived secrets on demand as a best practice. Lock down every environment with RBAC, versioning, and a full audit log of all changes.
* __Ditch `.env` files__: No more copying-and-pasting secrets or storing them in plaintext on dev computers. Developers can easily access secrets via CLI, API, Kubernetes operator, the Pulumi Cloud UI, and SDKs.
* __Use with or without Pulumi IaC__: Use Pulumi ESC independently, or use with Pulumi IaC to support storing secrets in config in a more secure way than using plaintext.

Pulumi ESC was born to address these problems and needs head on with the following features.

### Pulumi ESC Features

* __Centralized secrets management__: Access, share, and manage confidential information such as secrets, passwords, and API keys as well as configuration information such as network settings and deployment options.
* __Secrets orchestration__: Pull and sync configuration and secrets from any secrets store and consume in any application, tool, or CI/CD platform.
* __Composable environments__: Environments support importing one into another, allowing for easy composability and inheritance of shared secrets and configuration.
* __Versionable__: Every change to an environment as well as any of its secrets and configuration is versioned, so rolling back or accessing an old version is easy.
* __Role-based access control (RBAC)__: Role-based access controls (RBAC) makes it easy to secure your secrets and configurations by assigning permissions to users based on their role within your organization.
* __Dynamic Secrets__: Generate just-in-time, short-lived credentials that revoke access when the lease expires.
* __Audit Logging__: All actions taken on environments, secrets, or configuration values are fully logged for auditing.
* __Developer-friendly__: Developers can easily access secrets via CLI, API, Kubernetes operator, the Pulumi Cloud UI, and in-code with Typescript/Javascript, Python, and Go SDKs.

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
