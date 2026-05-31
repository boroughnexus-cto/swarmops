import * as universal from '../entries/pages/terminal/_session_/_page.js';

export const index = 20;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/terminal/_session_/_page.svelte.js')).default;
export { universal };
export const universal_id = "src/routes/terminal/[session]/+page.js";
export const imports = ["_app/immutable/nodes/20.Ekw7F-v2.js","_app/immutable/chunks/DsnmJJEf.js","_app/immutable/chunks/BkAhmUI_.js","_app/immutable/chunks/CDGuArWT.js","_app/immutable/chunks/CEjoAh5C.js","_app/immutable/chunks/DcZzMn2F.js","_app/immutable/chunks/C9jOwEcz.js","_app/immutable/chunks/B5PWudQi.js","_app/immutable/chunks/Bmg_b7yX.js","_app/immutable/chunks/D9vBqKUr.js","_app/immutable/chunks/BvYyoKJp.js","_app/immutable/chunks/C7n9xuSX.js","_app/immutable/chunks/CwPTCi79.js","_app/immutable/chunks/b6D2bu8U.js","_app/immutable/chunks/C0Z1ZdgU.js"];
export const stylesheets = [];
export const fonts = [];
