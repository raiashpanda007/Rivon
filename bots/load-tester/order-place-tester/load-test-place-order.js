#!/usr/bin/env node

/**
 * Autocannon load test for POST /api/rivon/markets/create-order
 *
 * Usage:
 *   TOKEN=<jwt> MARKET_ID=<uuid> node load-test-place-order.js [options]
 *
 * Options:
 *   --duration  1 | 5 | 10        minutes  (default: 1)
 *   --rate      10 | 20 | 50 | 100  req/s  (default: 10)
 *   --url       base URL           (default: http://localhost:8080)
 *
 * Examples:
 *   TOKEN=abc MARKET_ID=uuid node load-test-place-order.js --duration 5 --rate 50
 */

const autocannon = require("autocannon");

// ── Config from env / CLI args ────────────────────────────────────────────────
const args = process.argv.slice(2);

function arg(name, fallback) {
  const idx = args.indexOf(`--${name}`);
  return idx !== -1 ? args[idx + 1] : fallback;
}

const BASE_URL = arg("url", process.env.BASE_URL || "http://localhost:8080");
const TOKEN = process.env.TOKEN;
const MARKET_ID = process.env.MARKET_ID;

if (!TOKEN) {
  console.error("ERROR: TOKEN env var is required  (export TOKEN=<jwt>)");
  process.exit(1);
}
if (!MARKET_ID) {
  console.error("ERROR: MARKET_ID env var is required  (export MARKET_ID=<uuid>)");
  process.exit(1);
}

// ── Allowed modes ─────────────────────────────────────────────────────────────
const VALID_DURATIONS = [1, 2, 5, 10];        // minutes
const VALID_RATES = [10, 20, 50, 100, 500];   // req/sec

const durationMin = parseInt(arg("duration", "1"), 10);
const rate = parseInt(arg("rate", "10"), 10);

if (!VALID_DURATIONS.includes(durationMin)) {
  console.error(`ERROR: --duration must be one of ${VALID_DURATIONS.join(", ")} (minutes)`);
  process.exit(1);
}
if (!VALID_RATES.includes(rate)) {
  console.error(`ERROR: --rate must be one of ${VALID_RATES.join(", ")} (req/sec)`);
  process.exit(1);
}

const durationSec = durationMin * 60;

// ── Request body ──────────────────────────────────────────────────────────────
// Alternates BUY / SELL on each request to keep the orderbook active.
// Price and quantity are randomised within a tight band so fills actually happen.
function makeBody(i) {
  const side = i % 2 === 0 ? "BUY" : "SELL";
  // price band 95–105 so orders cross and generate fills
  const price = 95 + Math.floor(Math.random() * 11);
  const quantity = 1 + Math.floor(Math.random() * 5);
  return JSON.stringify({
    marketId: MARKET_ID,
    price,
    quantity,
    orderType: side,
  });
}

// ── Run ───────────────────────────────────────────────────────────────────────
console.log(`
╔══════════════════════════════════════════════════════╗
║          Rivon — Place Order Load Test               ║
╠══════════════════════════════════════════════════════╣
║  URL       : ${(BASE_URL + "/api/rivon/markets/create-order").padEnd(38)}║
║  Duration  : ${String(durationMin + " min").padEnd(38)}║
║  Rate      : ${String(rate + " req/sec").padEnd(38)}║
║  Market ID : ${MARKET_ID.padEnd(38)}║
╚══════════════════════════════════════════════════════╝
`);

let requestIndex = 0;

const instance = autocannon(
  {
    url: `${BASE_URL}/api/rivon/markets/create-order`,
    method: "POST",
    duration: durationSec,
    amount: undefined,          // duration-based, not count-based
    connections: Math.ceil(rate / 5), // ~5 req/conn keeps pipeline sane
    pipelining: 1,
    overallRate: rate,             // autocannon honours this as a token-bucket cap
    headers: {
      "content-type": "application/json",
      // JWT is sent as a cookie — mirrors the browser auth flow
      cookie: `access_token=${TOKEN}`,
    },
    setupClient(client) {
      client.setBody(makeBody(requestIndex++));
      client.on("response", () => {
        client.setBody(makeBody(requestIndex++));
      });
    },
  },
  (err, result) => {
    if (err) {
      console.error("Autocannon error:", err);
      process.exit(1);
    }
    printSummary(result, durationMin, rate);
  }
);

autocannon.track(instance, { renderProgressBar: true });

// ── Pretty summary ────────────────────────────────────────────────────────────
function printSummary(r, durMin, rateTarget) {
  console.log(`
┌─────────────────────────────────────────────────────┐
│                   RESULTS SUMMARY                   │
├──────────────────────────┬──────────────────────────┤
│  Duration                │ ${String(durMin + " min").padEnd(24)} │
│  Target rate             │ ${String(rateTarget + " req/sec").padEnd(24)} │
│  Actual req/sec (avg)    │ ${String(r.requests.average.toFixed(1) + " req/sec").padEnd(24)} │
│  Total requests          │ ${String(r.requests.total).padEnd(24)} │
├──────────────────────────┼──────────────────────────┤
│  Latency p50             │ ${String(r.latency.p50 + " ms").padEnd(24)} │
│  Latency p90             │ ${String(r.latency.p90 + " ms").padEnd(24)} │
│  Latency p99             │ ${String(r.latency.p99 + " ms").padEnd(24)} │
│  Latency max             │ ${String(r.latency.max + " ms").padEnd(24)} │
├──────────────────────────┼──────────────────────────┤
│  2xx responses           │ ${String(r["2xx"]).padEnd(24)} │
│  Non-2xx responses       │ ${String(r.non2xx).padEnd(24)} │
│  Errors                  │ ${String(r.errors).padEnd(24)} │
│  Timeouts                │ ${String(r.timeouts).padEnd(24)} │
├──────────────────────────┼──────────────────────────┤
│  Throughput (avg)        │ ${String((r.throughput.average / 1024).toFixed(1) + " KB/sec").padEnd(24)} │
└──────────────────────────┴──────────────────────────┘
`);

  if (r.non2xx > 0 || r.errors > 0) {
    console.warn("⚠  Non-2xx / errors detected — check server logs.");
  } else {
    console.log("✓  All responses were 2xx.");
  }
}
