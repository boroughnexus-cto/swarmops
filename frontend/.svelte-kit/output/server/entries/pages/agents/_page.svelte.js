import { O as head, D as pop, A as push, G as attr } from "../../../chunks/index2.js";
import { C as Card } from "../../../chunks/Card.js";
import { B as Button } from "../../../chunks/Button.js";
function _page($$payload, $$props) {
  push();
  let showAddAgentForm = false;
  let newAgent = { name: "", command: "", params: "" };
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Agents - SwarmOps</title>`;
  });
  $$payload.out.push(`<div class="space-y-6"><div class="flex items-center justify-between"><div><h1 class="text-3xl font-bold text-vanna-navy font-serif">Agents</h1> <p class="mt-2 text-slate-500">Configure and manage AI development agents</p></div> `);
  Button($$payload, {
    onclick: () => showAddAgentForm = true,
    variant: "primary",
    children: ($$payload2) => {
      $$payload2.out.push(`<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path></svg> Add Agent`);
    }
  });
  $$payload.out.push(`<!----></div> `);
  if (showAddAgentForm) {
    $$payload.out.push("<!--[-->");
    Card($$payload, {
      children: ($$payload2) => {
        $$payload2.out.push(`<h3 class="text-xl font-semibold text-vanna-navy mb-4">Add New Agent</h3> <form class="space-y-4"><div><label for="agent-name" class="block text-sm font-medium text-vanna-navy mb-2">Agent Name</label> <input id="agent-name" type="text"${attr("value", newAgent.name)} placeholder="e.g., Claude Assistant" class="input-field" required/></div> <div><label for="agent-command" class="block text-sm font-medium text-vanna-navy mb-2">Command</label> <input id="agent-command" type="text"${attr("value", newAgent.command)} placeholder="e.g., claude" class="input-field" required/></div> <div><label for="agent-params" class="block text-sm font-medium text-vanna-navy mb-2">Parameters (optional)</label> <input id="agent-params" type="text"${attr("value", newAgent.params)} placeholder="e.g., --model claude-3-sonnet" class="input-field"/></div> <div class="flex gap-3">`);
        Button($$payload2, {
          type: "submit",
          variant: "primary",
          children: ($$payload3) => {
            $$payload3.out.push(`<!---->Add Agent`);
          }
        });
        $$payload2.out.push(`<!----> `);
        Button($$payload2, {
          type: "button",
          onclick: () => showAddAgentForm = false,
          variant: "secondary",
          children: ($$payload3) => {
            $$payload3.out.push(`<!---->Cancel`);
          }
        });
        $$payload2.out.push(`<!----></div></form>`);
      }
    });
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="flex items-center justify-center py-12"><div class="animate-spin rounded-full h-8 w-8 border-b-2 border-vanna-teal"></div> <span class="ml-3 text-slate-500">Loading agents...</span></div>`);
  }
  $$payload.out.push(`<!--]--></div>`);
  pop();
}
export {
  _page as default
};
