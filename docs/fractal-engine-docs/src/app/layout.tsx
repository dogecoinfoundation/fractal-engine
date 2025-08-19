import '@/app/global.css';
import { RootProvider } from 'fumadocs-ui/provider';
import { Inter } from 'next/font/google';
import type { ReactNode } from 'react';
import type { Metadata } from 'next';

const inter = Inter({
  subsets: ['latin'],
});

export const metadata: Metadata = {
  metadataBase: new URL(process.env.NEXT_PUBLIC_SITE_URL ?? 'http://localhost:3000'),
  title: {
    default: 'Fractal Engine Docs',
    template: '%s | Fractal Engine Docs',
  },
  description: 'Documentation for Fractal Engine.',
  applicationName: 'Fractal Engine Docs',
  authors: [{ name: 'Dogecoin Foundation' }],
  creator: 'Dogecoin Foundation',
  keywords: ['Fractal Engine', 'Dogecoin', 'Docs', 'API', 'Architecture'],
  icons: {
    icon: '/icon.svg',
  },
  openGraph: {
    title: 'Fractal Engine Docs',
    description: 'Documentation for Fractal Engine.',
    url: '/',
    siteName: 'Fractal Engine Docs',
    locale: 'en_US',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Fractal Engine Docs',
    description: 'Documentation for Fractal Engine.',
  },
  robots: {
    index: true,
    follow: true,
  },
  alternates: {
    canonical: '/',
  },
};

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" className={inter.className} suppressHydrationWarning>
      <body className="flex flex-col min-h-screen">
        <RootProvider>{children}</RootProvider>
      </body>
    </html>
  );
}
