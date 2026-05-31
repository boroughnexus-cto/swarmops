import { E as ensure_array_like, Z as fallback, G as attr, J as escape_html, U as bind_props, D as pop, A as push } from "./index2.js";
function Breadcrumb($$payload, $$props) {
  push();
  let segments = fallback($$props["segments"], () => [], true);
  const each_array = ensure_array_like(
    // segments should be an array of objects like:
    // [
    //   { label: "Home", href: "/", icon: "banner" }, // supports icon: "banner" for SwarmOps Logo
    //   { label: "Projects", href: "/projects" },
    //   { label: "My Project", href: "/projects/1" }
    // ]
    // The last segment is automatically treated as current page (no link)
    segments
  );
  $$payload.out.push(`<nav class="flex items-center space-x-2 text-sm text-slate-400 mb-6" aria-label="Breadcrumb"><!--[-->`);
  for (let index = 0, $$length = each_array.length; index < $$length; index++) {
    let segment = each_array[index];
    if (index < segments.length - 1) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<a${attr("href", segment.href)} class="hover:text-vanna-teal transition-colors flex items-center gap-2">`);
      if (segment.icon === "banner") {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<img src="https://swarmops.io/static/images/banner.svg" alt="SwarmOps Logo" class="h-10 w-auto"/>`);
      } else {
        $$payload.out.push("<!--[!-->");
        if (index === 0) {
          $$payload.out.push("<!--[-->");
          $$payload.out.push(`<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H9L7 3H3a2 2 0 00-2 2v2z"></path></svg>`);
        } else {
          $$payload.out.push("<!--[!-->");
        }
        $$payload.out.push(`<!--]-->`);
      }
      $$payload.out.push(`<!--]--> `);
      if (segment.icon !== "banner") {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`${escape_html(segment.label)}`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]--></a> <svg class="w-4 h-4 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg>`);
    } else {
      $$payload.out.push("<!--[!-->");
      $$payload.out.push(`<span class="text-vanna-navy font-medium flex items-center gap-2" aria-current="page">`);
      if (segment.icon === "banner") {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<img src="https://swarmops.io/static/images/banner.svg" alt="SwarmOps Logo" class="h-10 w-auto"/>`);
      } else {
        $$payload.out.push("<!--[!-->");
        if (segments.length === 1) {
          $$payload.out.push("<!--[-->");
          $$payload.out.push(`<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H9L7 3H3a2 2 0 00-2 2v2z"></path></svg>`);
        } else {
          $$payload.out.push("<!--[!-->");
        }
        $$payload.out.push(`<!--]-->`);
      }
      $$payload.out.push(`<!--]--> `);
      if (segment.icon !== "banner") {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`${escape_html(segment.label)}`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]--></span>`);
    }
    $$payload.out.push(`<!--]-->`);
  }
  $$payload.out.push(`<!--]--></nav>`);
  bind_props($$props, { segments });
  pop();
}
export {
  Breadcrumb as B
};
