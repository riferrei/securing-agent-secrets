# Securing MCP servers

> Pointing an off-the-shelf MCP server at the same data, with 1Password securing *its* credential.

Part of **[What Your Agent Doesn't Know Can't Hurt You][series]**, a hands-on
series on keeping secrets out of an AI agent's reach.

---

The earlier steps hardened an agent *we* wrote: we owned the Redis call, split
the PII, scoped the credential. In practice you'll reach for **off-the-shelf MCP
servers** you didn't write, and you can't bolt your guardrails onto someone else's
server. What you *can* do is keep its secret out of plaintext and hand it only a
least-privilege credential. That's this step.

Unlike the others, this one has **no app of its own**. The agent is an MCP host
([Claude Desktop] here), and the tool layer is the official
[Redis MCP server][redis-mcp]. The database is the same Redis from
[the previous step][prev], reached with the same read-only, PII-blind `agent`
credential.

![Architecture](docs/architecture.png)

## What you need

1. **The Redis from [the previous step][prev]**, running. It provides the ACL and
   the seeded data. Check that branch out and bring it up (only Redis is
   required).
2. **A `redis-mcp` Server item** in the same `Agent Prod` vault. Everywhere else
   the client runs *inside* Compose and reaches Redis as `redis-prod`; here the
   MCP server runs on your host, spawned by Claude Desktop, so it reaches the same
   Redis over the published port. This item carries that host-side address, the
   *same* read-only `agent` credential as before:

   | Field | Value |
   |-------|-------|
   | `host` | `127.0.0.1` |
   | `port` | `6379` |
   | `username` | `agent` |
   | `password` | the same agent password from the previous step |
3. **The [1Password desktop app][op-desktop]** with CLI integration enabled
   (*Settings > Developer > Integrate with 1Password CLI*). This lets `op run`
   authenticate by biometric, so **no token lives in the config**.
4. **The Redis MCP server**, pinned to a compatible SDK (see [the gotcha](#the-gotcha)):
   ```bash
   uv tool install --with "mcp<2" redis-mcp-server
   ```

## The config (note what's *not* in it)

Add this to `~/Library/Application Support/Claude/claude_desktop_config.json`
(also saved in [`mcp/claude_desktop_config.json`](mcp/claude_desktop_config.json)):

```json
{
  "mcpServers": {
    "redis": {
      "command": "/opt/homebrew/bin/op",
      "args": ["run", "--", "/Users/riferrei/.local/bin/redis-mcp-server"],
      "env": {
        "REDIS_HOST": "op://Agent Prod/redis-mcp/host",
        "REDIS_PORT": "op://Agent Prod/redis-mcp/port",
        "REDIS_USERNAME": "op://Agent Prod/redis-mcp/username",
        "REDIS_PWD": "op://Agent Prod/redis-mcp/password"
      }
    }
  }
}
```

The command isn't the MCP server; it's `op run --`, which resolves the `op://`
references in memory and hands the real values to the server it spawns. The
config holds **only references, no connection details** — not just the password,
but the host and port too, matching how every earlier branch keeps the whole
Redis connection in the vault. The paths are absolute because Claude Desktop
doesn't inherit your shell `PATH`; set yours with `command -v op` and your `uv`
tools bin (typically `~/.local/bin/redis-mcp-server`).

Restart Claude Desktop and confirm the `redis` server shows **running** under
*Settings > Developer*.

## The gotcha

`redis-mcp-server` 0.5.0 imports `mcp.server.fastmcp`, which the `mcp` SDK
**removed in 2.0**. A plain `uv tool install` pulls `mcp` 2.0 and the server
crashes on startup:

```
ModuleNotFoundError: No module named 'mcp.server.fastmcp'
```

The `--with "mcp<2"` pin above avoids it.

## See it work

Talk to Claude; it picks the Redis MCP tools itself:

| You say | What happens |
| --- | --- |
| "Look up customer 1." | Business record returned ✅ |
| "What's their SSN?" | **Denied**: *No permissions to access a key* 🛑 |
| "Change their first name to HACKED." | **Denied**: *has no permissions to run the 'hset' command* 🛑 |

Even the denial masks the resolved credential (*User `<concealed by 1Password>`
has no permissions…*): `op run` scrubbing secrets out of the tool's own output.

## The point

Claude here has the **full, generic Redis toolbelt** (`hget`, `hset`, `delete`,
`scan`) on a server you didn't write and can't add guardrails to. It could ask
for anything. It still can't read one SSN or change one record, because the only
credential it was ever handed is the least-privilege, read-only one 1Password
injected just in time, never in plaintext.

And because that credential lives in the vault rather than the tool's config, it
stays disposable even here. If it leaks, you **rotate or revoke** it in 1Password
and the next `op run` picks up the change; the third-party server you can't edit
keeps working, now with a credential the old copy can no longer use. Control of
the secret never left the boundary 1Password owns.

**When you can't control the tool, control the credential.**

## What this doesn't solve

The agent trusts whatever MCP server it points at, and the secret didn't vanish;
it relocated to a boundary 1Password owns. Security is never finished.

---

**← Previous: [Controlling the blast radius][prev] · [Series overview][series]**

[series]: https://github.com/riferrei/securing-agent-secrets-1password
[prev]: https://github.com/riferrei/securing-agent-secrets-1password/tree/controlling-blast-radius
[Claude Desktop]: https://claude.ai/download
[redis-mcp]: https://github.com/redis/mcp-redis
[op-desktop]: https://1password.com/downloads
