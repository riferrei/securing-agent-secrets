# CLAUDE.md

Guidance for working in this branch.

## What this is

The `securing-mcp-servers` branch of *What Your Agent Doesn't Know Can't Hurt
You*. It is **different in kind from the other branches: there is no application
here.** The others build and harden a Go agent; this one shows the other path,
where the agent is an off-the-shelf MCP host and the tool layer is a server you
did not write.

The whole content is `README.md` and `mcp/claude_desktop_config.json`: how the
official Redis MCP server is secured with 1Password. Its Redis credential is the
same read-only, PII-blind `agent` credential from the `controlling-blast-radius`
branch, resolved by `op run` from an `op://` reference and never written in
plaintext.

## Do not

- Do not add application code here, or treat this as a runnable version of the
  agent. If you need the running app (and the Redis it talks to), check out
  `controlling-blast-radius`.
- Do not put a resolved secret in the config or anywhere on disk. The config
  holds `op://` references only.

## The series

Each step of the hardening journey is its own branch; `main` lists them. This is
the final one.
