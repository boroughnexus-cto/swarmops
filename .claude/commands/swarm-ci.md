Run the full SwarmOps CI gate: `go vet` then `go test -race`.

```bash
cd /home/sbarker/git-bnx/TKN/swarmops && make ci
```

Report pass/fail and any test output. If there are failures, identify the failing test names and the first error message. Do not attempt fixes unless asked.
