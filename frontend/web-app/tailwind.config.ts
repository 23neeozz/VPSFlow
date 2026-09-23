import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        sans: ["var(--font-jakarta)", "system-ui", "sans-serif"],
      },
      colors: {
        bc: {
          bg: "#000000",
          "bg-secondary": "#0A0A0A",
          card: "#111111",
          "card-hover": "#1A1A1A",
          sidebar: "#000000",
          primary: "#FFFFFF",
          "primary-hover": "#E5E5E5",
          "primary-light": "#F5F5F5",
          success: "#22C55E",
          warning: "#F59E0B",
          danger: "#EF4444",
          "text-primary": "#FFFFFF",
          "text-secondary": "#A3A3A3",
          "text-tertiary": "#737373",
          border: "rgba(255,255,255,0.10)",
          "border-hover": "rgba(255,255,255,0.22)",
        },
      },
      borderRadius: {
        btn: "16px",
        input: "14px",
        card: "24px",
        modal: "28px",
      },
      boxShadow: {
        card: "0 15px 40px rgba(0,0,0,.55)",
        "card-hover": "0 25px 80px rgba(255,255,255,.06)",
        glow: "0 0 40px rgba(255,255,255,.12)",
        "glow-sm": "0 0 20px rgba(255,255,255,.08)",
      },
      animation: {
        "fade-in": "fadeIn 0.35s ease-in-out",
        "slide-up": "slideUp 0.35s ease-in-out",
        float: "float 6s ease-in-out infinite",
        pulseGlow: "pulseGlow 4s ease-in-out infinite",
      },
      keyframes: {
        fadeIn: {
          "0%": { opacity: "0" },
          "100%": { opacity: "1" },
        },
        slideUp: {
          "0%": { opacity: "0", transform: "translateY(12px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        float: {
          "0%, 100%": { transform: "translateY(0)" },
          "50%": { transform: "translateY(-12px)" },
        },
        pulseGlow: {
          "0%, 100%": { opacity: "0.4" },
          "50%": { opacity: "0.8" },
        },
      },
      transitionDuration: {
        DEFAULT: "250ms",
      },
      transitionTimingFunction: {
        DEFAULT: "cubic-bezier(0.4, 0, 0.2, 1)",
      },
    },
  },
  plugins: [],
};

export default config;
