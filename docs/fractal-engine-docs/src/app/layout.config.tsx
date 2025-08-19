import { GithubInfo } from "@/components/GithubInfo";
import type { BaseLayoutProps } from "fumadocs-ui/layouts/shared";

/**
 * Shared layout configurations
 *
 * you can customise layouts individually from:
 * Home Layout: app/(home)/layout.tsx
 * Docs Layout: app/docs/layout.tsx
 */
export const baseOptions: BaseLayoutProps = {
  themeSwitch: { enabled: false },
  nav: {
    title: (
      <>
        <svg
          width="24"
          height="24"
          xmlns="http://www.w3.org/2000/svg"
          aria-label="Logo"
        >
          <circle cx={12} cy={12} r={12} fill="currentColor" />
        </svg>
        Fractal Engine
      </>
    ),
  },
  // see https://fumadocs.dev/docs/ui/navigation/links
  links: [
    {
      type: 'custom',
      children: (
        <GithubInfo owner="dogecoinfoundation" repo="fractal-engine" className="lg:-mx-2" />
      ),
    },
    {
      type: 'custom',
      children: (
        <GithubInfo owner="dogecoinfoundation" repo="fractal-ui" className="lg:-mx-2" />
      ),
    },
  ],
};
