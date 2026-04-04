import RecommendationCards from "@/components/Landing/Cards/RecommendationCards"
import MarketLists from "@/components/Landing/MarketLists"
import Link from "next/link"
import { Button } from "@workspace/ui/components/button"
import { ArrowUpRight } from "lucide-react"

export default function Page() {
  return (
    <div className="relative min-h-[calc(100vh-3.5rem)] flex flex-col overflow-hidden">
      {/* Ambient Glows */}
      <div className="absolute top-[0%] right-[0%] -z-10 w-[500px] h-[500px] rounded-full bg-orange-500/10 blur-[120px] pointer-events-none" />
      <div className="absolute top-[40%] left-[-10%] -z-10 w-[400px] h-[400px] rounded-full bg-blue-500/5 blur-[100px] pointer-events-none" />

      {/* System status bar */}
      <div className="shrink-0 border-b border-border bg-muted/15 px-4 md:px-6 py-1.5 flex items-center gap-5 overflow-x-auto">
        <div className="flex items-center gap-1.5 shrink-0">
          <span className="w-1 h-1 rounded-full bg-green-500 animate-pulse" />
          <span className="text-[10px] text-muted-foreground">
            ENGINE <span className="text-green-400">ONLINE</span>
          </span>
        </div>
        <div className="h-3 w-px bg-border shrink-0" />
        <span className="text-[10px] text-muted-foreground shrink-0">
          LATENCY <span className="text-orange-400">12ms</span>
        </span>
        <div className="h-3 w-px bg-border shrink-0" />
        <span className="text-[10px] text-muted-foreground shrink-0">
          MODE <span className="text-foreground">LIMIT</span>
        </span>
        <div className="ml-auto text-[10px] text-muted-foreground/40 shrink-0 hidden sm:block">
          RIVON · EXCHANGE · v1.0
        </div>
      </div>

      <div className="px-4 md:px-6 py-5 flex-1">

        {/* Page header */}
        <div className="flex flex-col md:flex-row md:items-end justify-between mb-8 pb-6 border-b border-border gap-6">
          <div>
            <div className="flex items-center gap-2 mb-3">
              <div className="px-2 py-1 rounded-sm bg-orange-500/10 border border-orange-500/20 flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-orange-500 animate-pulse shadow-[0_0_8px_rgba(249,115,22,0.8)]" />
                <span className="font-mono text-[10px] font-bold text-orange-500 tracking-widest">LIVE_EXCHANGE</span>
              </div>
            </div>
            <h1 className="text-5xl sm:text-6xl md:text-7xl font-black tracking-tighter text-foreground leading-none">
              Global <span className="text-transparent bg-clip-text bg-gradient-to-r from-orange-400 to-orange-600">Markets</span>
            </h1>
            <p className="font-mono text-xs text-muted-foreground mt-4 max-w-xl">
              Real-time trading execution across all sports markets. Deep liquidity and pure price discovery.
            </p>
          </div>

          {/* Aggregate stats */}
          <div className="hidden sm:grid grid-cols-3 divide-x divide-border border border-border bg-card">
            <div className="px-4 py-2 flex flex-col items-center">
              <span className="text-[9px] text-muted-foreground/60 tracking-widest mb-0.5">24H VOL</span>
              <span className="text-xs font-bold text-foreground tabular-nums">$248,401</span>
            </div>
            <div className="px-4 py-2 flex flex-col items-center">
              <span className="text-[9px] text-muted-foreground/60 tracking-widest mb-0.5">MARKETS</span>
              <span className="text-xs font-bold text-green-400 tabular-nums">24 OPEN</span>
            </div>
            <div className="px-4 py-2 flex flex-col items-center">
              <span className="text-[9px] text-muted-foreground/60 tracking-widest mb-0.5">SPREAD</span>
              <span className="text-xs font-bold text-foreground tabular-nums">0.02%</span>
            </div>
          </div>
        </div>

        {/* Recommendation panels row */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 mb-5">
          <RecommendationCards by="FAV" />
          <RecommendationCards by="TOP" />
          <RecommendationCards by="POPULAR" />
        </div>

        {/* All Markets */}
        <div className="mt-8">
          <div className="flex flex-col md:flex-row md:items-end justify-between mb-5 gap-4">
            <div>
              <div className="flex items-center gap-2 mb-2">
                <span className="w-1.5 h-1.5 rounded-full bg-orange-500 shadow-[0_0_8px_rgba(249,115,22,0.5)]" />
                <span className="font-mono text-[10px] text-orange-400 tracking-widest font-bold">MARKET_INDEX</span>
              </div>
              <h2 className="text-3xl sm:text-4xl font-extrabold text-foreground tracking-tight">Active Instruments</h2>
            </div>
            <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4 border ">
              <Link href="/leagues " className="border border-orange-500">
                <Button variant="outline" size="sm" className="h-8 font-mono text-[10px] pr-2 pl-3 bg-orange-500/5 text-orange-400 hover:text-orange-500 hover:bg-orange-500/10 rounded-sm shadow-[0_0_15px_rgba(249,115,22,0.1)] cursor-pointer border-orange-500">
                  CHECK_STANDINGS <ArrowUpRight className="w-3 h-3 ml-1" />
                </Button>
              </Link>
              <div className="flex items-center gap-4 bg-muted/20 px-3 py-1.5 border border-border rounded-sm">
                <span className="font-mono text-[10px] text-muted-foreground">SORTED BY <span className="text-foreground">CODE ↑</span></span>
              </div>
            </div>
          </div>
          <MarketLists />
        </div>

      </div>
    </div >
  )
}
