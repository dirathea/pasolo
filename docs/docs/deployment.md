---
sidebar_position: 2
---
# Deployment

Pasolo is available at [GitHub Release](https://github.com/dirathea/pasolo/releases) as a single binary, or as Container Image. To customize deployment, use environment variable.

# Required Environment Variable

| Key | Description | Example |
| --- | --- | --- |
| USER_ID | Your user id as the identifier for Passkey prompt | email@your.domain |
| USER_DISPLAY_NAME | Display name on your passkey identifier | John Doe |
| USER_NAME | Your username as identifier for Passkey prompt | johndoe |
| SERVER_PORT | Pasolo server port | "8080" |
| SERVER_DOMAIN | Your domain for authentication cookie | your.domain |
| SERVER_PROTOCOL | Pasolo server protocol. It is recommended to use https | https |
| COOKIE_NAME | Authentication cookie name | pasolo-auth |
| COOKIE_SECRET | JWT secret | secret |
| ENCRYPTION_KEY | Session and user data encryption key | secret |
| PASSKEY_ORIGIN | Pasolo server origin. Make sure this origin matches pasolo domain to make the passkey works | https://pasolo.your.domain |
| STORE_DATADIR | Path to store persistent data | /secret |


# Persistent Volumes

Pasolo also required persistent volume to store login session, as well as registered passkeys. `STORE_DATADIR` environment variable configures where the data should be stored.

```yaml
# example docker-compose.yml
services:
  auth:
    image: ghcr.io/dirathea/pasolo:latest
    env_file:
      - .env
    environment:
      STORE_DATADIR: /secret
    ports:
      - 8080
    volumes:
      - secret:/secret

volumes:
  secret:
```