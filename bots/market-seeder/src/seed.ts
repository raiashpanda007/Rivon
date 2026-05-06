import postgres from "postgres";
import bcrypt from "bcryptjs";
import { randomUUID } from "node:crypto";
import { mkdirSync } from "node:fs";
import { join } from "node:path";

type UserSeed = {
  id: string;
  email: string;
  name: string;
};

type AuthUser = UserSeed & {
  token: string;
  userId: string;
};

type Market = {
  id: string;
  marketName?: string;
  marketCode?: string;
  lastPrice?: number;
};

type MarketReport = {
  marketId: string;
  marketCode: string;
  targetOrders: number;
  sent: number;
  ok: number;
  filled: number;
  partiallyFilled: number;
  queued: number;
  accepted: number;
  errors: number;
};

type Report = {
  startedAt: string;
  finishedAt: string;
  config: Record<string, string | number | boolean>;
  totals: MarketReport;
  markets: MarketReport[];
};

const BASE_URL = process.env.BASE_URL ?? "http://localhost:8000";
const WS_URL = process.env.WS_URL ?? "ws://localhost:8003";
const DATABASE_POSTGRES_URL = process.env.DATABASE_POSTGRES_URL;
const PASSWORD = process.env.PASSWORD ?? "bot-password";
const USER_COUNT = parseIntEnv("USER_COUNT", 20);
const WALLET_BALANCE = parseIntEnv("WALLET_BALANCE", 500000000);
const ASSET_QTY = parseIntEnv("ASSET_QTY", 10000);
const PRICE_MIN_DOLLARS = parseFloatEnv("PRICE_MIN", 50);
const PRICE_MAX_DOLLARS = parseFloatEnv("PRICE_MAX", 500);
const QTY_MIN = parseIntEnv("QTY_MIN", 1);
const QTY_MAX = parseIntEnv("QTY_MAX", 5);
const ORDERS_PER_MARKET_MIN = parseIntEnv("ORDERS_PER_MARKET_MIN", 900);
const ORDERS_PER_MARKET_MAX = parseIntEnv("ORDERS_PER_MARKET_MAX", 1100);
const PAIR_RATIO = parseFloatEnv("PAIR_RATIO", 0.6);
const ORDER_JITTER_MS = parseIntEnv("ORDER_JITTER_MS", 10);
const MARKET_LIMIT = parseOptionalIntEnv("MARKET_LIMIT");
const WS_MODE = (process.env.WS_MODE ?? "per-market").toLowerCase();
const REPORT_DIR = process.env.REPORT_DIR ?? "reports";

if (!DATABASE_POSTGRES_URL) {
  throw new Error("DATABASE_POSTGRES_URL is required");
}

const PRICE_MIN_CENTS = Math.round(PRICE_MIN_DOLLARS * 100);
const PRICE_MAX_CENTS = Math.round(PRICE_MAX_DOLLARS * 100);

validateConfig();

