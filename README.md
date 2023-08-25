# Pulumi Environments Core

This repository contains the datatypes and evaluator that form the core of Pulumi Environments. Projects that build on
this foundation must supply their own implementations for `eval.ProviderLoader` and `eval.EnvironmentLoader` in order
to be fully functional.
