# VPSFlow Design System + UI Guidelines

A partir de este momento actuarás como Principal Product Designer, UI Designer, UX Designer y Frontend Architect de VPSFlow.

No quiero una landing genérica de hosting.

No quiero una copia de ningún producto existente.

Quiero una identidad visual completamente original, premium, minimalista y coherente para toda la marca VPSFlow.

VPSFlow es una empresa de infraestructura cloud moderna, orientada a empresas y desarrolladores. No vendemos simplemente hosting; ofrecemos una plataforma cloud profesional.

Toda la web deberá transmitir tecnología, rendimiento, estabilidad, confianza y simplicidad.

---

# Filosofía

El diseño debe transmitir:

• Premium
• Enterprise
• Moderno
• Elegante
• Tecnológico
• Minimalista
• Muy limpio
• Mucho espacio en blanco
• Muy pocas distracciones
• Muy buena jerarquía visual

El usuario debe sentir que está navegando por una empresa tecnológica de primer nivel.

No quiero un diseño típico de hosting lleno de iconos, tablas enormes y colores chillones.

---

# Inspiración

Quiero inspirarme en el nivel de calidad de:

• Vercel
• Stripe
• Linear
• Cloudflare
• Railway
• Render
• Raycast
• Arc Browser
• Apple

NO copiar ningún diseño.

Solo inspirarse en:

• Espaciado
• Jerarquía
• Calidad visual
• Tipografía
• Animaciones
• Experiencia de usuario

Todo debe ser completamente original.

---

# Identidad de marca

Nombre:

VPSFlow

Eslogan:

Infrastructure Without Limits.

---

# Tipografía

Utilizar exclusivamente:

Plus Jakarta Sans

Importarla desde Google Fonts.

No utilizar Inter.

No utilizar Poppins.

No utilizar Montserrat.

Toda la interfaz deberá utilizar Plus Jakarta Sans.

Jerarquía:

Hero

80px

H1

64px

H2

48px

H3

36px

H4

28px

Body

18px

Small

15px

Caption

13px

Peso:

400

500

600

700

800

---

# Colores

Background principal

#05070A

Background secundario

#0B1118

Cards

#101722

Cards Hover

#161F2E

Sidebar

#0B1018

Azul principal

#168BFF

Hover

#0E74E6

Azul claro

#56B5FF

Success

#22C55E

Warning

#F59E0B

Danger

#EF4444

Texto principal

#FFFFFF

Texto secundario

#A5B4C3

Texto terciario

#64748B

Border

rgba(255,255,255,.08)

Border Hover

rgba(255,255,255,.16)

---

# Glow

Utilizar glow azul muy suave.

Nunca exagerado.

Ejemplos:

rgba(22,139,255,.35)

rgba(56,189,248,.20)

---

# Espaciado

Sistema de 8px.

4

8

16

24

32

40

48

64

96

128

160

---

# Radius

Botones

16px

Inputs

14px

Cards

24px

Badges

999px

Modals

28px

---

# Sombras

Muy suaves.

Nada agresivo.

Cards

0 15px 40px rgba(0,0,0,.35)

Hover

0 25px 80px rgba(22,139,255,.15)

---

# Fondo

Nunca completamente negro.

Debe utilizar:

Radial Gradients

Glow

Noise muy ligero

Blur

Pequeñas partículas

Debe sentirse vivo.

---

# Navbar

Altura

80px

Glassmorphism ligero.

Blur.

Scroll inteligente.

Al hacer scroll:

Background sólido.

Border inferior.

Transición suave.

Distribución:

Logo izquierda

Navegación centro

Acciones derecha

---

# Logo

Crear carpeta:

public/img/

Dentro:

logo.png

favicon.ico

Toda la web debe utilizar:

/img/logo.png

Dejar preparado para reemplazar el logo fácilmente.

---

# Botones

Primary

Azul

Texto blanco

Hover

Glow

Elevación ligera

Scale 1.02

Secondary

Transparente

Border

Hover

Background oscuro

Ghost

Solo texto

