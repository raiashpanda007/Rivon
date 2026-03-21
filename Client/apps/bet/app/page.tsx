"use client"

import { useState } from "react"
import { Coins, TrendingUp, Clock } from "lucide-react"
import { Button } from "@workspace/ui/components/button"

const MOCK_MARKETS = [
    { id: 1, home: "Real Madrid", away: "Bayern Munich", homeOdds: "2.10", drawOdds: "3.40", awayOdds: "3.25", kickoff: "19:45", league: "UCL" },
    { id: 2, home: "Arsenal", away: "Man City", homeOdds: "2.80", drawOdds: "3.10", awayOdds: "2.50", kickoff: "20:00", league: "EPL" },
    { id: 3, home: "PSG", away: "Inter Milan", homeOdds: "1.95", drawOdds: "3.50", awayOdds: "3.80", kickoff: "20:45", league: "UCL" },
    { id: 4, home: "Man United", away: "Liverpool", homeOdds: "2.60", drawOdds: "3.20", awayOdds: "2.70", kickoff: "17:30", league: "EPL" },
]

const ACTIVE_BETS = [
    { market: "RMFC WIN", odds: "2.10", stake: "50.00", potential: "105.00", status: "open" },
    { market: "MCFC WIN", odds: "1.85", stake: "30.00", potential: "55.50", status: "open" },
]

