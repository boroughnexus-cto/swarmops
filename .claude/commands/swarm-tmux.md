Show the current state of SwarmOps tmux sessions and the service health.

```bash
tmux ls 2>/dev/null || echo "(no tmux sessions)"
echo "---"
systemctl --user is-active swarmops 2>/dev/null && echo "Service: active" || echo "Service: inactive/not found"
echo "---"
curl -sf http://localhost:8080/api/dashboard/stats 2>/dev/null | python3 -m json.tool 2>/dev/null || echo "(backend not responding)"
```

Report: active tmux sessions (names + pane counts), service status, and dashboard stats if available.
