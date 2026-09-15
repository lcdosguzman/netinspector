# Mejoras Pendientes De NetInspector

Este documento resume únicamente las mejoras que todavía faltan.

## Clean Code

### 1. Revisar Archivos No Usados

Revisar si `apps/dashboard/go.mod` sigue siendo necesario. Si no se usa, eliminarlo para evitar confusión.

## Prácticas De Ingeniería

### 1. Agregar Más Tests

Áreas útiles:

- parsing ARP.
- parsing SSDP.
- helpers de red local.
- mappers de Protobuf.
- mappers del frontend.
- exportación de reportes.

### 2. Agregar Metadata Del Proyecto

Archivos útiles:

- `LICENSE`
- `CONTRIBUTING.md`
- `.env.example`
- diagrama corto de arquitectura en el README.

### 3. Mejorar Logging

Usar `log/slog` para logs estructurados.

Eventos importantes:

- scan iniciado/finalizado
- interfaz seleccionada
- errores por fuente de discovery
- duración del scan
- cantidad de dispositivos
- reporte exportado

## Producto

### 1. Detección De Cambios En Dispositivos

Comparar un scan nuevo contra scans anteriores.

Detectar:

- dispositivo nuevo
- dispositivo desaparecido
- hostname cambiado
- MAC cambiada
- puerto nuevo
- puerto cerrado

### 2. Inspección Profunda De Dispositivos

Agregar una acción `Inspect` por dispositivo.

Podría incluir:

- scan extendido de puertos
- fingerprinting de servicios
- estimación de sistema operativo
- más lookups de hostname
- hints de riesgo o exposición

### 3. Puntaje De Confianza

Mostrar qué tan confiable es la clasificación del dispositivo.

Ejemplo:

- `TV - 87% confianza`
- Evidencia: Google Cast, vendor, puerto 8443, hostname.

### 4. Explicaciones De Privacidad

Agregar explicaciones claras cuando:

- una MAC es privada/randomizada.
- el OUI identifica el chip Wi-Fi, no siempre la marca real.
- mDNS/Bonjour no entrega datos.
- SSDP/UPnP no entrega datos.

### 5. Mejor Layout Del Grafo

Mejorar el layout actual de React Flow.

Ideas:

- layout automático con Dagre o ELK.
- agrupación por tipo de dispositivo.
- separación visual para router, clientes, IoT e impresoras.
- colores por confianza o estado.

### 6. Exportación PDF

JSON y CSV ya existen. Falta un PDF para compartir resultados con usuarios no técnicos.

Podría incluir:

- resumen del scan
- tabla de dispositivos
- snapshot de topología
- notas de identificación
- limitaciones de privacidad

## Orden Sugerido

1. Agregar detección de cambios.
2. Agregar inspección profunda.
3. Agregar puntaje de confianza.
4. Mejorar explicaciones de privacidad.
5. Mejorar layout del grafo.
6. Agregar exportación PDF.
