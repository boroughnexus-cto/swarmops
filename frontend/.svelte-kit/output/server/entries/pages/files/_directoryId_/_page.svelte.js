import { K as store_get, O as head, M as unsubscribe_stores, D as pop, A as push, J as escape_html } from "../../../../chunks/index2.js";
import { p as page } from "../../../../chunks/stores.js";
import { B as Breadcrumb } from "../../../../chunks/Breadcrumb.js";
function _page($$payload, $$props) {
  push();
  var $$store_subs;
  let directoryId, breadcrumbSegments;
  directoryId = store_get($$store_subs ??= {}, "$page", page).params.directoryId;
  breadcrumbSegments = [
    { label: "", href: "/", icon: "banner" },
    { label: "Directories", href: "/directories" },
    { label: "Files", href: `/files/${directoryId}` }
  ];
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Files - ${escape_html("Loading...")}</title>`;
  });
  $$payload.out.push(`<div class="space-y-4">`);
  Breadcrumb($$payload, { segments: breadcrumbSegments });
  $$payload.out.push(`<!----> `);
  {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="flex items-center justify-center min-h-64"><div class="animate-spin rounded-full h-12 w-12 border-b-2 border-vanna-teal"></div></div>`);
  }
  $$payload.out.push(`<!--]--></div>`);
  if ($$store_subs) unsubscribe_stores($$store_subs);
  pop();
}
export {
  _page as default
};
