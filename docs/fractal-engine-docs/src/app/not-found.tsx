import Link from "next/link";

export default function NotFound() {
  return (
    <main className="flex flex-1 items-center justify-center p-8">
      <div className="max-w-xl text-center space-y-4">
        <h1 className="text-2xl font-semibold">Page not found</h1>
        <p className="text-muted-foreground">
          The page you’re looking for doesn’t exist or has moved.
        </p>
        <div className="flex items-center justify-center gap-4">
          <Link href="/" className="underline underline-offset-4">
            Go home
          </Link>
          <Link href="/docs" className="underline underline-offset-4">
            Browse docs
          </Link>
        </div>
      </div>
    </main>
  );
}
