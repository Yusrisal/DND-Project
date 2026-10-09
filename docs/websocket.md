# WebSocket protocol

```
GET /ws?map_id=1
```

One connection per client, and **one room per campaign** — not per map. Chat
and rolls belong to the table rather than to whichever map happens to be open,
so they survive the DM switching maps. The campaign is read from the stored map,
never from the query string.

A client therefore receives token events for maps it is not looking at. Token
payloads carry `map_id`, and the client ignores the ones that are not its own.
For a group playing one campaign at a time this is a handful of extra frames and
no bookkeeping.

The session cookie is checked **before** the upgrade, so a signed-out client
gets a plain `401` the browser can surface rather than an invisible close frame.
The cookie is sent automatically: the app and the socket share an origin, so
there is no token in the query string.

`map_id` is required; without it the upgrade is refused with `400 bad_request`.
The `Origin` header must appear in `CORS_ORIGINS` or the upgrade fails — in
production that means the real hostname, not `localhost`.

## Frame format

Every frame in both directions is a text frame holding:

```json
{ "type": "token.moved", "payload": {} }
```

Unknown `type` values and unparseable frames get an `error` frame back; the
connection stays open.

## Server to client

| `type`          | Payload              | When                                 |
| --------------- | -------------------- | ------------------------------------ |
| `board`         | `Token[]`            | Once on connect: the requested map    |
| `chat.history`  | `ChatMessage[]`      | Once on connect: the last 50 messages |
| `roll.history`  | `DiceRoll[]`         | Once on connect: the last 50 rolls    |
| `token.created` | `Token`              | A token was added                     |
| `token.updated` | `Token`              | A token was edited                    |
| `token.moved`   | `Token`              | A token was dragged                   |
| `token.deleted` | `{ "id": 7 }`        | A token was removed                   |
| `chat.message`  | `ChatMessage`        | Somebody said something               |
| `roll`          | `DiceRoll`           | Somebody rolled                       |
| `attack`        | `DiceRoll`           | An attack was resolved                |
| `encounter`     | `Encounter`          | Initiative changed, and once on connect when a fight is underway |
| `error`         | `{ "message": "…" }` | The last client frame was rejected    |

The three history frames arrive in that order immediately after the upgrade.
`board` is the full state, so a client that reconnects does not need to
reconcile anything — it replaces what it has.

Payloads carry the stored row, so `x` and `y` are always the server values, and
a roll carries the breakdown the server produced.

## Client to server

| `type`         | Payload                                                        |
| -------------- | -------------------------------------------------------------- |
| `token.moved`  | `{ "id": 7, "x": 125, "y": 75 }`                                |
| `chat.message` | `{ "body": "I search the body" }`                               |
| `roll`         | `{ "expression": "2d20kh1+5", "label": "STR check", "character_id": 3 }` |
| `attack`       | `{ "item_id": 9, "target_token_id": 14, "advantage": true }`    |

Nothing else is accepted. Creating, editing and deleting tokens go through the
REST API, which broadcasts the result here.

```js
socket.send(JSON.stringify({ type: "token.moved", payload: { id: 7, x: 125, y: 75 } }))
socket.send(JSON.stringify({ type: "chat.message", payload: { body: "I search the body" } }))
socket.send(JSON.stringify({ type: "roll", payload: { expression: "1d20+5", label: "Perception" } }))
```

Chat and rolls are broadcast to the whole room **including the sender**, which
needs the server-assigned id and timestamp. Only `token.moved` skips the sender.

A roll is evaluated on the server: the client never sends a result, and
`character_id` is checked against the sheets the caller owns. Expressions and
bounds are in [characters-and-dice.md](characters-and-dice.md#expressions).

An attack names a weapon (`item_id`, which requires owning that sheet) or a
monster action (`action_id`, DM only), never a result. The server decides hit,
damage and the new hit points, then broadcasts the resolution followed by the
updated token. See [combat.md](combat.md#attacks).

The server checks three things before it writes:

1. The token exists.
2. The mover may move it — a player only owns the tokens assigned to them, while
   the DM may move anything. The stored `owner_user_id` is what counts, not
   anything the client claims.
3. The token really belongs to this room, re-read after the write.

Then it broadcasts `token.moved` to everyone **except the sender**, since the
sender already drew the drag locally. A rejected move gets an `error` frame and
no broadcast; the frontend reloads the board from
`GET /api/maps/{id}/tokens` so the optimistic drag does not linger on screen.

Rejection messages are `that token is not yours to move` and
`could not move token`.

## Delivery guarantees

Messages are best-effort, not durably queued:

- Each client has a 32-message buffer. A client that falls behind has updates
  dropped rather than stalling the room.
- A dropped or closed connection is reopened by the frontend after two seconds,
  and the fresh `board` frame resyncs it.

Treat `board` as the source of truth and every other frame as an optimisation.

## Client reference

[`frontend/src/hooks/useSession.ts`](../frontend/src/hooks/useSession.ts) is the
whole client: it opens the socket, applies the events above to React state,
filters token events down to the map on screen, merges chat and rolls into one
feed ordered by timestamp, and reconnects on close. It exposes `moveToken`,
`sendMessage` and `roll`.
