# Clock, two ways

Both pages render the complete initial clock and hand positions on the server.
The browser supplies its IANA timezone in the page URL (`?tz=Europe/Budapest`).
Opening a URL without `tz` adds the browser timezone and navigates once; tab links
preserve it. No cookies or loading placeholder.

- `/client`: vanilla JS adopts the rendered timestamp, then updates three CSS
  rotation variables and the labels every second. No further requests.
- `/stream`: Datastar opens `/events?tz=…` once. Go sends fresh clock HTML every
  second over that same SSE response, until the client disconnects.

```sh
go run .
# http://127.0.0.1:8091
```

Timezone rules include daylight saving time. The client version advances from
server time using elapsed browser time (network latency is not corrected).
Datastar v1.0.4 is vendored and loaded only on the SSE page.
