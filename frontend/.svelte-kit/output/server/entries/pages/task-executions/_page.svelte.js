import { O as head, F as attr_class, J as escape_html, D as pop, A as push, I as stringify } from "../../../chunks/index2.js";
import { C as Card } from "../../../chunks/Card.js";
import { B as Button } from "../../../chunks/Button.js";
function _page($$payload, $$props) {
  push();
  let kanbanColumns, totalTasks, completedTasks, runningTasks, waitingTasks;
  let taskExecutions = [];
  kanbanColumns = {
    pending: taskExecutions.filter((t) => t.status?.toLowerCase() === "pending"),
    running: taskExecutions.filter((t) => t.status?.toLowerCase() === "running"),
    waiting: taskExecutions.filter((t) => t.status?.toLowerCase() === "waiting"),
    completed: taskExecutions.filter((t) => t.status?.toLowerCase() === "completed"),
    failed: taskExecutions.filter((t) => t.status?.toLowerCase() === "failed")
  };
  totalTasks = taskExecutions.length;
  completedTasks = kanbanColumns.completed.length;
  runningTasks = kanbanColumns.running.length;
  waitingTasks = kanbanColumns.waiting.length;
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Task Executions - SwarmOps</title>`;
  });
  $$payload.out.push(`<div class="space-y-6"><div class="flex items-center justify-between"><div><h1 class="text-3xl font-bold text-vanna-navy font-serif">Task Executions</h1> <p class="mt-2 text-slate-500">Track and manage task executions and workflows</p></div> <div class="flex items-center space-x-3"><div class="flex items-center bg-vanna-cream/50 rounded-lg p-1"><button${attr_class(`px-3 py-1 text-sm font-medium rounded-md transition-colors ${stringify(
    "bg-white text-vanna-navy shadow-sm"
  )}`)}><svg class="w-4 h-4 mr-1 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2"></path></svg> Kanban</button> <button${attr_class(`px-3 py-1 text-sm font-medium rounded-md transition-colors ${stringify("text-slate-500 hover:text-vanna-navy")}`)}><svg class="w-4 h-4 mr-1 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16"></path></svg> List</button></div> `);
  Button($$payload, {
    href: "/projects",
    variant: "primary",
    children: ($$payload2) => {
      $$payload2.out.push(`<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path></svg> Execute Task`);
    }
  });
  $$payload.out.push(`<!----></div></div> <div class="grid grid-cols-1 md:grid-cols-4 gap-4">`);
  Card($$payload, {
    class: "p-4",
    children: ($$payload2) => {
      $$payload2.out.push(`<div class="flex items-center"><div class="flex-shrink-0"><div class="w-8 h-8 bg-vanna-navy/10 rounded-lg flex items-center justify-center"><svg class="w-4 h-4 text-vanna-navy" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path></svg></div></div> <div class="ml-3"><p class="text-sm font-medium text-slate-500">Total Tasks</p> <p class="text-lg font-semibold text-vanna-navy">${escape_html(totalTasks)}</p></div></div>`);
    }
  });
  $$payload.out.push(`<!----> `);
  Card($$payload, {
    class: "p-4",
    children: ($$payload2) => {
      $$payload2.out.push(`<div class="flex items-center"><div class="flex-shrink-0"><div class="w-8 h-8 bg-vanna-teal/10 rounded-lg flex items-center justify-center"><svg class="w-4 h-4 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg></div></div> <div class="ml-3"><p class="text-sm font-medium text-slate-500">Completed</p> <p class="text-lg font-semibold text-vanna-navy">${escape_html(completedTasks)}</p></div></div>`);
    }
  });
  $$payload.out.push(`<!----> `);
  Card($$payload, {
    class: "p-4",
    children: ($$payload2) => {
      $$payload2.out.push(`<div class="flex items-center"><div class="flex-shrink-0"><div class="w-8 h-8 bg-vanna-magenta/10 rounded-lg flex items-center justify-center"><div class="w-2 h-2 bg-vanna-magenta rounded-full animate-pulse"></div></div></div> <div class="ml-3"><p class="text-sm font-medium text-slate-500">Running</p> <p class="text-lg font-semibold text-vanna-navy">${escape_html(runningTasks)}</p></div></div>`);
    }
  });
  $$payload.out.push(`<!----> `);
  Card($$payload, {
    class: "p-4",
    children: ($$payload2) => {
      $$payload2.out.push(`<div class="flex items-center"><div class="flex-shrink-0"><div class="w-8 h-8 bg-vanna-orange/10 rounded-lg flex items-center justify-center"><svg class="w-4 h-4 text-vanna-orange" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg></div></div> <div class="ml-3"><p class="text-sm font-medium text-slate-500">Waiting</p> <p class="text-lg font-semibold text-vanna-navy">${escape_html(waitingTasks)}</p></div></div>`);
    }
  });
  $$payload.out.push(`<!----></div> `);
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
