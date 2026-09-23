import { FadeIn } from "@/components/motion/fade-in";
import { PageHeader } from "@/components/ui/page";
import VPSPanel from "./vps-panel";

export default function UserHomePage() {
  return (
    <FadeIn>
      <PageHeader
        title="Mis VPS"
        description="Gestiona las máquinas virtuales asignadas a tu cuenta. Inicia, detén o conéctate por escritorio remoto."
      />
      <VPSPanel />
    </FadeIn>
  );
}
