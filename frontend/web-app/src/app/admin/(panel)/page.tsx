import { FadeIn } from "@/components/motion/fade-in";
import { PageHeader } from "@/components/ui/page";
import AdminVPSPanel from "../admin-vps-panel";

export default function AdminHomePage() {
  return (
    <FadeIn>
      <PageHeader
        title="VPS de clientes"
        description="Crea VPS, asígnalas a clientes y gestiona el ciclo de vida."
      />
      <AdminVPSPanel />
    </FadeIn>
  );
}
