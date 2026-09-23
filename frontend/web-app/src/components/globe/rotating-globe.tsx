"use client";

import { useEffect, useRef } from "react";
import createGlobe from "cobe";

/** Region markers from DESIGN.md */
const REGIONS: { name: string; lat: number; lng: number }[] = [
  { name: "Madrid", lat: 40.4168, lng: -3.7038 },
  { name: "Barcelona", lat: 41.3874, lng: 2.1686 },
  { name: "París", lat: 48.8566, lng: 2.3522 },
  { name: "Frankfurt", lat: 50.1109, lng: 8.6821 },
  { name: "Amsterdam", lat: 52.3676, lng: 4.9041 },
  { name: "New York", lat: 40.7128, lng: -74.006 },
  { name: "Miami", lat: 25.7617, lng: -80.1918 },
  { name: "São Paulo", lat: -23.5505, lng: -46.6333 },
];

type RotatingGlobeProps = {
  className?: string;
  size?: number;
};

export function RotatingGlobe({ className = "", size = 480 }: RotatingGlobeProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const pointerInteracting = useRef<number | null>(null);
  const pointerInteractionMovement = useRef(0);
  const phiRef = useRef(0);

  useEffect(() => {
    if (!canvasRef.current) return;

    const globe = createGlobe(canvasRef.current, {
      devicePixelRatio: 2,
      width: size * 2,
      height: size * 2,
      phi: 0,
      theta: 0.22,
      dark: 1,
      diffuse: 1.15,
      mapSamples: 20000,
      mapBrightness: 6,
      mapBaseBrightness: 0.08,
      baseColor: [0.2, 0.2, 0.22],
      markerColor: [1, 1, 1],
      glowColor: [0.9, 0.9, 0.9],
      scale: 1.05,
      markers: REGIONS.map((r) => ({
        location: [r.lat, r.lng],
        size: 0.06,
      })),
    });

    let frameId = 0;
    const animate = () => {
      const canvas = canvasRef.current;
      if (!canvas) return;

      if (pointerInteracting.current === null) {
        phiRef.current += 0.004;
      }

      const dim = canvas.offsetWidth * 2;
      globe.update({
        phi: phiRef.current + pointerInteractionMovement.current,
        width: dim,
        height: dim,
      });
      frameId = requestAnimationFrame(animate);
    };

    frameId = requestAnimationFrame(animate);

    return () => {
      cancelAnimationFrame(frameId);
      globe.destroy();
    };
  }, [size]);

  return (
    <canvas
      ref={canvasRef}
      className={`aspect-square w-full max-w-[480px] cursor-grab active:cursor-grabbing ${className}`}
      style={{ width: size, height: size, contain: "layout paint size" }}
      onPointerDown={(e) => {
        pointerInteracting.current =
          e.clientX - pointerInteractionMovement.current;
        if (canvasRef.current) canvasRef.current.style.cursor = "grabbing";
      }}
      onPointerUp={() => {
        pointerInteracting.current = null;
        if (canvasRef.current) canvasRef.current.style.cursor = "grab";
      }}
      onPointerOut={() => {
        pointerInteracting.current = null;
        if (canvasRef.current) canvasRef.current.style.cursor = "grab";
      }}
      onMouseMove={(e) => {
        if (pointerInteracting.current !== null) {
          const delta = e.clientX - pointerInteracting.current;
          pointerInteractionMovement.current = delta / 120;
        }
      }}
      onTouchMove={(e) => {
        if (pointerInteracting.current !== null && e.touches[0]) {
          const delta = e.touches[0].clientX - pointerInteracting.current;
          pointerInteractionMovement.current = delta / 80;
        }
      }}
    />
  );
}
