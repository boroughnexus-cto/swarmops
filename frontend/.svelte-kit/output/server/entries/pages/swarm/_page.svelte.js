import { O as head, J as escape_html, D as pop, A as push } from "../../../chunks/index2.js";
import { o as onDestroy } from "../../../chunks/index-server.js";
import "@sveltejs/kit/internal";
import "../../../chunks/exports.js";
import "clsx";
import "../../../chunks/state.svelte.js";
function _page($$payload, $$props) {
  push();
  let sessions = [];
  onDestroy(() => {
  });
  let liveCount = sessions.filter((s) => s.live_agents > 0).length;
  let stuckCount = sessions.reduce((n, s) => n + s.stuck_agents, 0);
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Swarm — SwarmOps</title>`;
  });
  $$payload.out.push(`<div class="max-w-5xl mx-auto"><div class="flex items-center justify-between mb-6"><div><h1 class="text-2xl font-bold text-vanna-navy">Swarm Orchestrator</h1> <p class="text-sm text-slate-500 mt-1">Coordinate multiple Claude Code agents across your projects</p></div> <div class="flex items-center gap-2"><a href="/swarm/live" class="flex items-center gap-1.5 px-3 py-2 rounded-xl border border-vanna-teal/30 text-vanna-teal text-sm font-medium hover:bg-vanna-teal/10 transition-colors"><span class="w-2 h-2 rounded-full bg-vanna-teal animate-pulse"></span> Live View</a> <button type="button" class="flex items-center gap-2 bg-vanna-teal text-white px-4 py-2 rounded-xl font-medium text-sm hover:bg-vanna-teal/90 transition-colors shadow-vanna-subtle"><svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg> New Swarm</button></div></div> `);
  if (sessions.length > 0) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="flex items-center gap-4 mb-6 text-sm"><span class="text-slate-500">${escape_html(sessions.length)} session${escape_html(sessions.length !== 1 ? "s" : "")}</span> `);
    if (liveCount > 0) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<span class="flex items-center gap-1.5 text-vanna-teal font-medium"><span class="w-2 h-2 rounded-full bg-vanna-teal animate-pulse"></span> ${escape_html(liveCount)} active</span>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--> `);
    if (stuckCount > 0) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<span class="flex items-center gap-1.5 text-red-500 font-medium"><span class="w-2 h-2 rounded-full bg-red-500"></span> ${escape_html(stuckCount)} stuck</span>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="flex items-center justify-center py-16"><div class="animate-spin rounded-full h-8 w-8 border-b-2 border-vanna-teal"></div></div>`);
  }
  $$payload.out.push(`<!--]--></div>`);
  pop();
}
export {
  _page as default
};