async function main() {
  const startedAt = new Date().toISOString();
  const sql = postgres(DATABASE_POSTGRES_URL, { max: 1 });

  const passwordHash = await bcrypt.hash(PASSWORD, 10);
  const markets = await fetchMarkets(BASE_URL);
  const trimmedMarkets = MARKET_LIMIT ? markets.slice(0, MARKET_LIMIT) : markets;

  console.log(`Seeding ${USER_COUNT} users and ${trimmedMarkets.length} markets...`);
  const users = await seedUsers(sql, passwordHash, USER_COUNT, WALLET_BALANCE);
  await seedAssets(sql, users, trimmedMarkets, ASSET_QTY, PRICE_MIN_CENTS);

  console.log("Logging in users and warming wallets...");
  const authedUsers = await loginAndWarmUsers(users, BASE_URL, PASSWORD);

  const totals: MarketReport = {
    marketId: "TOTAL",
    marketCode: "TOTAL",
    targetOrders: 0,
    sent: 0,
    ok: 0,
    filled: 0,
    partiallyFilled: 0,
    queued: 0,
    accepted: 0,
    errors: 0,
  };

  const marketReports: MarketReport[] = [];

  for (let i = 0; i < trimmedMarkets.length; i += 1) {
    const market = trimmedMarkets[i];
    const marketCode = market.marketCode ?? market.marketName ?? market.id;
    const targetOrders = randomInt(ORDERS_PER_MARKET_MIN, ORDERS_PER_MARKET_MAX);
    const pairRatio = clamp(PAIR_RATIO + randomFloat(-0.1, 0.1), 0.2, 0.9);
    const pairedOrderCount = Math.floor((targetOrders * pairRatio) / 2) * 2;

    console.log(`\n[${i + 1}/${trimmedMarkets.length}] ${marketCode} -> ${targetOrders} orders`);

    const wsConnections = await openMarketWsConnections(
      WS_URL,
      market.id,
      authedUsers,
      WS_MODE
    );

    const report: MarketReport = {
      marketId: market.id,
      marketCode,
      targetOrders,
      sent: 0,
      ok: 0,
      filled: 0,
      partiallyFilled: 0,
      queued: 0,
      accepted: 0,
      errors: 0,
    };

    await placePairedOrders(
      market.id,
      pairedOrderCount,
      authedUsers,
      BASE_URL,
      report
    );

    const remaining = targetOrders - pairedOrderCount;
    await placeRandomOrders(market.id, remaining, authedUsers, BASE_URL, report);

    closeWsConnections(wsConnections, market.id);

    marketReports.push(report);
    accumulateReport(totals, report);

    console.log(
      `Done ${marketCode}: sent=${report.sent}, ok=${report.ok}, filled=${report.filled}, errors=${report.errors}`
    );
  }

  const finishedAt = new Date().toISOString();
  const report: Report = {
    startedAt,
    finishedAt,
    config: buildReportConfig(trimmedMarkets.length),
    totals,
    markets: marketReports,
  };

  await writeReport(report);
  await sql.end({ timeout: 5 });

  console.log("\nAll markets completed.");
}

function validateConfig() {
  if (PRICE_MIN_CENTS <= 0 || PRICE_MAX_CENTS <= 0 || PRICE_MIN_CENTS >= PRICE_MAX_CENTS) {
    throw new Error("Invalid PRICE_MIN/PRICE_MAX range");
  }
  if (QTY_MIN <= 0 || QTY_MAX <= 0 || QTY_MIN > QTY_MAX) {
    throw new Error("Invalid QTY_MIN/QTY_MAX range");
  }
  if (ORDERS_PER_MARKET_MIN <= 0 || ORDERS_PER_MARKET_MAX <= 0 || ORDERS_PER_MARKET_MIN > ORDERS_PER_MARKET_MAX) {
    throw new Error("Invalid ORDERS_PER_MARKET_MIN/ORDERS_PER_MARKET_MAX range");
  }
  if (PAIR_RATIO < 0 || PAIR_RATIO > 1) {
    throw new Error("PAIR_RATIO must be between 0 and 1");
  }
  if (!WS_URL.startsWith("ws")) {
    throw new Error("WS_URL must be a ws:// or wss:// URL");
  }
  if (WS_MODE !== "per-market" && WS_MODE !== "per-user") {
    throw new Error("WS_MODE must be per-market or per-user");
  }
}

function buildReportConfig(marketCount: number) {
  return {
    baseUrl: BASE_URL,
    wsUrl: WS_URL,
    userCount: USER_COUNT,
    marketCount,
    walletBalance: WALLET_BALANCE,
    assetQty: ASSET_QTY,
    priceMinDollars: PRICE_MIN_DOLLARS,
    priceMaxDollars: PRICE_MAX_DOLLARS,
    qtyMin: QTY_MIN,
    qtyMax: QTY_MAX,
    ordersMin: ORDERS_PER_MARKET_MIN,
    ordersMax: ORDERS_PER_MARKET_MAX,
    pairRatio: PAIR_RATIO,
    orderJitterMs: ORDER_JITTER_MS,
    wsMode: WS_MODE,
  };
}

