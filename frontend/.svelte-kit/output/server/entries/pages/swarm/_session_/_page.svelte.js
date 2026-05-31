import { K as store_get, O as head, M as unsubscribe_stores, D as pop, A as push, J as escape_html } from "../../../../chunks/index2.js";
import { p as page } from "../../../../chunks/stores.js";
import "@sveltejs/kit/internal";
import "../../../../chunks/exports.js";
import "clsx";
import "../../../../chunks/state.svelte.js";
function _page($$payload, $$props) {
  push();
  var $$store_subs;
  store_get($$store_subs ??= {}, "$page", page).params.session;
  let agents = [];
  agents.filter((a) => !!a.tmux_session);
  agents.filter((a) => !!a.repo_path && !a.tmux_session).length;
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>${escape_html("Swarm")} — SwarmOps</title>`;
  });
  {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="flex items-center justify-center py-20"><div class="animate-spin rounded-full h-8 w-8 border-b-2 border-vanna-teal"></div></div>`);
  }
  $$payload.out.push(`<!--]--> `);
  {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]-->`);
  if ($$store_subs) unsubscribe_stores($$store_subs);
  pop();
}
export {
  _page as default
};