Hover muy ligero

---

# Inputs

Oscuros

Focus azul

Glow

Placeholder gris

---

# Cards

Diseño limpio.

Mucho padding.

Hover:

TranslateY(-4px)

Border azul

Glow muy ligero

Nunca sombras exageradas.

---

# Hero

Debe ocupar prácticamente toda la pantalla.

Lado izquierdo:

Badge superior.

Título enorme.

Descripción.

Dos botones.

Badges inferiores.

Lado derecho:

NO utilizar imágenes típicas de hosting.

Crear una ilustración moderna.

Puede ser:

Cloud abstracta

Cluster

Nodos

Partículas

Anillos

Líneas

Objetos flotantes

Con animaciones suaves.

---

# Secciones

Hero

↓

Infraestructura

↓

Productos

↓

Regiones

↓

Dashboard Preview

↓

Automatización

↓

API

↓

Estadísticas

↓

Clientes

↓

Precios

↓

FAQ

↓

CTA Final

↓

Footer

---

# Productos

Cloud VPS

Dedicated Servers

Object Storage

Networking

Load Balancer

Backups

Firewall

Kubernetes

Cada uno en cards premium.

---

# Dashboard Preview

No utilizar capturas.

Crear un dashboard ficticio.

Muy moderno.

Dark mode.

CPU

RAM

Storage

Traffic

VMs

Graphs

Alerts

---

# API Section

Mostrar ejemplos de código.

Go

JavaScript

cURL

Terraform

Con syntax highlighting.

---

# Mapa

Mapa oscuro.

Puntos luminosos.

Madrid

Barcelona

París

Frankfurt

Amsterdam

New York

Miami

São Paulo

Animaciones.

---

# Estadísticas

15.000+

Virtual Machines

99.99%

Availability

10 Tbps

Network Capacity

24/7

Support

15

Regions

---

# FAQ

Accordion moderno.

Animaciones suaves.

---

# Footer

Minimalista.

Productos

Empresa

Recursos

Legal

Estado

API

GitHub

Contacto

Redes sociales

---

# Iconografía

Lucide Icons.

Siempre outline.

Nunca iconos rellenos.

---

# Animaciones

Muy importantes.

Utilizar Framer Motion.

Fade

Slide

Scale

Blur

Parallax ligero

Glow

Hover

Nunca exageradas.

Duración:

250ms

Ease:

ease-in-out

---

# Responsive

Desktop

Laptop

Tablet

Mobile

Todo perfectamente adaptado.

---

# Accesibilidad

Contraste AA/AAA.

Focus visible.

Navegación por teclado.

Aria labels.

SEO optimizado.

---

# Componentes

Button

Navbar

Footer

Hero

Badge

Card

FeatureCard

StatsCard

PricingCard

DashboardCard

Input

Select

Textarea

Accordion

Modal

Toast

Sidebar

Table

Pagination

CodeBlock

Metric

Chart

StatusBadge

UserMenu

Search

Breadcrumb

RegionCard

Loading

Skeleton

---

# Estructura

/public

    /img

        logo.png

        favicon.ico

        hero.webp

        dashboard-preview.webp

        world-map.svg

        particles.png

/src

    /components

    /sections

    /layouts

    /hooks

    /lib

    /styles

    /types

    /utils

---

# Tecnologías

Next.js

TypeScript

TailwindCSS

shadcn/ui

Framer Motion

Lucide React

React Hook Form

Zod

TanStack Query

---

# Reglas

No utilizar Bootstrap.

No utilizar jQuery.

No utilizar plantillas.

No utilizar componentes genéricos.

Todo debe ser completamente original.

Cada componente debe ser reutilizable.

Cada sección debe sentirse premium.

Cada animación debe tener un propósito.

Cada página debe transmitir que VPSFlow es una plataforma cloud enterprise de última generación.

La prioridad absoluta es crear una identidad visual coherente, moderna y memorable, con una calidad comparable a las mejores empresas SaaS del mundo, manteniendo una personalidad propia y diferenciada.