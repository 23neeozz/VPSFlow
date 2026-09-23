import Image from "next/image";
import Link from "next/link";

export function Logo({ className = "" }: { className?: string }) {
  return (
    <Link href="/" className={`inline-flex items-center ${className}`}>
      <Image
        src="/img/logo.svg"
        alt="Panel"
        width={120}
        height={36}
        priority
        className="h-8 w-auto"
      />
    </Link>
  );
}

export function LogoMark({ className = "" }: { className?: string }) {
  return (
    <div
      className={`flex h-9 w-9 items-center justify-center rounded-[10px] border border-bc-border bg-bc-card ${className}`}
    >
      <span className="text-sm font-bold text-bc-primary">P</span>
    </div>
  );
}
