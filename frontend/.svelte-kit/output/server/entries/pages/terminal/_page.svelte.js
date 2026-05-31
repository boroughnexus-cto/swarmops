import { F as attr_class, T as clsx, O as head, E as ensure_array_like, J as escape_html, D as pop, A as push } from "../../../chunks/index2.js";
import { g as goto } from "../../../chunks/client.js";
import { C as Card } from "../../../chunks/Card.js";
import { B as Button } from "../../../chunks/Button.js";
import { h as html } from "../../../chunks/html.js";
function Badge($$payload, $$props) {
  let {
    variant = "primary",
    size = "md",
    class: className = "",
    children
  } = $$props;
  const baseClasses = "inline-flex items-center font-medium rounded-full";
  const variantClasses = {
    primary: "bg-vanna-teal/10 text-vanna-teal",
    secondary: "bg-vanna-cream text-vanna-navy",
    success: "bg-vanna-teal/10 text-vanna-teal",
    danger: "bg-vanna-orange/10 text-vanna-orange",
    warning: "bg-vanna-orange/10 text-vanna-orange",
    info: "bg-vanna-teal/10 text-vanna-teal",
    gray: "bg-slate-100 text-slate-600",
    magenta: "bg-vanna-magenta/10 text-vanna-magenta"
  };
  const sizeClasses = {
    sm: "px-2 py-0.5 text-xs",
    md: "px-2.5 py-0.5 text-sm",
    lg: "px-3 py-1 text-sm"
  };
  const classes = `${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]} ${className}`;
  $$payload.out.push(`<span${attr_class(clsx(classes))}>`);
  children?.($$payload);
  $$payload.out.push(`<!----></span>`);
}
function _page($$payload, $$props) {
  push();
  let taskSessions, regularSessions;
  let sessions = [];
  let selectedSession = null;
  let showGlobalTerminal = false;
  let isConnected = false;
  let terminalReady = false;
  let connectionError = null;
  function attachToSession(session) {
    goto(`/terminal/${session.name}`);
  }
  function showGlobal() {
    selectedSession = null;
    showGlobalTerminal = true;
    initializeTerminal();
  }
  function initializeTerminal(sessionType) {
    isConnected = false;
    terminalReady = false;
    connectionError = null;
    if (!window.Terminal) {
      const script1 = document.createElement("script");
      script1.src = "https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.js";
      document.head.appendChild(script1);
      const script2 = document.createElement("script");
      script2.src = "https://cdn.jsdelivr.net/npm/xterm-addon-fit@0.8.0/lib/xterm-addon-fit.js";
      document.head.appendChild(script2);
      const script3 = document.createElement("script");
      script3.src = "https://cdn.jsdelivr.net/npm/xterm-addon-canvas@0.5.0/lib/xterm-addon-canvas.js";
      document.head.appendChild(script3);
      const script4 = document.createElement("script");
      script4.src = "https://cdn.jsdelivr.net/npm/@xterm/addon-unicode11@0.8.0/lib/addon-unicode11.js";
      document.head.appendChild(script4);
      script4.onload = () => createTerminal();
    } else {
      createTerminal();
    }
  }
  function createTerminal(sessionType) {
    {
      console.log("Terminal element not available yet, retrying...");
      setTimeout(() => createTerminal(), 100);
      return;
    }
  }
  function closeTerminal() {
    isConnected = false;
    terminalReady = false;
    connectionError = null;
    selectedSession = null;
    showGlobalTerminal = false;
  }
  function formatTime(timestamp) {
    const date = new Date(parseInt(timestamp) * 1e3);
    return date.toLocaleString();
  }
  function getSessionTypeIcon(session) {
    if (session.is_task) {
      return `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/>
			</svg>`;
    }
    return `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3"/>
		</svg>`;
  }
  taskSessions = sessions.filter((s) => s.is_task);
  regularSessions = sessions.filter((s) => !s.is_task);
  if (sessions.length > 0) {
    setTimeout(
      () => {
        const previewContainers = document.querySelectorAll(".terminal-preview");
        previewContainers.forEach((container) => {
          container.scrollTop = container.scrollHeight;
        });
      },
      10
    );
  }
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>SwarmOps Terminal</title>`;
    $$payload2.out.push(`<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css"/>`);
  });
  $$payload.out.push(`<div class="space-y-6"><div class="flex items-center justify-between"><div><h1 class="text-3xl font-bold text-vanna-navy font-serif">Terminal Sessions</h1> <p class="mt-2 text-slate-500">Manage tmux sessions and terminal access</p></div> `);
  Button($$payload, {
    onclick: showGlobal,
    variant: "primary",
    children: ($$payload2) => {
      $$payload2.out.push(`<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path></svg> Global Terminal`);
    }
  });
  $$payload.out.push(`<!----></div> `);
  if (selectedSession || showGlobalTerminal) {
    $$payload.out.push("<!--[-->");
    Card($$payload, {
      children: ($$payload2) => {
        $$payload2.out.push(`<div class="flex items-center justify-between mb-4"><div class="flex items-center gap-3">`);
        if (selectedSession) {
          $$payload2.out.push("<!--[-->");
          $$payload2.out.push(`<div class="flex items-center gap-2 text-vanna-orange">${html(getSessionTypeIcon(selectedSession))} <span class="font-mono">${escape_html(selectedSession.name)}</span> `);
          if (selectedSession.is_task) {
            $$payload2.out.push("<!--[-->");
            $$payload2.out.push(`<span class="text-sm text-slate-500">(${escape_html(selectedSession.task_name || `Task ${selectedSession.task_id}`)} • ${escape_html(selectedSession.agent_name || `Agent ${selectedSession.agent_id}`)})</span>`);
          } else {
            $$payload2.out.push("<!--[!-->");
          }
          $$payload2.out.push(`<!--]--></div>`);
        } else {
          $$payload2.out.push("<!--[!-->");
          $$payload2.out.push(`<div class="flex items-center gap-2 text-vanna-teal"><svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3"></path></svg> <span class="font-mono">Global Terminal</span></div>`);
        }
        $$payload2.out.push(`<!--]--></div> `);
        Button($$payload2, {
          onclick: closeTerminal,
          variant: "secondary",
          children: ($$payload3) => {
            $$payload3.out.push(`<!---->Close`);
          }
        });
        $$payload2.out.push(`<!----></div> <div class="bg-black rounded-lg border border-slate-300 p-4 shadow-inner"><div id="terminal" class="w-full h-[70vh] focus:outline-none"></div></div> <div class="mt-4 text-sm flex items-center gap-4">`);
        if (connectionError) {
          $$payload2.out.push("<!--[-->");
          $$payload2.out.push(`<div class="flex items-center gap-2 text-vanna-orange"><div class="w-2 h-2 bg-vanna-orange rounded-full"></div> <span>Error: ${escape_html(connectionError)}</span></div>`);
        } else {
          $$payload2.out.push("<!--[!-->");
          if (isConnected && terminalReady) {
            $$payload2.out.push("<!--[-->");
            $$payload2.out.push(`<div class="flex items-center gap-2 text-vanna-teal"><div class="w-2 h-2 bg-vanna-teal rounded-full animate-pulse"></div> <span>Connected</span></div>`);
          } else {
            $$payload2.out.push("<!--[!-->");
            if (isConnected && !terminalReady) {
              $$payload2.out.push("<!--[-->");
              $$payload2.out.push(`<div class="flex items-center gap-2 text-vanna-orange"><div class="w-2 h-2 bg-vanna-orange rounded-full animate-pulse"></div> <span>Initializing terminal...</span></div>`);
            } else {
              $$payload2.out.push("<!--[!-->");
              $$payload2.out.push(`<div class="flex items-center gap-2 text-vanna-orange"><div class="w-2 h-2 bg-vanna-orange rounded-full"></div> <span>Disconnected</span></div>`);
            }
            $$payload2.out.push(`<!--]-->`);
          }
          $$payload2.out.push(`<!--]-->`);
        }
        $$payload2.out.push(`<!--]--> `);
        if (selectedSession) {
          $$payload2.out.push("<!--[-->");
          $$payload2.out.push(`<span class="text-slate-500">Session: <span class="font-mono text-vanna-navy">${escape_html(selectedSession.name)}</span></span>`);
        } else {
          $$payload2.out.push("<!--[!-->");
          if (showGlobalTerminal) {
            $$payload2.out.push("<!--[-->");
            $$payload2.out.push(`<span class="text-slate-500">Global Terminal</span>`);
          } else {
            $$payload2.out.push("<!--[!-->");
          }
          $$payload2.out.push(`<!--]-->`);
        }
        $$payload2.out.push(`<!--]--></div>`);
      }
    });
  } else {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`<div class="space-y-8">`);
    if (taskSessions.length > 0) {
      $$payload.out.push("<!--[-->");
      const each_array = ensure_array_like(taskSessions);
      $$payload.out.push(`<div><h2 class="text-xl font-semibold text-vanna-navy mb-4 flex items-center gap-2"><svg class="w-6 h-6 text-vanna-orange" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"></path></svg> Task Sessions (${escape_html(taskSessions.length)})</h2> <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"><!--[-->`);
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let session = each_array[$$index];
        Card($$payload, {
          class: "card-hover cursor-pointer border-vanna-orange/30 hover:border-vanna-orange/60",
          onclick: () => {
            if (session.execution_id) {
              goto(`/task-executions/${session.execution_id}`);
            } else {
              attachToSession(session);
            }
          },
          children: ($$payload2) => {
            $$payload2.out.push(`<div class="flex items-center justify-between mb-3"><div class="flex items-center gap-2 text-vanna-orange">${html(getSessionTypeIcon(session))} <span class="font-mono text-sm">${escape_html(session.name)}</span></div> <span class="text-xs text-slate-500">${escape_html(formatTime(session.created))}</span></div> <div class="text-sm text-slate-500 mb-3 space-y-1"><div>Task: <span class="text-vanna-navy font-mono">${escape_html(session.task_name || `ID: ${session.task_id}`)}</span></div> <div>Agent: <span class="text-vanna-navy font-mono">${escape_html(session.agent_name || `ID: ${session.agent_id}`)}</span></div> `);
            if (session.execution_id) {
              $$payload2.out.push("<!--[-->");
              $$payload2.out.push(`<div class="text-xs text-vanna-teal font-medium">→ Click to view task execution</div>`);
            } else {
              $$payload2.out.push("<!--[!-->");
            }
            $$payload2.out.push(`<!--]--></div> <div class="terminal-preview bg-gray-900 rounded p-2 text-xs font-mono text-gray-300 h-32 overflow-y-auto overflow-x-auto scrollbar-hide svelte-cxfwhz" style="scroll-behavior: smooth;"><pre class="whitespace-pre font-mono" style="margin: 0;">${html(session.preview)}</pre></div> <div class="mt-3 flex justify-between items-center">`);
            Badge($$payload2, {
              variant: "warning",
              size: "sm",
              children: ($$payload3) => {
                $$payload3.out.push(`<!---->TASK SESSION`);
              }
            });
            $$payload2.out.push(`<!----> `);
            if (session.execution_id) {
              $$payload2.out.push("<!--[-->");
              Button($$payload2, {
                variant: "ghost",
                size: "xs",
                onclick: (e) => {
                  e.stopPropagation();
                  attachToSession(session);
                },
                children: ($$payload3) => {
                  $$payload3.out.push(`<svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3"></path></svg> Terminal`);
                }
              });
            } else {
              $$payload2.out.push("<!--[!-->");
            }
            $$payload2.out.push(`<!--]--></div>`);
          }
        });
      }
      $$payload.out.push(`<!--]--></div></div>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--> `);
    if (regularSessions.length > 0) {
      $$payload.out.push("<!--[-->");
      const each_array_1 = ensure_array_like(regularSessions);
      $$payload.out.push(`<div><h2 class="text-xl font-semibold text-vanna-navy mb-4 flex items-center gap-2"><svg class="w-6 h-6 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3"></path></svg> Regular Sessions (${escape_html(regularSessions.length)})</h2> <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"><!--[-->`);
      for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
        let session = each_array_1[$$index_1];
        Card($$payload, {
          class: "card-hover cursor-pointer",
          onclick: () => attachToSession(session),
          children: ($$payload2) => {
            $$payload2.out.push(`<div class="flex items-center justify-between mb-3"><div class="flex items-center gap-2 text-vanna-teal">${html(getSessionTypeIcon(session))} <span class="font-mono text-sm">${escape_html(session.name)}</span></div> <span class="text-xs text-slate-500">${escape_html(formatTime(session.created))}</span></div> <div class="terminal-preview bg-gray-900 rounded p-2 text-xs font-mono text-gray-300 h-32 overflow-y-auto overflow-x-auto scrollbar-hide svelte-cxfwhz" style="scroll-behavior: smooth;"><pre class="whitespace-pre font-mono" style="margin: 0;">${html(session.preview)}</pre></div> <div class="mt-3 flex justify-end">`);
            Badge($$payload2, {
              variant: "success",
              size: "sm",
              children: ($$payload3) => {
                $$payload3.out.push(`<!---->TMUX SESSION`);
              }
            });
            $$payload2.out.push(`<!----></div>`);
          }
        });
      }
      $$payload.out.push(`<!--]--></div></div>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--> `);
    {
      $$payload.out.push("<!--[-->");
      Card($$payload, {
        class: "text-center py-12",
        children: ($$payload2) => {
          $$payload2.out.push(`<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-vanna-teal mx-auto mb-4"></div> <p class="text-slate-500">Loading tmux sessions...</p>`);
        }
      });
    }
    $$payload.out.push(`<!--]--></div>`);
  }
  $$payload.out.push(`<!--]--></div>`);
  pop();
}
export {
  _page as default
};
