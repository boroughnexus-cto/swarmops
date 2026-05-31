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
  let directoryId, breadcrumbSegments;
  directoryId = store_get($$store_subs ??= {}, "$page", page).params.directoryId;
  breadcrumbSegments = [
    { label: "", href: "/", icon: "banner" },
    { label: "Directories", href: "/directories" },
    { label: "Dev Server", href: `/dev-server/${directoryId}` }
  ];
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Dev Server: ${escape_html("Loading...")} - SwarmOps</title>`;
    $$payload2.out.push(`<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css"/>`);
  });
  $$payload.out.push(`<div class="min-h-screen bg-gray-900 text-white"><div class="container mx-auto p-6">`);
  Breadcrumb($$payload, { segments: breadcrumbSegments });
  $$payload.out.push(`<!----> <div class="mb-6">`);
  {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="flex items-center gap-4"><div class="animate-spin rounded-full h-8 w-8 border-b-2 border-green-400"></div> <div><h1 class="text-2xl font-bold text-green-400">Loading Dev Server...</h1> <p class="text-gray-300">Connecting to terminal</p></div></div>`);
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
