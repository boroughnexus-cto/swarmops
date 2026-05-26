#!/usr/bin/env python3
"""Patch run.py with Haiku voice ack fast-path."""

path = '/home/hermes-admin/.hermes/hermes-agent/gateway/run.py'
content = open(path).read()
original_len = len(content)

# ── Change 1: insert _generate_haiku_voice_ack before _handle_voice_channel_input ──
insert_before = '    async def _handle_voice_channel_input(\n        self, guild_id: int, user_id: int, transcript: str\n    ):'
assert insert_before in content, "FAIL: _handle_voice_channel_input not found"

haiku_fn = r"""
    async def _generate_haiku_voice_ack(
        self, transcript: str, adapter, guild_id: int
    ) -> None:
        """Fire a Haiku request in parallel with the main agent to generate a
        short, tailored spoken acknowledgment phrase.  The phrase is converted
        to TTS and played in the voice channel within ~1-2 s of the user
        finishing their utterance, filling the silence while Sonnet processes.

        Falls back silently on any error so the main voice pipeline is unaffected.
        """
        try:
            import aiohttp as _aiohttp
            import json as _json

            # Resolve SwarmOps base URL from config (same proxy the main model uses)
            try:
                cfg = _load_gateway_config()
                custom_providers = cfg.get("custom_providers") or []
                base_url = next(
                    (
                        cp.get("base_url", "").rstrip("/")
                        for cp in custom_providers
                        if cp.get("base_url", "").strip()
                    ),
                    "http://nuc-ubuntu-dev.gate-hexatonic.ts.net:8080/v1",
                )
            except Exception:
                base_url = "http://nuc-ubuntu-dev.gate-hexatonic.ts.net:8080/v1"

            payload = {
                "model": "claude-haiku-4-5-20251001",
                "max_tokens": 40,
                "messages": [{"role": "user", "content": transcript[:300]}],
                "system": (
                    "You are Hannah Alexis, a warm and direct British EA. "
                    "The user just spoke a voice request. "
                    "Reply with ONLY a brief spoken acknowledgment phrase (5-15 words). "
                    "Make it natural and specific to their request — not generic filler. "
                    "Good examples: 'Sure, let me check the weather for you', "
                    "'On it, I\\'ll look that up right now', "
                    "'Great, give me just a moment to find that'. "
                    "For greetings or simple chitchat respond naturally and briefly. "
                    "Output the phrase ONLY — no quotes, no explanation, no punctuation at the end unless natural."
                ),
            }

            async with _aiohttp.ClientSession() as session:
                async with session.post(
                    f"{base_url}/chat/completions",
                    json=payload,
                    timeout=_aiohttp.ClientTimeout(total=6.0),
                ) as resp:
                    if resp.status != 200:
                        logger.debug(
                            "Haiku voice ack: upstream returned %d", resp.status
                        )
                        return
                    data = await resp.json()
                    ack_text = (
                        data.get("choices", [{}])[0]
                        .get("message", {})
                        .get("content", "")
                        .strip()
                        .strip("\"'")
                    )
                    if not ack_text:
                        return

            logger.info("Haiku voice ack: %r", ack_text)

            # Generate TTS for the acknowledgment phrase
            from tools.tts_tool import text_to_speech_tool, check_tts_requirements
            if not check_tts_requirements():
                return
            tts_result = await asyncio.to_thread(text_to_speech_tool, text=ack_text)
            tts_data = _json.loads(tts_result)
            tts_path = tts_data.get("file_path")
            if not tts_path or not Path(tts_path).exists():
                return

            try:
                await adapter.play_in_voice_channel(guild_id, tts_path)
            finally:
                try:
                    os.remove(tts_path)
                except OSError:
                    pass

        except asyncio.CancelledError:
            pass
        except Exception as e:
            logger.debug("Haiku voice ack failed: %s", e)

"""

idx = content.find(insert_before)
content = content[:idx] + haiku_fn + content[idx:]
print("✓ Change 1: _generate_haiku_voice_ack method inserted")

# ── Change 2: replace ack clip block in _handle_voice_channel_input ───────
# The current ack clip code:
old_ack = ('        # Play immediate ack clip for snappy voice UX\n'
           '        try:\n'
           '            import glob as _glob, random as _random\n'
           '            ack_dir = os.path.join(tempfile.gettempdir(), \'hermes_voice\', \'ack\')\n'
           '            ack_clips = _glob.glob(os.path.join(ack_dir, \'ack_*.mp3\'))\n'
           '            if ack_clips:\n'
           '                await adapter.play_in_voice_channel(guild_id, _random.choice(ack_clips))\n'
           '        except Exception as e:\n'
           '            logger.debug(\'Voice ack clip failed: %s\', e)\n'
           '\n'
           '        await adapter.handle_message(event)\n')

new_ack = ('        # Parallel Haiku fast-path: generates a short tailored ack phrase while\n'
           '        # the main Sonnet agent runs.  Fills the silence gap between the user\n'
           '        # finishing speaking and the full TTS response arriving (~5-20 s).\n'
           '        # Falls back to static ack clips if Haiku returns too slowly.\n'
           '        _haiku_ack_task = asyncio.create_task(\n'
           '            self._generate_haiku_voice_ack(transcript, adapter, guild_id)\n'
           '        )\n'
           '\n'
           '        # Start main agent (spawns background task, returns immediately)\n'
           '        await adapter.handle_message(event)\n'
           '\n'
           '        # Await Haiku ack — usually completes in 1-2 s (API + TTS + playback)\n'
           '        try:\n'
           '            await asyncio.wait_for(_haiku_ack_task, timeout=12.0)\n'
           '        except asyncio.TimeoutError:\n'
           '            logger.debug("Haiku voice ack timed out, falling back to ack clip")\n'
           '            _haiku_ack_task.cancel()\n'
           '            # Fallback: static ack clip\n'
           '            try:\n'
           '                import glob as _glob, random as _random\n'
           '                ack_dir = os.path.join(tempfile.gettempdir(), \'hermes_voice\', \'ack\')\n'
           '                ack_clips = _glob.glob(os.path.join(ack_dir, \'ack_*.mp3\'))\n'
           '                if ack_clips:\n'
           '                    await adapter.play_in_voice_channel(guild_id, _random.choice(ack_clips))\n'
           '            except Exception:\n'
           '                pass\n'
           '        except Exception as e:\n'
           '            logger.debug("Haiku voice ack error: %s", e)\n'
           '            _haiku_ack_task.cancel()\n')

assert old_ack in content, "FAIL: ack clip block not found in _handle_voice_channel_input"
content = content.replace(old_ack, new_ack, 1)
print("✓ Change 2: ack clip replaced with Haiku fast-path")

open(path, 'w').write(content)
print(f"\nrun.py patched OK ({original_len} -> {len(content)} bytes, +{len(content)-original_len})")