export default function Page() {
    const [selectedBets, setSelectedBets] = useState<Record<string, string>>({})

    const selectBet = (marketId: number, outcome: string) => {
        setSelectedBets(prev => {
            const key = String(marketId)
            if (prev[key] === outcome) {
                const next = { ...prev }
                delete next[key]
                return next
            }
            return { ...prev, [key]: outcome }
        })
    }

    return (
        <div className="relative min-h-[calc(100vh-3.5rem)] px-4 md:px-6 py-6 overflow-hidden">
            {/* Ambient Glows */}
            <div className="absolute top-[10%] left-[5%] -z-10 w-[400px] h-[400px] rounded-full bg-orange-500/10 blur-[120px] pointer-events-none" />
            <div className="absolute bottom-[20%] right-[10%] -z-10 w-[500px] h-[500px] rounded-full bg-orange-500/5 blur-[120px] pointer-events-none" />
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-end justify-between mb-8 pb-6 border-b border-border gap-6">
                <div>
                    <div className="flex items-center gap-2 mb-3">
                        <div className="px-2 py-1 rounded-sm bg-orange-500/10 border border-orange-500/20 flex items-center gap-2">
                            <span className="w-1.5 h-1.5 rounded-full bg-orange-500 animate-pulse shadow-[0_0_8px_rgba(249,115,22,0.8)]" />
                            <span className="font-mono text-[10px] font-bold text-orange-500 tracking-widest">LIVE_ODDS</span>
                        </div>
                    </div>
                    <h1 className="text-5xl sm:text-6xl md:text-7xl font-black tracking-tighter text-foreground leading-none">
                        Sportsbook
                    </h1>
                    <p className="font-mono text-xs text-muted-foreground mt-4 max-w-xl">
                        Direct betting access with algorithmically driven odds combined with real-time exchange liquidity.
                    </p>
                </div>
                <div className="hidden sm:flex items-center gap-2 px-4 py-2 border border-orange-500/30 bg-orange-500/5 rounded-sm text-xs font-mono font-bold text-orange-400 shadow-[0_0_15px_rgba(249,115,22,0.1)]">
                    <Coins className="w-4 h-4" />
                    BALANCE: ₹ 0.00
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
                {/* Main markets */}
                <div className="lg:col-span-2 space-y-4">
                    <div className="flex items-center gap-2 mb-2">
                        <span className="font-mono text-[10px] text-muted-foreground">UPCOMING_FIXTURES</span>
                        <span className="font-mono text-[10px] text-orange-400 ml-auto">{MOCK_MARKETS.length} MARKETS</span>
                    </div>

                    {MOCK_MARKETS.map((match) => (
                        <div
                            key={match.id}
                            className="terminal-panel overflow-hidden"
                        >
                            <div className="terminal-panel-header">
                                <span className="font-mono text-[9px] text-orange-400">{match.league}</span>
                                <div className="flex items-center gap-1 ml-auto">
                                    <Clock className="w-2.5 h-2.5 text-muted-foreground" />
                                    <span className="font-mono text-[9px] text-muted-foreground">{match.kickoff}</span>
                                </div>
                            </div>

                            <div className="px-3 py-3">
                                <div className="flex items-center justify-between mb-3">
                                    <div className="text-center flex-1">
                                        <p className="text-sm font-semibold">{match.home}</p>
                                    </div>
                                    <span className="font-mono text-xs text-muted-foreground px-3">VS</span>
                                    <div className="text-center flex-1">
                                        <p className="text-sm font-semibold">{match.away}</p>
                                    </div>
                                </div>

                                <div className="grid grid-cols-3 gap-2">
                                    {[
                                        { label: "1 HOME", odds: match.homeOdds, outcome: "home" },
                                        { label: "X DRAW", odds: match.drawOdds, outcome: "draw" },
                                        { label: "2 AWAY", odds: match.awayOdds, outcome: "away" },
                                    ].map(opt => {
                                        const isSelected = selectedBets[String(match.id)] === opt.outcome
                                        return (
                                            <button
                                                key={opt.outcome}
                                                onClick={() => selectBet(match.id, opt.outcome)}
                                                className={`flex flex-col items-center py-2 px-1 border rounded-sm transition-all cursor-pointer ${isSelected
                                                    ? "border-orange-500 bg-orange-500/10 shadow-[0_0_8px_rgba(249,115,22,0.2)]"
                                                    : "border-border bg-muted/20 hover:border-orange-500/40 hover:bg-orange-500/5"
                                                    }`}
                                            >
                                                <span className={`font-mono text-[9px] mb-0.5 ${isSelected ? "text-orange-400" : "text-muted-foreground"}`}>
                                                    {opt.label}
                                                </span>
                                                <span className={`font-mono text-sm font-bold ${isSelected ? "text-orange-500" : "text-foreground"}`}>
                                                    {opt.odds}
                                                </span>
                                            </button>
                                        )
                                    })}
                                </div>
                            </div>
                        </div>
                    ))}
                </div>

                {/* Right panel */}
                <div className="space-y-4">
                    {/* Bet slip */}
                    <div className="terminal-panel overflow-hidden">
                        <div className="terminal-panel-header">
                            <TrendingUp className="w-3 h-3 text-muted-foreground" />
                            <span className="font-mono text-[10px] text-muted-foreground">BET_SLIP</span>
                            <span className="font-mono text-[10px] text-orange-400 ml-auto">
                                {Object.keys(selectedBets).length} SEL
                            </span>
                        </div>

                        {Object.keys(selectedBets).length === 0 ? (
                            <div className="px-3 py-8 text-center">
                                <p className="font-mono text-[10px] text-muted-foreground/50">
                                    SELECT OUTCOMES TO BUILD YOUR BET SLIP
                                </p>
                            </div>
                        ) : (
                            <div className="p-3 space-y-2">
                                {Object.entries(selectedBets).map(([id, outcome]) => {
                                    const m = MOCK_MARKETS.find(x => String(x.id) === id)
                                    if (!m) return null
                                    const oddsMap = { home: m.homeOdds, draw: m.drawOdds, away: m.awayOdds }
                                    return (
                                        <div key={id} className="text-xs border border-border/60 rounded-sm p-2 bg-muted/10">
                                            <p className="font-mono text-[10px] text-muted-foreground">{m.home} vs {m.away}</p>
                                            <div className="flex justify-between items-center mt-1">
                                                <span className="text-orange-400 font-mono text-[10px] uppercase">{outcome}</span>
                                                <span className="font-mono font-bold">{oddsMap[outcome as keyof typeof oddsMap]}</span>
                                            </div>
                                        </div>
                                    )
                                })}
                                <Button className="w-full h-8 bg-orange-500 hover:bg-orange-600 text-white border-0 rounded-sm font-mono text-xs cursor-pointer mt-2">
                                    PLACE_BET
                                </Button>
                            </div>
                        )}
                    </div>

                    {/* Active bets */}
                    <div className="terminal-panel overflow-hidden">
                        <div className="terminal-panel-header">
                            <span className="dot-live" />
                            <span className="font-mono text-[10px] text-muted-foreground">ACTIVE_BETS</span>
                        </div>

                        {ACTIVE_BETS.map((bet, i) => (
                            <div key={i} className="terminal-data-row">
                                <div>
                                    <p className="font-mono text-xs font-semibold text-foreground">{bet.market}</p>
                                    <p className="font-mono text-[10px] text-muted-foreground">Stake: ${bet.stake}</p>
                                </div>
                                <div className="text-right">
                                    <p className="font-mono text-xs font-bold text-orange-400">@{bet.odds}</p>
                                    <p className="font-mono text-[10px] text-green-500">${bet.potential}</p>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    )
}
