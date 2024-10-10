# Traefik

To run pasolo with Traefik, here are the configuration sample. Dont forget to adjust to match your deployment

```yaml
http:
  routers:
    your-service-route:
      rule: "Host(`service.your.domain`) && PathPrefix(`/`)"
      service: your-service-backend
      middlewares:
        - pasolo-auth-redirect # redirects all unauthenticated to pasolo signin
      tls:
        certResolver: default
        domains:
          - main: "your.domain"
            sans:
              - "*.your.domain"
    pasolo-route:
      rule: "Host(`pasolo.your.domain`) && PathPrefix(`/`)"
      service: pasolo-backend
      tls:
        certResolver: default
        domains:
          - main: "your.domain"
            sans:
              - "*.your.domain"

  services:
    your-service-backend:
      loadBalancer:
        servers:
          - url: http://172.16.0.2:7555
    pasolo-backend:
      loadBalancer:
        servers:
          - url: http://172.16.0.1:4180

  middlewares:
    pasolo-auth-redirect:
      forwardAuth:
        address: https://pasolo.your.domain/
```