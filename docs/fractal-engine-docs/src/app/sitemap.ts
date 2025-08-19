import type { MetadataRoute } from "next";
import { source } from "@/lib/source";

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const now = new Date();

  const staticRoutes: MetadataRoute.Sitemap = [
    {
      url: new URL("/", siteUrl).toString(),
      lastModified: now,
      changeFrequency: "weekly",
      priority: 1,
    },
    {
      url: new URL("/docs", siteUrl).toString(),
      lastModified: now,
      changeFrequency: "weekly",
      priority: 0.8,
    },
  ];

  const params = await source.generateParams();
  const docRoutes: MetadataRoute.Sitemap = params.map(({ slug }) => {
    const path =
      Array.isArray(slug) && slug.length > 0
        ? `/docs/${slug.join("/")}`
        : "/docs";
    return {
      url: new URL(path, siteUrl).toString(),
      lastModified: now,
      changeFrequency: "weekly",
      priority: 0.6,
    };
  });

  // Deduplicate by URL (in case '/docs' appears multiple times)
  const byUrl = new Map<string, MetadataRoute.Sitemap[number]>();
  [...staticRoutes, ...docRoutes].forEach((entry) =>
    byUrl.set(entry.url, entry)
  );

  return Array.from(byUrl.values());
}
