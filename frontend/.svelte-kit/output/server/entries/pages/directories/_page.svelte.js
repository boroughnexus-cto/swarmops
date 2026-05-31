import { O as head, G as attr, J as escape_html, D as pop, A as push } from "../../../chunks/index2.js";
function _page($$payload, $$props) {
  push();
  let dirtyCount;
  let directories = [];
  let gitStatusMap = /* @__PURE__ */ new Map();
  let showDirtyOnly = false;
  let groupByProject = false;
  function getChangeCount(status) {
    if (!status) return 0;
    return (status.stagedFiles?.length || 0) + (status.unstagedFiles?.length || 0) + (status.untrackedFiles?.length || 0);
  }
  function isDirty(status) {
    return getChangeCount(status) > 0;
  }
  dirtyCount = directories.filter((dir) => isDirty(gitStatusMap.get(dir.id))).length;
  head($$payload, ($$payload2) => {
    $$payload2.title = `<title>Directories - SwarmOps</title>`;
  });
  $$payload.out.push(`<div class="space-y-6"><div><h1 class="text-3xl font-bold text-vanna-navy font-serif">Directories</h1> <p class="mt-2 text-slate-500">View git status across all projects</p></div> <div class="flex items-center gap-4"><label class="flex items-center gap-2 cursor-pointer"><input type="checkbox"${attr("checked", showDirtyOnly, true)} class="rounded border-slate-300 text-vanna-teal focus:ring-vanna-teal"/> <span class="text-sm text-slate-600">Show only dirty (${escape_html(dirtyCount)})</span></label> <label class="flex items-center gap-2 cursor-pointer"><input type="checkbox"${attr("checked", groupByProject, true)} class="rounded border-slate-300 text-vanna-teal focus:ring-vanna-teal"/> <span class="text-sm text-slate-600">Group by project</span></label></div> `);
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
