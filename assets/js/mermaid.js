// Renders ```mermaid code blocks. Only included on pages that have one.
import mermaid from "mermaid";

const dark = window.matchMedia("(prefers-color-scheme: dark)");

const palette = (isDark) => isDark
  ? { bg: "#141b16", fg: "#dde3d6", node: "#1b241e", line: "#5f7f5a", accent: "#9cc283", note: "#2a2418" }
  : { bg: "#f5f2e9", fg: "#1f2a22", node: "#ece7da", line: "#7c9a6a", accent: "#3f6b3a", note: "#f1e6d4" };

const blocks = [...document.querySelectorAll("pre.mermaid")];
const sources = blocks.map((el) => el.textContent);

async function render() {
  const c = palette(dark.matches);
  mermaid.initialize({
    startOnLoad: false,
    theme: "base",
    darkMode: dark.matches,
    fontFamily: '"Fira Sans", system-ui, sans-serif',
    themeVariables: {
      background: c.bg,
      primaryColor: c.node,
      primaryTextColor: c.fg,
      primaryBorderColor: c.accent,
      secondaryColor: c.node,
      tertiaryColor: c.bg,
      lineColor: c.line,
      textColor: c.fg,
      noteBkgColor: c.note,
      noteTextColor: c.fg,
    },
  });
  blocks.forEach((el, i) => {
    el.removeAttribute("data-processed");
    el.textContent = sources[i];
  });
  await mermaid.run({ nodes: blocks });
}

render();
dark.addEventListener("change", render);
