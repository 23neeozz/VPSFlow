import { redirect } from "next/navigation";
import { Logo } from "@/components/brand/logo";
import { FadeIn } from "@/components/motion/fade-in";
import OrgPicker from "@/app/organizations/org-picker";

export default function AdminOrganizationsPage({
  searchParams,
}: {
  searchParams: { next?: string };
}) {
  const next = searchParams.next || "/admin";
  return (
    <main className="mx-auto flex min-h-screen max-w-2xl flex-col justify-center px-6 py-16">
      <FadeIn>
        <Logo className="mb-8" />
        <h1 className="text-2xl font-bold tracking-tight text-bc-text-primary md:text-3xl">
          Organización · Admin
        </h1>
        <p className="mt-3 text-[15px] text-bc-text-secondary">
          Selecciona la organización para el panel de administración.
        </p>
        <div className="mt-8">
          <OrgPicker redirectTo={next} />
        </div>
      </FadeIn>
    </main>
  );
}
