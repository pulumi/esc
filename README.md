# Pulumi ESC (Environments, Secrets, Configurations)
Pulumi ESC is a solution to manage application and infrastructure environments, secrets, and configurations. Pulumi ESC is an easy-to-use single source of truth for all configurations with guardrails. Pulumi ESC reduces downtime over changed configurations because you can change a value once and have it updated everywhere. Pulumi ESC also enforces least-privileged access through role-based access controls and maintains compliance by logging all value accesses and changes for auditing. 

## How Pulumi ESC works
![Pulumi ESC Graphic V4_yes_numbers](https://github.com/pulumi/esc/assets/7718195/bebdf14b-be81-4509-b738-87b72ba42417)

1. Pulumi ESC enables you to define environments, which are collections of secrets and configurations. Each environment can be composed from  multiple environments.
2. Pulumi ESC supports a variety of configuration and secrets sources, and it has an extensible plugin model that allows third-party sources.
3. Pulumi ESC has a rich API that allows for easy integration.  Every value in an environment can be accessed from any execution environment.
4. Every environment can be locked down with RBAC, versioned, and audited. 

## Getting Started
<TBD Video>

See the [Get Started](https://www.pulumi.com/docs/esc/) guide to quickly start using Pulumi ESC. 

## Misc
This repository contains the datatypes and evaluator that form the core of Pulumi ESC. Projects that build on
this foundation must supply their own implementations for `eval.ProviderLoader` and `eval.EnvironmentLoader` in order
to be fully functional.
