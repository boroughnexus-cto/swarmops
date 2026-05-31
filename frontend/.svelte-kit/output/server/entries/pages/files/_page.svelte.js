import { O as head, D as pop, A as push } from "../../../chunks/index2.js";
import { B as Breadcrumb } from "../../../chunks/Breadcrumb.js";
function _page($$payload, $$props) {
  push();
  const breadcrumbSegments = [
    { label: "", href: "/", icon: "banner" },
    { label: "Files", href: "/files" }
  ];
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Files - SwarmOps</title>`;
  });
  $$payload.out.push(`<div class="space-y-6">`);
  Breadcrumb($$payload, { segments: breadcrumbSegments });
  $$payload.out.push(`<!----> <div><h1 class="text-3xl font-bold text-vanna-navy font-serif">File Browser</h1> <p class="mt-2 text-slate-500">Browse and edit files in your project directories</p></div> `);
  {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="flex items-center justify-center min-h-64"><div class="animate-spin rounded-full h-12 w-12 border-b-2 border-vanna-teal"></div></div>`);
  }
  $$payload.out.push(`<!--]--></div>`);
  pop();
}
export {
  _page as default
};
