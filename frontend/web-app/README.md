# VPSFlow Web App

MVP del panel web de VPSFlow, construido con **Next.js 14 (App Router) + TypeScript + Tailwind**.

Habla con el backend a través del **gateway** (`/api/v1/...`) usando rutas de proxy
server-side de Next.js, por lo que el navegador nunca contacta directamente con el
gateway y los tokens se guardan en cookies `httpOnly`.

## Funcionalidad

- Login y registro de usuarios (auth service).
- Selección y creación de organizaciones (tenant service).
- Listado de hipervisores con capacidad y heartbeat (cluster service).
- Listado, creación, arranque, parada y borrado de VMs (vm service), con
  auto-refresh mientras hay operaciones en curso.

## Requisitos

- Node.js 18.18+ (recomendado 20 LTS).
- El stack de VPSFlow corriendo (gateway en `http://127.0.0.1:8080` por defecto).

## Puesta en marcha

```bash
cd frontend/web-app
npm install
copy .env.local.example .env.local   # en PowerShell/CMD
npm run dev
```

Abre http://localhost:3000

## Configuración

`.env.local`:

```
VPSFLOW_GATEWAY_URL=http://127.0.0.1:8080
```

Solo se usa en el servidor Next.js (nunca se expone al navegador).
