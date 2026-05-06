import type { MarketData, OrderBookData, WsStatus } from "@/app/markets/[market]/types"

interface Props {
  market: MarketData | null
  orderBook: OrderBookData
  livePrice: number | null
  wsStatus: WsStatus
}

const STATUS_DOT: Record<WsStatus, string> = {
  live: "dot-live",
  connecting: "w-1.5 h-1.5 rounded-full bg-amber-500",
  disconnected: "dot-down",
}

const STATUS_LABEL: Record<WsStatus, string> = {
  live: "LIVE",
  connecting: "CONNECTING",
  disconnected: "DISCONNECTED",
}

export function OrderBook({ market, orderBook, livePrice, wsStatus }: Props) {
  const displayPrice = livePrice ?? market?.lastPrice ?? null
  const priceChange = displayPrice != null && market ? displayPrice - market.openPrice : 0
  const maxAskQty = orderBook.asks.reduce((m, r) => Math.max(m, r.quantity), 1)
  const maxBidQty = orderBook.bids.reduce((m, r) => Math.max(m, r.quantity), 1)

  return (
    <div className="terminal-panel rounded-none border-x-0 border-t-0 w-full flex-1 min-h-0 flex flex-col">
      <div className="terminal-panel-header">
        <span className={STATUS_DOT[wsStatus]} />
        <span className="font-mono text-[10px] text-muted-foreground uppercase tracking-widest">
          Order Book
        </span>
        <span className={`ml-auto font-mono text-[9px] uppercase tracking-widest ${
          wsStatus === "live" ? "text-orange-400" :
          wsStatus === "connecting" ? "text-amber-500" : "text-red-500"
        }`}>
          {STATUS_LABEL[wsStatus]}
        </span>
      </div>

      {/* Column labels */}
      <div className="grid grid-cols-3 px-3 py-1 border-b border-border/50">
        <span className="font-mono text-[10px] text-muted-foreground">Price</span>
        <span className="font-mono text-[10px] text-muted-foreground text-right">Qty</span>
        <span className="font-mono text-[10px] text-muted-foreground text-right">Total</span>
      </div>

      {/* Asks — reversed so lowest ask is nearest the price row */}
      <div className="flex flex-col-reverse overflow-y-auto flex-1 min-h-0">
        {orderBook.asks.length === 0 ? (
          <div className="flex items-center justify-center h-24 text-muted-foreground font-mono text-xs">
            {wsStatus === "connecting" ? "Connecting…" : "Awaiting data…"}
          </div>
        ) : (
          orderBook.asks.map((level, i) => (
            <div key={i} className="relative grid grid-cols-3 px-3 py-[3px] hover:bg-muted/20">
              <div
                className="absolute right-0 top-0 h-full bg-red-500/10 pointer-events-none z-0"
                style={{ width: `${(level.quantity / maxAskQty) * 80}%` }}
              />
              <span className="text-down font-mono text-xs z-10">{level.price}</span>
              <span className="font-mono text-xs text-foreground z-10 text-right">{level.quantity}</span>
              <span className="font-mono text-xs text-muted-foreground z-10 text-right">{level.total}</span>
            </div>
          ))
        )}
      </div>

      {/* Last price divider */}
      <div className="flex items-center gap-2 px-3 py-1.5 border-y border-orange-500/30 bg-orange-500/5 shrink-0">
        <span className="font-mono text-sm font-bold text-orange-400">
          {displayPrice ?? "—"}
        </span>
        <span className={`font-mono text-[10px] ${priceChange >= 0 ? "text-up" : "text-down"}`}>
          {priceChange >= 0 ? "▲" : "▼"}
        </span>
      </div>

      {/* Bids */}
      <div className="overflow-y-auto flex-1 min-h-0">
        {orderBook.bids.length === 0 ? (
          <div className="flex items-center justify-center h-24 text-muted-foreground font-mono text-xs">
            {wsStatus === "connecting" ? "Connecting…" : "Awaiting data…"}
          </div>
        ) : (
          orderBook.bids.map((level, i) => (
            <div key={i} className="relative grid grid-cols-3 px-3 py-[3px] hover:bg-muted/20">
              <div
                className="absolute right-0 top-0 h-full bg-green-500/10 pointer-events-none z-0"
                style={{ width: `${(level.quantity / maxBidQty) * 80}%` }}
              />
              <span className="text-up font-mono text-xs z-10">{level.price}</span>
              <span className="font-mono text-xs text-foreground z-10 text-right">{level.quantity}</span>
              <span className="font-mono text-xs text-muted-foreground z-10 text-right">{level.total}</span>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
