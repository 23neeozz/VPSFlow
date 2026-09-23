import { FadeIn } from "@/components/motion/fade-in";
import { VPSDetailPanel } from "@/components/vps/vps-detail-panel";

export default function UserVPSDetailPage({
  params,
}: {
  params: { id: string };
}) {
  return (
    <FadeIn>
      <VPSDetailPanel vpsId={params.id} backHref="/app" mode="user" />
    </FadeIn>
  );
}
