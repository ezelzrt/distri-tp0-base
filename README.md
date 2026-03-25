## Aclaración
Durante ejercicios anteriores ya se agrego paralelización en el servidor para atender múltiples clientes en simultáneo. Después en el ejercicio 7, para resolverlo se usaron barreras para sincronizar el sorteo con la consulta de ganadores. Por lo mencionado, en este ejercicio no fue necesario agregar nuevas funcionalidades o cambiar el codigo en sí.

Igualmente, se agrego una sección de [concurrencia y sincronizacion](#concurrencia-y-sincronización) para explicar brevemente el uso y la implementación realizada.

<br>

# TP0: Docker + Comunicaciones + Concurrencia

## Ejercicio 8

### Protocolo Implementado (Go/Python)

Formato de cada mensaje:

- 1 byte: `Type`
- 2 bytes: `Agency ID` en big-endian (numero de agencia)
- 1 byte: `EOF flag` (0 o 1)
- 4 bytes: `Length` en big-endian (tamaño del payload)
- `Length` bytes: `Payload` (UTF-8)

Header total: 8 bytes

Tipos de mensaje:

- 0: batch de apuestas
- 1: ACK
- 2: consulta de ganadores
- 3: respuesta de ganadores

Criterio de uso:

- El cliente envía Type 0 para apuestas y marca EOF en el último envío.
- El servidor responde ACK por cada batch:
  - Payload 0: éxito
  - Payload 1: error
- Terminadas las apuestas y luego del sorteo, el cliente envía Type 2.
- El servidor responde Type 3 con lista de DNI ganadores de esa agencia.



#### Diagrama del formato del mensaje:

```
0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|     Type      |           Agency ID           |   EOF Flag    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                        Payload Length                         |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
.                                                               .
.                   Payload (Variable Length)                   .
.                                                               .
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

#### Decisiones de Diseño
Se decidió un formato de mensaje con un header fijo de 8 bytes para simplificar la implementación y mantener la claridad del protocolo (en lugar de empaquetar el flag EOF y el Agency ID en un solo byte). Esto para priorizar un código limpio y sencillo por sobre una, tal vez, micro-optimización de espacio prematura.

#### Concurrencia y sincronización
Para atender varios clientes al mismo tiempo, el servidor se ejecuta con múltiples hilos, y cada conexión se procesa en un hilo independiente.

Este enfoque es adecuado en Python para este caso porque la carga principal del servidor está en operaciones de entrada/salida (red y archivos), no en cálculo intensivo de CPU. Por eso, aunque exista la limitación del GIL, el modelo sigue siendo efectivo para este tipo de aplicación.

A nivel de sincronización, se usaron primitivas básicas para mantener consistencia y evitar bloqueos:

- Locks para proteger recursos compartidos y evitar condiciones de carrera.
- Una barrera para coordinar el punto en que todas las agencias terminaron de enviar apuestas antes de habilitar la etapa de ganadores.
- Un evento de parada para coordinar el cierre ordenado del servidor cuando recibe señales del sistema.

---

### Ejecución de la solución

1) Archivos de configuración

- `client/config.yaml`: cliente (server address, log level, batch.maxAmount).
- `server/config.ini`: servidor (puerto, backlog, log level).

2) CSV de apuestas

- El cliente lee `./data/agency-<id>.csv`.

3) Levantar el sistema

```bash
make docker-compose-up
# o:
# docker compose -f docker-compose-dev.yaml up -d --build
```

4) Verificar logs

```bash
make docker-compose-logs
# o:
# docker compose -f docker-compose-dev.yaml logs -f
```

Debe verse:
- `action: apuesta_recibida | result: success | cantidad: N`
- `action: sorteo | result: success`
- `action: consulta_ganadores | result: success | cant_ganadores: N`

5) Detener el sistema

```bash
make docker-compose-down
```

