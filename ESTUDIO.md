# Guía de estudio — para que este código sea tuyo

El código lo escribió Claude como punto de partida. **Antes de mencionarlo en
el CV o en una entrevista, tienes que poder explicar cada línea y haberlo
extendido tú.** Un entrevistador que usa Go y Kotlin a diario va a abrir el repo
y preguntarte por qué algo está hecho así. Esta guía es el camino.

## Día 1 — Go

1. Lee `go/queue/queue.go` de arriba abajo. Responde por escrito:
   - ¿Por qué `List` devuelve `[]Item` (copias) y no `[]*Item`?
   - ¿Qué pasaría sin el `defer s.mu.Unlock()`?
   - ¿Para qué sirve `seq` si ya existe `CreatedAt`?
2. Corre `go test -race ./...`. Quita el `s.mu.Lock()` de `Add`, vuelve a correr
   y mira qué detecta `-race`. Restáuralo.
3. **Ejercicio tuyo:** agrega `DELETE /venues/{venue}/queue/{itemId}` (la mesa
   que pidió la canción puede retirarla; otra mesa recibe `403`). Con su prueba.

## Día 2 — Kotlin

1. Lee `kotlin/src/main/kotlin/jukebox/Queue.kt`. Compara con Go:
   - `sealed class QueueError` frente a los `var Err...` de Go: ¿qué te da el
     `when` exhaustivo en `Main.kt` que Go no te da?
   - `data class`, `?:`, `getOrPut`, `compareByDescending().thenBy()`.
2. Corre `./gradlew test`. Quita un `@Synchronized` y corre la prueba de
   concurrencia varias veces.
3. **Ejercicio tuyo:** el mismo `DELETE` del día 1, ahora en Kotlin.

## Día 3 — Tuyo

1. **Ejercicio tuyo:** límite de canciones por hora para todo el local (por
   ejemplo, 20), en los dos lenguajes. Decide tú el código de error.
2. Prepara una respuesta de dos minutos a: *"¿Qué te gustó y qué no de cada
   lenguaje después de escribir lo mismo en los dos?"*
3. Recién ahí, sube el repo a GitHub y envía el CV.

## Lo que puedes decir en la entrevista

> "No tengo Go ni Kotlin en producción. Para prepararme escribí el mismo
> servicio en los dos — la cola de canciones de un producto real que construyo
> — y lo extendí con [lo que hayas agregado]. Mi Java de producción está en
> class mediators sobre WSO2."

Es verdad, y es una respuesta fuerte.
