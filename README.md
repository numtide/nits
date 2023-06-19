<h1 align="center">
  <br>
  <img src="docs/assets/logo.png" alt="logo" width="200">
  <br>
  Nits — Nix & Nats
  <br>
  <br>
</h1>

**Status: highly experimental**

# Motivation

This project began as a directed learning exercise, taking everyday use cases and functionality from the world of [Nix](https://nixos.org) and seeing how to implement and integrate them with [NATS](https://nats.io).

Over time this has evolved towards a specific use case: _pull-based orchestration of NixOS hosts_.

Several options are already available for deploying NixOS across a series of hosts ([Colmena](https://github.com/zhaofengli/colmena), [deploy-rs](https://github.com/serokell/deploy-rs) et al.). However, these tools are inherently _push-based_, supporting _'online'_ deployments that require hosts to be contactable during rollout.

In some cases, it is desirable to have more of a _pull-based_ deployment, in which the system closure for a given host is updated ahead of time and applied when that host subsequently checks in. For example, you may have a series of devices deployed in an area with little network coverage.

We feel that NATS is a logical platform on top of which to build a pull-based deployment mechanism. As such, this project will work towards that goal, developing what is needed to meet it and see what valuable sub-components and side-projects may emerge along the way.

# Usage

The project is currently in a state of rapid iteration. We will update this section and add more documentation when it has reached a steady state.

# Preview

![](./docs/assets/deploy.gif)
