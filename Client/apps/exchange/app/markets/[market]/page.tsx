"use client"

import { useEffect, useState } from "react"
import { motion } from "framer-motion"
import { ArrowUpRight, ArrowDownRight, Activity } from "lucide-react"
import { Button } from "@workspace/ui/components/button"
import ApiCaller, { RequestType } from "@workspace/api-caller"

interface MarketInfo {
    id: string
    marketName: string
    marketCode: string
    lastPrice: number
    openPrice: number
    status: string
    volume24h: number
    teamDetails?: {
        name: string
        emblem: string
    }
}

interface OrderEntry {
    price: number
    qty: number
}

interface TradeEntry {
    price: number
    qty: number
    side: "BUY" | "SELL"
    time: string
}

function generateOrderBook() {
    const mid = 4.23 + (Math.random() - 0.5) * 0.1
    const asks: OrderEntry[] = Array.from({ length: 8 }, (_, i) => ({
        price: parseFloat((mid + (i + 1) * 0.01).toFixed(2)),
        qty: Math.floor(Math.random() * 300) + 20
    })).reverse()
    const bids: OrderEntry[] = Array.from({ length: 8 }, (_, i) => ({
        price: parseFloat((mid - (i + 1) * 0.01).toFixed(2)),
        qty: Math.floor(Math.random() * 300) + 20
    }))
    return { asks, bids, mid: parseFloat(mid.toFixed(2)) }
}

function generateTrades(): TradeEntry[] {
    const now = new Date()
    return Array.from({ length: 12 }, (_, i) => {
        const d = new Date(now.getTime() - i * 3000)
        const side = Math.random() > 0.5 ? "BUY" : "SELL"
        return {
            price: parseFloat((4.23 + (Math.random() - 0.5) * 0.15).toFixed(2)),
            qty: Math.floor(Math.random() * 200) + 5,
            side,
            time: d.toTimeString().slice(0, 8)
        }
    })
}

function generateSparkline(base: number = 4.23, points: number = 80): number[] {
    const data: number[] = [base]
    for (let i = 1; i < points; i++) {
        data.push(parseFloat((data[i - 1] + (Math.random() - 0.5) * 0.04).toFixed(3)))
    }
    return data
}

function Sparkline({ data, isUp }: { data: number[]; isUp: boolean }) {
    if (data.length < 2) return null
    const W = 600
    const H = 160
    const padX = 8
    const padY = 16
    const min = Math.min(...data)
    const max = Math.max(...data)
    const range = max - min || 0.01
    const pts = data.map((v, i) => {
        const x = padX + (i / (data.length - 1)) * (W - padX * 2)
        const y = padY + (1 - (v - min) / range) * (H - padY * 2)
        return [x, y] as [number, number]
    })
    const linePath = "M " + pts.map(([x, y]) => `${x},${y}`).join(" L ")
    const fillPath = `${linePath} L ${pts[pts.length - 1][0]},${H} L ${pts[0][0]},${H} Z`
    const color = isUp ? "#22c55e" : "#ef4444"
    const gid = `sg-${isUp ? "up" : "dn"}`
    const lastPt = pts[pts.length - 1]
    const firstPrice = data[0]
    const lastPrice = data[data.length - 1]

    return (
        <svg viewBox={`0 0 ${W} ${H}`} className="w-full h-full" preserveAspectRatio="none">
            <defs>
                <linearGradient id={gid} x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor={color} stopOpacity="0.18" />
                    <stop offset="100%" stopColor={color} stopOpacity="0" />
                </linearGradient>
            </defs>
            {/* Horizontal reference lines */}
            {[0.25, 0.5, 0.75].map((t) => (
                <line
                    key={t}
                    x1={padX} y1={padY + t * (H - padY * 2)}
                    x2={W - padX} y2={padY + t * (H - padY * 2)}
                    stroke="currentColor"
                    strokeOpacity="0.06"
                    strokeWidth="1"
                    strokeDasharray="4 4"
                    className="text-foreground"
                />
            ))}
            {/* Price labels */}
            <text x={W - padX + 2} y={padY + 4} fontSize="9" fill={color} opacity="0.7" textAnchor="start">
                {max.toFixed(2)}
            </text>
            <text x={W - padX + 2} y={H - padY} fontSize="9" fill={color} opacity="0.7" textAnchor="start">
                {min.toFixed(2)}
            </text>
            {/* Fill */}
            <path d={fillPath} fill={`url(#${gid})`} />
            {/* Line */}
            <path d={linePath} stroke={color} strokeWidth="1.5" fill="none" strokeLinejoin="round" />
            {/* Current price dot */}
            <circle cx={lastPt[0]} cy={lastPt[1]} r="3" fill={color} />
            <circle cx={lastPt[0]} cy={lastPt[1]} r="6" fill={color} opacity="0.2" />
            {/* Open/close diff */}
            <text x={padX} y={padY - 4} fontSize="9" fill="currentColor" opacity="0.4" className="text-foreground">
                O: {firstPrice.toFixed(2)}
            </text>
            <text x={W / 2} y={padY - 4} fontSize="9" fill={color} textAnchor="middle">
                {lastPrice > firstPrice ? "+" : ""}{(lastPrice - firstPrice).toFixed(3)}
            </text>
        </svg>
    )
}

