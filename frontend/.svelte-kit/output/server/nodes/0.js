import * as universal from '../entries/pages/_layout.ts.js';

export const index = 0;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/_layout.svelte.js')).default;
export { universal };
export const universal_id = "src/routes/+layout.ts";
export const imports = ["_app/immutable/nodes/0.CN01ejZH.js","_app/immutable/chunks/DsnmJJEf.js","_app/immutable/chunks/CEjoAh5C.js","_app/immutable/chunks/CDGuArWT.js","_app/immutable/chunks/DcZzMn2F.js","_app/immutable/chunks/B5PWudQi.js","_app/immutable/chunks/BvYyoKJp.js","_app/immutable/chunks/C0Z1ZdgU.js","_app/immutable/chunks/C7n9xuSX.js","_app/immutable/chunks/CwPTCi79.js","_app/immutable/chunks/yg1RgZC-.js"];
export const stylesheets = ["_app/immutable/assets/0.BsdtvrFJ.css"];
export const fonts = [];
