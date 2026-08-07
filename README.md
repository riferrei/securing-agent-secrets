# A vault as the source of truth

> Resolving the credential just in time from 1Password, never on disk and never in the model's context.

Part of **[What Your Agent Doesn't Know Can't Hurt You][series]**, a hands-on
series on keeping secrets out of an AI agent's reach.

---

Same app as the [previous step][prev], with one thing changed: the Valkey
connection is gone from `.env` and out of the repo. It lives in a 1Password vault
now, and the app resolves it just in time at startup through the 1Password SDK,
authenticated as a service account. Nothing sensitive touches disk or version
control.

`cat .env` tells the story. In the previous step these were literals, the
password among them in plaintext. Here every one is a *reference*, and the whole
connection, host and port included, is the vault's to hand out:

```ini
VALKEY_HOST=op://Agent Prod/valkey-prod/host
VALKEY_PORT=op://Agent Prod/valkey-prod/port
VALKEY_USER=op://Agent Prod/valkey-prod/username
VALKEY_PASSWORD=op://Agent Prod/valkey-prod/password
```

The only secret in play is the service account token; it lives in your shell as
`OP_SERVICE_ACCOUNT_TOKEN` and is never written anywhere.

## How it works

![Architecture](docs/architecture.png)

At startup the app asks 1Password for the connection, builds the Valkey client,
and never logs or returns the resolved value.

## Prerequisites

- **Docker**
- The **[1Password CLI][op-cli]** (`op`)
- A **1Password account** on a plan with service accounts (Business or Teams)

## 1Password setup

The repo references secrets by path, so set up a matching vault and item:

1. A **vault**. This walkthrough uses one named `Agent Prod`.
2. A **Server item** in it named `valkey-prod`, with these fields:

   | Field | Value |
   |-------|-------|
   | `host` | `valkey-prod` |
   | `port` | `6379` |
   | `username` | `default` |
   | `password` | any password you choose |

3. A **service account** with read access to that vault. Export its token:

   ```bash
   export OP_SERVICE_ACCOUNT_TOKEN=ops_...
   ```

Named your vault or item differently? Update the `op://` paths in `.env` and
`op.env` to match.

## Run it

```bash
op run --env-file=op.env -- docker compose up --build -d
```

`op.env` holds a *reference*, not a secret. `op run` resolves it in memory and
hands the value to Valkey as its password, never writing it to disk. The service
account token passes through to the app, which resolves the same connection
through the SDK. The app seeds the sample customers on first start; open
**http://localhost:8080**. Stop with `docker compose down`.

## Prove the secret is gone

```bash
cat .env                        # only op:// references
git grep -i 'valkey.*password'  # nothing sensitive is tracked
```

A secret scanner finds nothing to flag, and that's *deterministic*, not a matter
of discipline.

## When it leaks anyway

The point of a vault isn't only that the secret is hidden; it's that the secret
is *disposable*. The service account token that authenticates all of this is
short-lived and centrally controlled. In the 1Password console it carries an
expiry, and two buttons: **Rotate Token** issues a new one and invalidates the
old, and **Revoke Token** kills access outright. Neither touches this repo or a
redeploy: the app resolves whatever the vault hands it at the next startup.

That is the difference from the previous step. There, a leaked credential meant
rewriting git history and rotating the database password by hand, with no way to
revoke the exposure in the meantime. Here, a suspected leak is a single click,
and rotation on a schedule is the default rather than an emergency.

## What this doesn't solve

The app's *access* is unchanged. It still connects to Valkey with full reach: it
can read every field of every customer, PII included. Moving the secret into a
vault does nothing about an over-broad identity. The next step scopes it.

---

**← Previous: [Environment variables as the source of truth][prev] · Next: [Controlling the blast radius][next] →**

[series]: https://github.com/riferrei/securing-agent-secrets-1password
[prev]: https://github.com/riferrei/securing-agent-secrets-1password/tree/env-vars-as-source-truth
[next]: https://github.com/riferrei/securing-agent-secrets-1password/tree/controlling-blast-radius
[op-cli]: https://developer.1password.com/docs/cli/
