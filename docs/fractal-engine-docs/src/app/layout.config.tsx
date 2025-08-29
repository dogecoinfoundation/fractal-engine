import { GithubInfo } from "@/components/GithubInfo";
import Image from "next/image";
import type { BaseLayoutProps } from "fumadocs-ui/layouts/shared";

import FractalEngineLogo from "../../public/static/image/fractal.svg"

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
      <Image src={FractalEngineLogo} alt="Logo" width={192} />
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
