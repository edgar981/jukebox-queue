package jukebox

import java.time.Instant
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicInteger
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class QueueStoreTest {
    private val fixed = Instant.parse("2026-10-08T20:00:00Z")
    private fun store() = QueueStore { fixed }

    @Test fun `duplicate song is rejected`() {
        val s = store()
        s.add("bar", "s1", "A", "t1")
        assertFailsWith<QueueError.DuplicateSong> { s.add("bar", "s1", "A", "t2") }
    }

    @Test fun `table limit`() {
        val s = store()
        s.add("bar", "s1", "A", "t1")
        s.add("bar", "s2", "B", "t1")
        assertFailsWith<QueueError.TableLimit> { s.add("bar", "s3", "C", "t1") }
    }

    @Test fun `votes reorder and ties keep arrival order`() {
        val s = store()
        val a = s.add("bar", "s1", "A", "t1")
        val b = s.add("bar", "s2", "B", "t2")
        val c = s.add("bar", "s3", "C", "t3")
        s.vote("bar", c.id, "t9")
        assertEquals(listOf(c.id, a.id, b.id), s.list("bar").map { it.id })
    }

    @Test fun `one vote per table`() {
        val s = store()
        val it = s.add("bar", "s1", "A", "t1")
        s.vote("bar", it.id, "t2")
        assertFailsWith<QueueError.AlreadyVoted> { s.vote("bar", it.id, "t2") }
    }

    @Test fun `advance pops head and frees table slot`() {
        val s = store()
        val a = s.add("bar", "s1", "A", "t1")
        s.add("bar", "s2", "B", "t1")
        assertEquals(a.id, s.advance("bar").id)
        s.add("bar", "s3", "C", "t1") // la mesa recupera un cupo
    }

    @Test fun `advance on empty queue`() {
        assertFailsWith<QueueError.Empty> { store().advance("bar") }
    }

    @Test fun `venues are isolated`() {
        val s = store()
        s.add("bar-a", "s1", "A", "t1")
        s.add("bar-b", "s1", "A", "t1") // la misma canción en otro local sí se permite
    }

    @Test fun `concurrent adds respect the duplicate rule`() {
        val s = QueueStore()
        val pool = Executors.newFixedThreadPool(16)
        val wins = AtomicInteger()
        repeat(50) { i ->
            pool.submit {
                runCatching { s.add("bar", "same-song", "X", "t$i") }.onSuccess { wins.incrementAndGet() }
            }
        }
        pool.shutdown()
        pool.awaitTermination(10, TimeUnit.SECONDS)
        assertEquals(1, wins.get())
    }
}
