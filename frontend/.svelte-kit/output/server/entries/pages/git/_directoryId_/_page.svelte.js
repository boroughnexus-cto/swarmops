import { K as store_get, R as copy_payload, S as assign_payload, M as unsubscribe_stores, D as pop, A as push, O as head, J as escape_html } from "../../../../chunks/index2.js";
import { p as page } from "../../../../chunks/stores.js";
import { B as Breadcrumb } from "../../../../chunks/Breadcrumb.js";
function _page($$payload, $$props) {
  push();
  var $$store_subs;
  let directoryId, breadcrumbSegments;
  let tasks = [];
  directoryId = store_get($$store_subs ??= {}, "$page", page).params.directoryId;
  breadcrumbSegments = [
    { label: "", href: "/", icon: "banner" },
    { label: "Git Review", href: `/git/${directoryId}` }
  ];
  tasks.filter((t) => t.status !== "done");
  let $$settled = true;
  let $$inner_payload;
  function $$render_inner($$payload2) {
    head($$payload2, ($$payload3) => {
      $$payload3.title = `<title>Git Review - ${escape_html("Loading...")}</title>`;
    });
    $$payload2.out.push(`<div class="space-y-6">`);
    Breadcrumb($$payload2, { segments: breadcrumbSegments });
    $$payload2.out.push(`<!----> `);
    {
      $$payload2.out.push("<!--[-->");
      $$payload2.out.push(`<div class="flex items-center justify-center min-h-64"><div class="animate-spin rounded-full h-12 w-12 border-b-2 border-vanna-teal"></div></div>`);
    }
    $$payload2.out.push(`<!--]--></div>`);
  }
  do {
    $$settled = true;
    $$inner_payload = copy_payload($$payload);
    $$render_inner($$inner_payload);
  } while (!$$settled);
  assign_payload($$payload, $$inner_payload);
  if ($$store_subs) unsubscribe_stores($$store_subs);
  pop();
}
export {
  _page as default
};
