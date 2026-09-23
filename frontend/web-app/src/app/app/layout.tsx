import { redirect } from "next/navigation";
import { isAuthenticated } from "@/lib/session";

export default function UserAreaRootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  if (!isAuthenticated()) {
    redirect("/login");
  }
  return children;
}
