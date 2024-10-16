# Introduction

Pasolo is an authentication server for single users with passkeys. Pasolo runs alongside with other reverse proxy system such as Caddy, Traefik,  Nginx, using `forward_auth` or `external_auth` functionality.

<div style="position: relative; padding-bottom: 64.86486486486486%; height: 0;"><iframe src="https://www.loom.com/embed/4da6df49c2af4eb6a1007b87c7e4ed9b?sid=ec0ddc44-4f87-4a41-95e1-8e0ce9c6a071" frameborder="0" webkitallowfullscreen mozallowfullscreen allowfullscreen style="position: absolute; top: 0; left: 0; width: 100%; height: 100%;"></iframe></div>

## Why?

Inspired by a post in [r/selfhosted](https://www.reddit.com/r/selfhosted/comments/1f7fith/passkeys/) and other similar project like [Vouch Proxy](https://github.com/vouch/vouch-proxy), [oauth2-proxy](https://oauth2-proxy.github.io/oauth2-proxy/) and [Ory oathkeeper](https://www.ory.sh/docs/oathkeeper), Pasolo developed for self-hosted / home lab use cases, where the user is very limited -- in this case 1 person -- but want some degree of authentication on their setup.

## How It Works

![Pasolo Network Illustration](/docs/static/img/pasolo-network-illustration.png)

Pasolo runs alongside with your load balancer, works the best when it runs as one of your subdomain.

1. Client tries to access `app.your.domain`, the request received by Load Balancer.
2. Instead of forwarded directly to `app.your.domain`, the request forwarded to `pasolo.your.domain`.
3. Pasolo validate the request via request cookies. When pasolo find the request doesn't contains cookies or contains invalid cookies, it return 401 (Not Authorized)
4. The Load Balancer receive the 401 and act accordingly. It is recommended to configure the Load Balancer to forward the user to pasolo login page -- in this example `pasolo.your.domain/login`
5. Client login using passkeys that has been registered on the setup process, then redirected to `app.your.domain`.
6. Same as step 2, load balancer forward it to pasolo
7. Pasolo validate the request and return success 200
8. Then Load Balancer forward the request to `app.your.domain`
9. `app.your.domain` now reply the request as usual.

Please note that the redirection to Pasolo `/login` page only happen when no session found on the request, or the existing session is invalid.

## Getting Started

To get started, choose your desired reverse proxy to use, then configure it to use pasolo for authentication