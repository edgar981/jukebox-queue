# jukebox-queue — especificación

La cola de canciones de un bar: las mesas piden canciones, votan, y el local
avanza la cola. **La misma especificación está implementada en Go (`go/`) y en
Kotlin (`kotlin/`)**, con las mismas pruebas, para comparar los dos lenguajes.

Todo vive en memoria y es seguro bajo concurrencia. No hay base de datos a
propósito: el foco es el lenguaje, no la infraestructura.

## Reglas

1. Una canción no puede estar dos veces en la cola de un local → `409`.
2. Una mesa tiene como máximo **2 pedidos pendientes** en un local → `429`.
3. Una mesa vota una sola vez por ítem → `409`.
4. Orden de la cola: **más votos primero; a igualdad de votos, el más antiguo
   primero.**
5. Avanzar saca el primero de la cola y lo devuelve como "sonando ahora"; con la
   cola vacía → `404`.
6. Cada local tiene su propia cola (multi-tenant): lo que pasa en un local no
   afecta a otro.

## API

| Método | Ruta | Cuerpo | Éxito |
| --- | --- | --- | --- |
| `GET` | `/health` | — | `200 {"status":"ok"}` |
| `POST` | `/venues/{venue}/queue` | `{"songId","title","tableId"}` | `201` el ítem creado |
| `GET` | `/venues/{venue}/queue` | — | `200` lista ordenada |
| `POST` | `/venues/{venue}/queue/{itemId}/votes` | `{"tableId"}` | `200` el ítem actualizado |
| `POST` | `/venues/{venue}/queue/advance` | — | `200` el ítem que pasa a sonar |

Un ítem: `{"id","songId","title","tableId","votes","createdAt"}`.

Errores: `{"error": "<mensaje>"}` con `400` (cuerpo inválido o campos vacíos),
`404` (ítem o cola vacía), `409` (duplicado o voto repetido), `429` (límite de
la mesa).
