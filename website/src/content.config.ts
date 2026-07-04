import { defineCollection } from "astro:content";
import { glob } from "astro/loaders";

// Docs live in the repo's top-level docs/ folder so they stay
// single-sourced between GitHub and the website.
const docs = defineCollection({
  loader: glob({ pattern: "**/*.md", base: "../docs" }),
});

export const collections = { docs };
