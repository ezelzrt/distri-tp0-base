# TP0: Docker + Comunicaciones + Concurrencia

## Ejercicio 7

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

