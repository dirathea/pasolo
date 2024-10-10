# Caddy

To run pasolo with Caddy, here are the configuration sample. Dont forget to adjust to match your deployment

```Caddyfile
# Serve your app
your.app.domain {
	forward_auth / pasolo.domain {
		uri /validate
		copy_headers Remote-User Remote-Groups Remote-Name Remote-Email
		
		@error status 401
		handle_response @error {
			redir * https://pasolo.domain/login?rd={scheme}://{host}{uri} 302
		}
	}

	reverse_proxy your.app
}
```