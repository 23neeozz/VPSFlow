import { redirect } from "next/navigation";
import { Logo } from "@/components/brand/logo";
import { FadeIn } from "@/components/motion/fade-in";
import OrgPicker from "@/app/organizations/org-picker";

export default function AppOrganizationsPage({
  searchParams,
}: {
  searchParams: { next?: string };
}) {
  const next = searchParams.next || "/app";
  return (
    <main className="mx-auto flex min-h-screen max-w-2xl flex-col justify-center px-6 py-16">
      <FadeIn>
        <Logo className="mb-8" />
        <h1 className="text-2xl font-bold tracking-tight text-bc-text-primary md:text-3xl">
          Organización · Cliente
        </h1>
        <p className="mt-3 text-[15px] text-bc-text-secondary">
          Selecciona el espacio de trabajo de tu panel de cliente.
        </p>
        <div className="mt-8">
          <OrgPicker redirectTo={next} />
        </div>
      </FadeIn>
    </main>
  );
}
