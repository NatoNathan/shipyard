import { getCollection, type CollectionEntry } from "astro:content";

export type DocEntry = CollectionEntry<"docs">;

/** Pull the display title from the doc's first `# ` heading. */
export function docTitle(entry: DocEntry): string {
  const match = entry.body?.match(/^#\s+(.+)$/m);
  if (match) return match[1].replace(/`/g, "").trim();
  return entry.id.split("/").pop()!.replace(/-/g, " ");
}

export interface SidebarGroup {
  label: string;
  items: { id: string; title: string }[];
}

const GUIDE_ORDER = ["configuration", "consignment-format", "tag-generation"];

export async function sidebarGroups(): Promise<SidebarGroup[]> {
  const entries = await getCollection("docs");

  const guides = entries
    .filter((e) => !e.id.startsWith("reference/"))
    .sort((a, b) => {
      const ai = GUIDE_ORDER.indexOf(a.id);
      const bi = GUIDE_ORDER.indexOf(b.id);
      return (ai === -1 ? 99 : ai) - (bi === -1 ? 99 : bi);
    });

  const reference = entries
    .filter((e) => e.id.startsWith("reference/"))
    .sort((a, b) => a.id.localeCompare(b.id));

  return [
    {
      label: "Guides",
      items: guides.map((e) => ({ id: e.id, title: docTitle(e) })),
    },
    {
      label: "Command Reference",
      items: reference.map((e) => ({
        id: e.id,
        title: "shipyard " + e.id.replace("reference/", ""),
      })),
    },
  ];
}
