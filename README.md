# What Your Agent Doesn't Know Can't Hurt You

> Keeping credentials, and the sensitive data behind them, out of an AI agent's reach.

An AI agent needs secrets to do real work, but its reasoning context is an
exfiltration-prone surface: anything the model can see, an attacker can try to
talk it into revealing. Guarding the secret more carefully *inside* that context
is the wrong instinct. The better approach is to keep it out of the context
entirely: resolve it just in time, scope the identity that fetches it, and shrink
what a compromised agent can actually do.

This repository walks that approach one step at a time, around a single small,
realistic app. Each step lives on its own branch and is a complete, runnable
version of the app, safer than the one before. Read them in order to follow the
reasoning, or check one out and run it yourself.

## The example app

A customer-support assistant. Ask it about a customer in natural language; it
looks them up in Redis and answers. The secret it depends on is the Redis
connection. The architecture never changes across branches, only how the secret
is handled:

```
UI  →  REST API  →  Agent (LLM + tools)  →  Redis
```

Each branch will improve security with small code changes to the application: the security moves
into the vault, usage of ACL, and data model changes. The `.env` goes from holding
literals to holding `op://` references, and that's most of it. Safer without
slowing the developer down is the whole point.

You will notice that the data model evolves along with the new security enhancements. Early branches
store everything in one `customer:NNNN` record, PII included, mirroring how teams actually start;
branch 3 splits PII into its own key so a scoped identity can be denied it. This approach is used
to demonstrate why security should not be a deployment concern, but rather a development concern.

The app runs locally with Docker and uses [Ollama] as LLM running locally, so the core  walkthrough
has no external LLM costs and nothing leaves your machine.

## Learning path

| # | Branch | What it covers |
|---|--------|----------------|
| 1 | [`env-vars-as-source-truth`][b1] | The credential in a `.env` file: the setup that *feels* safe, and why it isn't |
| 2 | [`vaults-as-source-truth`][b2] | Resolving it just in time from 1Password, never on disk, never in the model's context |
| 3 | [`controlling-blast-radius`][b3] | A read-only, PII-blind identity: even a hijacked agent can't read SSNs or write |
| 4 | [`securing-mcp-servers`][b4] | Pointing an off-the-shelf MCP server at the same data, with 1Password securing *its* credential |

Each branch's README goes deep, and each closes with a candid *"What this
doesn't solve"*, the residual weakness that motivates the next step, so the
progression reads as intentional rather than as drift. Here's the through-line.

### 1 · The credential in an environment variable

The Redis credential lives in a `.env` file. This isn't a strawman; it's what
most teams actually ship: the value is in an env var, `.env` is git-ignored, and
everything works. It feels safe. It isn't: the credential is plaintext on disk,
one stray `git add .` from landing in history, where a leaked secret is
effectively public and rotation is the only real fix.

### 2 · A vault as the source of truth

The credential moves into 1Password and is resolved just in time through the
1Password SDK, authenticated as a service account. It never touches disk: there
is nothing in `.env` to leak and nothing on the box to steal. It lives in memory
only for the instant it's used.

### 3 · Controlling the blast radius

Shrink what the agent can *do*. The customer record splits into two keys,
business fields and PII (email, phone, SSN, address), and the agent connects as a
read-only Redis ACL user scoped to the business keys only. It cannot read PII and
cannot write anything; Redis refuses, not the agent. Even a fully hijacked agent
comes up empty on SSNs.

### 4 · Securing an off-the-shelf MCP server

In practice you'll adopt MCP servers you didn't build and can't add guardrails
to. Here the agent is an MCP host (for example Claude Desktop) and the tool layer
is the official [Redis MCP server][redis-mcp]. 1Password secures *its* credential:
the Redis connection is an `op://` reference resolved by `op run` at spawn time
(no secret in the config), and it's the same least-privilege credential from step 3.
The host has the full generic Redis toolbelt and still can't read one SSN or
change one record.

## Getting started

Each branch has its own README with exact prerequisites and commands. Start at
the beginning:

```bash
git clone https://github.com/riferrei/securing-agent-secrets.git
cd securing-agent-secrets
git checkout env-vars-as-source-truth
```

Step 1 needs only **Docker**. Steps 2 through 4 add the **[1Password CLI][op-cli]**
and a 1Password account. Note that 1Password offers trial accounts for you to play
with the product before committing with a credit card.

[Ollama]: https://ollama.com
[redis-mcp]: https://github.com/redis/mcp-redis
[op-cli]: https://developer.1password.com/docs/cli/
[b1]: https://github.com/riferrei/securing-agent-secrets/tree/env-vars-as-source-truth
[b2]: https://github.com/riferrei/securing-agent-secrets/tree/vaults-as-source-truth
[b3]: https://github.com/riferrei/securing-agent-secrets/tree/controlling-blast-radius
[b4]: https://github.com/riferrei/securing-agent-secrets/tree/securing-mcp-servers