async function fetchMarkets(baseUrl: string): Promise<Market[]> {
  const res = await fetch(`${baseUrl}/api/rivon/markets`);
  if (!res.ok) {
    throw new Error(`Failed to fetch markets: ${res.status}`);
  }
  const body = await res.json();
  if (!body?.data || !Array.isArray(body.data)) {
    throw new Error("Markets response is missing data array");
  }
  return body.data as Market[];
}

async function seedUsers(
  sql: ReturnType<typeof postgres>,
  passwordHash: string,
  userCount: number,
  walletBalance: number
): Promise<UserSeed[]> {
  const users: UserSeed[] = [];

  await sql.begin(async (tx) => {
    for (let i = 1; i <= userCount; i += 1) {
      const email = `seed-bot-${String(i).padStart(3, "0")}@rivon.local`;
      const name = `Seed Bot ${String(i).padStart(2, "0")}`;
      const userId = randomUUID();

      const created = await tx<UserSeed[]>`
        INSERT INTO users (
          id,
          type,
          name,
          email,
          provider,
          password_hash,
          verified
        ) VALUES (
          ${userId},
          'user',
          ${name},
          ${email},
          'credentials',
          ${passwordHash},
          true
        )
        ON CONFLICT (email, provider)
        DO UPDATE SET
          name = EXCLUDED.name,
          password_hash = EXCLUDED.password_hash,
          verified = true
        RETURNING id, email, name;
      `;

      const user = created[0];
      users.push({ id: user.id, email: user.email, name: user.name });

      await tx`
        INSERT INTO wallets (id, user_id, balance)
        VALUES (${randomUUID()}, ${user.id}, ${walletBalance})
        ON CONFLICT (user_id)
        DO UPDATE SET balance = EXCLUDED.balance, updated_at = NOW();
      `;
    }
  });

  return users;
}

async function seedAssets(
  sql: ReturnType<typeof postgres>,
  users: UserSeed[],
  markets: Market[],
  qty: number,
  avgCost: number
) {
  await sql.begin(async (tx) => {
    for (const user of users) {
      for (const market of markets) {
        await tx`
          INSERT INTO assets (
            id,
            user_id,
            market_id,
            quantity,
            avg_cost,
            updated_at
          ) VALUES (
            ${randomUUID()},
            ${user.id},
            ${market.id},
            ${qty},
            ${avgCost},
            NOW()
          )
          ON CONFLICT (user_id, market_id)
          DO UPDATE SET
            quantity = GREATEST(assets.quantity, EXCLUDED.quantity),
            avg_cost = EXCLUDED.avg_cost,
            updated_at = NOW();
        `;
      }
    }
  });
}

async function loginAndWarmUsers(users: UserSeed[], baseUrl: string, password: string) {
  const authed: AuthUser[] = [];

  for (const user of users) {
    const loggedIn = await loginUser(user, baseUrl, password);
    await warmWallet(baseUrl, loggedIn.token);
    authed.push(loggedIn);
  }

  return authed;
}

async function loginUser(user: UserSeed, baseUrl: string, password: string): Promise<AuthUser> {
  const res = await fetch(`${baseUrl}/api/rivon/auth/credentials/signin`, {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify({ email: user.email, password }),
  });

  if (!res.ok) {
    throw new Error(`Login failed for ${user.email}: ${res.status}`);
  }

  const accessToken = extractAccessToken(res);
  if (!accessToken) {
    throw new Error(`Missing access_token cookie for ${user.email}`);
  }

  const body = await res.json();
  const userId = body?.data?.id ?? user.id;

  return {
    ...user,
    token: accessToken,
    userId,
  };
}