export default function Page({ params }: { params: Promise<{ market: string }> }) {
    const [marketId, setMarketId] = useState<string>("")
    const [market, setMarket] = useState<MarketInfo | null>(null)
    const [orderBook, setOrderBook] = useState(generateOrderBook())
    const [trades, setTrades] = useState<TradeEntry[]>(generateTrades())
    const [orderSide, setOrderSide] = useState<"BUY" | "SELL">("BUY")
    const [orderType, setOrderType] = useState<"LIMIT" | "MARKET">("LIMIT")
    const [orderPrice, setOrderPrice] = useState("")
    const [orderQty, setOrderQty] = useState("")
    const [priceUp, setPriceUp] = useState(true)
    const [sparkline, setSparkline] = useState<number[]>(() => generateSparkline())

    useEffect(() => {
        params.then(p => setMarketId(p.market))
    }, [params])

    useEffect(() => {
        if (!marketId) return
        const fetch = async () => {
            try {
                const res = await ApiCaller<undefined, MarketInfo>({
                    requestType: RequestType.GET,
                    paths: ["api", "rivon", "markets", marketId],
                    body: undefined
                })
                if (res.ok) setMarket(res.response.data)
            } catch {
                // Market info not loaded — show mock data
            }
        }
        fetch()
    }, [marketId])

    // Simulate live order book + sparkline updates
    useEffect(() => {
        const interval = setInterval(() => {
            const newBook = generateOrderBook()
            setPriceUp(newBook.mid >= orderBook.mid)
            setOrderBook(newBook)
            setSparkline(prev => {
                const next = [...prev.slice(1), parseFloat((prev[prev.length - 1] + (Math.random() - 0.5) * 0.04).toFixed(3))]
                return next
            })
            setTrades(prev => {
                const now = new Date()
                const entry: TradeEntry = {
                    price: newBook.mid,
                    qty: Math.floor(Math.random() * 150) + 5,
                    side: Math.random() > 0.5 ? "BUY" : "SELL",
                    time: now.toTimeString().slice(0, 8)
                }
                return [entry, ...prev.slice(0, 15)]
            })
        }, 2000)
        return () => clearInterval(interval)
    }, [orderBook.mid])

    const displayCode = market?.marketCode ?? marketId?.slice(0, 6).toUpperCase() ?? "—"
    const displayName = market?.marketName ?? "Loading…"
    const lastPrice = market?.lastPrice ?? orderBook.mid
    const openPrice = market?.openPrice ?? 4.10
    const priceChange = lastPrice - openPrice
    const priceChangePct = openPrice > 0 ? (priceChange / openPrice) * 100 : 0
    const vol24h = market?.volume24h ?? 0

    const maxAskQty = Math.max(...orderBook.asks.map(a => a.qty), 1)
    const maxBidQty = Math.max(...orderBook.bids.map(b => b.qty), 1)
    const bestAsk = orderBook.asks[orderBook.asks.length - 1]?.price ?? 0
    const bestBid = orderBook.bids[0]?.price ?? 0
    const spread = bestAsk > 0 && bestBid > 0 ? (bestAsk - bestBid).toFixed(2) : "—"

    return (
        <div className="flex flex-col h-[calc(100vh-3.5rem)] bg-background overflow-hidden">

            {/* ── Market header bar ─────────────────────────────────── */}
            <div className="flex items-center gap-5 px-4 py-2 border-b border-border bg-card/80 shrink-0 flex-wrap gap-y-1.5">
                <div className="flex items-center gap-2.5">
                    {market?.teamDetails?.emblem && (
                        <img src={market.teamDetails.emblem} alt="" className="w-5 h-5 object-contain" />
                    )}
                    <div>
                        <span className="text-[9px] text-muted-foreground/60 leading-none block">{displayCode}</span>
                        <p className="text-xs font-semibold text-foreground leading-tight">{displayName}</p>
                    </div>
                </div>

                <div className="h-4 w-px bg-border" />

                <motion.div
                    key={lastPrice}
                    initial={{ opacity: 0.5 }}
                    animate={{ opacity: 1 }}
                    className={`text-base font-bold tabular-nums ${priceUp ? "text-green-400" : "text-red-400"}`}
                >
                    ${lastPrice.toFixed(2)}
                </motion.div>

                <div className={`flex items-center gap-0.5 text-xs font-bold ${priceChangePct >= 0 ? "text-green-500" : "text-red-500"}`}>
                    {priceChangePct >= 0
                        ? <ArrowUpRight className="w-3.5 h-3.5" />
                        : <ArrowDownRight className="w-3.5 h-3.5" />
                    }
                    {Math.abs(priceChangePct).toFixed(2)}%
                </div>

                <div className="h-4 w-px bg-border hidden sm:block" />

                <div className="hidden sm:flex items-center gap-5 text-[10px] text-muted-foreground">
                    <span>OPEN <span className="text-foreground tabular-nums">${openPrice.toFixed(2)}</span></span>
                    <span>VOL <span className="text-foreground tabular-nums">{vol24h.toLocaleString()}</span></span>
                    <span>SPREAD <span className="text-orange-400 tabular-nums">{spread}</span></span>
                    <div className="flex items-center gap-1">
                        <span className={`w-1.5 h-1.5 rounded-full ${market?.status === "open" ? "bg-green-500 animate-pulse" : "bg-zinc-500"}`} />
                        <span className="capitalize">{market?.status ?? "open"}</span>
                    </div>
                </div>
            </div>

            {/* ── Main trading layout ───────────────────────────────── */}
            <div className="flex flex-1 overflow-hidden divide-x divide-border">

                {/* ── Left: Order Book ─────────────────────────────── */}
                <div className="w-52 flex flex-col shrink-0 overflow-hidden hidden md:flex">
                    <div className="terminal-panel-header shrink-0 justify-between">
                        <div className="flex items-center gap-2">
                            <span className="w-1.5 h-1.5 rounded-full bg-orange-500 animate-pulse" />
                            <span className="text-[10px] text-muted-foreground">ORDER_BOOK</span>
                        </div>
                        <span className="text-[9px] text-muted-foreground/40">SPD {spread}</span>
                    </div>

                    <div className="flex-1 overflow-y-auto text-xs">
                        {/* Asks (sell) */}
                        <div className="px-2 pt-1.5 pb-0.5 flex justify-between">
                            <span className="text-[9px] text-muted-foreground/50">ASKS</span>
                            <span className="text-[9px] text-muted-foreground/50">QTY</span>
                        </div>
                        {orderBook.asks.map((ask, i) => (
                            <div key={i} className="relative flex justify-between items-center px-2 py-[3px] hover:bg-red-500/5 group">
                                <div
                                    className="absolute right-0 top-0 bottom-0 bg-red-500/8"
                                    style={{ width: `${(ask.qty / maxAskQty) * 100}%` }}
                                />
                                <span className="relative text-red-400 tabular-nums">{ask.price.toFixed(2)}</span>
                                <span className="relative text-muted-foreground tabular-nums">{ask.qty}</span>
                            </div>
                        ))}

                        {/* Spread mid */}
                        <motion.div
                            key={orderBook.mid}
                            initial={{ opacity: 0.6 }}
                            animate={{ opacity: 1 }}
                            className={`flex items-center justify-between px-2 py-1.5 border-y border-border my-0.5 ${priceUp ? "bg-green-500/5" : "bg-red-500/5"}`}
                        >
                            <span className={`font-bold text-sm tabular-nums ${priceUp ? "text-green-400" : "text-red-400"}`}>
                                {priceUp ? "▲" : "▼"} {orderBook.mid.toFixed(2)}
                            </span>
                            <span className="text-[9px] text-muted-foreground/50">MID</span>
                        </motion.div>

                        {/* Bids (buy) */}
                        <div className="px-2 pt-0.5 pb-0.5 flex justify-between">
                            <span className="text-[9px] text-muted-foreground/50">BIDS</span>
                            <span className="text-[9px] text-muted-foreground/50">QTY</span>
                        </div>
                        {orderBook.bids.map((bid, i) => (
                            <div key={i} className="relative flex justify-between items-center px-2 py-[3px] hover:bg-green-500/5">
                                <div
                                    className="absolute right-0 top-0 bottom-0 bg-green-500/8"
                                    style={{ width: `${(bid.qty / maxBidQty) * 100}%` }}
                                />
                                <span className="relative text-green-400 tabular-nums">{bid.price.toFixed(2)}</span>
                                <span className="relative text-muted-foreground tabular-nums">{bid.qty}</span>
                            </div>
                        ))}
                    </div>
                </div>

                {/* ── Center: Sparkline chart + Trade history ────────── */}
                <div className="flex-1 flex flex-col overflow-hidden">
                    {/* Chart area */}
                    <div className="flex-1 flex flex-col border-b border-border bg-background relative overflow-hidden">
                        <div className="absolute inset-0 bg-terminal-grid opacity-40" />

                        {/* Chart header */}
                        <div className="relative shrink-0 flex items-center justify-between px-3 py-2 border-b border-border/50 bg-card/30">
                            <div className="flex items-center gap-3">
                                <span className="text-[10px] text-muted-foreground">PRICE_CHART</span>
                                <span className="text-[9px] px-1.5 py-0.5 border border-border text-muted-foreground/50">1M</span>
                                <span className="text-[9px] text-orange-400 px-1.5 py-0.5 border border-orange-500/30 bg-orange-500/5">LIVE</span>
                            </div>
                            <span className={`text-[10px] tabular-nums font-bold ${priceUp ? "text-green-400" : "text-red-400"}`}>
                                {priceUp ? "▲" : "▼"} ${orderBook.mid.toFixed(2)}
                            </span>
                        </div>

                        {/* Sparkline */}
                        <div className="relative flex-1 p-2 overflow-hidden">
                            <Sparkline data={sparkline} isUp={priceUp} />
                        </div>
                    </div>

                    {/* Recent trades */}
                    <div className="h-44 overflow-hidden flex flex-col shrink-0">
                        <div className="terminal-panel-header shrink-0">
                            <Activity className="w-3 h-3 text-muted-foreground" />
                            <span className="text-[10px] text-muted-foreground">RECENT_TRADES</span>
                        </div>
                        <div className="overflow-y-auto flex-1 text-xs">
                            <table className="w-full">
                                <thead>
                                    <tr className="border-b border-border/50">
                                        <th className="px-3 py-1 text-left text-[9px] text-muted-foreground/50 font-normal">TIME</th>
                                        <th className="px-3 py-1 text-left text-[9px] text-muted-foreground/50 font-normal">SIDE</th>
                                        <th className="px-3 py-1 text-right text-[9px] text-muted-foreground/50 font-normal">PRICE</th>
                                        <th className="px-3 py-1 text-right text-[9px] text-muted-foreground/50 font-normal">QTY</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {trades.map((trade, i) => (
                                        <tr key={i} className="border-b border-border/30 hover:bg-muted/10">
                                            <td className="px-3 py-[3px] text-muted-foreground/50 text-[10px]">{trade.time}</td>
                                            <td className={`px-3 py-[3px] font-bold text-[10px] ${trade.side === "BUY" ? "text-green-400" : "text-red-400"}`}>
                                                {trade.side}
                                            </td>
                                            <td className="px-3 py-[3px] text-right text-foreground tabular-nums">{trade.price.toFixed(2)}</td>
                                            <td className="px-3 py-[3px] text-right text-muted-foreground tabular-nums">{trade.qty}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>

                {/* ── Right: Trade Panel ───────────────────────────── */}
                <div className="w-64 flex flex-col shrink-0 overflow-y-auto hidden lg:flex">
                    <div className="terminal-panel-header shrink-0">
                        <span className="text-[10px] text-muted-foreground">PLACE_ORDER</span>
                    </div>

                    <div className="p-3 space-y-3 flex-1">
                        {/* Order type toggle */}
                        <div className="flex rounded-sm overflow-hidden border border-border">
                            <button
                                onClick={() => setOrderType("LIMIT")}
                                className={`flex-1 py-1 text-[10px] font-bold transition-colors ${orderType === "LIMIT"
                                    ? "bg-orange-500/15 text-orange-400 border-r border-orange-500/20"
                                    : "bg-transparent text-muted-foreground hover:text-foreground border-r border-border"
                                }`}
                            >
                                LIMIT
                            </button>
                            <button
                                onClick={() => setOrderType("MARKET")}
                                className={`flex-1 py-1 text-[10px] font-bold transition-colors ${orderType === "MARKET"
                                    ? "bg-orange-500/15 text-orange-400"
                                    : "bg-transparent text-muted-foreground hover:text-foreground"
                                }`}
                            >
                                MARKET
                            </button>
                        </div>

                        {/* Side toggle */}
                        <div className="flex rounded-sm overflow-hidden border border-border">
                            <button
                                onClick={() => setOrderSide("BUY")}
                                className={`flex-1 py-1.5 text-xs font-bold transition-colors ${orderSide === "BUY"
                                    ? "bg-green-500 text-white"
                                    : "bg-transparent text-muted-foreground hover:text-foreground"
                                }`}
                            >
                                BUY
                            </button>
                            <button
                                onClick={() => setOrderSide("SELL")}
                                className={`flex-1 py-1.5 text-xs font-bold transition-colors ${orderSide === "SELL"
                                    ? "bg-red-500 text-white"
                                    : "bg-transparent text-muted-foreground hover:text-foreground"
                                }`}
                            >
                                SELL
                            </button>
                        </div>

                        {/* Price input — hidden for MARKET orders */}
                        {orderType === "LIMIT" && (
                            <div className="space-y-1">
                                <label className="text-[10px] text-muted-foreground">PRICE (USD)</label>
                                <div className="relative">
                                    <span className="absolute left-2.5 top-1/2 -translate-y-1/2 text-xs text-muted-foreground">$</span>
                                    <input
                                        type="number"
                                        value={orderPrice}
                                        onChange={e => setOrderPrice(e.target.value)}
                                        placeholder={orderBook.mid.toFixed(2)}
                                        className="w-full pl-6 pr-3 py-1.5 bg-input border border-border rounded-sm text-xs focus:outline-none focus:border-orange-500/50 focus:ring-1 focus:ring-orange-500/20 transition-colors"
                                    />
                                </div>
                            </div>
                        )}

                        {orderType === "MARKET" && (
                            <div className="px-2.5 py-2 border border-border/50 bg-muted/20 rounded-sm">
                                <span className="text-[9px] text-muted-foreground/50 block mb-0.5">MARKET PRICE</span>
                                <span className={`text-sm font-bold tabular-nums ${priceUp ? "text-green-400" : "text-red-400"}`}>
                                    ${orderBook.mid.toFixed(2)}
                                </span>
                            </div>
                        )}

                        {/* Quantity input */}
                        <div className="space-y-1">
                            <label className="text-[10px] text-muted-foreground">QUANTITY</label>
                            <input
                                type="number"
                                value={orderQty}
                                onChange={e => setOrderQty(e.target.value)}
                                placeholder="0"
                                className="w-full px-3 py-1.5 bg-input border border-border rounded-sm text-xs focus:outline-none focus:border-orange-500/50 focus:ring-1 focus:ring-orange-500/20 transition-colors"
                            />
                        </div>

                        {/* Total estimate */}
                        <div className="flex items-center justify-between py-2 border-y border-border">
                            <span className="text-[10px] text-muted-foreground">EST. TOTAL</span>
                            <span className="text-xs font-bold tabular-nums">
                                ${((parseFloat(orderPrice) || orderBook.mid) * (parseFloat(orderQty) || 0)).toFixed(2)}
                            </span>
                        </div>

                        {/* Submit */}
                        <Button
                            className={`w-full h-9 rounded-sm text-xs tracking-wide cursor-pointer ${orderSide === "BUY"
                                ? "bg-green-500 hover:bg-green-600 text-white border-0"
                                : "bg-red-500 hover:bg-red-600 text-white border-0"
                                }`}
                        >
                            {orderSide}_{orderType}
                        </Button>

                        <p className="text-[9px] text-muted-foreground/40 text-center">
                            ORDER SUBMISSION · REQUIRES AUTH
                        </p>
                    </div>

                    {/* Market stats */}
                    <div className="border-t border-border">
                        <div className="terminal-panel-header">
                            <span className="text-[10px] text-muted-foreground">MARKET_STATS</span>
                        </div>
                        <div className="text-[10px] divide-y divide-border/50">
                            <div className="flex justify-between px-3 py-1.5">
                                <span className="text-muted-foreground">24H HIGH</span>
                                <span className="text-green-400 tabular-nums">${(lastPrice * 1.04).toFixed(2)}</span>
                            </div>
                            <div className="flex justify-between px-3 py-1.5">
                                <span className="text-muted-foreground">24H LOW</span>
                                <span className="text-red-400 tabular-nums">${(lastPrice * 0.96).toFixed(2)}</span>
                            </div>
                            <div className="flex justify-between px-3 py-1.5">
                                <span className="text-muted-foreground">OPEN</span>
                                <span className="text-foreground tabular-nums">${openPrice.toFixed(2)}</span>
                            </div>
                            <div className="flex justify-between px-3 py-1.5">
                                <span className="text-muted-foreground">BEST BID</span>
                                <span className="text-green-400 tabular-nums">${bestBid.toFixed(2)}</span>
                            </div>
                            <div className="flex justify-between px-3 py-1.5">
                                <span className="text-muted-foreground">BEST ASK</span>
                                <span className="text-red-400 tabular-nums">${bestAsk.toFixed(2)}</span>
                            </div>
                            <div className="flex justify-between px-3 py-1.5">
                                <span className="text-muted-foreground">24H VOL</span>
                                <span className="text-foreground tabular-nums">{vol24h.toLocaleString()}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}
