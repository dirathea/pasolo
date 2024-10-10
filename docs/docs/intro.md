---
sidebar_position: 1
---

# Introduction

Pasolo is an authentication server for single users with passkeys. Pasolo runs alongside with other reverse proxy system such as Caddy, Traefik,  Nginx, using `forward_auth` or `external_auth` functionality.

## Why?

Inspired by a post in [r/selfhosted](https://www.reddit.com/r/selfhosted/comments/1f7fith/passkeys/) and other similar project like [Vouch Proxy](https://github.com/vouch/vouch-proxy), [oauth2-proxy](https://oauth2-proxy.github.io/oauth2-proxy/) and [Ory oathkeeper](https://www.ory.sh/docs/oathkeeper), Pasolo developed for self-hosted / home lab use cases, where the user is very limited -- in this case 1 person -- but want some degree of authentication on their setup.

## Getting Started

To get started, choose your desired reverse proxy to use, then configure it to use pasolo for authentication