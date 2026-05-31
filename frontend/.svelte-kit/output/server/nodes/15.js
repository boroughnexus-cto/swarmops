import * as universal from '../entries/pages/swarm/_session_/_page.js';

export const index = 15;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/swarm/_session_/_page.svelte.js')).default;
export { universal };
export const universal_id = "src/routes/swarm/[session]/+page.js";
export const imports = ["_app/immutable/nodes/15.DxKM06La.js","_app/immutable/chunks/DsnmJJEf.js","_app/immutable/chunks/CEjoAh5C.js","_app/immutable/chunks/CDGuArWT.js","_app/immutable/chunks/DcZzMn2F.js","_app/immutable/chunks/C0Z1ZdgU.js","_app/immutable/chunks/B5PWudQi.js","_app/immutable/chunks/D7W0yNa4.js","_app/immutable/chunks/BvYyoKJp.js","_app/immutable/chunks/C7n9xuSX.js","_app/immutable/chunks/CwPTCi79.js","_app/immutable/chunks/Bmg_b7yX.js","_app/immutable/chunks/DskCAEaq.js","_app/immutable/chunks/D9Z9MdNV.js","_app/immutable/chunks/B-iO80yE.js"];
export const stylesheets = [];
export const fonts = [];
