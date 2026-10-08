package jukebox

import io.ktor.client.request.get
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.HttpStatusCode
import io.ktor.http.contentType
import io.ktor.server.testing.testApplication
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import kotlin.test.Test
import kotlin.test.assertEquals

class ApiTest {
    @Test fun `http flow`() = testApplication {
        application { module() }
        suspend fun post(path: String, body: String = "{}") = client.post(path) {
            contentType(ContentType.Application.Json); setBody(body)
        }

        assertEquals(HttpStatusCode.OK, client.get("/health").status)

        val created = post("/venues/bar/queue", """{"songId":"s1","title":"A","tableId":"t1"}""")
        assertEquals(HttpStatusCode.Created, created.status)
        val id = Json.parseToJsonElement(created.bodyAsText()).jsonObject["id"]!!.jsonPrimitive.content

        assertEquals(HttpStatusCode.Conflict, post("/venues/bar/queue", """{"songId":"s1","title":"A","tableId":"t2"}""").status)
        assertEquals(HttpStatusCode.OK, post("/venues/bar/queue/$id/votes", """{"tableId":"t2"}""").status)
        assertEquals(HttpStatusCode.Conflict, post("/venues/bar/queue/$id/votes", """{"tableId":"t2"}""").status)
        assertEquals(HttpStatusCode.OK, post("/venues/bar/queue/advance").status)
        assertEquals(HttpStatusCode.NotFound, post("/venues/bar/queue/advance").status)
    }

    @Test fun `bad requests`() = testApplication {
        application { module() }
        val r1 = client.post("/venues/bar/queue") { contentType(ContentType.Application.Json); setBody("not json") }
        assertEquals(HttpStatusCode.BadRequest, r1.status)
        val r2 = client.post("/venues/bar/queue") {
            contentType(ContentType.Application.Json); setBody("""{"songId":"","title":"A","tableId":"t1"}""")
        }
        assertEquals(HttpStatusCode.BadRequest, r2.status)
    }

    @Test fun `table limit is 429`() = testApplication {
        application { module() }
        suspend fun add(song: String) = client.post("/venues/bar/queue") {
            contentType(ContentType.Application.Json); setBody("""{"songId":"$song","title":"X","tableId":"t1"}""")
        }
        add("s1"); add("s2")
        assertEquals(HttpStatusCode.TooManyRequests, add("s3").status)
    }
}
