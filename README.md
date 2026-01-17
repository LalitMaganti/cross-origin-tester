# Cross-Origin Isolation Tester

A tool for testing COOP, COEP, CORP, and Document-Isolation-Policy headers to understand cross-origin isolation behavior in browsers.

## Purpose

Getting `crossOriginIsolated: true` is required for using powerful web features like:
- `SharedArrayBuffer`
- High-resolution `performance.now()` timers
- `Atomics.wait()` on the main thread

This tool helps you understand:
1. **How to isolate your top-level page** using COOP + COEP headers
2. **How to isolate cross-origin iframes** (spoiler: COOP doesn't work in iframes!)
3. **How CORP headers affect resource loading** with COEP enabled
4. **Document-Isolation-Policy** - Chrome's solution for iframe isolation

## Key Findings

### COOP only works at top-level

`Cross-Origin-Opener-Policy` only applies to top-level browsing contexts. If you set COOP + COEP on an iframe's content, the iframe will **not** be cross-origin isolated.

### For iframe isolation, use Document-Isolation-Policy

```
Document-Isolation-Policy: isolate-and-require-corp
```

This Chrome 125+ feature is designed specifically for iframe isolation. It works without COOP, allowing pages to be embedded while still achieving isolation.

### COEP requires CORP on cross-origin resources

When your page has `Cross-Origin-Embedder-Policy: require-corp`, all cross-origin resources (fetches, iframes, images, etc.) must have:
- `Cross-Origin-Resource-Policy: cross-origin` header, OR
- Proper CORS headers with `crossorigin` attribute

## Running Locally

```bash
go build -o server .
./server -port 8000 -alt-port 8001
```

This starts two servers for true cross-origin testing:
- `http://localhost:8000` - Main page
- `http://localhost:8001` - Cross-origin target

## Usage

1. **Configure your page's headers** - Set COOP, COEP, CORP, DIP for the main page
2. **Test iframe embedding** - Try different header combinations on cross-origin iframes
3. **Test fetch requests** - See how COEP affects cross-origin fetch with different CORP settings

## Header Reference

| Header | Purpose |
|--------|----------|
| `Cross-Origin-Opener-Policy: same-origin` | Isolates window.opener relationships (top-level only) |
| `Cross-Origin-Embedder-Policy: require-corp` | Requires CORP on all cross-origin resources |
| `Cross-Origin-Embedder-Policy: credentialless` | Loads cross-origin without credentials |
| `Cross-Origin-Resource-Policy: cross-origin` | Allows resource to be loaded cross-origin with COEP |
| `Document-Isolation-Policy: isolate-and-require-corp` | Enables isolation in iframes (Chrome 125+) |

## Iframe Attributes

| Attribute | Purpose |
|-----------|----------|
| `credentialless` | Loads iframe without credentials |
| `allow="cross-origin-isolated"` | Permissions policy for isolation |

## License

MIT
