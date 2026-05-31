import { O as head, J as escape_html, D as pop, A as push } from "../../../chunks/index2.js";
import { B as Button } from "../../../chunks/Button.js";
function _page($$payload, $$props) {
  push();
  let executions = [];
  let showRecent = false;
  executions.filter((e) => e.status === "running");
  executions.filter((e) => e.status === "Waiting" || e.status === "waiting");
  let recent = executions.filter((e) => ["completed", "rejected", "failed"].includes(e.status)).slice(0, 5);
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Kanban - SwarmOps</title>`;
  });
  $$payload.out.push(`<div class="space-y-6"><div class="flex items-center justify-between mb-8"><div><h1 class="text-3xl font-bold text-vanna-navy font-serif">Agent Activity</h1> <p class="mt-2 text-slate-500">Live view of what each agent is working on</p></div> <div class="flex items-center gap-3">`);
  if (recent.length > 0) {
    $$payload.out.push("<!--[-->");
    Button($$payload, {
      variant: "ghost",
      size: "sm",
      onclick: () => showRecent = !showRecent,
      children: ($$payload2) => {
        $$payload2.out.push(`<!---->${escape_html(showRecent ? "Hide" : "Show")} Recent (${escape_html(recent.length)})`);
      }
    });
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> <div class="flex items-center gap-2 text-sm text-slate-500"><div class="w-2 h-2 bg-vanna-teal rounded-full animate-pulse"></div> Polling every 5s</div></div></div> `);
  {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="text-center py-12 text-slate-500">Loading agent activity...</div>`);
  }
  $$payload.out.push(`<!--]--></div>`);
  pop();
}
export {
  _page as default
};
