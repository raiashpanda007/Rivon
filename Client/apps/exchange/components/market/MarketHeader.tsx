import type { MarketData } from "@/app/markets/[market]/types"

interface Props {
  market: MarketData | null
  livePrice?: number | null
}

export function MarketHeader({ market, livePrice }: Props) {
  const displayPrice = livePrice ?? market?.lastPrice ?? null
  const priceChange = displayPrice != null && market ? displayPrice - market.openPrice : 0
  const priceChangePct =
    market && market.openPrice > 0 ? (priceChange / market.openPrice) * 100 : 0

  return (
    <div className="terminal-panel rounded-none border-x-0 border-t-0">
      <div className="flex items-center gap-5 px-4 py-2.5 overflow-x-auto">
        {market?.teamDetails?.emblem && (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            src={market.teamDetails.emblem}
            alt={market.teamDetails.name}
            width={28}
            height={28}
            className="rounded-sm shrink-0 object-contain"
          />
        )}

        <div className="flex items-center gap-2 shrink-0">
          <span className="font-mono font-semibold text-sm text-foreground">
            {market?.marketName ?? "—"}
          </span>
          <span className="font-mono text-[10px] px-1.5 py-0.5 bg-orange-500/10 text-orange-400 border border-orange-500/30 rounded">
            {market?.marketCode ?? "—"}
          </span>
        </div>

        <div className="flex items-center gap-2 shrink-0">
          <span className="font-mono text-xl font-bold text-orange-400">
            {displayPrice ?? "—"}
          </span>
          <span className={`font-mono text-xs ${priceChange >= 0 ? "text-up" : "text-down"}`}>
            {priceChange >= 0 ? "+" : ""}
            {priceChange} ({priceChangePct.toFixed(2)}%)
          </span>
        </div>

        <div className="hidden md:flex items-center gap-1.5 shrink-0">
          <span className="font-mono text-[10px] text-muted-foreground">24H VOL</span>
          <span className="font-mono text-xs">{market?.volume24h?.toLocaleString() ?? "—"}</span>
        </div>

        <div className="hidden md:flex items-center gap-1.5 shrink-0">
          <span className="font-mono text-[10px] text-muted-foreground">OPEN</span>
          <span className="font-mono text-xs">{market?.openPrice ?? "—"}</span>
        </div>

        <div className="flex items-center gap-1.5 shrink-0 ml-auto">
          <span
            className={`w-1.5 h-1.5 rounded-full ${
              market?.status === "open" ? "bg-green-500 animate-pulse" : "bg-muted-foreground"
            }`}
          />
          <span className="font-mono text-[10px] text-muted-foreground uppercase">
            {market?.status ?? "—"}
          </span>
        </div>
      </div>
    </div>
  )
}
