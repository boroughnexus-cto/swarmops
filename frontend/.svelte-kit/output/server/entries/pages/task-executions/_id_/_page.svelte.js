import { K as store_get, O as head, M as unsubscribe_stores, D as pop, A as push, J as escape_html } from "../../../../chunks/index2.js";
import { p as page } from "../../../../chunks/stores.js";
import "@sveltejs/kit/internal";
import "../../../../chunks/exports.js";
import "clsx";
import "../../../../chunks/state.svelte.js";
import { B as Breadcrumb } from "../../../../chunks/Breadcrumb.js";
function _page($$payload, $$props) {
  push();
  var $$store_subs;
  let breadcrumbSegments, executionId;
  breadcrumbSegments = [
    { label: "", href: "/", icon: "banner" },
    { label: "Projects", href: "/projects" },
    {
      label: "Project",
      href: "/projects"
    },
    {
      label: "Task",
      href: "#"
    },
    {
      label: `Agent`,
      href: `/task-executions/${store_get($$store_subs ??= {}, "$page", page).params.id}`
    }
  ];
  executionId = store_get($$store_subs ??= {}, "$page", page).params.id;
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Task Execution ${escape_html(executionId)} - SwarmOps</title>`;
    $$payload2.out.push(`<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css"/>`);
  });
  $$payload.out.push(`<div class="min-h-screen"><div class="container mx-auto p-6">`);
  Breadcrumb($$payload, { segments: breadcrumbSegments });
  $$payload.out.push(`<!----> <div class="mb-6">`);
  {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="border-b border-slate-200 pb-6 mb-8"><div class="flex items-center gap-4"><div class="animate-spin rounded-full h-8 w-8 border-b-2 border-vanna-teal"></div> <div><h1 class="text-2xl font-bold text-vanna-navy font-serif">Loading Task Execution...</h1> <p class="text-slate-500">Fetching execution details</p></div></div></div>`);
  }
  $$payload.out.push(`<!--]--></div> `);
  {
    $$payload.out.push("<!--[!-->");
    {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]-->`);
  }
  $$payload.out.push(`<!--]--></div></div>`);
  if ($$store_subs) unsubscribe_stores($$store_subs);
  pop();
}
export {
  _page as default
};
