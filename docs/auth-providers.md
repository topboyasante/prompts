# Auth Providers

## Supported Providers

- `github`
- `google`

## Endpoints

- `GET /v1/auth/{provider}/login`
- `GET /v1/auth/{provider}/callback`

## Account Linking Policy

- Auto-link is enabled when provider email is verified.
- If email is missing or unverified, account creation is still allowed.
- Provider identities are stored in `user_identities`.

## Provider Notes

- GitHub may not return a public email from `/user`; the API checks `/user/emails` for verified addresses.
- Google returns identity data from OpenID userinfo endpoint.
