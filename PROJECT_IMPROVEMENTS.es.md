# Mejoras Del Proyecto NetInspector

Este documento reúne mejoras técnicas y de producto para que NetInspector sea más limpio, más fácil de mantener y más atractivo como proyecto de portafolio.

Leyenda de estado:

- `Completado`: ya implementado.
- `Pendiente`: sigue recomendado.

## Fortalezas Actuales

- Estructura clara del backend en Go usando `cmd/`, `internal/` y paquetes orientados por dominio.
- Contrato moderno de API con Protocol Buffers y Connect-RPC.
- Dashboard en Next.js con cliente TypeScript generado desde Protobuf.
- Grafo de topología interactivo usando React Flow.
- Descubrimiento real de red mediante TCP probing, lectura de ARP cache, enriquecimiento con mDNS/Bonjour y discovery SSDP/UPnP.
- Exportación de scans guardados en formatos JSON y CSV.
- Cobertura básica de tests para el scanner.
- README funcional y `.gitignore` que excluye documentos locales de planificación.

## Mejoras De Clean Code

### 1. Separar Responsabilidades Del Scanner

`internal/scanner/tcp.go` contiene actualmente orquestación del scan, TCP probing, merge de dispositivos, helpers de IP y helpers de nombres de servicios.

Estructura sugerida:

- `scanner/scanner.go`: orquestación del scan.
- `scanner/tcp_probe.go`: probing de puertos TCP.
- `scanner/discovery.go`: interfaces compartidas de discovery.
- `scanner/ip.go`: funciones helper de IP/CIDR.
- `scanner/ports.go`: puertos por defecto y nombres de servicios.

Esto haría que el scanner sea más fácil de extender y testear.

### 2. Agregar Interfaces De Discovery

Estado: `Completado`

Una interfaz común ya define las fuentes de discovery del scanner:

```go
type Discoverer interface {
	Discover(ctx context.Context, local network.LocalNetwork) ([]Device, error)
}
```

Implementaciones actuales:

- `TCPDiscoverer`
- `ARPDiscoverer`
- `MDNSDiscoverer`
- `SSDPDiscoverer`

Esto haría que la arquitectura sea más extensible y esté mejor alineada con principios de clean architecture.

### 3. Centralizar La Lógica Del Caso De Uso De Scan

La API JSON de compatibilidad y el servicio Connect-RPC contienen lógica de orquestación de scan similar.

Mejora recomendada:

- Crear un servicio interno de aplicación, por ejemplo `internal/service/scan_service.go`.
- Hacer que los handlers JSON y Connect-RPC llamen a ese mismo servicio.
- Mantener en la capa de API solo el código específico del transporte.

Esto reduce duplicación y hace que el backend sea más fácil de evolucionar.

### 4. Separar La Página Del Dashboard

Estado: `Completado`

`apps/dashboard/app/page.tsx` estaba haciendo demasiadas cosas: manejo de estado, llamadas a API, render de topología, inspector de dispositivo, consola de eventos y mapeo de datos.

Estructura frontend implementada:

- `components/ScanControls.tsx`
- `components/NetworkSummary.tsx`
- `components/TopologyGraph.tsx`
- `components/DeviceInspector.tsx`
- `components/EventConsole.tsx`
- `lib/netinspector-client.ts`
- `lib/mappers.ts`
- `types/network.ts`

Esto hace que el frontend sea más mantenible y más fácil de testear.

### 5. Remover Archivos No Usados

Revisar `apps/dashboard/go.mod`. Un dashboard Next.js normalmente no debería necesitar un archivo de módulo Go dentro de la carpeta de la app, salvo que exista una razón específica.

Si no se usa, conviene eliminarlo para evitar confusión.

## Mejoras De Prácticas De Ingeniería

### 1. Agregar Un Makefile

Estado: `Completado`

El `Makefile` ya permite tener comandos estándar simples para contribuidores y desarrollo local.

Comandos disponibles:

```sh
make dev
make test
make lint
make build
make proto
make check
```

Esto mejora la experiencia de desarrollo y hace que el repositorio se vea más profesional.

### 2. Agregar CI Con GitHub Actions

Estado: `Completado`

El CI ya corre en pull requests y pushes a `main`:

- `go test ./...`
- lint del dashboard
- build del dashboard
- validación de generación Protobuf

