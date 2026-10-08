package jukebox

import kotlinx.serialization.Serializable
import java.time.Instant

/** Cuántos pedidos pendientes puede tener una mesa. */
const val MAX_PENDING_PER_TABLE = 2

/** Errores del dominio. La capa HTTP los traduce a códigos. */
sealed class QueueError(message: String) : Exception(message) {
    class DuplicateSong : QueueError("song already in queue")
    class TableLimit : QueueError("table has reached its pending request limit")
    class AlreadyVoted : QueueError("table already voted for this item")
    class NotFound : QueueError("item not found")
    class Empty : QueueError("queue is empty")
}

/** Lo que ve el cliente de un ítem. */
@Serializable
data class ItemView(
    val id: String,
    val songId: String,
    val title: String,
    val tableId: String,
    val votes: Int,
    val createdAt: String,
)

private class Item(
    val id: String,
    val songId: String,
    val title: String,
    val tableId: String,
    val createdAt: Instant,
    val seq: Long, // desempata dos ítems creados en el mismo instante
) {
    val voters = mutableSetOf<String>()
    val votes get() = voters.size

    fun view() = ItemView(id, songId, title, tableId, votes, createdAt.toString())
}

/** Una cola por local. Todas las operaciones están sincronizadas sobre el store. */
class QueueStore(private val clock: () -> Instant = Instant::now) {
    private val venues = mutableMapOf<String, MutableList<Item>>()
    private var seq = 0L

    private val order = compareByDescending<Item> { it.votes }.thenBy { it.seq }

    private fun venue(id: String) = venues.getOrPut(id) { mutableListOf() }

    @Synchronized
    fun add(venueId: String, songId: String, title: String, tableId: String): ItemView {
        val items = venue(venueId)
        if (items.any { it.songId == songId }) throw QueueError.DuplicateSong()
        if (items.count { it.tableId == tableId } >= MAX_PENDING_PER_TABLE) throw QueueError.TableLimit()

        seq++
        val item = Item("itm_$seq", songId, title, tableId, clock(), seq)
        items += item
        return item.view()
    }

    @Synchronized
    fun vote(venueId: String, itemId: String, tableId: String): ItemView {
        val item = venue(venueId).find { it.id == itemId } ?: throw QueueError.NotFound()
        if (!item.voters.add(tableId)) throw QueueError.AlreadyVoted()
        return item.view()
    }

    @Synchronized
    fun list(venueId: String): List<ItemView> = venue(venueId).sortedWith(order).map { it.view() }

    @Synchronized
    fun advance(venueId: String): ItemView {
        val items = venue(venueId)
        val head = items.minWithOrNull(order) ?: throw QueueError.Empty()
        items.remove(head)
        return head.view()
    }
}
