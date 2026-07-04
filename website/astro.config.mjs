import { defineConfig } from "astro/config";
import tailwindcss from "@tailwindcss/vite";
import { visit } from "unist-util-visit";

// Docs are authored with relative .md links (they must keep working on
// GitHub). Pages are served at /docs/<id>/ — one directory level deeper
// than the source file — so rewrite `foo.md` links to `../foo/`.
function remarkRewriteMdLinks() {
  return (tree) => {
    visit(tree, "link", (node) => {
      if (/^[a-z]+:/i.test(node.url)) return;
      const match = node.url.match(/^(.*)\.md(#.*)?$/);
      if (!match) return;
      const [, path, hash = ""] = match;
      node.url = "../" + path.replace(/^\.\//, "") + "/" + hash;
    });
  };
}

export default defineConfig({
  site: "https://natonathan.github.io",
  base: "/shipyard",
  vite: {
    plugins: [tailwindcss()],
  },
  markdown: {
    remarkPlugins: [remarkRewriteMdLinks],
    shikiConfig: {
      theme: "night-owl",
    },
  },
});
