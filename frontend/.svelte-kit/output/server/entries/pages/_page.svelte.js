import { Q as element, D as pop, A as push, F as attr_class, G as attr, J as escape_html, I as stringify, O as head, E as ensure_array_like } from "../../chunks/index2.js";
import { C as Card } from "../../chunks/Card.js";
import { B as Button } from "../../chunks/Button.js";
function StatsCard($$payload, $$props) {
  push();
  let {
    title,
    value,
    icon,
    color = "teal",
    trend,
    href,
    loading = false
  } = $$props;
  const colorClasses = {
    teal: {
      bg: "bg-vanna-teal",
      text: "text-vanna-teal",
      lightBg: "bg-vanna-teal/10"
    },
    orange: {
      bg: "bg-vanna-orange",
      text: "text-vanna-orange",
      lightBg: "bg-vanna-orange/10"
    },
    magenta: {
      bg: "bg-vanna-magenta",
      text: "text-vanna-magenta",
      lightBg: "bg-vanna-magenta/10"
    },
    navy: {
      bg: "bg-vanna-navy",
      text: "text-vanna-navy",
      lightBg: "bg-vanna-navy/10"
    },
    green: {
      bg: "bg-green-500",
      text: "text-green-600",
      lightBg: "bg-green-50"
    },
    purple: {
      bg: "bg-purple-500",
      text: "text-purple-600",
      lightBg: "bg-purple-50"
    }
  };
  function getIcon(iconName) {
    const icons = {
      terminal: "M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z",
      projects: "M19 11H5m14 0L5 7l14 4m0 0L5 19l14-4",
      tasks: "M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4",
      agents: "M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z",
      users: "M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z",
      chart: "M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
    };
    return icons[iconName] || icons.chart;
  }
  const Component = href ? "a" : "div";
  element(
    $$payload,
    Component,
    () => {
      $$payload.out.push(`${attr("href", href)}${attr_class(`bg-white/80 backdrop-blur-sm rounded-2xl border border-slate-200/60 p-6 shadow-vanna-card hover:shadow-vanna-feature hover:-translate-y-1 transition-all duration-200 ${stringify(href ? "cursor-pointer" : "")}`)}`);
    },
    () => {
      $$payload.out.push(`<div class="flex items-center"><div class="flex-shrink-0"><div${attr_class(`w-12 h-12 ${stringify(colorClasses[color].lightBg)} rounded-lg flex items-center justify-center`)}><svg${attr_class(`w-6 h-6 ${stringify(colorClasses[color].text)}`)} fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"${attr("d", getIcon(icon))}></path></svg></div></div> <div class="ml-4 flex-1"><div class="flex items-center justify-between"><div><p class="text-sm font-medium text-slate-500">${escape_html(title)}</p> <div class="text-2xl font-bold text-vanna-navy">`);
      if (loading) {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<div class="animate-pulse bg-slate-200 h-8 w-16 rounded"></div>`);
      } else {
        $$payload.out.push("<!--[!-->");
        $$payload.out.push(`${escape_html(value)}`);
      }
      $$payload.out.push(`<!--]--></div></div> `);
      if (trend) {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<div${attr_class(`flex items-center text-sm ${stringify(trend.isPositive ? "text-vanna-teal" : "text-vanna-orange")}`)}><svg class="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20">`);
        if (trend.isPositive) {
          $$payload.out.push("<!--[-->");
          $$payload.out.push(`<path fill-rule="evenodd" d="M3.293 9.707a1 1 0 010-1.414l6-6a1 1 0 011.414 0l6 6a1 1 0 01-1.414 1.414L11 5.414V17a1 1 0 11-2 0V5.414L4.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd"></path>`);
        } else {
          $$payload.out.push("<!--[!-->");
          $$payload.out.push(`<path fill-rule="evenodd" d="M16.707 10.293a1 1 0 010 1.414l-6 6a1 1 0 01-1.414 0l-6-6a1 1 0 111.414-1.414L9 14.586V3a1 1 0 012 0v11.586l4.293-4.293a1 1 0 011.414 0z" clip-rule="evenodd"></path>`);
        }
        $$payload.out.push(`<!--]--></svg> ${escape_html(Math.abs(trend.value))}%</div>`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]--></div></div></div>`);
    }
  );
  pop();
}
function _page($$payload, $$props) {
  push();
  let stats = {
    active_sessions: 0,
    projects: 0,
    task_executions: 0,
    agents: 0,
    git_changes_awaiting_review: [],
    agents_waiting_for_input: [],
    remote_ports: [],
    directory_dev_servers: []
  };
  let loading = true;
  let newPortNumber = "";
  let creatingTunnel = false;
  async function loadDashboardStats() {
    try {
      const response = await fetch("/api/dashboard/stats");
      if (response.ok) {
        stats = await response.json();
      }
    } catch (error) {
      console.error("Failed to load dashboard stats:", error);
    } finally {
      loading = false;
    }
  }
  async function createTunnel() {
    const port = parseInt(newPortNumber);
    if (!port || port <= 0 || port > 65535) {
      alert("Please enter a valid port number (1-65535)");
      return;
    }
    creatingTunnel = true;
    try {
      const response = await fetch("/api/remote-ports", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ port })
      });
      if (response.ok) {
        newPortNumber = "";
        await loadDashboardStats();
      } else {
        const error = await response.text();
        alert("Failed to create tunnel: " + error);
      }
    } catch (error) {
      console.error("Failed to create tunnel:", error);
      alert("Failed to create tunnel");
    } finally {
      creatingTunnel = false;
    }
  }
  async function stopTunnel(id) {
    try {
      const response = await fetch(`/api/remote-ports/${id}`, { method: "DELETE" });
      if (response.ok) {
        await loadDashboardStats();
      } else {
        alert("Failed to stop tunnel");
      }
    } catch (error) {
      console.error("Failed to stop tunnel:", error);
      alert("Failed to stop tunnel");
    }
  }
  async function stopDevServer(id) {
    try {
      const response = await fetch(`/api/directory-dev-servers/${id}`, { method: "DELETE" });
      if (response.ok) {
        await loadDashboardStats();
      } else {
        alert("Failed to stop dev server");
      }
    } catch (error) {
      console.error("Failed to stop dev server:", error);
      alert("Failed to stop dev server");
    }
  }
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Dashboard - SwarmOps</title>`;
  });
  $$payload.out.push(`<div class="space-y-6"><div class="mb-8"><h1 class="text-3xl font-bold text-vanna-navy font-serif">Dashboard</h1> <p class="mt-2 text-slate-500">Welcome to your development environment management platform</p></div> <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">`);
  StatsCard($$payload, {
    title: "Active Sessions",
    value: loading ? "..." : stats.active_sessions,
    icon: "terminal",
    color: "teal",
    href: "/terminal",
    loading
  });
  $$payload.out.push(`<!----> `);
  StatsCard($$payload, {
    title: "Projects",
    value: loading ? "..." : stats.projects,
    icon: "projects",
    color: "navy",
    href: "/projects",
    loading
  });
  $$payload.out.push(`<!----> `);
  StatsCard($$payload, {
    title: "Task Executions",
    value: loading ? "..." : stats.task_executions,
    icon: "tasks",
    color: "magenta",
    href: "/task-executions",
    loading
  });
  $$payload.out.push(`<!----> `);
  StatsCard($$payload, {
    title: "Agents",
    value: loading ? "..." : stats.agents,
    icon: "agents",
    color: "orange",
    href: "/agents",
    loading
  });
  $$payload.out.push(`<!----></div> `);
  Card($$payload, {
    children: ($$payload2) => {
      $$payload2.out.push(`<div class="flex items-center justify-between mb-4"><div class="flex items-center"><svg class="w-5 h-5 mr-2 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"></path></svg> <h2 class="text-lg font-semibold text-vanna-navy">Remote Ports</h2> `);
      if (stats.remote_ports && stats.remote_ports.length > 0) {
        $$payload2.out.push("<!--[-->");
        $$payload2.out.push(`<span class="ml-2 px-2 py-1 text-xs font-medium bg-vanna-teal/10 text-vanna-teal rounded-full">${escape_html(stats.remote_ports.length)}</span>`);
      } else {
        $$payload2.out.push("<!--[!-->");
      }
      $$payload2.out.push(`<!--]--></div></div> <div class="flex items-center gap-3 mb-4"><div class="flex-1 max-w-xs"><input type="number"${attr("value", newPortNumber)} placeholder="Port number (e.g., 3000)" class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-vanna-teal/50 focus:border-vanna-teal"/></div> `);
      Button($$payload2, {
        variant: "primary",
        onclick: createTunnel,
        disabled: creatingTunnel,
        children: ($$payload3) => {
          if (creatingTunnel) {
            $$payload3.out.push("<!--[-->");
            $$payload3.out.push(`<svg class="w-4 h-4 mr-2 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg> Starting...`);
          } else {
            $$payload3.out.push("<!--[!-->");
            $$payload3.out.push(`<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg> Start Tunnel`);
          }
          $$payload3.out.push(`<!--]-->`);
        }
      });
      $$payload2.out.push(`<!----></div> `);
      if (stats.remote_ports && stats.remote_ports.length > 0) {
        $$payload2.out.push("<!--[-->");
        const each_array = ensure_array_like(stats.remote_ports);
        $$payload2.out.push(`<div class="space-y-3"><!--[-->`);
        for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
          let tunnel = each_array[$$index];
          $$payload2.out.push(`<div class="flex items-center justify-between p-3 bg-vanna-cream/30 rounded-lg"><div class="flex-1"><div class="font-medium text-vanna-navy font-mono text-sm">localhost:${escape_html(tunnel.port)}</div> `);
          if (tunnel.external_url) {
            $$payload2.out.push("<!--[-->");
            $$payload2.out.push(`<a${attr("href", tunnel.external_url)} target="_blank" rel="noopener noreferrer" class="text-sm text-vanna-teal hover:underline break-all">${escape_html(tunnel.external_url)}</a>`);
          } else {
            $$payload2.out.push("<!--[!-->");
            $$payload2.out.push(`<div class="text-sm text-slate-400 italic">Waiting for URL...</div>`);
          }
          $$payload2.out.push(`<!--]--></div> <div class="flex items-center space-x-2">`);
          if (tunnel.status === "connected") {
            $$payload2.out.push("<!--[-->");
            $$payload2.out.push(`<span class="px-2 py-1 text-xs font-medium bg-green-100 text-green-700 rounded">connected</span>`);
          } else {
            $$payload2.out.push("<!--[!-->");
            if (tunnel.status === "starting") {
              $$payload2.out.push("<!--[-->");
              $$payload2.out.push(`<span class="px-2 py-1 text-xs font-medium bg-yellow-100 text-yellow-700 rounded animate-pulse">starting</span>`);
            } else {
              $$payload2.out.push("<!--[!-->");
              $$payload2.out.push(`<span class="px-2 py-1 text-xs font-medium bg-red-100 text-red-700 rounded">${escape_html(tunnel.status)}</span>`);
            }
            $$payload2.out.push(`<!--]-->`);
          }
          $$payload2.out.push(`<!--]--> `);
          if (tunnel.external_url) {
            $$payload2.out.push("<!--[-->");
            $$payload2.out.push(`<button class="p-2 text-slate-500 hover:text-vanna-teal hover:bg-vanna-teal/10 rounded-lg transition-colors" title="Copy URL" aria-label="Copy URL to clipboard"><svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path></svg></button>`);
          } else {
            $$payload2.out.push("<!--[!-->");
          }
          $$payload2.out.push(`<!--]--> `);
          Button($$payload2, {
            size: "sm",
            variant: "danger",
            onclick: () => stopTunnel(tunnel.id),
            children: ($$payload3) => {
              $$payload3.out.push(`<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg> Stop`);
            }
          });
          $$payload2.out.push(`<!----></div></div>`);
        }
        $$payload2.out.push(`<!--]--></div>`);
      } else {
        $$payload2.out.push("<!--[!-->");
        if (!loading) {
          $$payload2.out.push("<!--[-->");
          $$payload2.out.push(`<p class="text-sm text-slate-500 text-center py-4">No active tunnels. Enter a port number above to expose a local service remotely.</p>`);
        } else {
          $$payload2.out.push("<!--[!-->");
        }
        $$payload2.out.push(`<!--]-->`);
      }
      $$payload2.out.push(`<!--]-->`);
    }
  });
  $$payload.out.push(`<!----> `);
  if (stats.directory_dev_servers && stats.directory_dev_servers.length > 0) {
    $$payload.out.push("<!--[-->");
    Card($$payload, {
      children: ($$payload2) => {
        const each_array_1 = ensure_array_like(stats.directory_dev_servers);
        $$payload2.out.push(`<div class="flex items-center justify-between mb-4"><div class="flex items-center"><svg class="w-5 h-5 mr-2 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"></path></svg> <h2 class="text-lg font-semibold text-vanna-navy">Running Dev Servers</h2> <span class="ml-2 px-2 py-1 text-xs font-medium bg-green-100 text-green-700 rounded-full">${escape_html(stats.directory_dev_servers.length)}</span></div></div> <div class="space-y-3"><!--[-->`);
        for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
          let devServer = each_array_1[$$index_1];
          $$payload2.out.push(`<div class="flex items-center justify-between p-3 bg-vanna-cream/30 rounded-lg"><div class="flex-1"><div class="text-xs font-semibold text-vanna-teal uppercase tracking-wide mb-1">${escape_html(devServer.project_name)}</div> <div class="font-medium text-vanna-navy font-mono text-sm truncate">${escape_html(devServer.directory_path)}</div> <div class="text-xs text-slate-500 mt-1">Session: ${escape_html(devServer.tmux_session_id)}</div></div> <div class="flex items-center space-x-2"><span class="px-2 py-1 text-xs font-medium bg-green-100 text-green-700 rounded flex items-center gap-1"><span class="w-1.5 h-1.5 bg-green-500 rounded-full animate-pulse"></span> ${escape_html(devServer.status)}</span> `);
          Button($$payload2, {
            size: "sm",
            variant: "ghost",
            onclick: () => window.location.href = `/dev-server/${devServer.base_directory_id}`,
            children: ($$payload3) => {
              $$payload3.out.push(`<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path></svg> Terminal`);
            }
          });
          $$payload2.out.push(`<!----> `);
          Button($$payload2, {
            size: "sm",
            variant: "danger",
            onclick: () => stopDevServer(devServer.id),
            children: ($$payload3) => {
              $$payload3.out.push(`<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg> Stop`);
            }
          });
          $$payload2.out.push(`<!----></div></div>`);
        }
        $$payload2.out.push(`<!--]--></div>`);
      }
    });
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">`);
  if (!loading && stats.git_changes_awaiting_review.length > 0) {
    $$payload.out.push("<!--[-->");
    Card($$payload, {
      children: ($$payload2) => {
        const each_array_2 = ensure_array_like(stats.git_changes_awaiting_review);
        $$payload2.out.push(`<div class="flex items-center mb-4"><svg class="w-5 h-5 mr-2 text-vanna-orange" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"></path></svg> <h2 class="text-lg font-semibold text-vanna-navy">Git Changes Awaiting Review</h2> <span class="ml-2 px-2 py-1 text-xs font-medium bg-vanna-orange/10 text-vanna-orange rounded-full">${escape_html(stats.git_changes_awaiting_review.length)}</span></div> <div class="space-y-3"><!--[-->`);
        for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
          let item = each_array_2[$$index_2];
          $$payload2.out.push(`<div class="flex items-center justify-between p-3 bg-vanna-cream/30 rounded-lg"><div class="flex-1"><div class="font-medium text-vanna-navy font-mono text-sm">${escape_html(item.task_name)}</div> <div class="text-sm text-slate-500"><span class="inline-flex items-center gap-1"><svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14M5 12h14"></path></svg> ${escape_html(item.agent)}</span></div></div> <div class="flex items-center space-x-2"><span class="px-2 py-1 text-xs font-medium bg-vanna-orange/10 text-vanna-orange rounded">uncommitted</span> `);
          Button($$payload2, {
            size: "sm",
            variant: "primary",
            onclick: () => window.location.href = `/git/${item.id}`,
            children: ($$payload3) => {
              $$payload3.out.push(`<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"></path></svg> Review`);
            }
          });
          $$payload2.out.push(`<!----></div></div>`);
        }
        $$payload2.out.push(`<!--]--></div>`);
      }
    });
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if (!loading && stats.agents_waiting_for_input.length > 0) {
    $$payload.out.push("<!--[-->");
    Card($$payload, {
      children: ($$payload2) => {
        const each_array_3 = ensure_array_like(stats.agents_waiting_for_input);
        $$payload2.out.push(`<div class="flex items-center mb-4"><svg class="w-5 h-5 mr-2 text-vanna-orange animate-pulse" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"></path></svg> <h2 class="text-lg font-semibold text-vanna-navy">Agents Waiting for Input</h2> <span class="ml-2 px-2 py-1 text-xs font-medium bg-vanna-orange/10 text-vanna-orange rounded-full">${escape_html(stats.agents_waiting_for_input.length)}</span></div> <div class="space-y-3"><!--[-->`);
        for (let $$index_3 = 0, $$length = each_array_3.length; $$index_3 < $$length; $$index_3++) {
          let execution = each_array_3[$$index_3];
          $$payload2.out.push(`<div class="flex items-center justify-between p-3 bg-vanna-cream/30 rounded-lg"><div class="flex-1"><div class="text-xs font-semibold text-vanna-teal uppercase tracking-wide mb-1">${escape_html(execution.project_name)}</div> <div class="font-medium text-vanna-navy">${escape_html(execution.task_name)}</div> <div class="text-sm text-slate-500">Agent: ${escape_html(execution.agent)}</div></div> <div class="flex items-center space-x-2"><span class="px-2 py-1 text-xs font-medium bg-vanna-orange/10 text-vanna-orange rounded animate-pulse">Waiting</span> `);
          Button($$payload2, {
            size: "sm",
            variant: "primary",
            onclick: () => window.location.href = `/task-executions/${execution.id}`,
            children: ($$payload3) => {
              $$payload3.out.push(`<!---->Check Session`);
            }
          });
          $$payload2.out.push(`<!----></div></div>`);
        }
        $$payload2.out.push(`<!--]--></div>`);
      }
    });
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div></div>`);
  pop();
}
export {
  _page as default
};
