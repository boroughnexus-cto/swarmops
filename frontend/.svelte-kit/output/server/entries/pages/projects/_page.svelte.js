import { O as head, D as pop, A as push, G as attr } from "../../../chunks/index2.js";
import "@sveltejs/kit/internal";
import "../../../chunks/exports.js";
import "../../../chunks/state.svelte.js";
import { C as Card } from "../../../chunks/Card.js";
import { B as Button } from "../../../chunks/Button.js";
function _page($$payload, $$props) {
  push();
  let showCreateForm = false;
  let newProject = { name: "" };
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Projects - SwarmOps</title>`;
  });
  $$payload.out.push(`<div class="space-y-6"><div class="flex items-center justify-between"><div><h1 class="text-3xl font-bold text-vanna-navy font-serif">Projects</h1> <p class="mt-2 text-slate-500">Manage your development projects and repositories</p></div> `);
  Button($$payload, {
    onclick: () => showCreateForm = true,
    variant: "primary",
    children: ($$payload2) => {
      $$payload2.out.push(`<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path></svg> New Project`);
    }
  });
  $$payload.out.push(`<!----></div> `);
  if (showCreateForm) {
    $$payload.out.push("<!--[-->");
    Card($$payload, {
      children: ($$payload2) => {
        $$payload2.out.push(`<h3 class="text-xl font-semibold text-vanna-navy mb-4">Create New Project</h3> <form class="space-y-4"><div><label for="project-name" class="block text-sm font-medium text-vanna-navy mb-2">Project Name</label> <input id="project-name" type="text"${attr("value", newProject.name)} placeholder="Enter project name" class="input-field" required/></div> <div class="flex gap-3">`);
        Button($$payload2, {
          type: "submit",
          variant: "primary",
          children: ($$payload3) => {
            $$payload3.out.push(`<!---->Create Project`);
          }
        });
        $$payload2.out.push(`<!----> `);
        Button($$payload2, {
          type: "button",
          onclick: () => showCreateForm = false,
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
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="flex items-center justify-center py-12"><div class="animate-spin rounded-full h-8 w-8 border-b-2 border-vanna-teal"></div></div>`);
  }
  $$payload.out.push(`<!--]--></div>`);
  pop();
}
export {
  _page as default
};
