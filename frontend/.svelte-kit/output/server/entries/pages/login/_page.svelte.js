import { O as head, D as pop, A as push } from "../../../chunks/index2.js";
import "@sveltejs/kit/internal";
import "../../../chunks/exports.js";
import "clsx";
import "../../../chunks/state.svelte.js";
import "../../../chunks/auth.js";
import { C as Card } from "../../../chunks/Card.js";
function _page($$payload, $$props) {
  push();
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Login - Web Terminal</title>`;
  });
  $$payload.out.push(`<div class="min-h-screen flex items-center justify-center p-4 -mt-16"><div class="w-full max-w-md">`);
  Card($$payload, {
    class: "p-8",
    children: ($$payload2) => {
      $$payload2.out.push(`<div class="text-center mb-8"><div class="w-16 h-16 bg-vanna-teal/10 rounded-full flex items-center justify-center mx-auto mb-4"><svg class="w-8 h-8 text-vanna-teal" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path></svg></div> <h1 class="text-2xl font-bold text-vanna-navy">Web Terminal</h1> <p class="text-vanna-navy/60 mt-2">`);
      {
        $$payload2.out.push("<!--[-->");
        $$payload2.out.push(`Checking authentication...`);
      }
      $$payload2.out.push(`<!--]--></p></div> `);
      {
        $$payload2.out.push("<!--[!-->");
      }
      $$payload2.out.push(`<!--]--> `);
      {
        $$payload2.out.push("<!--[-->");
        $$payload2.out.push(`<div class="flex justify-center py-8"><svg class="animate-spin h-8 w-8 text-vanna-teal" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg></div>`);
      }
      $$payload2.out.push(`<!--]-->`);
    }
  });
  $$payload.out.push(`<!----> <p class="text-center text-xs text-vanna-navy/40 mt-6">Secured with WebAuthn passkey authentication</p></div></div>`);
  pop();
}
export {
  _page as default
};
