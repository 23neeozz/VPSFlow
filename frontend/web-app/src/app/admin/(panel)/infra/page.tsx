import { FadeIn } from "@/components/motion/fade-in";
import { PageHeader } from "@/components/ui/page";
import HypervisorPanel from "../../hypervisor-panel";

export default function AdminInfraPage() {
  return (
    <FadeIn>
      <PageHeader
        title="Infraestructura"
        description="Estado de hipervisores, capacidad del cluster y salud de los nodos."
      />
      <HypervisorPanel />
    </FadeIn>
  );
}
