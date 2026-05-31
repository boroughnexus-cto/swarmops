import { F as attr_class, D as pop, A as push, I as stringify } from "./index2.js";
function Card($$payload, $$props) {
  push();
  let {
    class: className = "",
    padding = "md",
    shadow = "vanna",
    border = true,
    rounded = "2xl",
    hover = false,
    onclick,
    children
  } = $$props;
  const paddingClasses = { none: "", sm: "p-3", md: "p-6", lg: "p-8" };
  const shadowClasses = {
    none: "",
    sm: "shadow-sm",
    md: "shadow-md",
    lg: "shadow-lg",
    vanna: "shadow-vanna-card"
  };
  const roundedClasses = {
    none: "",
    sm: "rounded-sm",
    md: "rounded-md",
    lg: "rounded-lg",
    xl: "rounded-xl",
    "2xl": "rounded-2xl"
  };
  const hoverClasses = hover ? "hover:-translate-y-1 hover:shadow-vanna-feature transition-all duration-200 cursor-pointer" : "";
  if (onclick) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<div${attr_class(`bg-white/80 backdrop-blur-sm ${stringify(border ? "border border-slate-200/60" : "")} ${stringify(shadowClasses[shadow])} ${stringify(roundedClasses[rounded])} ${stringify(paddingClasses[padding])} ${stringify(hoverClasses)} ${stringify(className)}`)} role="button" tabindex="0">`);
    children?.($$payload);
    $$payload.out.push(`<!----></div>`);
  } else {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`<div${attr_class(`bg-white/80 backdrop-blur-sm ${stringify(border ? "border border-slate-200/60" : "")} ${stringify(shadowClasses[shadow])} ${stringify(roundedClasses[rounded])} ${stringify(paddingClasses[padding])} ${stringify(hover ? hoverClasses : "")} ${stringify(className)}`)}>`);
    children?.($$payload);
    $$payload.out.push(`<!----></div>`);
  }
  $$payload.out.push(`<!--]-->`);
  pop();
}
export {
  Card as C
};
