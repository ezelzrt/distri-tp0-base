# TP0: Docker + Comunicaciones + Concurrencia

## Ejercicio 5 - Implementación de protocolo de apuestas

### Protocolo Implementado (Go/Python)

Formato de cada mensaje:

- 1 byte: `Type`
  - `0` = apuesta (bet)
  - `1` = confirmación (ack)
- 4 bytes: `Length` en big-endian (tamaño del payload)
- `Length` bytes: `Payload` (CSV UTF-8): `nombre;apellido;documento;nacimiento;numero`

#### Diagrama ASCII del paquete

```
+------+------------+----------------------------------------------------+
| Type | Length     | Payload                                            |
| (1B) | (4B BE)    | (N bytes)                                          |
+------+------------+----------------------------------------------------+
| 0x00 | 0x0000001F |  "Pepe;Martinez;12345678;1999-01-11;9091"         |
+------+------------+----------------------------------------------------+
```

- `Type`=0: envio de apuesta
- `Type`=1: ACK de server (puede payload vacío o "OK")

### Ejecución de la solución

1) Archivos de configuración

- `client/config.yaml`: configuración cliente (server address, log level).
- `server/config.ini`: configuración servidor (puerto, backlog, log level).
- `client/.env`: variables de apuesta (CLI_NOMBRE, CLI_APELLIDO, CLI_DOCUMENTO, CLI_NACIMIENTO, CLI_NUMERO).

2) .env

Copia el `.env.example` a `.env` antes de arrancar:

```bash
cp .env.example .env
```

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

Debería de verse algo como:
- `client1 | action: config | result: success ...`
- `server | action: apuesta_almacenada | result: success | dni: ...`
- `client1 | action: apuesta_enviada | result: success | dni: ...`

5) Detener el sistema

```bash
make docker-compose-down
```
