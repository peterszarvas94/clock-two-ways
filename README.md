# Clock, two ways

Two small Go clock demos: one ticks in the browser with vanilla JavaScript, the
other receives HTML from the server through Datastar and Server-Sent Events.
Both start with a complete server-rendered clock, including the hand positions.

## Run locally

Requires Go 1.24 or newer. No Node.js, frontend build, account, or API key is needed.

```sh
go run .
```

Open [http://127.0.0.1:8091](http://127.0.0.1:8091) and switch between the tabs:

- **Client ticking** (`/client`): vanilla JS adopts the initial server timestamp,
  then updates the labels and three CSS custom properties every second. There
  are no further clock requests.
- **Server SSE** (`/stream`): Datastar opens one `/events` connection. The server
  sends a fresh clock HTML fragment every second, and Datastar morphs it into
  the page. The connection stays open until the visitor leaves the page.

To listen on a different address:

```sh
go run . -addr :8080
```

## How it works

The browser supplies its IANA timezone in the URL, for example
`/client?tz=America%2FNew_York`. When the timezone is missing, a small script adds
it and navigates once. Tab links preserve it. No cookies are used. Without
JavaScript, the page remains a static server-rendered snapshot (UTC if no
explicit timezone is supplied).

The server renders raw values such as `--hour: 1; --minute: 5; --second: 32`.
CSS `calc()` converts those values into hand rotations, including the minute
and second contributions to the hour hand. The same HTML and CSS serve both
examples; only the update mechanism differs.

The client version advances from server time using elapsed browser time. It
does not correct for initial network latency or periodically synchronize with
the server. These are intentionally small proofs of concept, with one-second
ticks rather than smooth animation. Timezone data is embedded in the Go binary.

Datastar v1.0.4 is vendored in `static/datastar.js` and loaded only on the SSE tab.
There are no CDN requests. For an SSE deployment behind a proxy, allow long-lived
responses and disable response buffering.

## Check and build

```sh
go test -race ./...
go vet ./...
go build -o clock-two-ways .
./clock-two-ways
```

The tests cover server-rendered initial clocks, timezone validation, and multiple
clock updates over a single SSE response.

## Inspiration and license

Inspiration: [A Clock That Doesn't Snap — Ethan Niser](https://ethanniser.com/blog/a-clock-that-doesnt-snap/).

This project is licensed under the [MIT license](LICENSE). Datastar retains its
[own MIT license](static/DATASTAR-LICENSE.md).
