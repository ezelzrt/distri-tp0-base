# TP0: Docker + Comunicaciones + Concurrencia

## Ejercicio 6 - Procesamiento por batch

### Protocolo Implementado (Go/Python)

Formato de cada mensaje:

- 1 byte: `Type`
  - `0` = batch de apuestas
  - `1` = ACK
- 4 bytes: `Length` en big-endian (tamaño del payload)
- `Length` bytes: `Payload` (UTF-8)

Para la request batch:
- payload = apuestas CSV separadas por `\n`, cada apuesta con formato `A,B,00000000,2000-01-01,0` (nombre, apellido, documento, nacimiento, numero).

Para el ACK:
- payload `b"0"` = batch exitoso
- payload `b"1"` = batch con error

#### Diagrama ASCII del paquete batch

```
+------+------------+--------------------------------------------------------------+
| Type | Length     | Payload                                                      |
| (1B) | (4B BE)    | (N bytes)                                                    |
+------+------------+--------------------------------------------------------------+
| 0x00 | 0x0000004B | "A,B,00000001,2000-01-01,1234\nC,D,00000002,2000-01-01,5678" |
+------+------------+--------------------------------------------------------------+
```

#### Diagrama ASCII del paquete ACK

```
+------+------------+------------------+
| Type | Length     | Payload          |
| (1B) | (4B BE)    | (N bytes)        |
+------+------------+------------------+
| 0x01 | 0x00000001 | "0" (success)    |
+------+------------+------------------+
```

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
- `client1 | action: config | result: success ...`
- `server | action: apuesta_recibida | result: success | cantidad: N`

5) Detener el sistema

```bash
make docker-compose-down
```