Esto demuestra disciplina profesional de entrega y protege el proyecto contra regresiones.

### 3. Agregar Más Tests

Áreas útiles para testear:

- normalización de hostnames mDNS e inferencia de tipo de dispositivo.
- casos borde de parsing de ARP.
- helpers de detección de red local.
- funciones de mapeo Protobuf.
- mappers del frontend.

### 4. Agregar Metadata Del Proyecto

Archivos útiles para el repositorio:

- `LICENSE`
- `CONTRIBUTING.md`
- `.env.example`
- badge de GitHub Actions en README.
- diagrama corto de arquitectura en README.

### 5. Mejorar Logging

Usar `log/slog` de Go para logs estructurados.

Ejemplos:

- scan iniciado
- scan finalizado
- interfaz seleccionada
- errores de fuentes de discovery
- duración del scan
- cantidad de dispositivos

Los logs estructurados hacen que el backend se sienta más preparado para producción.

## Mejoras De Producto

### 1. Historial De Scans Con SQLite

Estado: `Completado`

Los scans reales anteriores ahora se guardan localmente con SQLite.

Esto permite:

- historial de scans
- comparación entre scans
- primer avistamiento / último avistamiento de un dispositivo
- datos persistentes en el dashboard

Esta es una de las mejoras con mayor valor para portafolio.

### 2. Detección De Cambios En Dispositivos

Comparar el scan actual contra scans anteriores.

Eventos útiles:

- nuevo dispositivo detectado
- dispositivo desaparecido
- hostname cambiado
- MAC cambiada
- nuevo puerto abierto
- puerto cerrado

Esto convierte la app de un scanner de una sola ejecución en una herramienta útil de monitoreo.

### 3. Inspección Profunda De Dispositivos

Agregar una acción `Inspect` por dispositivo.

Detalles posibles:

- scan extendido de puertos
- fingerprinting de servicios
- sistema operativo estimado
- lookups adicionales de hostname
- hints de riesgo o exposición

Esto puede usar el RPC `InspectDevice` que ya existe en el contrato Protobuf.

### 4. Puntaje De Confianza

Mostrar un puntaje de confianza para la identificación.

Ejemplo:

- `TV - 87% confidence`
- Evidencia: servicio Google Cast, hint de vendor, puerto 8443, hostname.

Esto haría que la clasificación de dispositivos sea más transparente y profesional.

### 5. Discovery SSDP/UPnP

Estado: `Completado`

El discovery SSDP ahora complementa mDNS/Bonjour enviando una búsqueda multicast UPnP controlada y mezclando las respuestas dentro del resultado del scan.

Ayuda a identificar:

- smart TVs
- routers
- consolas
- dispositivos multimedia
- dispositivos IoT

### 6. Exportación De Reportes

Estado: `Completado`

Los scans guardados ahora pueden exportarse desde el backend y descargarse desde el dashboard.

Formatos implementados:

- JSON
- CSV

Formato futuro:

- resumen PDF

Esto hace que la aplicación sea más útil para diagnósticos reales.

### 7. Explicaciones Con Enfoque En Privacidad

Agregar explicaciones más claras cuando:

- las direcciones MAC son privadas/randomizadas.
- OUI identifica el fabricante del chip Wi-Fi, no necesariamente la marca comercial del producto.
- mDNS/Bonjour no entrega datos disponibles.

Esto demuestra criterio técnico responsable.

### 8. Mejor Layout Del Grafo

React Flow ya está implementado, pero el layout actual es determinístico y simple.

Mejoras futuras:

- layout automático con Dagre o ELK
- agrupación por tipo de dispositivo
- separación visual para router, clientes, IoT e impresoras
- colores por confianza o estado

## Próximos Pasos Recomendados

Completado:

1. Separar el dashboard en componentes.
2. Agregar un `Makefile`.
3. Agregar CI con GitHub Actions.
4. Refactorizar las fuentes de discovery del scanner detrás de interfaces.
5. Agregar historial de scans con SQLite.
6. Agregar discovery SSDP/UPnP.
7. Agregar exportación de reportes JSON y CSV.

Orden sugerido para continuar:

1. Agregar detección de cambios en dispositivos.
2. Agregar inspección profunda de dispositivos.
3. Agregar exportación de reportes PDF.

Este orden mejora primero la mantenibilidad y luego suma funcionalidad profesional de producto.
