import { redirect } from "next/navigation";
import { isAuthenticated } from "@/lib/session";
import { Logo } from "@/components/brand/logo";
import { FadeIn } from "@/components/motion/fade-in";
import { AuthGlobePanel } from "@/components/globe/auth-globe-panel";
import LoginForm from "./login-form";

export default function LoginPage() {
  if (isAuthenticated()) {
    redirect("/");
  }
  return (
    <main className="flex min-h-screen items-center justify-center p-6 md:p-10">
      <div className="grid w-full max-w-5xl gap-8 lg:grid-cols-2 lg:gap-12">
        <FadeIn className="flex flex-col justify-center">
          <Logo className="mb-10" />
          <h1 className="text-3xl font-bold tracking-tight text-bc-text-primary md:text-4xl">
            Accede a tu panel
          </h1>
          <p className="mt-3 text-[15px] text-bc-text-secondary">
            Gestiona tus VPS y recursos cloud desde un único lugar.
          </p>
          <div className="mt-8">
            <LoginForm />
          </div>
        </FadeIn>
        <FadeIn delay={0.1}>
          <AuthGlobePanel />
        </FadeIn>
      </div>
    </main>
  );
}
