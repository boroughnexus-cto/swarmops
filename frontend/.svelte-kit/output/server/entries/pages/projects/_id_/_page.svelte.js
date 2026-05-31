import { A as push, P as getContext, G as attr, F as attr_class, T as clsx, U as bind_props, D as pop, J as escape_html, I as stringify, B as setContext, N as attr_style, V as maybe_selected, K as store_get, R as copy_payload, S as assign_payload, M as unsubscribe_stores, O as head, E as ensure_array_like } from "../../../../chunks/index2.js";
import { p as page } from "../../../../chunks/stores.js";
import { B as Breadcrumb } from "../../../../chunks/Breadcrumb.js";
import { B as Button } from "../../../../chunks/Button.js";
import { C as Card } from "../../../../chunks/Card.js";
import { h as html } from "../../../../chunks/html.js";
function Input($$payload, $$props) {
  push();
  let {
    type = "text",
    value = void 0,
    placeholder,
    disabled = false,
    required = false,
    readonly = false,
    class: className = "",
    id,
    name,
    autocomplete,
    min,
    max,
    step,
    pattern,
    size = "md",
    error = false,
    onInput,
    onChange,
    onFocus,
    onBlur
  } = $$props;
  const formFieldContext = getContext("formField");
  let hasError = error || !!formFieldContext?.error;
  let computedDescribedBy = formFieldContext?.describedBy;
  const sizeClasses = {
    sm: "px-3 py-1.5 text-sm",
    md: "px-3 py-2 text-sm",
    lg: "px-4 py-3 text-base"
  };
  const baseClasses = "w-full rounded-lg border transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed";
  const normalClasses = "border-slate-300 bg-white text-vanna-navy focus:border-vanna-teal focus:ring-vanna-teal placeholder-slate-400";
  const errorClasses = "border-vanna-orange bg-vanna-orange/5 text-vanna-navy focus:border-vanna-orange focus:ring-vanna-orange";
  let classes = `${baseClasses} ${sizeClasses[size]} ${hasError ? errorClasses : normalClasses} ${className}`;
  $$payload.out.push(`<input${attr("type", type)}${attr("id", id)}${attr("name", name)}${attr("placeholder", placeholder)}${attr("disabled", disabled, true)}${attr("required", required, true)}${attr("readonly", readonly, true)}${attr("autocomplete", autocomplete)}${attr("min", min)}${attr("max", max)}${attr("step", step)}${attr("pattern", pattern)}${attr("value", value)}${attr_class(clsx(classes))}${attr("aria-invalid", hasError || void 0)}${attr("aria-describedby", computedDescribedBy || void 0)}/>`);
  bind_props($$props, { value });
  pop();
}
function Textarea($$payload, $$props) {
  push();
  let {
    value = void 0,
    placeholder,
    disabled = false,
    required = false,
    readonly = false,
    class: className = "",
    id,
    name,
    rows = 3,
    cols,
    resize = "vertical",
    error = false,
    onInput,
    onChange,
    onFocus,
    onBlur
  } = $$props;
  const formFieldContext = getContext("formField");
  let hasError = error || !!formFieldContext?.error;
  let computedDescribedBy = formFieldContext?.describedBy;
  const resizeClasses = {
    none: "resize-none",
    both: "resize",
    horizontal: "resize-x",
    vertical: "resize-y"
  };
  const baseClasses = "w-full rounded-lg border transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed px-3 py-2 text-sm";
  const normalClasses = "border-slate-300 bg-white text-vanna-navy focus:border-vanna-teal focus:ring-vanna-teal placeholder-slate-400";
  const errorClasses = "border-vanna-orange bg-vanna-orange/5 text-vanna-navy focus:border-vanna-orange focus:ring-vanna-orange";
  let classes = `${baseClasses} ${resizeClasses[resize]} ${hasError ? errorClasses : normalClasses} ${className}`;
  $$payload.out.push(`<textarea${attr("id", id)}${attr("name", name)}${attr("placeholder", placeholder)}${attr("disabled", disabled, true)}${attr("required", required, true)}${attr("readonly", readonly, true)}${attr("rows", rows)}${attr("cols", cols)}${attr_class(clsx(classes))}${attr("aria-invalid", hasError || void 0)}${attr("aria-describedby", computedDescribedBy || void 0)}>`);
  const $$body = escape_html(value);
  if ($$body) {
    $$payload.out.push(`${$$body}`);
  }
  $$payload.out.push(`</textarea>`);
  bind_props($$props, { value });
  pop();
}
function Modal($$payload, $$props) {
  push();
  let {
    open = false,
    title,
    size = "md",
    onClose,
    class: className = "",
    children
  } = $$props;
  const sizeClasses = {
    sm: "max-w-md",
    md: "max-w-lg",
    lg: "max-w-2xl",
    xl: "max-w-4xl",
    "2xl": "max-w-6xl"
  };
  if (open) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="fixed inset-0 z-50 overflow-y-auto" role="dialog" aria-modal="true" tabindex="-1"><div class="fixed inset-0 bg-vanna-navy/50 backdrop-blur-sm transition-opacity" aria-hidden="true"></div> <div class="flex min-h-full items-center justify-center p-4"><div${attr_class(`relative w-full ${stringify(sizeClasses[size])} transform transition-all`)}><div${attr_class(`relative bg-white rounded-2xl shadow-vanna-feature border border-slate-200/60 ${stringify(className)}`)}>`);
    if (title) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<div class="flex items-center justify-between p-6 border-b border-slate-200"><h3 class="text-xl font-semibold text-vanna-navy font-serif">${escape_html(title)}</h3> `);
      if (onClose) {
        $$payload.out.push("<!--[-->");
        $$payload.out.push(`<button type="button" class="text-slate-400 hover:text-vanna-navy hover:bg-vanna-cream/50 rounded-lg p-2 transition-colors" aria-label="Close modal"><svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg></button>`);
      } else {
        $$payload.out.push("<!--[!-->");
      }
      $$payload.out.push(`<!--]--></div>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--> <div class="p-6">`);
    children?.($$payload);
    $$payload.out.push(`<!----></div></div></div></div></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]-->`);
  pop();
}
function FormField($$payload, $$props) {
  push();
  let {
    label,
    id,
    required = false,
    error,
    description,
    class: className = "",
    children
  } = $$props;
  const errorId = id ? `${id}-error` : void 0;
  const descriptionId = id ? `${id}-description` : void 0;
  let describedBy = [error && errorId, description && descriptionId].filter(Boolean).join(" ") || void 0;
  setContext("formField", {
    get error() {
      return error;
    },
    get describedBy() {
      return describedBy;
    }
  });
  $$payload.out.push(`<div${attr_class(`space-y-2 ${stringify(className)}`)}>`);
  if (label) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<label${attr("for", id)} class="block text-sm font-medium text-vanna-navy">${escape_html(label)} `);
    if (required) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<span class="text-vanna-orange ml-1" aria-hidden="true">*</span> <span class="sr-only">(required)</span>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--></label>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> <div class="relative">`);
  children?.($$payload);
  $$payload.out.push(`<!----></div> `);
  if (description) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<p${attr("id", descriptionId)} class="text-sm text-slate-500">${escape_html(description)}</p>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if (error) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<p${attr("id", errorId)} class="text-sm text-vanna-orange flex items-center gap-1" role="alert"><svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg> ${escape_html(error)}</p>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div>`);
  pop();
}
function Select($$payload, $$props) {
  push();
  let {
    value = void 0,
    placeholder,
    disabled = false,
    required = false,
    class: className = "",
    id,
    name,
    size = "md",
    error = false,
    onChange,
    onFocus,
    onBlur,
    children
  } = $$props;
  const sizeClasses = {
    sm: "px-3 py-1.5 text-sm",
    md: "px-3 py-2 text-sm",
    lg: "px-4 py-3 text-base"
  };
  const baseClasses = "w-full rounded-lg border transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed appearance-none bg-no-repeat bg-right bg-[length:16px_16px] pr-10";
  const normalClasses = "border-slate-300 bg-white text-vanna-navy focus:border-vanna-teal focus:ring-vanna-teal";
  const errorClasses = "border-vanna-orange bg-vanna-orange/5 text-vanna-navy focus:border-vanna-orange focus:ring-vanna-orange";
  const chevronIcon = "data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='m6 8 4 4 4-4'/%3e%3c/svg%3e";
  let classes = `${baseClasses} ${sizeClasses[size]} ${error ? errorClasses : normalClasses} ${className}`;
  let backgroundImage = `url("${chevronIcon}")`;
  $$payload.out.push(`<div class="relative"><select${attr("id", id)}${attr("name", name)}${attr("disabled", disabled, true)}${attr("required", required, true)}${attr_class(clsx(classes))}${attr_style(`background-image: ${stringify(backgroundImage)};`)}>`);
  $$payload.select_value = value;
  if (placeholder) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<option value=""${maybe_selected($$payload, "")} disabled hidden>${escape_html(placeholder)}</option>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]-->`);
  children?.($$payload);
  $$payload.out.push(`<!---->`);
  $$payload.select_value = void 0;
  $$payload.out.push(`</select></div>`);
  bind_props($$props, { value });
  pop();
}
function StatusBadge($$payload, $$props) {
  push();
  let { status, size = "md", variant = "dot", class: className = "" } = $$props;
  const statusConfig = {
    todo: {
      label: "To Do",
      color: "slate",
      dotColor: "bg-slate-400",
      solidColor: "bg-slate-100 text-slate-600",
      outlineColor: "border-slate-300 text-slate-600"
    },
    in_progress: {
      label: "In Progress",
      color: "teal",
      dotColor: "bg-vanna-teal",
      solidColor: "bg-vanna-teal/10 text-vanna-teal",
      outlineColor: "border-vanna-teal/30 text-vanna-teal"
    },
    done: {
      label: "Done",
      color: "teal",
      dotColor: "bg-vanna-teal",
      solidColor: "bg-vanna-teal/10 text-vanna-teal",
      outlineColor: "border-vanna-teal/30 text-vanna-teal"
    },
    running: {
      label: "Running",
      color: "teal",
      dotColor: "bg-vanna-teal",
      solidColor: "bg-vanna-teal/10 text-vanna-teal",
      outlineColor: "border-vanna-teal/30 text-vanna-teal"
    },
    completed: {
      label: "Completed",
      color: "teal",
      dotColor: "bg-vanna-teal",
      solidColor: "bg-vanna-teal/10 text-vanna-teal",
      outlineColor: "border-vanna-teal/30 text-vanna-teal"
    },
    failed: {
      label: "Failed",
      color: "orange",
      dotColor: "bg-vanna-orange",
      solidColor: "bg-vanna-orange/10 text-vanna-orange",
      outlineColor: "border-vanna-orange/30 text-vanna-orange"
    },
    pending: {
      label: "Pending",
      color: "slate",
      dotColor: "bg-slate-400",
      solidColor: "bg-slate-100 text-slate-400",
      outlineColor: "border-slate-300 text-slate-400"
    },
    waiting: {
      label: "Waiting",
      color: "slate",
      dotColor: "bg-slate-400",
      solidColor: "bg-slate-100 text-slate-500",
      outlineColor: "border-slate-300 text-slate-500"
    },
    starting: {
      label: "Starting",
      color: "teal",
      dotColor: "bg-vanna-teal/70",
      solidColor: "bg-vanna-teal/10 text-vanna-teal",
      outlineColor: "border-vanna-teal/30 text-vanna-teal"
    }
  };
  const sizeClasses = {
    sm: "px-2 py-0.5 text-xs",
    md: "px-2.5 py-1 text-sm",
    lg: "px-3 py-1.5 text-base"
  };
  const dotSizeClasses = { sm: "w-1.5 h-1.5", md: "w-2 h-2", lg: "w-2.5 h-2.5" };
  let config = statusConfig[status] || statusConfig.pending;
  let baseClasses = `inline-flex items-center gap-1.5 rounded-full font-medium ${sizeClasses[size]}`;
  let variantClasses = {
    dot: `${config.solidColor}`,
    solid: `${config.solidColor}`,
    outline: `border ${config.outlineColor} bg-transparent`
  }[variant];
  $$payload.out.push(`<span${attr_class(`${stringify(baseClasses)} ${stringify(variantClasses)} ${stringify(className)}`)}>`);
  if (variant === "dot" || variant === "solid") {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div${attr_class(`rounded-full ${stringify(dotSizeClasses[size])} ${stringify(config.dotColor)}`)}></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> ${escape_html(config.label)}</span>`);
  pop();
}
function IconButton($$payload, $$props) {
  let {
    variant = "ghost",
    size = "md",
    disabled = false,
    loading = false,
    class: className = "",
    type = "button",
    title,
    onclick,
    children
  } = $$props;
  const baseClasses = "inline-flex items-center justify-center font-medium rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed";
  const variantClasses = {
    primary: "bg-vanna-teal hover:bg-vanna-teal/90 text-white focus:ring-vanna-teal",
    secondary: "bg-vanna-cream hover:bg-vanna-cream/80 text-vanna-navy focus:ring-vanna-teal",
    success: "bg-vanna-teal hover:bg-vanna-teal/90 text-white focus:ring-vanna-teal",
    danger: "bg-vanna-orange hover:bg-vanna-orange/90 text-white focus:ring-vanna-orange",
    warning: "bg-vanna-orange hover:bg-vanna-orange/90 text-white focus:ring-vanna-orange",
    info: "bg-vanna-teal hover:bg-vanna-teal/90 text-white focus:ring-vanna-teal",
    ghost: "bg-transparent hover:bg-vanna-cream/50 text-vanna-navy focus:ring-vanna-teal"
  };
  const sizeClasses = { xs: "p-1", sm: "p-1.5", md: "p-2", lg: "p-2.5" };
  const iconSizeClasses = { xs: "w-3 h-3", sm: "w-4 h-4", md: "w-5 h-5", lg: "w-6 h-6" };
  let classes = `${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]} ${className}`;
  $$payload.out.push(`<button${attr("type", type)}${attr("disabled", disabled, true)}${attr("title", title)}${attr_class(clsx(classes))}${attr("aria-label", title)}>`);
  if (loading) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<svg${attr_class(`animate-spin ${stringify(iconSizeClasses[size])}`)} fill="none" viewBox="0 0 24 24" aria-hidden="true"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>`);
  } else {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`<div${attr_class(iconSizeClasses[size])}>`);
    children?.($$payload);
    $$payload.out.push(`<!----></div>`);
  }
  $$payload.out.push(`<!--]--></button>`);
}
function EmptyState($$payload, $$props) {
  let {
    title,
    description,
    icon = "generic",
    class: className = "",
    children
  } = $$props;
  const icons = {
    tasks: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/>`,
    folder: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"/>`,
    agents: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>`,
    search: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>`,
    generic: `<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"/>`
  };
  $$payload.out.push(`<div${attr_class(`text-center py-12 ${stringify(className)}`)}><div class="w-16 h-16 bg-vanna-teal/10 rounded-xl flex items-center justify-center mx-auto mb-6"><svg class="w-8 h-8 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24">${html(icons[icon])}</svg></div> <h3 class="text-lg font-semibold text-vanna-navy mb-2">${escape_html(title)}</h3> `);
  if (description) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<p class="text-slate-500 mb-6 max-w-md mx-auto">${escape_html(description)}</p>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--> `);
  if (children) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div class="flex flex-col sm:flex-row gap-3 justify-center items-center">`);
    children?.($$payload);
    $$payload.out.push(`<!----></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
  }
  $$payload.out.push(`<!--]--></div>`);
}
function _page($$payload, $$props) {
  push();
  var $$store_subs;
  let breadcrumbSegments, projectId, todoTasks, inProgressTasks, toVerifyTasks, doneTasks;
  let project = null;
  let loading = true;
  let error = null;
  let showCreateTaskForm = false;
  let showCreateDirectoryForm = false;
  let newTask = {
    title: "",
    description: "",
    status: "todo",
    baseDirectoryId: ""
  };
  let newDirectory = {
    path: "",
    gitInitialized: false,
    setupCommands: "",
    teardownCommands: "",
    devServerSetupCommands: "",
    devServerTeardownCommands: ""
  };
  let editingDirectoryId = "";
  let editDirectory = {
    path: "",
    gitInitialized: false,
    setupCommands: "",
    teardownCommands: "",
    devServerSetupCommands: "",
    devServerTeardownCommands: ""
  };
  let selectedTask = null;
  let showTaskModal = false;
  let availableAgents = [];
  let selectedAgents = [];
  let executingTask = false;
  let showEditTaskModal = false;
  let editingTask = null;
  let editTaskForm = { title: "", description: "", status: "", baseDirectoryId: "" };
  let taskExecutions = /* @__PURE__ */ new Map();
  let deletingTasks = /* @__PURE__ */ new Set();
  let updatingTasks = /* @__PURE__ */ new Set();
  let deletingProject = false;
  let refreshing = false;
  let directoryGitStatus = /* @__PURE__ */ new Map();
  let devServerStatusMap = /* @__PURE__ */ new Map();
  let startingDevServer = /* @__PURE__ */ new Set();
  let stoppingDevServer = /* @__PURE__ */ new Set();
  const columns = [
    { id: "todo", title: "To Do", color: "bg-slate-400" },
    {
      id: "in_progress",
      title: "In Progress",
      color: "bg-vanna-magenta"
    },
    {
      id: "to_verify",
      title: "To Verify",
      color: "bg-vanna-orange"
    },
    { id: "done", title: "Done", color: "bg-vanna-teal" }
  ];
  function setDefaultBaseDirectory() {
    if (project && project.baseDirectories && project.baseDirectories.length === 1) {
      newTask.baseDirectoryId = project.baseDirectories[0].base_directory_id;
    }
  }
  async function loadProject() {
    try {
      if (!project) {
        loading = true;
      } else {
        refreshing = true;
      }
      error = null;
      const response = await fetch(`/api/projects/${projectId}`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      project = await response.json();
      console.log("=== RAW PROJECT DATA ===");
      console.log("Raw project from API:", project);
      console.log("Raw tasks:", project.tasks);
      if (project.tasks) {
        project.tasks = project.tasks.map((task) => ({ ...task, status: task.status || "todo" }));
        console.log("=== PROCESSED TASKS ===");
        console.log("Tasks after processing:", project.tasks);
      }
      project = { ...project };
      loading = false;
      refreshing = false;
    } catch (err) {
      console.error("Failed to load project:", err);
      error = err.message;
      loading = false;
      refreshing = false;
    }
  }
  function getTasksByStatus(status) {
    console.log("getTasksByStatus called with status:", status);
    console.log("Project exists:", !!project);
    console.log("Project.tasks exists:", !!(project && project.tasks));
    console.log("Project.tasks length:", project?.tasks?.length || 0);
    if (!project || !project.tasks) {
      console.log("Returning empty array - no project or tasks");
      return [];
    }
    console.log("All tasks with statuses:", project.tasks.map((t) => ({ id: t.id, title: t.title, status: t.status })));
    const filteredTasks = project.tasks.filter((task) => {
      console.log(`Task ${task.id} (${task.title}) has status '${task.status}', looking for '${status}', match: ${task.status === status}`);
      return task.status === status;
    });
    console.log(`Found ${filteredTasks.length} tasks with status '${status}':`, filteredTasks.map((t) => ({ id: t.id, title: t.title })));
    return filteredTasks;
  }
  async function loadTaskExecutions() {
    if (!project || !project.tasks) return;
    try {
      for (const task of project.tasks) {
        const response = await fetch(`/api/task-executions?task_id=${task.id}`);
        if (response.ok) {
          const executions = await response.json();
          taskExecutions.set(task.id, executions);
        }
      }
      taskExecutions = new Map(taskExecutions);
    } catch (error2) {
      console.error("Failed to load task executions:", error2);
    }
  }
  async function deleteDirectory(directoryId) {
    if (!confirm("Are you sure you want to delete this directory? This action cannot be undone.")) {
      return;
    }
    try {
      const url = `/api/projects/${projectId}/base-directories/${directoryId}`;
      console.log("Deleting directory:", url);
      const response = await fetch(url, { method: "DELETE" });
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      await loadProject();
      console.log("Directory deleted successfully");
    } catch (error2) {
      console.error("Failed to delete directory:", error2);
      alert("Failed to delete directory. Please try again.");
    }
  }
  function startEditDirectory(dir) {
    editingDirectoryId = dir.base_directory_id;
    editDirectory = {
      path: dir.path,
      gitInitialized: (dir.gitInitialized ?? dir.git_initialized) || false,
      setupCommands: dir.setupCommands ?? dir.setup_commands ?? "",
      teardownCommands: dir.teardownCommands ?? dir.teardown_commands ?? "",
      devServerSetupCommands: dir.devServerSetupCommands ?? dir.dev_server_setup_commands ?? "",
      devServerTeardownCommands: dir.devServerTeardownCommands ?? dir.dev_server_teardown_commands ?? ""
    };
  }
  function cancelEditDirectory() {
    editingDirectoryId = "";
    editDirectory = {
      path: "",
      gitInitialized: false,
      setupCommands: "",
      teardownCommands: "",
      devServerSetupCommands: "",
      devServerTeardownCommands: ""
    };
  }
  async function saveEditDirectory() {
    if (!editingDirectoryId) return;
    try {
      const url = `/api/projects/${projectId}/base-directories/${editingDirectoryId}`;
      const response = await fetch(url, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(editDirectory)
      });
      if (!response.ok) {
        const text = await response.text();
        console.error("Update failed:", text);
        throw new Error(`HTTP ${response.status}`);
      }
      await loadProject();
      cancelEditDirectory();
    } catch (err) {
      console.error("Failed to update directory:", err);
      alert("Failed to update directory.");
    }
  }
  function closeTaskModal() {
    selectedTask = null;
    showTaskModal = false;
    selectedAgents = [];
    availableAgents = [];
  }
  function openEditTaskModal(task) {
    editingTask = task;
    editTaskForm = {
      title: task.title,
      description: task.description,
      status: task.status,
      baseDirectoryId: task.baseDirectory?.base_directory_id || ""
    };
    showEditTaskModal = true;
  }
  function closeEditTaskModal() {
    editingTask = null;
    showEditTaskModal = false;
    editTaskForm = { title: "", description: "", status: "", baseDirectoryId: "" };
  }
  async function startTaskExecution() {
    if (selectedAgents.length === 0) {
      alert("Please select at least one agent to execute the task.");
      return;
    }
    executingTask = true;
    try {
      for (const agent of selectedAgents) {
        const response = await fetch("/api/task-executions", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ task_id: selectedTask.id, agent_id: agent.id })
        });
        if (!response.ok) {
          throw new Error(`Failed to start execution with ${agent.name}: ${response.status}`);
        }
      }
      alert(`Task execution started with ${selectedAgents.length} agent(s)!`);
      await loadProject();
      await loadTaskExecutions();
      closeTaskModal();
    } catch (error2) {
      console.error("Failed to start task execution:", error2);
      alert("Failed to start task execution. Please try again.");
    } finally {
      executingTask = false;
    }
  }
  async function deleteTask(task) {
    if (deletingTasks.has(task.id)) return;
    const confirmed = confirm(`Are you sure you want to delete "${task.title}"? This will:

