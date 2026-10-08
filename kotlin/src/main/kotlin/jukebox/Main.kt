package jukebox

import io.ktor.http.HttpStatusCode
import io.ktor.serialization.kotlinx.json.json
import io.ktor.server.application.Application
import io.ktor.server.application.install
import io.ktor.server.engine.embeddedServer
import io.ktor.server.netty.Netty
import io.ktor.server.plugins.BadRequestException
import io.ktor.server.plugins.contentnegotiation.ContentNegotiation
import io.ktor.server.plugins.statuspages.StatusPages
import io.ktor.server.request.receive
import io.ktor.server.response.respond
import io.ktor.server.routing.get
import io.ktor.server.routing.post
import io.ktor.server.routing.route
import io.ktor.server.routing.routing
import kotlinx.serialization.Serializable

@Serializable data class AddRequest(val songId: String = "", val title: String = "", val tableId: String = "")
@Serializable data class VoteRequest(val tableId: String = "")
@Serializable data class ErrorBody(val error: String)

fun Application.module(store: QueueStore = QueueStore()) {
    install(ContentNegotiation) { json() }
    install(StatusPages) {
        exception<QueueError> { call, e ->
            val status = when (e) {
                is QueueError.DuplicateSong, is QueueError.AlreadyVoted -> HttpStatusCode.Conflict
                is QueueError.TableLimit -> HttpStatusCode.TooManyRequests
                is QueueError.NotFound, is QueueError.Empty -> HttpStatusCode.NotFound
            }
            call.respond(status, ErrorBody(e.message ?: "error"))
        }
        exception<BadRequestException> { call, _ ->
            call.respond(HttpStatusCode.BadRequest, ErrorBody("invalid JSON body"))
        }
    }

    routing {
        get("/health") { call.respond(mapOf("status" to "ok")) }

        route("/venues/{venue}/queue") {
            post {
                val body = call.receive<AddRequest>()
                if (listOf(body.songId, body.title, body.tableId).any { it.isBlank() }) {
                    return@post call.respond(HttpStatusCode.BadRequest, ErrorBody("songId, title and tableId are required"))
                }
                val item = store.add(call.parameters["venue"]!!, body.songId, body.title, body.tableId)
                call.respond(HttpStatusCode.Created, item)
            }
            get { call.respond(store.list(call.parameters["venue"]!!)) }
            post("/{itemId}/votes") {
                val body = call.receive<VoteRequest>()
                if (body.tableId.isBlank()) {
                    return@post call.respond(HttpStatusCode.BadRequest, ErrorBody("tableId is required"))
                }
                call.respond(store.vote(call.parameters["venue"]!!, call.parameters["itemId"]!!, body.tableId))
            }
            post("/advance") { call.respond(store.advance(call.parameters["venue"]!!)) }
        }
    }
}

fun main() {
    val port = System.getenv("PORT")?.toIntOrNull() ?: 8081
    embeddedServer(Netty, port = port) { module() }.start(wait = true)
}
