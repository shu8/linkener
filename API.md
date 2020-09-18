# Linkener REST API

<sub>[Back to README](./README.md)</sub>

The API is split up into 2 main endpoints. By default, it can be found at `/api` of your server's root (e.g. `http://localhost:3000/api/urls/`), but this can be configured with the `api_root` config option.

- `/urls/` (all methods require authentication with a valid access token)
- `/auth/` (some methods require authentication with a valid access token)

Where stated, endpoints will require an access token in the `Authorization` header, e.g. `Authorization: "YOUR_ACCESS_TOKEN"`.

**Note:** if the Linkener instance has the `auth_enabled` config option set to _false_, an access token is **not** required.

## `urls` endpoints

### `GET /urls/`

_Get all Short URLs._ **Access token required.**

Request: empty body

Response: an array of objects representing each short URL, e.g:

```json
[
  {
    "slug": "blog",
    "url": "https://blog.sjain.dev/mlh-fellowship/",
    "date_created": "2020-09-15T17:21:21.7320076+01:00",
    "allowed_visits": 50,
    "visits": [
        {"referer": ""}
    ],
    "password": ""
  },
  ...
]
```

### `POST /urls/`

_Create a new Short URL._ **Access token required.**

Request: JSON object with fields: `url` (required), `allowed_visits` (optional), `password` (optional) and **one of** either `slug` (a custom slug) or `slug_length` (the length of a random slug to generate). e.g:

```json
{
    "slug": "blog",
    "url": "https://blog.sjain.dev/mlh-fellowship/",
    "allowed_visits": 50,
    "password": ""
},
```

Response: the new Short URL record, e.g:

```json
{
    "slug": "blog",
    "url": "https://blog.sjain.dev/mlh-fellowship/",
    "date_created": "2020-09-15T17:21:21.7320076+01:00",
    "allowed_visits": 50,
    "visits": [
        {"referer": ""}
    ],
    "password": ""
},
```

### `GET /urls/{slug}/`

_Get a specific short URL._ **Access token required.**

Request: empty body

Response: a JSON object representing a single short URL, e.g:

```json
{
    "slug": "blog",
    "url": "https://blog.sjain.dev/mlh-fellowship/",
    "date_created": "2020-09-15T17:21:21.7320076+01:00",
    "allowed_visits": 50,
    "visits": [
        {"referer": ""}
    ],
    "password": ""
},
```

### `DELETE /urls/{slug}/`

_Delete a specific short URL._ **Access token required.**

Request: empty body

Response: `plain/text` body; status 200 on success

### `PUT /urls/{slug}/`

_Edit a specific short URL._ **Access token required.**

Request: JSON object with fields: `url` (required), `allowed_visits` (required), `password` (optional, absence means no change, `""` means no password). e.g:

```json
{
    "url": "https://blog.sjain.dev/mlh-fellowship/",
    "allowed_visits": 50,
    "password": "YOUR_PASSWORD"
},
```

Response: `plain/text` body; status 200 on success

## `auth` endpoints

### `/users`

_Create a new user._ **Note: this endpoint is disabled if the Linkener instance has `registration_enabled=false`**.

Request: JSON object with `username` and `password` keys.

Response: `plain/text` body; status 200 on success

### `/new_token`

_Create a new access token for the given user credentials._

Request: JSON object with `username` and `password` keys, representing the credentials for the user for whom the access token should be generated.

Response: `plain/text` body; status 200 and access token in body on success

### `/revoke_token`

_Revokes the given access token._

Request: JSON object with `access_token` key.

Response: `plain/text` body; status 200 on success

### `/users/{username}`

_Edit the authorized user._ **Access token required belonging to user to be edited.**

Request: JSON object with `password` field representing new password for the authorized user.

Response: `plain/text` body; status 200 for success