• Delete all task executions
• Clean up all associated resources
• Remove the task permanently

This action cannot be undone.`);
    if (!confirmed) return;
    try {
      deletingTasks = /* @__PURE__ */ new Set([...deletingTasks, task.id]);
      const response = await fetch(`/api/tasks/${task.id}`, { method: "DELETE" });
      if (response.ok) {
        await loadProject();
        await loadTaskExecutions();
      } else {
        const errorData = await response.text();
        alert(`Failed to delete task: ${errorData}`);
      }
    } catch (err) {
      console.error("Failed to delete task:", err);
      alert("Failed to delete task");
    } finally {
      deletingTasks = new Set([...deletingTasks].filter((id) => id !== task.id));
    }
  }
  async function deleteProject() {
    if (deletingProject) return;
    const confirmed = confirm(`Are you sure you want to delete the "${project.name}" project? This will:

• Delete all tasks and their executions
• Clean up all associated resources
• Remove all base directories
• Delete the project permanently

This action cannot be undone.`);
    if (!confirmed) return;
    try {
      deletingProject = true;
      const response = await fetch(`/api/projects/${projectId}`, { method: "DELETE" });
      if (response.ok) {
        window.location.href = "/";
      } else {
        const errorData = await response.text();
        alert(`Failed to delete project: ${errorData}`);
      }
    } catch (err) {
      console.error("Failed to delete project:", err);
      alert("Failed to delete project");
    } finally {
      deletingProject = false;
    }
  }
  breadcrumbSegments = [
    { label: "", href: "/", icon: "banner" },
    { label: "Projects", href: "/projects" },
    {
      label: project?.name || "Loading...",
      href: `/projects/${store_get($$store_subs ??= {}, "$page", page).params.id}`
    }
  ];
  projectId = store_get($$store_subs ??= {}, "$page", page).params.id;
  if (project && project.tasks) {
    console.log("=== PROJECT TASKS DEBUG ===");
    console.log("Total tasks:", project.tasks.length);
    console.log("Task statuses:", project.tasks.map((t) => ({ id: t.id, title: t.title, status: t.status })));
    console.log("Unique statuses in tasks:", [...new Set(project.tasks.map((t) => t.status))]);
  }
  todoTasks = project && !loading ? getTasksByStatus("todo") : [];
  inProgressTasks = project && !loading ? getTasksByStatus("in_progress") : [];
  toVerifyTasks = project && !loading ? getTasksByStatus("to_verify") : [];
  doneTasks = project && !loading ? getTasksByStatus("done") : [];
  {
    console.log("=== KANBAN FILTER DEBUG ===");
    console.log("Todo tasks:", todoTasks.length, todoTasks);
    console.log("In Progress tasks:", inProgressTasks.length, inProgressTasks);
    console.log("Done tasks:", doneTasks.length, doneTasks);
    console.log("Columns config:", columns);
  }
  let $$settled = true;
  let $$inner_payload;
  function $$render_inner($$payload2) {
    head($$payload2, ($$payload3) => {
      $$payload3.title = `<title>${escape_html(project?.name || "Project")} - SwarmOps</title>`;
    });
    $$payload2.out.push(`<div class="space-y-6">`);
    Breadcrumb($$payload2, { segments: breadcrumbSegments });
    $$payload2.out.push(`<!----> `);
    if (loading) {
      $$payload2.out.push("<!--[-->");
      $$payload2.out.push(`<div class="flex items-center justify-center min-h-64"><div class="animate-spin rounded-full h-12 w-12 border-b-2 border-vanna-teal"></div></div>`);
    } else {
      $$payload2.out.push("<!--[!-->");
      if (error) {
        $$payload2.out.push("<!--[-->");
        Card($$payload2, {
          class: "border-vanna-orange/30 bg-vanna-orange/5",
          children: ($$payload3) => {
            $$payload3.out.push(`<div class="flex"><svg class="w-5 h-5 text-vanna-orange" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg> <div class="ml-3"><h3 class="text-sm font-medium text-vanna-orange">Error loading project</h3> <div class="mt-2 text-sm text-vanna-orange/80"><p>${escape_html(error)}</p></div></div></div>`);
          }
        });
      } else {
        $$payload2.out.push("<!--[!-->");
        if (project) {
          $$payload2.out.push("<!--[-->");
          $$payload2.out.push(`<div class="border-b border-slate-200 pb-6 mb-8"><div class="flex items-center justify-between"><div class="min-w-0 flex-1"><div class="flex items-center gap-4 mb-2"><div class="w-12 h-12 bg-vanna-teal rounded-xl flex items-center justify-center"><svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path></svg></div> <div><h1 class="text-2xl font-bold text-vanna-navy font-serif sm:text-3xl">${escape_html(project.name)}</h1> <p class="mt-1 text-sm text-slate-500">${escape_html((project.tasks || []).length)} tasks • ${escape_html((project.baseDirectories || []).length)} directories</p></div></div></div> <div class="flex items-center gap-3 ml-6"><div class="flex bg-vanna-cream/50 rounded-lg p-1"><button${attr_class(`px-3 py-1 rounded text-sm transition-colors ${stringify(
            "bg-vanna-teal text-white"
          )}`)}>Kanban</button> <button${attr_class(`px-3 py-1 rounded text-sm transition-colors ${stringify("text-vanna-navy hover:text-vanna-teal")}`)}>List</button></div> `);
          Button($$payload2, {
            variant: "success",
            onclick: () => showCreateDirectoryForm = true,
            children: ($$payload3) => {
              $$payload3.out.push(`<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"></path></svg> Add Directory`);
            }
          });
          $$payload2.out.push(`<!----> `);
          Button($$payload2, {
            variant: "primary",
            onclick: () => {
              showCreateTaskForm = true;
              setDefaultBaseDirectory();
            },
            children: ($$payload3) => {
              $$payload3.out.push(`<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path></svg> New Task`);
            }
          });
          $$payload2.out.push(`<!----> `);
          Button($$payload2, {
            variant: "danger",
            onclick: deleteProject,
            disabled: deletingProject,
            loading: deletingProject,
            children: ($$payload3) => {
              $$payload3.out.push(`<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>${escape_html(deletingProject ? "Deleting..." : "Delete Project")}`);
            }
          });
          $$payload2.out.push(`<!----></div></div></div>`);
        } else {
          $$payload2.out.push("<!--[!-->");
        }
        $$payload2.out.push(`<!--]-->`);
      }
      $$payload2.out.push(`<!--]-->`);
    }
    $$payload2.out.push(`<!--]--> `);
    Modal($$payload2, {
      open: showCreateTaskForm,
      title: "Create New Task",
      onClose: () => showCreateTaskForm = false,
      children: ($$payload3) => {
        $$payload3.out.push(`<form class="space-y-6">`);
        FormField($$payload3, {
          label: "Task Title",
          id: "task-title",
          required: true,
          children: ($$payload4) => {
            Input($$payload4, {
              id: "task-title",
              type: "text",
              placeholder: "Enter task title",
              required: true,
              get value() {
                return newTask.title;
              },
              set value($$value) {
                newTask.title = $$value;
                $$settled = false;
              }
            });
          }
        });
        $$payload3.out.push(`<!----> `);
        FormField($$payload3, {
          label: "Description",
          id: "task-description",
          children: ($$payload4) => {
            Textarea($$payload4, {
              id: "task-description",
              placeholder: "Enter task description",
              rows: 3,
              get value() {
                return newTask.description;
              },
              set value($$value) {
                newTask.description = $$value;
                $$settled = false;
              }
            });
          }
        });
        $$payload3.out.push(`<!----> `);
        FormField($$payload3, {
          label: "Base Directory",
          id: "task-base-directory",
          required: true,
          children: ($$payload4) => {
            Select($$payload4, {
              id: "task-base-directory",
              placeholder: "Select a base directory...",
              required: true,
              get value() {
                return newTask.baseDirectoryId;
              },
              set value($$value) {
                newTask.baseDirectoryId = $$value;
                $$settled = false;
              },
              children: ($$payload5) => {
                if (project && project.baseDirectories) {
                  $$payload5.out.push("<!--[-->");
                  const each_array = ensure_array_like(project.baseDirectories);
                  $$payload5.out.push(`<!--[-->`);
                  for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
                    let directory = each_array[$$index];
                    $$payload5.out.push(`<option${attr("value", directory.base_directory_id)}${maybe_selected($$payload5, directory.base_directory_id)}>${escape_html(directory.path)}</option>`);
                  }
                  $$payload5.out.push(`<!--]-->`);
                } else {
                  $$payload5.out.push("<!--[!-->");
                }
                $$payload5.out.push(`<!--]-->`);
              },
              $$slots: { default: true }
            });
          }
        });
        $$payload3.out.push(`<!----> `);
        FormField($$payload3, {
          label: "Initial Status",
          id: "task-status",
          children: ($$payload4) => {
            Select($$payload4, {
              id: "task-status",
              get value() {
                return newTask.status;
              },
              set value($$value) {
                newTask.status = $$value;
                $$settled = false;
              },
              children: ($$payload5) => {
                const each_array_1 = ensure_array_like(columns);
                $$payload5.out.push(`<!--[-->`);
                for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
                  let column = each_array_1[$$index_1];
                  $$payload5.out.push(`<option${attr("value", column.id)}${maybe_selected($$payload5, column.id)}>${escape_html(column.title)}</option>`);
                }
                $$payload5.out.push(`<!--]-->`);
              },
              $$slots: { default: true }
            });
          }
        });
        $$payload3.out.push(`<!----> <div class="flex gap-3 pt-4">`);
        Button($$payload3, {
          type: "submit",
          variant: "primary",
          children: ($$payload4) => {
            $$payload4.out.push(`<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path></svg> Create Task`);
          }
        });
        $$payload3.out.push(`<!----> `);
        Button($$payload3, {
          type: "button",
          variant: "secondary",
          onclick: () => showCreateTaskForm = false,
          children: ($$payload4) => {
            $$payload4.out.push(`<!---->Cancel`);
          }
        });
        $$payload3.out.push(`<!----></div></form>`);
      }
    });
    $$payload2.out.push(`<!----> `);
    Modal($$payload2, {
      open: showCreateDirectoryForm,
      title: "Add Base Directory",
      onClose: () => showCreateDirectoryForm = false,
      children: ($$payload3) => {
        $$payload3.out.push(`<form class="space-y-6">`);
        FormField($$payload3, {
          label: "Directory Path",
          id: "directory-path",
          required: true,
          children: ($$payload4) => {
            Input($$payload4, {
              id: "directory-path",
              type: "text",
              placeholder: "/path/to/project/directory",
              required: true,
              get value() {
                return newDirectory.path;
              },
              set value($$value) {
                newDirectory.path = $$value;
                $$settled = false;
              }
            });
          }
        });
        $$payload3.out.push(`<!----> <div class="flex items-center"><input id="git-initialized" type="checkbox"${attr("checked", newDirectory.gitInitialized, true)} class="w-4 h-4 text-vanna-teal bg-white border-slate-300 rounded focus:ring-vanna-teal focus:ring-2"/> <label for="git-initialized" class="ml-2 text-sm text-vanna-navy">Git initialized</label></div> <div class="grid grid-cols-1 md:grid-cols-2 gap-4">`);
        FormField($$payload3, {
          label: "Setup Commands",
          id: "setup-commands",
          children: ($$payload4) => {
            Textarea($$payload4, {
              id: "setup-commands",
              placeholder: "npm install",
              rows: 3,
              get value() {
                return newDirectory.setupCommands;
              },
              set value($$value) {
                newDirectory.setupCommands = $$value;
                $$settled = false;
              }
            });
          }
        });
        $$payload3.out.push(`<!----> `);
        FormField($$payload3, {
          label: "Teardown Commands",
          id: "teardown-commands",
          children: ($$payload4) => {
            Textarea($$payload4, {
              id: "teardown-commands",
              placeholder: "npm run clean",
              rows: 3,
              get value() {
                return newDirectory.teardownCommands;
              },
              set value($$value) {
                newDirectory.teardownCommands = $$value;
                $$settled = false;
              }
            });
          }
        });
        $$payload3.out.push(`<!----> `);
        FormField($$payload3, {
          label: "Dev Server Setup Commands",
          id: "dev-server-setup",
          children: ($$payload4) => {
            Textarea($$payload4, {
              id: "dev-server-setup",
              placeholder: "npm run dev",
              rows: 3,
              get value() {
                return newDirectory.devServerSetupCommands;
              },
              set value($$value) {
                newDirectory.devServerSetupCommands = $$value;
                $$settled = false;
              }
            });
          }
        });
        $$payload3.out.push(`<!----> `);
        FormField($$payload3, {
          label: "Dev Server Teardown Commands",
          id: "dev-server-teardown",
          children: ($$payload4) => {
            Textarea($$payload4, {
              id: "dev-server-teardown",
              placeholder: "pkill -f 'npm run dev'",
              rows: 3,
              get value() {
                return newDirectory.devServerTeardownCommands;
              },
              set value($$value) {
                newDirectory.devServerTeardownCommands = $$value;
                $$settled = false;
              }
            });
          }
        });
        $$payload3.out.push(`<!----></div> <div class="flex gap-3 pt-4">`);
        Button($$payload3, {
          type: "submit",
          variant: "success",
          children: ($$payload4) => {
            $$payload4.out.push(`<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path></svg> Add Directory`);
          }
        });
        $$payload3.out.push(`<!----> `);
        Button($$payload3, {
          type: "button",
          variant: "secondary",
          onclick: () => showCreateDirectoryForm = false,
          children: ($$payload4) => {
            $$payload4.out.push(`<!---->Cancel`);
          }
        });
        $$payload3.out.push(`<!----></div></form>`);
      }
    });
    $$payload2.out.push(`<!----> `);
    if (project && (project.baseDirectories || []).length > 0) {
      $$payload2.out.push("<!--[-->");
      Card($$payload2, {
        children: ($$payload3) => {
          const each_array_2 = ensure_array_like(project.baseDirectories);
          $$payload3.out.push(`<h3 class="text-lg font-semibold text-vanna-navy mb-4 flex items-center gap-2"><svg class="w-5 h-5 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"></path></svg> Base Directories</h3> <div class="grid grid-cols-1 md:grid-cols-2 gap-4"><!--[-->`);
          for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
            let directory = each_array_2[$$index_2];
            $$payload3.out.push(`<div class="bg-vanna-cream/30 rounded-lg p-4 border border-slate-200"><div class="flex items-center justify-between mb-2"><div class="flex items-center gap-2"><svg class="w-4 h-4 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"></path></svg> <span class="font-mono text-sm text-vanna-teal break-all">${escape_html(directory.path)}</span> `);
            if (directory.gitInitialized) {
              $$payload3.out.push("<!--[-->");
              $$payload3.out.push(`<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-vanna-teal/10 text-vanna-teal">Git</span>`);
            } else {
              $$payload3.out.push("<!--[!-->");
            }
            $$payload3.out.push(`<!--]--></div> <div class="flex items-center gap-2">`);
            IconButton($$payload3, {
              onclick: () => startEditDirectory(directory),
              variant: "ghost",
              size: "sm",
              title: "Edit directory",
              children: ($$payload4) => {
                $$payload4.out.push(`<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path></svg>`);
              }
            });
            $$payload3.out.push(`<!----> `);
            IconButton($$payload3, {
              onclick: () => deleteDirectory(directory.base_directory_id),
              variant: "ghost",
              size: "sm",
              class: "text-vanna-orange hover:text-vanna-orange/80",
              title: "Delete directory",
              children: ($$payload4) => {
                $$payload4.out.push(`<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>`);
              }
            });
            $$payload3.out.push(`<!----></div></div> `);
            if (directory.setupCommands || directory.setup_commands || directory.devServerSetupCommands || directory.dev_server_setup_commands) {
              $$payload3.out.push("<!--[-->");
              $$payload3.out.push(`<div class="text-xs text-slate-500 space-y-1">`);
              if (directory.setupCommands || directory.setup_commands) {
                $$payload3.out.push("<!--[-->");
                $$payload3.out.push(`<div><strong>Setup:</strong> ${escape_html(directory.setupCommands || directory.setup_commands)}</div>`);
              } else {
                $$payload3.out.push("<!--[!-->");
              }
              $$payload3.out.push(`<!--]--> `);
              if (directory.devServerSetupCommands || directory.dev_server_setup_commands) {
                $$payload3.out.push("<!--[-->");
                $$payload3.out.push(`<div><strong>Dev Server:</strong> ${escape_html(directory.devServerSetupCommands || directory.dev_server_setup_commands)}</div>`);
              } else {
                $$payload3.out.push("<!--[!-->");
              }
              $$payload3.out.push(`<!--]--></div>`);
            } else {
              $$payload3.out.push("<!--[!-->");
            }
            $$payload3.out.push(`<!--]--> `);
            if (directoryGitStatus.has(directory.id)) {
              $$payload3.out.push("<!--[-->");
              const gitStatus = directoryGitStatus.get(directory.id);
              if (gitStatus.isRepo) {
                $$payload3.out.push("<!--[-->");
                $$payload3.out.push(`<div class="mt-3 pt-3 border-t border-slate-200"><div class="flex items-center justify-between"><div class="flex items-center gap-2 text-sm"><svg class="w-4 h-4 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14M5 12h14"></path></svg> <span class="text-vanna-navy">${escape_html(gitStatus.currentBranch || "main")}</span> `);
                if (gitStatus.ahead > 0) {
                  $$payload3.out.push("<!--[-->");
                  $$payload3.out.push(`<span class="text-xs bg-vanna-teal/10 text-vanna-teal px-1.5 py-0.5 rounded">${escape_html(gitStatus.ahead)} ahead</span>`);
                } else {
                  $$payload3.out.push("<!--[!-->");
                }
                $$payload3.out.push(`<!--]--> `);
                if (gitStatus.behind > 0) {
                  $$payload3.out.push("<!--[-->");
                  $$payload3.out.push(`<span class="text-xs bg-vanna-orange/10 text-vanna-orange px-1.5 py-0.5 rounded">${escape_html(gitStatus.behind)} behind</span>`);
                } else {
                  $$payload3.out.push("<!--[!-->");
                }
                $$payload3.out.push(`<!--]--></div> `);
                if (gitStatus.isDirty) {
                  $$payload3.out.push("<!--[-->");
                  $$payload3.out.push(`<a${attr("href", `/git/${stringify(directory.id)}`)} class="inline-flex items-center gap-1 px-2 py-1 text-xs font-medium text-vanna-orange bg-vanna-orange/10 rounded hover:bg-vanna-orange/20 transition-colors"><svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"></path></svg> ${escape_html((gitStatus.stagedFiles?.length || 0) + (gitStatus.unstagedFiles?.length || 0) + (gitStatus.untrackedFiles?.length || 0))} changes</a>`);
                } else {
                  $$payload3.out.push("<!--[!-->");
                  $$payload3.out.push(`<span class="text-xs text-vanna-teal">Clean</span>`);
                }
                $$payload3.out.push(`<!--]--></div></div>`);
              } else {
                $$payload3.out.push("<!--[!-->");
              }
              $$payload3.out.push(`<!--]-->`);
            } else {
              $$payload3.out.push("<!--[!-->");
            }
            $$payload3.out.push(`<!--]--> `);
            if (directory.devServerSetupCommands || directory.dev_server_setup_commands) {
              $$payload3.out.push("<!--[-->");
              $$payload3.out.push(`<div class="mt-3 pt-3 border-t border-slate-200"><div class="flex items-center justify-between"><div class="flex items-center gap-2 text-sm"><svg class="w-4 h-4 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"></path></svg> <span class="text-vanna-navy">Dev Server</span> `);
              if (devServerStatusMap.get(directory.id)?.running) {
                $$payload3.out.push("<!--[-->");
                $$payload3.out.push(`<span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-vanna-teal/10 text-vanna-teal"><span class="w-1.5 h-1.5 bg-vanna-teal rounded-full animate-pulse"></span> Running</span>`);
              } else {
                $$payload3.out.push("<!--[!-->");
                $$payload3.out.push(`<span class="text-xs text-slate-500">Stopped</span>`);
              }
              $$payload3.out.push(`<!--]--></div> <div class="flex items-center gap-2">`);
              if (devServerStatusMap.get(directory.id)?.running) {
                $$payload3.out.push("<!--[-->");
                $$payload3.out.push(`<a${attr("href", `/dev-server/${stringify(directory.id)}`)} class="inline-flex items-center gap-1 px-2 py-1 text-xs font-medium text-vanna-teal bg-vanna-teal/10 rounded hover:bg-vanna-teal/20 transition-colors"><svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path></svg> Terminal</a> <button${attr("disabled", stoppingDevServer.has(directory.id), true)} class="inline-flex items-center gap-1 px-2 py-1 text-xs font-medium text-vanna-orange bg-vanna-orange/10 rounded hover:bg-vanna-orange/20 transition-colors disabled:opacity-50">`);
                if (stoppingDevServer.has(directory.id)) {
                  $$payload3.out.push("<!--[-->");
                  $$payload3.out.push(`<div class="animate-spin rounded-full h-3 w-3 border-b border-current"></div>`);
                } else {
                  $$payload3.out.push("<!--[!-->");
                  $$payload3.out.push(`<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"></path></svg>`);
                }
                $$payload3.out.push(`<!--]--> Stop</button>`);
              } else {
                $$payload3.out.push("<!--[!-->");
                $$payload3.out.push(`<button${attr("disabled", startingDevServer.has(directory.id), true)} class="inline-flex items-center gap-1 px-2 py-1 text-xs font-medium text-vanna-teal bg-vanna-teal/10 rounded hover:bg-vanna-teal/20 transition-colors disabled:opacity-50">`);
                if (startingDevServer.has(directory.id)) {
                  $$payload3.out.push("<!--[-->");
                  $$payload3.out.push(`<div class="animate-spin rounded-full h-3 w-3 border-b border-current"></div>`);
                } else {
                  $$payload3.out.push("<!--[!-->");
                  $$payload3.out.push(`<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>`);
                }
                $$payload3.out.push(`<!--]--> Start</button>`);
              }
              $$payload3.out.push(`<!--]--></div></div></div>`);
            } else {
              $$payload3.out.push("<!--[!-->");
            }
            $$payload3.out.push(`<!--]--> `);
            if (editingDirectoryId === directory.base_directory_id) {
              $$payload3.out.push("<!--[-->");
              $$payload3.out.push(`<div class="mt-3 space-y-3 border-t border-slate-200 pt-3">`);
              FormField($$payload3, {
                label: "Path",
                id: `path-${directory.base_directory_id}`,
                children: ($$payload4) => {
                  Input($$payload4, {
                    id: `path-${directory.base_directory_id}`,
                    get value() {
                      return editDirectory.path;
                    },
                    set value($$value) {
                      editDirectory.path = $$value;
                      $$settled = false;
                    }
                  });
                }
              });
              $$payload3.out.push(`<!----> <div class="flex items-center gap-2"><input${attr("id", `git-${directory.base_directory_id}`)} type="checkbox"${attr("checked", editDirectory.gitInitialized, true)} class="w-4 h-4 text-vanna-teal bg-white border-slate-300 rounded focus:ring-vanna-teal focus:ring-2"/> <label${attr("for", `git-${directory.base_directory_id}`)} class="text-sm text-vanna-navy">Git initialized</label></div> `);
              FormField($$payload3, {
                label: "Setup Commands",
                id: `setup-${directory.base_directory_id}`,
                children: ($$payload4) => {
                  Textarea($$payload4, {
                    id: `setup-${directory.base_directory_id}`,
                    rows: 2,
                    get value() {
                      return editDirectory.setupCommands;
                    },
                    set value($$value) {
                      editDirectory.setupCommands = $$value;
                      $$settled = false;
                    }
                  });
                }
              });
              $$payload3.out.push(`<!----> `);
              FormField($$payload3, {
                label: "Teardown Commands",
                id: `teardown-${directory.base_directory_id}`,
                children: ($$payload4) => {
                  Textarea($$payload4, {
                    id: `teardown-${directory.base_directory_id}`,
                    rows: 2,
                    get value() {
                      return editDirectory.teardownCommands;
                    },
                    set value($$value) {
                      editDirectory.teardownCommands = $$value;
                      $$settled = false;
                    }
                  });
                }
              });
              $$payload3.out.push(`<!----> `);
              FormField($$payload3, {
                label: "Dev Server Setup Commands",
                id: `dev-setup-${directory.base_directory_id}`,
                children: ($$payload4) => {
                  Textarea($$payload4, {
                    id: `dev-setup-${directory.base_directory_id}`,
                    rows: 2,
                    get value() {
                      return editDirectory.devServerSetupCommands;
                    },
                    set value($$value) {
                      editDirectory.devServerSetupCommands = $$value;
                      $$settled = false;
                    }
                  });
                }
              });
              $$payload3.out.push(`<!----> `);
              FormField($$payload3, {
                label: "Dev Server Teardown Commands",
                id: `dev-teardown-${directory.base_directory_id}`,
                children: ($$payload4) => {
                  Textarea($$payload4, {
                    id: `dev-teardown-${directory.base_directory_id}`,
                    rows: 2,
                    get value() {
                      return editDirectory.devServerTeardownCommands;
                    },
                    set value($$value) {
                      editDirectory.devServerTeardownCommands = $$value;
                      $$settled = false;
                    }
                  });
                }
              });
              $$payload3.out.push(`<!----> <div class="flex gap-2">`);
              Button($$payload3, {
                variant: "primary",
                onclick: saveEditDirectory,
                children: ($$payload4) => {
                  $$payload4.out.push(`<!---->Save`);
                }
              });
              $$payload3.out.push(`<!----> `);
              Button($$payload3, {
                variant: "secondary",
                onclick: cancelEditDirectory,
                children: ($$payload4) => {
                  $$payload4.out.push(`<!---->Cancel`);
                }
              });
              $$payload3.out.push(`<!----></div></div>`);
            } else {
              $$payload3.out.push("<!--[!-->");
            }
            $$payload3.out.push(`<!--]--></div>`);
          }
          $$payload3.out.push(`<!--]--></div>`);
        }
      });
    } else {
      $$payload2.out.push("<!--[!-->");
    }
    $$payload2.out.push(`<!--]--> `);
    if (loading) {
      $$payload2.out.push("<!--[-->");
      $$payload2.out.push(`<div class="flex items-center justify-center py-12"><div class="animate-spin rounded-full h-8 w-8 border-b-2 border-vanna-teal"></div></div>`);
    } else {
      $$payload2.out.push("<!--[!-->");
      if (error) {
        $$payload2.out.push("<!--[-->");
        $$payload2.out.push(`<div class="text-center py-12"><div class="w-16 h-16 bg-vanna-orange/10 rounded-xl flex items-center justify-center mx-auto mb-4"><svg class="w-8 h-8 text-vanna-orange" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg></div> <h3 class="text-xl font-semibold text-vanna-navy mb-2">Failed to Load Project</h3> <p class="text-slate-500 mb-4">Unable to load project details</p></div>`);
      } else {
        $$payload2.out.push("<!--[!-->");
        if (project) {
          $$payload2.out.push("<!--[-->");
          {
            $$payload2.out.push("<!--[-->");
            $$payload2.out.push(`<div class="grid grid-cols-1 md:grid-cols-3 gap-6">`);
            Card($$payload2, {
              children: ($$payload3) => {
                const each_array_3 = ensure_array_like(todoTasks);
                $$payload3.out.push(`<div class="flex items-center gap-2 mb-4"><div class="w-3 h-3 rounded-full bg-slate-400"></div> <h2 class="font-semibold text-vanna-navy">To Do</h2> <span class="text-sm text-slate-500">(${escape_html(todoTasks.length)})</span></div> <div class="space-y-3"><!--[-->`);
                for (let $$index_5 = 0, $$length = each_array_3.length; $$index_5 < $$length; $$index_5++) {
                  let task = each_array_3[$$index_5];
                  const each_array_4 = ensure_array_like(columns);
                  $$payload3.out.push(`<div class="bg-white rounded-lg p-3 border border-slate-200 hover:border-vanna-teal/30 hover:shadow-md transition-all group"><div class="flex items-start justify-between mb-1"><button type="button" class="font-medium text-vanna-navy flex-1 cursor-pointer hover:text-vanna-teal transition-colors text-left">${escape_html(task.title)}</button> <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity ml-2">`);
                  IconButton($$payload3, {
                    onclick: () => openEditTaskModal(task),
                    variant: "ghost",
                    size: "xs",
                    class: "text-vanna-teal hover:text-vanna-teal/80",
                    title: "Edit task",
                    children: ($$payload4) => {
                      $$payload4.out.push(`<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path></svg>`);
                    }
                  });
                  $$payload3.out.push(`<!----> <div class="relative"><select${attr("disabled", updatingTasks.has(task.id), true)} class="bg-white border border-slate-300 rounded px-2 py-1 text-xs text-vanna-navy hover:bg-vanna-cream/30 transition-colors disabled:opacity-50" aria-label="Change task status">`);
                  $$payload3.select_value = task.status;
                  $$payload3.out.push(`<!--[-->`);
                  for (let $$index_3 = 0, $$length2 = each_array_4.length; $$index_3 < $$length2; $$index_3++) {
                    let col = each_array_4[$$index_3];
                    $$payload3.out.push(`<option${attr("value", col.id)}${maybe_selected($$payload3, col.id)}>${escape_html(col.title)}</option>`);
                  }
                  $$payload3.out.push(`<!--]-->`);
                  $$payload3.select_value = void 0;
                  $$payload3.out.push(`</select></div> `);
                  IconButton($$payload3, {
                    onclick: () => deleteTask(task),
                    disabled: deletingTasks.has(task.id),
                    variant: "ghost",
                    size: "xs",
                    class: "text-vanna-orange hover:text-vanna-orange/80",
                    title: "Delete task",
                    children: ($$payload4) => {
                      if (deletingTasks.has(task.id)) {
                        $$payload4.out.push("<!--[-->");
                        $$payload4.out.push(`<div class="animate-spin rounded-full h-3 w-3 border-b border-current"></div>`);
                      } else {
                        $$payload4.out.push("<!--[!-->");
                        $$payload4.out.push(`<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>`);
                      }
                      $$payload4.out.push(`<!--]-->`);
                    }
                  });
                  $$payload3.out.push(`<!----></div></div> <button type="button" class="cursor-pointer text-left w-full">`);
                  if (task.description) {
                    $$payload3.out.push("<!--[-->");
                    $$payload3.out.push(`<p class="text-sm text-slate-600 mb-2">${escape_html(task.description)}</p>`);
                  } else {
                    $$payload3.out.push("<!--[!-->");
                  }
                  $$payload3.out.push(`<!--]--> `);
                  if (taskExecutions.has(task.id) && taskExecutions.get(task.id).length > 0) {
                    $$payload3.out.push("<!--[-->");
                    const each_array_5 = ensure_array_like(taskExecutions.get(task.id));
                    $$payload3.out.push(`<div class="mb-2"><div class="flex flex-wrap gap-1"><!--[-->`);
                    for (let $$index_4 = 0, $$length2 = each_array_5.length; $$index_4 < $$length2; $$index_4++) {
                      let execution = each_array_5[$$index_4];
                      $$payload3.out.push(`<a${attr("href", `/task-executions/${stringify(execution.id)}`)} class="inline-flex items-center gap-1 bg-vanna-magenta/10 hover:bg-vanna-magenta/20 text-vanna-magenta px-2 py-1 rounded text-xs transition-colors"><svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg> ${escape_html(execution.agent_name)}</a>`);
                    }
                    $$payload3.out.push(`<!--]--></div></div>`);
                  } else {
                    $$payload3.out.push("<!--[!-->");
                  }
                  $$payload3.out.push(`<!--]--> <div class="text-xs text-slate-500 mt-2">📁 ${escape_html(task.baseDirectory?.path || "No base directory")}</div></button></div>`);
                }
                $$payload3.out.push(`<!--]--> `);
                if (todoTasks.length === 0) {
                  $$payload3.out.push("<!--[-->");
                  EmptyState($$payload3, {
                    children: ($$payload4) => {
                      $$payload4.out.push(`<svg class="w-8 h-8 mx-auto mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path></svg> <p class="text-sm">No tasks</p>`);
                    }
                  });
                } else {
                  $$payload3.out.push("<!--[!-->");
                }
                $$payload3.out.push(`<!--]--></div>`);
              }
            });
            $$payload2.out.push(`<!----> `);
            Card($$payload2, {
              children: ($$payload3) => {
                const each_array_6 = ensure_array_like(inProgressTasks);
                $$payload3.out.push(`<div class="flex items-center gap-2 mb-4"><div class="w-3 h-3 rounded-full bg-vanna-magenta"></div> <h2 class="font-semibold text-vanna-navy">In Progress</h2> <span class="text-sm text-slate-500">(${escape_html(inProgressTasks.length)})</span></div> <div class="space-y-3"><!--[-->`);
                for (let $$index_8 = 0, $$length = each_array_6.length; $$index_8 < $$length; $$index_8++) {
                  let task = each_array_6[$$index_8];
                  const each_array_7 = ensure_array_like(columns);
                  $$payload3.out.push(`<div class="bg-white rounded-lg p-3 border border-slate-200 hover:border-vanna-teal/30 hover:shadow-md transition-all group"><div class="flex items-start justify-between mb-1"><button type="button" class="font-medium text-vanna-navy flex-1 cursor-pointer hover:text-vanna-teal transition-colors text-left">${escape_html(task.title)}</button> <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity ml-2">`);
                  IconButton($$payload3, {
                    onclick: () => openEditTaskModal(task),
                    variant: "ghost",
                    size: "xs",
                    class: "text-vanna-teal hover:text-vanna-teal/80",
                    title: "Edit task",
                    children: ($$payload4) => {
                      $$payload4.out.push(`<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path></svg>`);
                    }
                  });
                  $$payload3.out.push(`<!----> <div class="relative"><select${attr("disabled", updatingTasks.has(task.id), true)} class="bg-white border border-slate-300 rounded px-2 py-1 text-xs text-vanna-navy hover:bg-vanna-cream/30 transition-colors disabled:opacity-50" aria-label="Change task status">`);
                  $$payload3.select_value = task.status;
                  $$payload3.out.push(`<!--[-->`);
                  for (let $$index_6 = 0, $$length2 = each_array_7.length; $$index_6 < $$length2; $$index_6++) {
                    let col = each_array_7[$$index_6];
                    $$payload3.out.push(`<option${attr("value", col.id)}${maybe_selected($$payload3, col.id)}>${escape_html(col.title)}</option>`);
                  }
                  $$payload3.out.push(`<!--]-->`);
                  $$payload3.select_value = void 0;
                  $$payload3.out.push(`</select></div> `);
                  IconButton($$payload3, {
                    onclick: () => deleteTask(task),
                    disabled: deletingTasks.has(task.id),
                    variant: "ghost",
                    size: "xs",
                    class: "text-vanna-orange hover:text-vanna-orange/80",
                    title: "Delete task",
                    children: ($$payload4) => {
                      if (deletingTasks.has(task.id)) {
                        $$payload4.out.push("<!--[-->");
                        $$payload4.out.push(`<div class="animate-spin rounded-full h-3 w-3 border-b border-current"></div>`);
                      } else {
                        $$payload4.out.push("<!--[!-->");
                        $$payload4.out.push(`<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>`);
                      }
                      $$payload4.out.push(`<!--]-->`);
                    }
                  });
                  $$payload3.out.push(`<!----></div></div> <button type="button" class="cursor-pointer text-left w-full">`);
                  if (task.description) {
                    $$payload3.out.push("<!--[-->");
                    $$payload3.out.push(`<p class="text-sm text-slate-600 mb-2">${escape_html(task.description)}</p>`);
                  } else {
                    $$payload3.out.push("<!--[!-->");
                  }
                  $$payload3.out.push(`<!--]--> `);
                  if (taskExecutions.has(task.id) && taskExecutions.get(task.id).length > 0) {
                    $$payload3.out.push("<!--[-->");
                    const each_array_8 = ensure_array_like(taskExecutions.get(task.id));
                    $$payload3.out.push(`<div class="mb-2"><div class="flex flex-wrap gap-1"><!--[-->`);
                    for (let $$index_7 = 0, $$length2 = each_array_8.length; $$index_7 < $$length2; $$index_7++) {
                      let execution = each_array_8[$$index_7];
                      $$payload3.out.push(`<a${attr("href", `/task-executions/${stringify(execution.id)}`)} class="inline-flex items-center gap-1 bg-vanna-magenta/10 hover:bg-vanna-magenta/20 text-vanna-magenta px-2 py-1 rounded text-xs transition-colors"><svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg> ${escape_html(execution.agent_name)}</a>`);
                    }
                    $$payload3.out.push(`<!--]--></div></div>`);
                  } else {
                    $$payload3.out.push("<!--[!-->");
                  }
                  $$payload3.out.push(`<!--]--> <div class="text-xs text-slate-500 mt-2">📁 ${escape_html(task.baseDirectory?.path || "No base directory")}</div></button></div>`);
                }
                $$payload3.out.push(`<!--]--> `);
                if (inProgressTasks.length === 0) {
                  $$payload3.out.push("<!--[-->");
                  EmptyState($$payload3, {
                    children: ($$payload4) => {
                      $$payload4.out.push(`<svg class="w-8 h-8 mx-auto mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path></svg> <p class="text-sm">No tasks</p>`);
                    }
                  });
                } else {
                  $$payload3.out.push("<!--[!-->");
                }
                $$payload3.out.push(`<!--]--></div>`);
              }
            });
            $$payload2.out.push(`<!----> <div class="flex flex-col gap-6">`);
            Card($$payload2, {
              children: ($$payload3) => {
                const each_array_9 = ensure_array_like(toVerifyTasks);
                $$payload3.out.push(`<div class="flex items-center gap-2 mb-4"><div class="w-3 h-3 rounded-full bg-vanna-orange"></div> <h2 class="font-semibold text-vanna-navy">To Verify</h2> <span class="text-sm text-slate-500">(${escape_html(toVerifyTasks.length)})</span></div> <div class="space-y-3"><!--[-->`);
                for (let $$index_11 = 0, $$length = each_array_9.length; $$index_11 < $$length; $$index_11++) {
                  let task = each_array_9[$$index_11];
                  const each_array_10 = ensure_array_like(columns);
                  $$payload3.out.push(`<div class="bg-white rounded-lg p-3 border border-slate-200 hover:border-vanna-teal/30 hover:shadow-md transition-all group"><div class="flex items-start justify-between mb-1"><button type="button" class="font-medium text-vanna-navy flex-1 cursor-pointer hover:text-vanna-teal transition-colors text-left">${escape_html(task.title)}</button> <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity ml-2">`);
                  IconButton($$payload3, {
                    onclick: () => openEditTaskModal(task),
                    variant: "ghost",
                    size: "xs",
                    class: "text-vanna-teal hover:text-vanna-teal/80",
                    title: "Edit task",
                    children: ($$payload4) => {
                      $$payload4.out.push(`<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path></svg>`);
                    }
                  });
                  $$payload3.out.push(`<!----> <div class="relative"><select${attr("disabled", updatingTasks.has(task.id), true)} class="bg-white border border-slate-300 rounded px-2 py-1 text-xs text-vanna-navy hover:bg-vanna-cream/30 transition-colors disabled:opacity-50" aria-label="Change task status">`);
                  $$payload3.select_value = task.status;
                  $$payload3.out.push(`<!--[-->`);
                  for (let $$index_9 = 0, $$length2 = each_array_10.length; $$index_9 < $$length2; $$index_9++) {
                    let col = each_array_10[$$index_9];
                    $$payload3.out.push(`<option${attr("value", col.id)}${maybe_selected($$payload3, col.id)}>${escape_html(col.title)}</option>`);
                  }
                  $$payload3.out.push(`<!--]-->`);
                  $$payload3.select_value = void 0;
                  $$payload3.out.push(`</select></div> `);
                  IconButton($$payload3, {
                    onclick: () => deleteTask(task),
                    disabled: deletingTasks.has(task.id),
                    variant: "ghost",
                    size: "xs",
                    class: "text-vanna-orange hover:text-vanna-orange/80",
                    title: "Delete task",
                    children: ($$payload4) => {
                      if (deletingTasks.has(task.id)) {
                        $$payload4.out.push("<!--[-->");
                        $$payload4.out.push(`<div class="animate-spin rounded-full h-3 w-3 border-b border-current"></div>`);
                      } else {
                        $$payload4.out.push("<!--[!-->");
                        $$payload4.out.push(`<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>`);
                      }
                      $$payload4.out.push(`<!--]-->`);
                    }
                  });
                  $$payload3.out.push(`<!----></div></div> <button type="button" class="cursor-pointer text-left w-full">`);
                  if (task.description) {
                    $$payload3.out.push("<!--[-->");
                    $$payload3.out.push(`<p class="text-sm text-slate-600 mb-2">${escape_html(task.description)}</p>`);
                  } else {
                    $$payload3.out.push("<!--[!-->");
                  }
                  $$payload3.out.push(`<!--]--> `);
                  if (taskExecutions.has(task.id) && taskExecutions.get(task.id).length > 0) {
                    $$payload3.out.push("<!--[-->");
                    const each_array_11 = ensure_array_like(taskExecutions.get(task.id));
                    $$payload3.out.push(`<div class="mb-2"><div class="flex flex-wrap gap-1"><!--[-->`);
                    for (let $$index_10 = 0, $$length2 = each_array_11.length; $$index_10 < $$length2; $$index_10++) {
                      let execution = each_array_11[$$index_10];
                      $$payload3.out.push(`<a${attr("href", `/task-executions/${stringify(execution.id)}`)} class="inline-flex items-center gap-1 bg-vanna-magenta/10 hover:bg-vanna-magenta/20 text-vanna-magenta px-2 py-1 rounded text-xs transition-colors"><svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg> ${escape_html(execution.agent_name)}</a>`);
                    }
                    $$payload3.out.push(`<!--]--></div></div>`);
                  } else {
                    $$payload3.out.push("<!--[!-->");
                  }
                  $$payload3.out.push(`<!--]--> <div class="text-xs text-slate-500 mt-2">📁 ${escape_html(task.baseDirectory?.path || "No base directory")}</div></button></div>`);
                }
                $$payload3.out.push(`<!--]--> `);
                if (toVerifyTasks.length === 0) {
                  $$payload3.out.push("<!--[-->");
                  EmptyState($$payload3, {
                    children: ($$payload4) => {
                      $$payload4.out.push(`<svg class="w-8 h-8 mx-auto mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path></svg> <p class="text-sm">No tasks</p>`);
                    }
                  });
                } else {
                  $$payload3.out.push("<!--[!-->");
                }
                $$payload3.out.push(`<!--]--></div>`);
              }
            });
            $$payload2.out.push(`<!----> `);
            Card($$payload2, {
              children: ($$payload3) => {
                const each_array_12 = ensure_array_like(doneTasks);
                $$payload3.out.push(`<div class="flex items-center gap-2 mb-4"><div class="w-3 h-3 rounded-full bg-vanna-teal"></div> <h2 class="font-semibold text-vanna-navy">Done</h2> <span class="text-sm text-slate-500">(${escape_html(doneTasks.length)})</span></div> <div class="space-y-3"><!--[-->`);
                for (let $$index_14 = 0, $$length = each_array_12.length; $$index_14 < $$length; $$index_14++) {
                  let task = each_array_12[$$index_14];
                  const each_array_13 = ensure_array_like(columns);
                  $$payload3.out.push(`<div class="bg-white rounded-lg p-3 border border-slate-200 hover:border-vanna-teal/30 hover:shadow-md transition-all group"><div class="flex items-start justify-between mb-1"><button type="button" class="font-medium text-vanna-navy flex-1 cursor-pointer hover:text-vanna-teal transition-colors text-left">${escape_html(task.title)}</button> <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity ml-2">`);
                  IconButton($$payload3, {
                    onclick: () => openEditTaskModal(task),
                    variant: "ghost",
                    size: "xs",
                    class: "text-vanna-teal hover:text-vanna-teal/80",
                    title: "Edit task",
                    children: ($$payload4) => {
                      $$payload4.out.push(`<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path></svg>`);
                    }
                  });
                  $$payload3.out.push(`<!----> <div class="relative"><select${attr("disabled", updatingTasks.has(task.id), true)} class="bg-white border border-slate-300 rounded px-2 py-1 text-xs text-vanna-navy hover:bg-vanna-cream/30 transition-colors disabled:opacity-50" aria-label="Change task status">`);
                  $$payload3.select_value = task.status;
                  $$payload3.out.push(`<!--[-->`);
                  for (let $$index_12 = 0, $$length2 = each_array_13.length; $$index_12 < $$length2; $$index_12++) {
                    let col = each_array_13[$$index_12];
                    $$payload3.out.push(`<option${attr("value", col.id)}${maybe_selected($$payload3, col.id)}>${escape_html(col.title)}</option>`);
                  }
                  $$payload3.out.push(`<!--]-->`);
                  $$payload3.select_value = void 0;
                  $$payload3.out.push(`</select></div> `);
                  IconButton($$payload3, {
                    onclick: () => deleteTask(task),
                    disabled: deletingTasks.has(task.id),
                    variant: "ghost",
                    size: "xs",
                    class: "text-vanna-orange hover:text-vanna-orange/80",
                    title: "Delete task",
                    children: ($$payload4) => {
                      if (deletingTasks.has(task.id)) {
                        $$payload4.out.push("<!--[-->");
                        $$payload4.out.push(`<div class="animate-spin rounded-full h-3 w-3 border-b border-current"></div>`);
                      } else {
                        $$payload4.out.push("<!--[!-->");
                        $$payload4.out.push(`<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>`);
                      }
                      $$payload4.out.push(`<!--]-->`);
                    }
                  });
                  $$payload3.out.push(`<!----></div></div> <button type="button" class="cursor-pointer text-left w-full">`);
                  if (task.description) {
                    $$payload3.out.push("<!--[-->");
                    $$payload3.out.push(`<p class="text-sm text-slate-600 mb-2">${escape_html(task.description)}</p>`);
                  } else {
                    $$payload3.out.push("<!--[!-->");
                  }
                  $$payload3.out.push(`<!--]--> `);
                  if (taskExecutions.has(task.id) && taskExecutions.get(task.id).length > 0) {
                    $$payload3.out.push("<!--[-->");
                    const each_array_14 = ensure_array_like(taskExecutions.get(task.id));
                    $$payload3.out.push(`<div class="mb-2"><div class="flex flex-wrap gap-1"><!--[-->`);
                    for (let $$index_13 = 0, $$length2 = each_array_14.length; $$index_13 < $$length2; $$index_13++) {
                      let execution = each_array_14[$$index_13];
                      $$payload3.out.push(`<a${attr("href", `/task-executions/${stringify(execution.id)}`)} class="inline-flex items-center gap-1 bg-vanna-magenta/10 hover:bg-vanna-magenta/20 text-vanna-magenta px-2 py-1 rounded text-xs transition-colors"><svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg> ${escape_html(execution.agent_name)}</a>`);
                    }
                    $$payload3.out.push(`<!--]--></div></div>`);
                  } else {
                    $$payload3.out.push("<!--[!-->");
                  }
                  $$payload3.out.push(`<!--]--> <div class="text-xs text-slate-500 mt-2">📁 ${escape_html(task.baseDirectory?.path || "No base directory")}</div></button></div>`);
                }
                $$payload3.out.push(`<!--]--> `);
                if (doneTasks.length === 0) {
                  $$payload3.out.push("<!--[-->");
                  EmptyState($$payload3, {
                    children: ($$payload4) => {
                      $$payload4.out.push(`<svg class="w-8 h-8 mx-auto mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path></svg> <p class="text-sm">No tasks</p>`);
                    }
                  });
                } else {
                  $$payload3.out.push("<!--[!-->");
                }
                $$payload3.out.push(`<!--]--></div>`);
              }
            });
            $$payload2.out.push(`<!----></div></div>`);
          }
          $$payload2.out.push(`<!--]-->`);
        } else {
          $$payload2.out.push("<!--[!-->");
        }
        $$payload2.out.push(`<!--]-->`);
      }
      $$payload2.out.push(`<!--]-->`);
    }
    $$payload2.out.push(`<!--]--></div> `);
    Modal($$payload2, {
      open: showTaskModal && selectedTask,
      title: "Task Details",
      size: "xl",
      onClose: closeTaskModal,
      children: ($$payload3) => {
        if (selectedTask) {
          $$payload3.out.push("<!--[-->");
          $$payload3.out.push(`<div class="space-y-4"><div><h3 class="text-lg font-medium text-vanna-navy mb-2">${escape_html(selectedTask.title)}</h3> <div class="flex items-center gap-2 mb-3">`);
          StatusBadge($$payload3, { status: selectedTask.status || "pending" });
          $$payload3.out.push(`<!----></div></div> `);
          if (selectedTask.description) {
            $$payload3.out.push("<!--[-->");
            $$payload3.out.push(`<div><h4 class="text-sm font-medium text-vanna-navy mb-2">Description</h4> <p class="text-slate-600">${escape_html(selectedTask.description)}</p></div>`);
          } else {
            $$payload3.out.push("<!--[!-->");
          }
          $$payload3.out.push(`<!--]--> <div><h4 class="text-sm font-medium text-vanna-navy mb-2">Base Directory</h4> <div class="bg-vanna-cream/30 rounded-lg p-3 border border-slate-200"><div class="flex items-center gap-2 mb-2"><svg class="w-4 h-4 text-vanna-teal" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H7.5L5 5H3v2z"></path></svg> <span class="font-mono text-sm text-vanna-teal">${escape_html(selectedTask.baseDirectory?.path || "No path")}</span> `);
          if (selectedTask.baseDirectory?.git_initialized) {
            $$payload3.out.push("<!--[-->");
            $$payload3.out.push(`<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-vanna-teal/10 text-vanna-teal">Git</span>`);
          } else {
            $$payload3.out.push("<!--[!-->");
          }
          $$payload3.out.push(`<!--]--></div> `);
          if (selectedTask.baseDirectory?.setup_commands || selectedTask.baseDirectory?.dev_server_setup_commands) {
            $$payload3.out.push("<!--[-->");
            $$payload3.out.push(`<div class="text-xs text-slate-500 space-y-1">`);
            if (selectedTask.baseDirectory?.setup_commands) {
              $$payload3.out.push("<!--[-->");
              $$payload3.out.push(`<div><strong>Setup:</strong> ${escape_html(selectedTask.baseDirectory.setup_commands)}</div>`);
            } else {
              $$payload3.out.push("<!--[!-->");
            }
            $$payload3.out.push(`<!--]--> `);
            if (selectedTask.baseDirectory?.dev_server_setup_commands) {
              $$payload3.out.push("<!--[-->");
              $$payload3.out.push(`<div><strong>Dev Server:</strong> ${escape_html(selectedTask.baseDirectory.dev_server_setup_commands)}</div>`);
            } else {
              $$payload3.out.push("<!--[!-->");
            }
            $$payload3.out.push(`<!--]--></div>`);
          } else {
            $$payload3.out.push("<!--[!-->");
          }
          $$payload3.out.push(`<!--]--></div></div> `);
          if (taskExecutions.has(selectedTask.id) && taskExecutions.get(selectedTask.id).length > 0) {
            $$payload3.out.push("<!--[-->");
            const each_array_19 = ensure_array_like(taskExecutions.get(selectedTask.id));
            $$payload3.out.push(`<div><h4 class="text-sm font-medium text-vanna-navy mb-3">Active Executions</h4> <div class="space-y-2"><!--[-->`);
            for (let $$index_19 = 0, $$length = each_array_19.length; $$index_19 < $$length; $$index_19++) {
              let execution = each_array_19[$$index_19];
              $$payload3.out.push(`<div class="bg-vanna-orange/5 rounded-lg p-3 border border-vanna-orange/30"><div class="flex items-center justify-between"><div class="flex items-center gap-2"><svg class="w-4 h-4 text-vanna-orange" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3"></path></svg> <span class="text-vanna-navy font-medium">${escape_html(execution.agent_name)}</span> `);
              StatusBadge($$payload3, { status: execution.status || "pending", size: "sm" });
              $$payload3.out.push(`<!----></div> `);
              Button($$payload3, {
                href: `/task-executions/${stringify(execution.id)}`,
                variant: "primary",
                size: "xs",
                children: ($$payload4) => {
                  $$payload4.out.push(`<!---->View Execution`);
                }
              });
              $$payload3.out.push(`<!----></div> <div class="text-xs text-slate-500 mt-2">Started: ${escape_html(new Date(execution.created_at.Time).toLocaleString())}</div></div>`);
            }
            $$payload3.out.push(`<!--]--></div></div>`);
          } else {
            $$payload3.out.push("<!--[!-->");
          }
          $$payload3.out.push(`<!--]--> <div class="pt-4 border-t border-slate-200"><h4 class="text-sm font-medium text-vanna-navy mb-3">Execute Task</h4> `);
          if (availableAgents.length > 0) {
            $$payload3.out.push("<!--[-->");
            const each_array_20 = ensure_array_like(availableAgents);
            $$payload3.out.push(`<div class="bg-vanna-cream/30 rounded-lg p-4"><p class="text-slate-600 mb-3">Select agents to execute this task:</p> <div class="space-y-2 mb-4 max-h-40 overflow-y-auto"><!--[-->`);
            for (let $$index_20 = 0, $$length = each_array_20.length; $$index_20 < $$length; $$index_20++) {
              let agent = each_array_20[$$index_20];
              $$payload3.out.push(`<label class="flex items-center gap-3 p-2 rounded hover:bg-vanna-cream/50 cursor-pointer"><input type="checkbox"${attr("checked", selectedAgents.some((a) => a.id === agent.id), true)} class="w-4 h-4 text-vanna-teal bg-white border-slate-300 rounded focus:ring-vanna-teal focus:ring-2"/> <div class="flex-1"><div class="font-medium text-vanna-navy">${escape_html(agent.name)}</div> <div class="text-sm text-slate-500 font-mono">${escape_html(agent.command)} ${escape_html(agent.params)}</div></div></label>`);
            }
            $$payload3.out.push(`<!--]--></div> `);
            if (selectedAgents.length > 0) {
              $$payload3.out.push("<!--[-->");
              $$payload3.out.push(`<div class="text-sm text-slate-600 mb-3">Selected: ${escape_html(selectedAgents.map((a) => a.name).join(", "))}</div>`);
            } else {
              $$payload3.out.push("<!--[!-->");
            }
            $$payload3.out.push(`<!--]--> `);
            Button($$payload3, {
              onclick: startTaskExecution,
              disabled: selectedAgents.length === 0 || executingTask,
              loading: executingTask,
              variant: "primary",
              class: "w-full",
              children: ($$payload4) => {
                $$payload4.out.push(`<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h1m4 0h1m-6 4h1m4 0h1m-6-8h1m4 0h1m-2-4h.01M12 16h.01M12 8h.01M12 12h.01"></path></svg>${escape_html(executingTask ? "Starting Execution..." : `Start Execution (${selectedAgents.length})`)}`);
              }
            });
            $$payload3.out.push(`<!----></div>`);
          } else {
            $$payload3.out.push("<!--[!-->");
            $$payload3.out.push(`<div class="bg-vanna-cream/30 rounded-lg p-4 text-center"><p class="text-slate-600 mb-3">No agents configured</p> `);
            Button($$payload3, {
              href: "/agents",
              variant: "warning",
              children: ($$payload4) => {
                $$payload4.out.push(`<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path></svg> Configure Agents`);
              }
            });
            $$payload3.out.push(`<!----></div>`);
          }
          $$payload3.out.push(`<!--]--></div></div>`);
        } else {
          $$payload3.out.push("<!--[!-->");
        }
        $$payload3.out.push(`<!--]-->`);
      }
    });
    $$payload2.out.push(`<!----> `);
    Modal($$payload2, {
      open: showEditTaskModal && editingTask,
      title: "Edit Task",
      size: "lg",
      onClose: closeEditTaskModal,
      children: ($$payload3) => {
        if (editingTask) {
          $$payload3.out.push("<!--[-->");
          $$payload3.out.push(`<form class="space-y-6">`);
          FormField($$payload3, {
            label: "Task Title",
            id: "edit-task-title",
            required: true,
            children: ($$payload4) => {
              Input($$payload4, {
                id: "edit-task-title",
                type: "text",
                placeholder: "Enter task title",
                required: true,
                get value() {
                  return editTaskForm.title;
                },
                set value($$value) {
                  editTaskForm.title = $$value;
                  $$settled = false;
                }
              });
            }
          });
          $$payload3.out.push(`<!----> `);
          FormField($$payload3, {
            label: "Description",
            id: "edit-task-description",
            children: ($$payload4) => {
              Textarea($$payload4, {
                id: "edit-task-description",
                placeholder: "Enter task description",
                rows: 4,
                get value() {
                  return editTaskForm.description;
                },
                set value($$value) {
                  editTaskForm.description = $$value;
                  $$settled = false;
                }
              });
            }
          });
          $$payload3.out.push(`<!----> `);
          FormField($$payload3, {
            label: "Status",
            id: "edit-task-status",
            children: ($$payload4) => {
              Select($$payload4, {
                id: "edit-task-status",
                get value() {
                  return editTaskForm.status;
                },
                set value($$value) {
                  editTaskForm.status = $$value;
                  $$settled = false;
                },
                children: ($$payload5) => {
                  const each_array_21 = ensure_array_like(columns);
                  $$payload5.out.push(`<!--[-->`);
                  for (let $$index_21 = 0, $$length = each_array_21.length; $$index_21 < $$length; $$index_21++) {
                    let column = each_array_21[$$index_21];
                    $$payload5.out.push(`<option${attr("value", column.id)}${maybe_selected($$payload5, column.id)}>${escape_html(column.title)}</option>`);
                  }
                  $$payload5.out.push(`<!--]-->`);
                },
                $$slots: { default: true }
              });
            }
          });
          $$payload3.out.push(`<!----> `);
          FormField($$payload3, {
            label: "Base Directory",
            id: "edit-task-base-directory",
            required: true,
            children: ($$payload4) => {
              Select($$payload4, {
                id: "edit-task-base-directory",
                placeholder: "Select a base directory...",
                required: true,
                get value() {
                  return editTaskForm.baseDirectoryId;
                },
                set value($$value) {
                  editTaskForm.baseDirectoryId = $$value;
                  $$settled = false;
                },
                children: ($$payload5) => {
                  if (project && project.baseDirectories) {
                    $$payload5.out.push("<!--[-->");
                    const each_array_22 = ensure_array_like(project.baseDirectories);
                    $$payload5.out.push(`<!--[-->`);
                    for (let $$index_22 = 0, $$length = each_array_22.length; $$index_22 < $$length; $$index_22++) {
                      let directory = each_array_22[$$index_22];
                      $$payload5.out.push(`<option${attr("value", directory.base_directory_id)}${maybe_selected($$payload5, directory.base_directory_id)}>${escape_html(directory.path)}</option>`);
                    }
                    $$payload5.out.push(`<!--]-->`);
                  } else {
                    $$payload5.out.push("<!--[!-->");
                  }
                  $$payload5.out.push(`<!--]-->`);
                },
                $$slots: { default: true }
              });
            }
          });
          $$payload3.out.push(`<!----> <div class="flex gap-3 pt-4">`);
          Button($$payload3, {
            type: "submit",
            disabled: updatingTasks.has(editingTask.id),
            loading: updatingTasks.has(editingTask.id),
            variant: "primary",
            children: ($$payload4) => {
              $$payload4.out.push(`<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>${escape_html(updatingTasks.has(editingTask.id) ? "Saving..." : "Save Changes")}`);
            }
          });
          $$payload3.out.push(`<!----> `);
          Button($$payload3, {
            type: "button",
            onclick: closeEditTaskModal,
            variant: "secondary",
            children: ($$payload4) => {
              $$payload4.out.push(`<!---->Cancel`);
            }
          });
          $$payload3.out.push(`<!----></div></form>`);
        } else {
          $$payload3.out.push("<!--[!-->");
        }
        $$payload3.out.push(`<!--]-->`);
      }
    });
    $$payload2.out.push(`<!---->`);
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
