Build SwarmOps and restart the systemd user service on this machine.

Steps:
1. Run `make build` in `/home/sbarker/git-bnx/TKN/swarmops`
2. Copy the binary to the deployment path: `cp swarmops ~/swarmops/swarmops`
3. Run `make restart` (stops port 8080, restarts `swarmops.service`, health-checks)

```bash
cd /home/sbarker/git-bnx/TKN/swarmops && make build && cp swarmops ~/swarmops/swarmops && make restart
```

Report the git commit hash that was built and whether the health check passed.

**Warning:** Do not run this if the user has an active TUI session — it will disconnect them from the backend.
