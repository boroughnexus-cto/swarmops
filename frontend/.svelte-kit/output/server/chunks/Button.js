import { G as attr, F as attr_class, T as clsx } from "./index2.js";
function Button($$payload, $$props) {
  let {
    variant = "primary",
    size = "md",
    disabled = false,
    loading = false,
    class: className = "",
    href,
    type = "button",
    onclick,
    children
  } = $$props;
  const baseClasses = "inline-flex items-center justify-center font-medium rounded-full transition-all focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed";
  const variantClasses = {
    primary: "bg-vanna-teal hover:bg-vanna-teal/90 text-white shadow-vanna-subtle focus:ring-vanna-teal",
    secondary: "bg-vanna-cream hover:bg-vanna-cream/80 text-vanna-navy border border-vanna-teal/30 focus:ring-vanna-teal",
    success: "bg-vanna-teal hover:bg-vanna-teal/90 text-white focus:ring-vanna-teal",
    danger: "bg-vanna-orange hover:bg-vanna-orange/90 text-white focus:ring-vanna-orange",
    warning: "bg-vanna-orange hover:bg-vanna-orange/90 text-white focus:ring-vanna-orange",
    info: "bg-vanna-teal hover:bg-vanna-teal/90 text-white focus:ring-vanna-teal",
    ghost: "bg-transparent hover:bg-vanna-cream/50 text-vanna-navy focus:ring-vanna-teal",
    outline: "border border-vanna-navy/30 bg-transparent hover:bg-vanna-cream/30 text-vanna-navy focus:ring-vanna-teal"
  };
  const sizeClasses = {
    xs: "px-2.5 py-1.5 text-xs h-7",
    sm: "px-3 py-2 text-sm h-9",
    md: "px-4 py-2 text-sm h-11",
    lg: "px-4 py-2 text-base h-12",
    xl: "px-6 py-3 text-base h-14"
  };
  const classes = `${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]} ${className}`;
  if (href) {
    $$payload.out.push("<!--[-->");
    $$payload.out.push(`<a${attr("href", href)}${attr_class(clsx(classes))}>`);
    if (loading) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--> `);
    children?.($$payload);
    $$payload.out.push(`<!----></a>`);
  } else {
    $$payload.out.push("<!--[!-->");
    $$payload.out.push(`<button${attr("type", type)}${attr("disabled", disabled, true)}${attr_class(clsx(classes))}>`);
    if (loading) {
      $$payload.out.push("<!--[-->");
      $$payload.out.push(`<svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>`);
    } else {
      $$payload.out.push("<!--[!-->");
    }
    $$payload.out.push(`<!--]--> `);
    children?.($$payload);
    $$payload.out.push(`<!----></button>`);
  }
  $$payload.out.push(`<!--]-->`);
}
export {
  Button as B
};
