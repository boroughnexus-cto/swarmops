#!/usr/bin/env bash
#
# deploy.sh — the ONE sanctioned way to deploy SwarmOps to production.
#
# Invariant: production (~/swarmops) only ever runs a clean build of a
# CI-validated commit. There is no other sanctioned path — the GitHub Actions
# deploy job, `make deploy`, and `swarmops redeploy` all funnel through this
# script, so a build can never again be cut from a stale branch/worktree.
#
# Sequence: lock → (dirty guard) → fetch → reset --hard to target → clean →
# backup → atomic build → restart → health-check → consistent rollback on fail.
#
# Env:
#   SWARMOPS_DEPLOY_DIR   deploy checkout (default: $HOME/swarmops)
#   DEPLOY_SHA            exact commit to deploy (CI passes github.sha so the
#                         deployed artifact is exactly what the gate validated;
#                         unset → origin/main, the intended manual semantics)
#   GITHUB_ACTIONS / CI   when set, the dirty-tree guard is skipped
#   PORT                  health-check port (default: 8080)
# Args:
#   --force               skip the dirty-tree guard for manual runs
set -euo pipefail

DEPLOY_DIR="${SWARMOPS_DEPLOY_DIR:-$HOME/swarmops}"
PORT="${PORT:-8080}"
HEALTH_PATH="/api/dashboard/stats" # exercises the DB, not just the HTTP server
FORCE=0
[[ "${1:-}" == "--force" ]] && FORCE=1

log() { echo "deploy: $*" >&2; }
audit() { logger -t swarmops-deploy "$*" 2>/dev/null || true; }

# health_check returns 0 once the service answers HEALTH_PATH, else 1 after
# 10 attempts (1s apart, 5s per request).
health_check() {
	local i
	for ((i = 0; i < 10; i++)); do
		if curl -fsS -m 5 "http://localhost:${PORT}${HEALTH_PATH}" >/dev/null 2>&1; then
			return 0
		fi
		sleep 1
	done
	return 1
}

[[ -d "$DEPLOY_DIR" ]] || {
	log "deploy dir $DEPLOY_DIR does not exist"
	exit 1
}
cd "$DEPLOY_DIR"
git rev-parse --is-inside-work-tree >/dev/null 2>&1 || {
	log "$DEPLOY_DIR is not a git checkout"
	exit 1
}

# Serialize every caller (CI / make deploy / swarmops redeploy) through one lock.
exec 9>"$DEPLOY_DIR/.deploy.lock"
if ! flock -w 300 9; then
	log "another deploy holds the lock (waited 300s) — aborting"
	exit 1
fi

# Dirty-tree guard. CI resets unconditionally; interactive callers must be clean
# (or pass --force) so a deploy never silently discards local tracked edits.
if [[ -z "${GITHUB_ACTIONS:-}${CI:-}" && $FORCE -eq 0 ]]; then
	if ! git diff --quiet || ! git diff --cached --quiet; then
		log "$DEPLOY_DIR has uncommitted tracked changes — commit/stash or pass --force"
		exit 1
	fi
fi

OLD_SHA="$(git rev-parse HEAD 2>/dev/null || echo "")"
TARGET="${DEPLOY_SHA:-origin/main}"

log "fetching origin; deploying ${TARGET}"
git fetch --prune origin
git reset --hard "$TARGET"
git clean -fd # remove stale untracked files (keeps gitignored runtime state: *.db, binaries)
NEW_SHA="$(git rev-parse --short HEAD)"

# Back up last-known-good binaries (the currently-running pair, pre-build) so we
# can roll back. Track whether a real backup exists — never claim a rollback we
# can't perform (e.g. first-ever deploy).
HAVE_PREV=0
if [[ -f swarmops && -f quota-proxy ]]; then
	cp -f swarmops swarmops.prev
	cp -f quota-proxy quota-proxy.prev
	HAVE_PREV=1
fi

log "building $NEW_SHA"
make build

log "restarting service"
systemctl --user restart swarmops

if health_check; then
	audit "deployed $NEW_SHA"
	log "$NEW_SHA deployed and healthy"
	exit 0
fi

# --- Deploy unhealthy: roll back to a consistent previous state -------------
log "health check failed for $NEW_SHA"
if [[ $HAVE_PREV -ne 1 || -z "$OLD_SHA" ]]; then
	audit "DEPLOY FAILED $NEW_SHA — no previous build to roll back to; service may be down"
	log "no previous build to roll back to — manual intervention required"
	exit 1
fi

log "rolling back to $OLD_SHA"
audit "ROLLBACK starting: $NEW_SHA unhealthy, restoring $OLD_SHA"
git reset --hard "$OLD_SHA" # restore tracked files so sources match the restored binaries
git clean -fd
cp -f swarmops.prev swarmops
cp -f quota-proxy.prev quota-proxy
if ! systemctl --user restart swarmops; then
	audit "ROLLBACK restart FAILED after $NEW_SHA — service down"
	log "rollback restart failed — service down, manual intervention required"
	exit 1
fi
if health_check; then
	audit "ROLLBACK succeeded to $OLD_SHA (after failed $NEW_SHA)"
	log "rolled back to $OLD_SHA and healthy"
else
	audit "ROLLBACK to $OLD_SHA still UNHEALTHY (after failed $NEW_SHA)"
	log "rollback restored $OLD_SHA but service is still unhealthy"
fi
exit 1
