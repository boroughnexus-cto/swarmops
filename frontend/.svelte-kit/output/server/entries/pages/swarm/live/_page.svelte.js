import { N as attr_style, G as attr, F as attr_class, J as escape_html, I as stringify, D as pop, A as push, E as ensure_array_like, O as head } from "../../../../chunks/index2.js";
import { o as onDestroy } from "../../../../chunks/index-server.js";
function IsoWorker($$payload, $$props) {
  push();
  let { agent, tasks = [], col, row } = $$props;
  const TW = 80;
  const TH = 40;
  function tileToScreen(c, r) {
    return { x: (c - r) * (TW / 2), y: (c + r) * (TH / 2) };
  }
  const roleConfig = {
    orchestrator: {
      color: "#ec4899",
      monitorColor: "#ec4899",
      label: "Orchestrator"
    },
    "senior-dev": {
      color: "#14b8a6",
      monitorColor: "#14b8a6",
      label: "Senior Dev"
    },
    "qa-agent": { color: "#eab308", monitorColor: "#eab308", label: "QA" },
    "devops-agent": { color: "#3b82f6", monitorColor: "#3b82f6", label: "DevOps" },
    researcher: {
      color: "#a855f7",
      monitorColor: "#a855f7",
      label: "Researcher"
    },
    worker: { color: "#64748b", monitorColor: "#64748b", label: "Worker" }
  };
  const roleEmoji = {
    orchestrator: "🧠",
    "senior-dev": "🧑‍💻",
    "qa-agent": "🔬",
    "devops-agent": "⚙️",
    researcher: "📚",
    worker: "👷"
  };
  const statusAnim = {
    coding: "anim-typing",
    testing: "anim-typing",
    thinking: "anim-think",
    waiting: "anim-idle",
    stuck: "anim-panic",
    done: "anim-idle",
    idle: "anim-idle"
  };
  const statusGlow = {
    coding: "rgba(74,222,128,0.5)",
    testing: "rgba(234,179,8,0.5)",
    thinking: "rgba(96,165,250,0.5)",
    waiting: "rgba(249,115,22,0.4)",
    stuck: "rgba(239,68,68,0.7)",
    done: "rgba(74,222,128,0.3)",
    idle: "rgba(100,116,139,0.2)"
  };
  let rc = roleConfig[agent.role] ?? roleConfig.worker;
  let emoji = roleEmoji[agent.role] ?? "👷";
  let anim = statusAnim[agent.status] ?? "anim-idle";
  let glow = statusGlow[agent.status] ?? "rgba(100,116,139,0.2)";
  let isLive = !!agent.tmux_session;
  let currentTask = tasks.find((t) => t.id === agent.current_task_id) ?? null;
  let terminalHref = isLive ? `/terminal/${agent.tmux_session}` : null;
  let pos = tileToScreen(col, row);
  let zIdx = row * 100 + col;
  function firstName(name) {
    return name.split(" ")[0];
  }
  let bubble = agent.status === "thinking" ? "💭" : agent.status === "stuck" ? "!!" : agent.status === "coding" ? "</>" : agent.status === "testing" ? "🧪" : agent.status === "done" ? "✓" : null;
  $$payload.out.push(`<div class="iso-tile svelte-19syhuc"${attr_style(` left: ${stringify(pos.x)}px; top: ${stringify(pos.y)}px; z-index: ${stringify(zIdx)}; `)}><svg class="desk-svg svelte-19syhuc" width="88" height="64" viewBox="0 0 88 64" fill="none"><polygon points="44,4 84,24 44,44 4,24"${attr("fill", rc.color)} fill-opacity="0.18"${attr("stroke", rc.color)} stroke-opacity="0.5" stroke-width="1" class="svelte-19syhuc"></polygon><polygon points="4,24 44,44 44,54 4,34" fill="rgba(0,0,0,0.45)"${attr("stroke", rc.color)} stroke-opacity="0.3" stroke-width="0.5" class="svelte-19syhuc"></polygon><polygon points="84,24 44,44 44,54 84,34" fill="rgba(0,0,0,0.3)"${attr("stroke", rc.color)} stroke-opacity="0.3" stroke-width="0.5" class="svelte-19syhuc"></polygon><rect x="32" y="10" width="26" height="18" rx="2"${attr("fill", rc.monitorColor)} fill-opacity="0.12"${attr("stroke", rc.monitorColor)} stroke-opacity="0.7" stroke-width="1.2" transform="skewX(-28) translate(12,0)" class="svelte-19syhuc"></rect><rect x="34" y="12" width="22" height="14" rx="1"${attr("fill", rc.monitorColor)} fill-opacity="0.25" transform="skewX(-28) translate(12,0)" class="svelte-19syhuc"></rect><rect x="43" y="28" width="3" height="6"${attr("fill", rc.color)} fill-opacity="0.4" transform="skewX(-28) translate(12,0)" class="svelte-19syhuc"></rect></svg> <a${attr("href", terminalHref ?? void 0)}${attr_class(`worker ${stringify(anim)} ${stringify(isLive ? "" : "worker-offline")}`, "svelte-19syhuc")}${attr_style(`--glow: ${stringify(glow)}; --rc: ${stringify(rc.color)}`)}${attr("title", `${stringify(agent.name)} (${stringify(rc.label)}) — ${stringify(agent.status)}${stringify(currentTask ? " · " + currentTask.title : "")}`)}><span class="worker-emoji svelte-19syhuc">${escape_html(emoji)}</span></a> `);
  if (bubble && isLive) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div${attr_class(`bubble ${stringify(agent.status === "stuck" ? "bubble-panic" : "")}`, "svelte-19syhuc")}>${escape_html(bubble)}</div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if ((agent.status === "coding" || agent.status === "testing") && isLive) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="monitor-glow svelte-19syhuc"${attr_style(`--glow: ${stringify(glow)}`)}></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> <div class="nametag svelte-19syhuc"${attr_style(`--rc: ${stringify(rc.color)}`)}><span class="nametag-name svelte-19syhuc">${escape_html(firstName(agent.name))}</span> `);
  if (currentTask) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="nametag-task svelte-19syhuc">${escape_html(currentTask.title)}</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div></div>`);
  pop();
}
function IsoRoom($$payload, $$props) {
  push();
  let { sessionId, sessionName, agents, tasks, color } = $$props;
  const TW = 80;
  const TH = 40;
  const COLS = 4;
  let agentTiles = agents.map((a, i) => ({ agent: a, col: i % COLS, row: Math.floor(i / COLS) }));
  let rows = Math.max(1, Math.ceil(agents.length / COLS));
  let canvasW = (COLS + rows) * TW + 40;
  let canvasH = (COLS + rows) * TH + 140;
  let floorTiles = Array.from({ length: rows }, (_, r) => Array.from({ length: COLS }, (_2, c) => ({ col: c, row: r }))).flat();
  function tileToScreen(c, r) {
    return {
      x: (c - r) * (TW / 2) + canvasW / 2 - TW / 2,
      y: (c + r) * (TH / 2) + 60
    };
  }
  let hasStuck = agents.some((a) => a.status === "stuck");
  let liveCount = agents.filter((a) => !!a.tmux_session).length;
  const each_array = ensure_array_like(floorTiles);
  const each_array_1 = ensure_array_like(agentTiles);
  $$payload.out.push(`<section${attr_class(`iso-room ${stringify(hasStuck ? "room-alert" : "")}`, "svelte-1x3mbqs")}${attr_style(`--rc: ${stringify(color)}`)}><div class="room-header svelte-1x3mbqs"><span class="room-name svelte-1x3mbqs">${escape_html(sessionName)}</span> <div class="room-pills svelte-1x3mbqs">`);
  if (liveCount > 0) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="pill pill-live svelte-1x3mbqs"><span class="live-dot svelte-1x3mbqs"></span>${escape_html(liveCount)} live</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if (hasStuck) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="pill pill-stuck svelte-1x3mbqs">⚠ stuck</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if (agents.length === 0) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="pill pill-empty svelte-1x3mbqs">empty</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div></div> <div class="iso-canvas-wrap svelte-1x3mbqs"><div class="iso-canvas svelte-1x3mbqs"${attr_style(`width: ${stringify(canvasW)}px; height: ${stringify(canvasH)}px`)}><!--[-->`);
  for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
    let tile = each_array[$$index];
    const pos = tileToScreen(tile.col, tile.row);
    const zIdx = tile.row * 100 + tile.col;
    $$payload.out.push(`<svg class="floor-tile svelte-1x3mbqs"${attr("width", TW + 2)}${attr("height", TH + 2)}${attr("viewBox", `-1 -1 ${stringify(TW + 2)} ${stringify(TH + 2)}`)}${attr_style(`left: ${stringify(pos.x - TW / 2)}px; top: ${stringify(pos.y - TH / 2)}px; z-index: ${stringify(zIdx - 1)}`)}><polygon${attr("points", `${stringify(TW / 2)},1 ${stringify(TW)},${stringify(TH / 2)} ${stringify(TW / 2)},${stringify(TH)} 0,${stringify(TH / 2)}`)} fill="rgba(255,255,255,0.02)"${attr("stroke", color)} stroke-opacity="0.12" stroke-width="0.5" class="svelte-1x3mbqs"></polygon></svg>`);
  }
  $$payload.out.push(`<!--]--> <!--[-->`);
  for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
    let at = each_array_1[$$index_1];
    const pos = tileToScreen(at.col, at.row);
    $$payload.out.push(`<div class="worker-wrap svelte-1x3mbqs"${attr_style(`left: ${stringify(pos.x)}px; top: ${stringify(pos.y)}px; z-index: ${stringify(at.row * 100 + at.col + 10)}`)}>`);
    IsoWorker($$payload, { agent: at.agent, tasks, col: 0, row: 0 });
    $$payload.out.push(`<!----></div>`);
  }
  $$payload.out.push(`<!--]--> `);
  if (agents.length === 0) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="empty-msg svelte-1x3mbqs"${attr_style(`top: ${stringify(canvasH / 2 - 20)}px`)}>No agents — <a${attr("href", `/swarm/${stringify(sessionId)}`)} class="svelte-1x3mbqs">add one</a></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div></div></section>`);
  pop();
}
function _page($$payload, $$props) {
  push();
  let sessionData = {};
  let sessionColors = {};
  const sockets = /* @__PURE__ */ new Map();
  let allAgents = Object.values(sessionData).flatMap((s) => s.agents);
  let liveCount = allAgents.filter((a) => !!a.tmux_session).length;
  let codingCount = allAgents.filter((a) => a.status === "coding").length;
  let thinkingCount = allAgents.filter((a) => a.status === "thinking").length;
  let stuckCount = allAgents.filter((a) => a.status === "stuck").length;
  let waitingCount = allAgents.filter((a) => a.status === "waiting").length;
  let sessions = Object.values(sessionData);
  let hasStuck = stuckCount > 0;
  onDestroy(() => {
    for (const ws of sockets.values()) ws.close();
  });
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Swarm Live — SwarmOps</title>`;
  });
  $$payload.out.push(`<div class="floor-wrap svelte-1etkbwf"><header${attr_class(`hud ${stringify(hasStuck ? "hud-alert" : "")}`, "svelte-1etkbwf")}><div class="hud-left svelte-1etkbwf"><span class="hud-title svelte-1etkbwf">RC SWARM LIVE</span> <span class="hud-sep svelte-1etkbwf">│</span> <a href="/swarm" class="hud-back svelte-1etkbwf">← Sessions</a></div> <div class="hud-stats svelte-1etkbwf">`);
  if (liveCount > 0) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="stat stat-live svelte-1etkbwf"><span class="stat-dot live-dot svelte-1etkbwf"></span> ${escape_html(liveCount)} live</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if (codingCount > 0) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="stat stat-coding svelte-1etkbwf">⚡ ${escape_html(codingCount)} coding</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if (thinkingCount > 0) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="stat stat-thinking svelte-1etkbwf">💭 ${escape_html(thinkingCount)} thinking</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if (waitingCount > 0) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="stat stat-waiting svelte-1etkbwf">⏸ ${escape_html(waitingCount)} waiting</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if (stuckCount > 0) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="stat stat-stuck svelte-1etkbwf">⚠ ${escape_html(stuckCount)} stuck</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if (liveCount === 0 && allAgents.length === 0) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="stat stat-idle svelte-1etkbwf">No active agents</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div></header> <main class="floor svelte-1etkbwf">`);
  if (sessions.length === 0) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="empty-floor svelte-1etkbwf"><p class="empty-title svelte-1etkbwf">No swarm sessions</p> <a href="/swarm" class="empty-link svelte-1etkbwf">Create one →</a></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
    const each_array = ensure_array_like(sessions);
    $$payload.out.push(`<!--[-->`);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let s = each_array[$$index];
      const color = sessionColors[s.session.id] ?? "#14b8a6";
      IsoRoom($$payload, {
        sessionId: s.session.id,
        sessionName: s.session.name,
        agents: s.agents,
        tasks: s.tasks,
        color
      });
    }
    $$payload.out.push(`<!--]-->`);
  }
  $$payload.out.push(`<!--]--></main></div>`);
  pop();
}
export {
  _page as default
};