async function warmWallet(baseUrl: string, token: string) {
  const res = await fetch(`${baseUrl}/api/rivon/wallet/me`, {
    headers: { cookie: `access_token=${token}` },
  });
  if (!res.ok) {
    throw new Error(`wallet/me failed: ${res.status}`);
  }
}

function extractAccessToken(res: Response): string | null {
  const headerAny = res.headers as unknown as { getSetCookie?: () => string[] };
  const setCookies = headerAny.getSetCookie?.() ?? (res.headers.get("set-cookie") ? [res.headers.get("set-cookie") as string] : []);

  for (const cookie of setCookies) {
    const match = cookie.match(/access_token=([^;]+)/);
    if (match) return match[1];
  }
  return null;
}

async function openMarketWsConnections(
  wsUrl: string,
  marketId: string,
  users: AuthUser[],
  wsMode: string
): Promise<WebSocket[]> {
  if (wsMode === "per-user") {
    const sockets: WebSocket[] = [];
    for (const user of users) {
      const ws = await openMarketWs(wsUrl, marketId, user.userId);
      sockets.push(ws);
    }
    return sockets;
  }

  const representative = users[0];
  return [await openMarketWs(wsUrl, marketId, representative.userId)];
}

function closeWsConnections(connections: WebSocket[], marketId: string) {
  for (const ws of connections) {
    try {
      ws.send(JSON.stringify({ type: "UNSUBSCRIBE_MARKET", payload: { marketID: marketId } }));
    } catch {
      // ignore
    }
    ws.close();
  }
}

function openMarketWs(wsUrl: string, marketId: string, userId: string): Promise<WebSocket> {
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(wsUrl);
    let settled = false;

    const timeout = setTimeout(() => {
      if (settled) return;
      settled = true;
      ws.close();
      reject(new Error(`WS subscribe timeout for market ${marketId}`));
    }, 5000);

    ws.onopen = () => {
      ws.send(
        JSON.stringify({
          type: "SUBSCRIBE_MARKET",
          payload: { marketID: marketId, userID: userId },
        })
      );
    };

    ws.onmessage = (event) => {
      if (settled) return;
      try {
        const msg = JSON.parse(String(event.data));
        if (
          msg?.type === "SUBSCRIBED" ||
          msg?.type === "ORDERBOOK_DATA" ||
          msg?.type === "ORDERBOOK_UPDATE"
        ) {
          settled = true;
          clearTimeout(timeout);
          resolve(ws);
        }
      } catch {
        // ignore
      }
    };

    ws.onerror = () => {
      if (settled) return;
      settled = true;
      clearTimeout(timeout);
      reject(new Error(`WS error for market ${marketId}`));
    };

    ws.onclose = () => {
      if (settled) return;
      settled = true;
      clearTimeout(timeout);
      reject(new Error(`WS closed before subscribe for market ${marketId}`));
    };
  });
}

async function placePairedOrders(
  marketId: string,
  count: number,
  users: AuthUser[],
  baseUrl: string,
  report: MarketReport
) {
  const pairs = Math.floor(count / 2);
  for (let i = 0; i < pairs; i += 1) {
    const { price, quantity } = randomPriceQty();
    const [seller, buyer] = pickTwoUsers(users);

    await placeOrder(baseUrl, seller, marketId, "SELL", price, quantity, report);
    await jitter();
    await placeOrder(baseUrl, buyer, marketId, "BUY", price, quantity, report);
    await jitter();

    if ((i + 1) % 50 === 0) {
      console.log(`  paired orders: ${i + 1}/${pairs}`);
    }
  }
}

async function placeRandomOrders(
  marketId: string,
  count: number,
  users: AuthUser[],
  baseUrl: string,
  report: MarketReport
) {
  for (let i = 0; i < count; i += 1) {
    const { price, quantity } = randomPriceQty();
    const side = Math.random() > 0.5 ? "BUY" : "SELL";
    const user = pickRandom(users);

    await placeOrder(baseUrl, user, marketId, side, price, quantity, report);
    await jitter();

    if ((i + 1) % 100 === 0) {
      console.log(`  random orders: ${i + 1}/${count}`);
    }
  }
}

