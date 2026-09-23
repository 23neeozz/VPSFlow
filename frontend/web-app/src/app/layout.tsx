import type { Metadata } from "next";
import { Plus_Jakarta_Sans } from "next/font/google";
import "./globals.css";

const jakarta = Plus_Jakarta_Sans({
  subsets: ["latin"],
  variable: "--font-jakarta",
  weight: ["400", "500", "600", "700", "800"],
  display: "swap",
});

export const metadata: Metadata = {
  title: "VPSFlow — Infraestructura Cloud y VPS",
  description: "Panel de gestión de VPS e infraestructura cloud VPSFlow.",
  icons: { icon: "/img/favicon.ico" },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="es" className={jakarta.variable}>
      <body className="min-h-screen">
        <div className="bc-mesh" aria-hidden />
        <div className="bc-noise" aria-hidden />
        <div className="relative z-10">{children}</div>
      </body>
    </html>
  );
}
