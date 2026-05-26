#!/usr/bin/env python3
"""Patch discord.py with persistent dedup + voice buffer clear post-TTS."""

path = '/home/hermes-admin/.hermes/hermes-agent/gateway/platforms/discord.py'
content = open(path).read()
original_len = len(content)

# ── Change 1: add _dedup_persist_path after _dedup init ──────────────────
old1 = '        self._dedup = MessageDeduplicator()\n'
new1 = ('        self._dedup = MessageDeduplicator()\n'
        '        self._dedup_persist_path = Path.home() / ".hermes" / "discord_dedup_cache.json"\n')
assert old1 in content, "FAIL: dedup init not found"
content = content.replace(old1, new1, 1)
print("✓ Change 1: _dedup_persist_path added")

# ── Change 2: persist dedup entry in on_message ───────────────────────────
old2 = ('            @self._client.event\n'
        '            async def on_message(message: DiscordMessage):\n'
        '                # Dedup: Discord RESUME replays events after reconnects (#4777)\n'
        '                if adapter_self._dedup.is_duplicate(str(message.id)):\n'
        '                    return\n')
new2 = ('            @self._client.event\n'
        '            async def on_message(message: DiscordMessage):\n'
        '                # Dedup: Discord RESUME replays events after reconnects (#4777)\n'
        '                if adapter_self._dedup.is_duplicate(str(message.id)):\n'
        '                    return\n'
        '                # Persist so dedup survives gateway restarts (RESUME replay fix)\n'
        '                asyncio.create_task(asyncio.to_thread(\n'
        '                    adapter_self._persist_dedup_entry, str(message.id)\n'
        '                ))\n')
assert old2 in content, "FAIL: on_message dedup check not found"
content = content.replace(old2, new2, 1)
print("✓ Change 2: on_message dedup persist added")

# ── Change 3: load persistent dedup after bot ready ───────────────────────
old3 = ('            self._running = True\n'
        '            return True\n'
        '\n'
        '        except asyncio.TimeoutError:\n'
        '            logger.error("[%s] Timeout waiting for connection to Discord", self.name, exc_info=True)\n')
new3 = ('            self._running = True\n'
        "            # Load persisted dedup cache so we don't replay events after restart\n"
        '            await self._load_persistent_dedup()\n'
        '            return True\n'
        '\n'
        '        except asyncio.TimeoutError:\n'
        '            logger.error("[%s] Timeout waiting for connection to Discord", self.name, exc_info=True)\n')
assert old3 in content, "FAIL: running=True/TimeoutError block not found"
content = content.replace(old3, new3, 1)
print("✓ Change 3: _load_persistent_dedup() call added in connect()")

# ── Change 4a: add _barged_in flag before while loop ─────────────────────
old4a = '            while not done.is_set():\n'
new4a = '            _barged_in = False\n            while not done.is_set():\n'
assert old4a in content, "FAIL: while not done.is_set() not found"
content = content.replace(old4a, new4a, 1)
print("✓ Change 4a: _barged_in flag added")

# ── Change 4b: set _barged_in on barge-in + clear buffers after TTS ───────
barge_marker = '                        logger.info("Barge-in detected (frame-level) \u2014 stopping TTS playback")\n                        vc.stop()\n                        break\n'
new_barge = '                        logger.info("Barge-in detected (frame-level) \u2014 stopping TTS playback")\n                        vc.stop()\n                        _barged_in = True\n                        break\n'
assert barge_marker in content, "FAIL: barge-in vc.stop/break not found"
content = content.replace(barge_marker, new_barge, 1)

# Add buffer clear + reset after the while loop
old4b = ('\n            self._reset_voice_timeout(guild_id)\n'
         '            return True\n'
         '        finally:\n'
         '            pass  # Receiver stays active')
new4b = ('\n'
         '            # If TTS played to completion (no barge-in), clear any stale audio\n'
         '            # that accumulated in the receiver buffer during processing. Prevents\n'
         '            # phantom utterances being fed to STT after Hannah finishes speaking.\n'
         '            if not _barged_in and receiver:\n'
         '                with receiver._lock:\n'
         '                    receiver._buffers.clear()\n'
         '                    receiver._last_packet_time.clear()\n'
         '\n'
         '            self._reset_voice_timeout(guild_id)\n'
         '            return True\n'
         '        finally:\n'
         '            pass  # Receiver stays active')
assert old4b in content, "FAIL: reset_voice_timeout/finally block not found"
content = content.replace(old4b, new4b, 1)
print("✓ Change 4b: voice buffer clear after TTS added")

# ── Change 5: insert new methods after _voice_timeout_handler ─────────────
insert_marker = '    async def _voice_timeout_handler(self, guild_id: int) -> None:\n'
idx = content.find(insert_marker)
assert idx != -1, "FAIL: _voice_timeout_handler not found"
method_end = content.find('\n    async def ', idx + len(insert_marker))
assert method_end != -1, "FAIL: next method after _voice_timeout_handler not found"

new_methods = r"""
    async def _load_persistent_dedup(self) -> None:
        """Load persisted Discord message IDs into the dedup cache.

        Survives gateway restarts so Discord RESUME replays are caught even
        when the in-memory dedup was lost. Entries older than 24 h are dropped.
        """
        try:
            if not self._dedup_persist_path.exists():
                return
            import time as _time
            now = _time.time()
            cutoff = now - 86400  # 24 hours
            data = json.loads(self._dedup_persist_path.read_text())
            loaded = 0
            for msg_id, ts in data.items():
                if ts > cutoff:
                    self._dedup._seen[msg_id] = ts
                    loaded += 1
            if loaded:
                logger.info("[%s] Loaded %d message IDs from persistent dedup cache", self.name, loaded)
            # Compact: drop stale entries so the file doesn't grow unboundedly
            recent = {k: v for k, v in data.items() if v > cutoff}
            if len(recent) != len(data):
                self._dedup_persist_path.write_text(json.dumps(recent))
        except Exception as e:
            logger.debug("[%s] Failed to load persistent dedup cache: %s", self.name, e)

    def _persist_dedup_entry(self, msg_id: str) -> None:
        """Append msg_id to the on-disk dedup cache (runs in a thread pool)."""
        try:
            import time as _time
            now = _time.time()
            if self._dedup_persist_path.exists():
                data = json.loads(self._dedup_persist_path.read_text())
            else:
                data = {}
            data[msg_id] = now
            # Atomic write via temp file
            tmp = self._dedup_persist_path.with_suffix(".tmp")
            tmp.write_text(json.dumps(data))
            tmp.replace(self._dedup_persist_path)
        except Exception as e:
            logger.debug("[%s] Failed to persist dedup entry %s: %s", self.name, msg_id, e)

"""

content = content[:method_end] + new_methods + content[method_end:]
print("✓ Change 5: _load_persistent_dedup / _persist_dedup_entry methods added")

open(path, 'w').write(content)
print(f"\ndiscord.py patched OK ({original_len} -> {len(content)} bytes, +{len(content)-original_len})")
