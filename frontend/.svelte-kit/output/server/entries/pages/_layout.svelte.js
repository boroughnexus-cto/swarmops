import { E as ensure_array_like, F as attr_class, G as attr, I as stringify, J as escape_html, D as pop, K as store_get, M as unsubscribe_stores, A as push, N as attr_style, O as head } from "../../chunks/index2.js";
import { p as page } from "../../chunks/stores.js";
import "@sveltejs/kit/internal";
import "../../chunks/exports.js";
import "clsx";
import "../../chunks/state.svelte.js";
import "../../chunks/auth.js";
const favicon = "data:image/svg+xml,%3csvg%20xmlns='http://www.w3.org/2000/svg'%20width='107'%20height='128'%20viewBox='0%200%20107%20128'%3e%3ctitle%3esvelte-logo%3c/title%3e%3cpath%20d='M94.157%2022.819c-10.4-14.885-30.94-19.297-45.792-9.835L22.282%2029.608A29.92%2029.92%200%200%200%208.764%2049.65a31.5%2031.5%200%200%200%203.108%2020.231%2030%2030%200%200%200-4.477%2011.183%2031.9%2031.9%200%200%200%205.448%2024.116c10.402%2014.887%2030.942%2019.297%2045.791%209.835l26.083-16.624A29.92%2029.92%200%200%200%2098.235%2078.35a31.53%2031.53%200%200%200-3.105-20.232%2030%2030%200%200%200%204.474-11.182%2031.88%2031.88%200%200%200-5.447-24.116'%20style='fill:%23ff3e00'/%3e%3cpath%20d='M45.817%20106.582a20.72%2020.72%200%200%201-22.237-8.243%2019.17%2019.17%200%200%201-3.277-14.503%2018%2018%200%200%201%20.624-2.435l.49-1.498%201.337.981a33.6%2033.6%200%200%200%2010.203%205.098l.97.294-.09.968a5.85%205.85%200%200%200%201.052%203.878%206.24%206.24%200%200%200%206.695%202.485%205.8%205.8%200%200%200%201.603-.704L69.27%2076.28a5.43%205.43%200%200%200%202.45-3.631%205.8%205.8%200%200%200-.987-4.371%206.24%206.24%200%200%200-6.698-2.487%205.7%205.7%200%200%200-1.6.704l-9.953%206.345a19%2019%200%200%201-5.296%202.326%2020.72%2020.72%200%200%201-22.237-8.243%2019.17%2019.17%200%200%201-3.277-14.502%2017.99%2017.99%200%200%201%208.13-12.052l26.081-16.623a19%2019%200%200%201%205.3-2.329%2020.72%2020.72%200%200%201%2022.237%208.243%2019.17%2019.17%200%200%201%203.277%2014.503%2018%2018%200%200%201-.624%202.435l-.49%201.498-1.337-.98a33.6%2033.6%200%200%200-10.203-5.1l-.97-.294.09-.968a5.86%205.86%200%200%200-1.052-3.878%206.24%206.24%200%200%200-6.696-2.485%205.8%205.8%200%200%200-1.602.704L37.73%2051.72a5.42%205.42%200%200%200-2.449%203.63%205.79%205.79%200%200%200%20.986%204.372%206.24%206.24%200%200%200%206.698%202.486%205.8%205.8%200%200%200%201.602-.704l9.952-6.342a19%2019%200%200%201%205.295-2.328%2020.72%2020.72%200%200%201%2022.237%208.242%2019.17%2019.17%200%200%201%203.277%2014.503%2018%2018%200%200%201-8.13%2012.053l-26.081%2016.622a19%2019%200%200%201-5.3%202.328'%20style='fill:%23fff'/%3e%3c/svg%3e";
function Sidebar($$payload, $$props) {
  push();
  var $$store_subs;
  let {
    collapsed = false,
    navItems,
    mobileOpen = false
  } = $$props;
  function isActive(href) {
    return store_get($$store_subs ??= {}, "$page", page).url.pathname === href || store_get($$store_subs ??= {}, "$page", page).url.pathname.startsWith(href) && href !== "/";
  }
  function getIcon(iconName) {
    const icons = {
      dashboard: "M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z",
      terminal: "M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z",
      projects: "M19 11H5m14 0L5 7l14 4m0 0L5 19l14-4",
      tasks: "M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4",
      agents: "M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z",
      settings: "M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z",
      swarm: "M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z",
      kanban: "M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2",
      directories: "M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z",
      files: "M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
    };
    return icons[iconName] || icons.dashboard;
  }
  const each_array = ensure_array_like(navItems || []);
  if (mobileOpen) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="fixed inset-0 z-30 bg-vanna-navy/50 backdrop-blur-sm lg:hidden" role="button" tabindex="0" aria-label="Close menu"></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> <aside${attr_class(`fixed left-0 top-0 z-40 h-screen transition-all duration-300 ease-in-out ${stringify(collapsed ? "w-16" : "w-64")} bg-white/95 backdrop-blur-sm border-r border-slate-200 shadow-vanna-card lg:translate-x-0 ${stringify(mobileOpen ? "translate-x-0" : "-translate-x-full")}`)}><div class="h-full flex flex-col"><div class="flex items-center justify-between px-4 py-4 border-b border-slate-200/60">`);
  if (collapsed) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="w-8 h-8 bg-vanna-teal rounded-lg flex items-center justify-center mx-auto"><span class="text-white font-bold text-sm">R</span></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`<img src="https://swarmops.io/static/images/banner.svg" class="h-8" alt="SwarmOps Logo"/>`);
  }
  $$payload.out.push(`<!--]--> <button class="lg:hidden p-1.5 text-slate-500 hover:text-vanna-navy hover:bg-vanna-cream/50 rounded-lg transition-colors" aria-label="Close sidebar"><svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg></button></div> <nav class="flex-1 px-3 py-4 overflow-y-auto"><ul class="space-y-1"><!--[-->`);
  for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
    let item = each_array[$$index];
    $$payload.out.push(`<li><a${attr("href", item.href)}${attr_class(`flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-200 group ${stringify(isActive(item.href) ? "bg-vanna-teal text-white shadow-vanna-subtle" : "text-slate-600 hover:bg-vanna-cream/50 hover:text-vanna-navy")}`)}><svg${attr_class(`w-5 h-5 flex-shrink-0 transition-colors ${stringify(isActive(item.href) ? "text-white" : "text-slate-400 group-hover:text-vanna-teal")}`)} fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"${attr("d", getIcon(item.icon))}></path></svg> `);
    if (!collapsed) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<span class="flex-1 font-medium text-sm">${escape_html(item.label)}</span> `);
      if (item.badge) {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<span${attr_class(`inline-flex items-center justify-center min-w-[1.25rem] h-5 px-1.5 text-xs font-semibold rounded-full ${stringify(isActive(item.href) ? "bg-white/20 text-white" : "bg-vanna-teal/10 text-vanna-teal")}`)}>${escape_html(item.badge)}</span>`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]-->`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--></a></li>`);
  }
  $$payload.out.push(`<!--]--></ul></nav> <div class="px-3 py-4 border-t border-slate-200/60">`);
  if (!collapsed) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="px-3 py-2 bg-vanna-cream/30 rounded-xl"><p class="text-xs text-slate-500">SwarmOps v1.0.0</p> <p class="text-xs text-vanna-teal">Connected</p></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`<div class="flex justify-center"><div class="w-2 h-2 bg-vanna-teal rounded-full"></div></div>`);
  }
  $$payload.out.push(`<!--]--></div></div></aside>`);
  if ($$store_subs) unsubscribe_stores($$store_subs);
  pop();
}
function Header($$payload, $$props) {
  push();
  let {
    sidebarCollapsed = false,
    agentsWaitingForInput = []
  } = $$props;
  let waitingCount = agentsWaitingForInput.length;
  $$payload.out.push(`<header${attr_class(`fixed top-0 right-0 z-30 h-16 bg-white/80 backdrop-blur-md border-b border-slate-200/60 transition-all duration-300 lg:left-${stringify(sidebarCollapsed ? "16" : "64")} left-0`)} style="left: 0; width: 100%;"><div class="hidden lg:block absolute inset-0 transition-all duration-300"${attr_style(`left: ${stringify(sidebarCollapsed ? "4rem" : "16rem")};`)}></div> <div class="relative h-full px-4 lg:px-6"><div${attr_class(`flex items-center justify-between h-full lg:ml-${stringify(sidebarCollapsed ? "16" : "64")}`)} style="margin-left: 0;"><div class="flex items-center gap-3"><button class="lg:hidden p-2 text-slate-500 hover:text-vanna-navy hover:bg-vanna-cream/50 rounded-lg transition-colors" aria-label="Open menu"><svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path></svg></button> <button class="hidden lg:flex items-center p-2 text-slate-500 hover:text-vanna-navy hover:bg-vanna-cream/50 rounded-lg transition-colors"${attr_style(`margin-left: ${stringify(sidebarCollapsed ? "4rem" : "16rem")};`)} aria-label="Toggle sidebar"><svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h8m-8 6h16"></path></svg></button> <div class="lg:hidden"><img src="https://swarmops.io/static/images/banner.svg" class="h-6" alt="SwarmOps"/></div></div> <div class="flex items-center gap-2"><div class="relative hidden md:block"><div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none"><svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path></svg></div> <input type="text" class="w-64 pl-10 pr-4 py-2 text-sm text-vanna-navy bg-vanna-cream/30 border-0 rounded-xl placeholder-slate-400 focus:bg-white focus:ring-2 focus:ring-vanna-teal/30 transition-all" placeholder="Search..."/></div> <div class="relative"><button class="relative p-2 text-slate-500 hover:text-vanna-navy hover:bg-vanna-cream/50 rounded-lg transition-colors" aria-label="Notifications"><svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"></path></svg> `);
  if (waitingCount > 0) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<span class="absolute -top-0.5 -right-0.5 w-4 h-4 text-xs font-bold text-white bg-vanna-orange rounded-full flex items-center justify-center animate-pulse">${escape_html(waitingCount)}</span>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></button> `);
  {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div> <div class="relative"><button class="flex items-center gap-2 p-1.5 hover:bg-vanna-cream/50 rounded-xl transition-colors" aria-label="User menu"><div class="w-8 h-8 bg-gradient-to-br from-vanna-teal to-vanna-navy rounded-lg flex items-center justify-center"><span class="text-white font-semibold text-sm">U</span></div></button> `);
  {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div></div></div></div></header>`);
  pop();
}
function _layout($$payload, $$props) {
  push();
  var $$store_subs;
  let { children } = $$props;
  let sidebarCollapsed = false;
  let mobileMenuOpen = false;
  let stats = {
    agents_waiting_for_input: []
  };
  let navItems = [
    { label: "Dashboard", href: "/", icon: "dashboard" },
    { label: "Swarm", href: "/swarm", icon: "swarm" },
    { label: "Kanban", href: "/kanban", icon: "kanban" },
    {
      label: "Terminal",
      href: "/terminal",
      icon: "terminal",
      badge: void 0
    },
    {
      label: "Projects",
      href: "/projects",
      icon: "projects",
      badge: void 0
    },
    {
      label: "Task Executions",
      href: "/task-executions",
      icon: "tasks",
      badge: void 0
    },
    {
      label: "Agents",
      href: "/agents",
      icon: "agents",
      badge: void 0
    },
    {
      label: "Directories",
      href: "/directories",
      icon: "directories"
    },
    { label: "Files", href: "/files", icon: "files" },
    { label: "Settings", href: "/settings", icon: "settings" }
  ];
  let isLoginPage = store_get($$store_subs ??= {}, "$page", page).url.pathname === "/login";
  head($$payload, ($$payload2) => {
    $$payload2.out.push(`<link rel="icon"${attr("href", favicon)}/>`);
  });
  $$payload.out.push(`<div class="min-h-screen bg-gradient-to-b from-vanna-cream via-white to-vanna-cream relative overflow-hidden"><div class="fixed inset-0 pointer-events-none"><div class="absolute -top-40 -left-40 w-[600px] h-[600px] rounded-full bg-vanna-teal/10 blur-[180px]"></div> <div class="absolute -bottom-40 -right-40 w-[500px] h-[500px] rounded-full bg-vanna-cream blur-[150px]"></div> <div class="absolute top-1/2 right-0 w-[300px] h-[300px] rounded-full bg-vanna-magenta/5 blur-[120px]"></div> <div class="absolute inset-0 dot-pattern opacity-30"></div></div> <div class="relative z-10">`);
  if (isLoginPage) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<main class="min-h-screen">`);
    children?.($$payload);
    $$payload.out.push(`<!----></main>`);
  } else {
    $$payload.out.push("<!--[!-->");
    Sidebar($$payload, {
      navItems,
      collapsed: sidebarCollapsed,
      mobileOpen: mobileMenuOpen
    });
    $$payload.out.push(`<!----> `);
    Header($$payload, {
      sidebarCollapsed,
      agentsWaitingForInput: stats.agents_waiting_for_input
    });
    $$payload.out.push(`<!----> <main${attr_class("transition-all duration-300 pt-16 lg:ml-64", void 0, { "lg:ml-16": sidebarCollapsed })}><div class="p-4 lg:p-6">`);
    children?.($$payload);
    $$payload.out.push(`<!----></div></main>`);
  }
  $$payload.out.push(`<!--]--></div></div>`);
  if ($$store_subs) unsubscribe_stores($$store_subs);
  pop();
}
export {
  _layout as default
};