async function placeOrder(
  baseUrl: string,
  user: AuthUser,
  marketId: string,
  side: "BUY" | "SELL",
  price: number,
  quantity: number,
  report: MarketReport
) {
  report.sent += 1;

  const res = await fetch(`${baseUrl}/api/rivon/markets/create-order`, {
    method: "POST",
    headers: {
      "content-type": "application/json",
      cookie: `access_token=${user.token}`,
    },
    body: JSON.stringify({
      marketId,
      price,
      quantity,
      orderType: side,
    }),
  });

  if (res.status === 401 || res.status === 403) {
    const relogged = await loginUser({ id: user.id, email: user.email, name: user.name }, BASE_URL, PASSWORD);
    user.token = relogged.token;
    user.userId = relogged.userId;
    await warmWallet(baseUrl, user.token);
    return placeOrder(baseUrl, user, marketId, side, price, quantity, report);
  }

  if (!res.ok) {
    report.errors += 1;
    return;
  }

  report.ok += 1;

  const body = await res.json().catch(() => null);
  const status = body?.data?.status as string | undefined;

  switch (status) {
    case "filled":
      report.filled += 1;
      break;
    case "partially_filled":
      report.partiallyFilled += 1;
      break;
    case "queued":
      report.queued += 1;
      break;
    case "accepted":
      report.accepted += 1;
      break;
    default:
      break;
  }
}

function randomPriceQty() {
  return {
    price: randomInt(PRICE_MIN_CENTS, PRICE_MAX_CENTS),
    quantity: randomInt(QTY_MIN, QTY_MAX),
  };
}

function pickRandom<T>(items: T[]): T {
  return items[randomInt(0, items.length - 1)];
}

function pickTwoUsers(users: AuthUser[]) {
  const first = pickRandom(users);
  let second = pickRandom(users);
  while (second.id === first.id) {
    second = pickRandom(users);
  }
  return [first, second] as const;
}

function accumulateReport(totals: MarketReport, report: MarketReport) {
  totals.targetOrders += report.targetOrders;
  totals.sent += report.sent;
  totals.ok += report.ok;
  totals.filled += report.filled;
  totals.partiallyFilled += report.partiallyFilled;
  totals.queued += report.queued;
  totals.accepted += report.accepted;
  totals.errors += report.errors;
}

async function writeReport(report: Report) {
  mkdirSync(REPORT_DIR, { recursive: true });
  const fileName = `seed-report-${new Date().toISOString().replace(/[:.]/g, "-")}.json`;
  const filePath = join(REPORT_DIR, fileName);
  await Bun.write(filePath, JSON.stringify(report, null, 2));
  console.log(`Report written to ${filePath}`);
}

function parseIntEnv(name: string, fallback: number) {
  const raw = process.env[name];
  if (!raw) return fallback;
  const val = Number.parseInt(raw, 10);
  if (Number.isNaN(val)) throw new Error(`Invalid ${name}`);
  return val;
}

function parseOptionalIntEnv(name: string): number | undefined {
  const raw = process.env[name];
  if (!raw) return undefined;
  const val = Number.parseInt(raw, 10);
  if (Number.isNaN(val)) throw new Error(`Invalid ${name}`);
  return val;
}

function parseFloatEnv(name: string, fallback: number) {
  const raw = process.env[name];
  if (!raw) return fallback;
  const val = Number.parseFloat(raw);
  if (Number.isNaN(val)) throw new Error(`Invalid ${name}`);
  return val;
}

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value));
}

function randomInt(min: number, max: number) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function randomFloat(min: number, max: number) {
  return Math.random() * (max - min) + min;
}

function jitter() {
  if (ORDER_JITTER_MS <= 0) return Promise.resolve();
  const extra = randomInt(0, ORDER_JITTER_MS);
  return sleep(ORDER_JITTER_MS + extra);
}

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
