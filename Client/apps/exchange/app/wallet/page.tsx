"use client"

import { useState, useEffect } from "react"
import ApiCaller, { RequestType } from "@workspace/api-caller"
import Loading from "@/components/Loading"
import { ArrowUpRight, ArrowDownRight, TrendingUp, TrendingDown } from "lucide-react"
import Link from "next/link"

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8000"
const PAGE_SIZE = 50

interface WalletInfo {
  id: string
  userId: string
  balance: number
}

interface AssetWithMarket {
  marketId: string
  marketName: string
  marketCode: string
  emblem: string
  quantity: number
  lockedQty: number
  avgCost: number
  currentPrice: number
}

interface Transaction {
  id: string
  type: "credit" | "debit"
  amount: number
  balanceBefore: number
  balanceAfter: number
  orderId: string | null
  tradeId: string | null
  createdAt: string
  marketName: string | null
  marketCode: string | null
}

// --- helpers ---------------------------------------------------------------

function fmt(n: number) {
  return `$${(n / 100).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function fmtCompact(n: number) {
  const v = n / 100
  if (Math.abs(v) >= 1_000_000) return `$${(v / 1_000_000).toFixed(2)}M`
  if (Math.abs(v) >= 1_000) return `$${(v / 1_000).toFixed(1)}K`
  return fmt(n)
}

function timeAgo(iso: string) {
  const diff = Date.now() - new Date(iso).getTime()
  const m = Math.floor(diff / 60000)
  if (m < 1) return "just now"
  if (m < 60) return `${m}m ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  return new Date(iso).toLocaleDateString()
}

// ---------------------------------------------------------------------------

function PortfolioSummary({ assets }: { assets: AssetWithMarket[] }) {
  const totalValue = assets.reduce((s, a) => s + a.quantity * a.currentPrice, 0)
  const totalCost  = assets.reduce((s, a) => s + a.quantity * a.avgCost, 0)
  const pnl        = totalValue - totalCost
  const pnlPct     = totalCost > 0 ? (pnl / totalCost) * 100 : 0
  const isUp       = pnl >= 0

  return (
    <div className="grid grid-cols-3 divide-x divide-border border border-border rounded-sm bg-card">
      <div className="px-4 py-3 flex flex-col gap-0.5">
        <span className="font-mono text-[9px] text-muted-foreground uppercase tracking-widest">Portfolio Value</span>
        <span className="font-mono text-lg font-bold text-foreground tabular-nums">{fmtCompact(totalValue)}</span>
      </div>
      <div className="px-4 py-3 flex flex-col gap-0.5">
        <span className="font-mono text-[9px] text-muted-foreground uppercase tracking-widest">Cost Basis</span>
        <span className="font-mono text-lg font-bold text-muted-foreground tabular-nums">{fmtCompact(totalCost)}</span>
      </div>
      <div className="px-4 py-3 flex flex-col gap-0.5">
        <span className="font-mono text-[9px] text-muted-foreground uppercase tracking-widest">Unrealised P&amp;L</span>
        <div className="flex items-center gap-1">
          {isUp
            ? <TrendingUp className="w-3.5 h-3.5 text-green-400 shrink-0" />
            : <TrendingDown className="w-3.5 h-3.5 text-red-400 shrink-0" />}
          <span className={`font-mono text-lg font-bold tabular-nums ${isUp ? "text-green-400" : "text-red-400"}`}>
            {isUp ? "+" : ""}{fmtCompact(pnl)}
          </span>
          <span className={`font-mono text-[10px] tabular-nums ${isUp ? "text-green-500/70" : "text-red-500/70"}`}>
            ({isUp ? "+" : ""}{pnlPct.toFixed(2)}%)
          </span>
        </div>
      </div>
    </div>
  )
}

function AssetCard({ asset }: { asset: AssetWithMarket }) {
  const pnlPerUnit = asset.currentPrice - asset.avgCost
  const totalPnl   = pnlPerUnit * asset.quantity
  const pnlPct     = asset.avgCost > 0 ? (pnlPerUnit / asset.avgCost) * 100 : 0
  const isUp       = totalPnl >= 0
  const available  = asset.quantity - asset.lockedQty

  return (
    <Link href={`/markets/${asset.marketId}`} className="block">
      <div className="terminal-panel p-4 flex flex-col gap-3 hover:border-border/60 transition-colors cursor-pointer">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            {asset.emblem && (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={asset.emblem} alt={asset.marketName} width={28} height={28}
                className="rounded-sm object-contain bg-background border border-border p-0.5 shrink-0" />
            )}
            <div>
              <span className="font-mono text-xs font-semibold text-foreground">{asset.marketName}</span>
              <div className="font-mono text-[9px] px-1.5 py-0.5 bg-orange-500/10 text-orange-400 border border-orange-500/30 rounded-sm inline-block ml-1.5">
                {asset.marketCode}
              </div>
            </div>
          </div>
          <div className={`flex items-center gap-0.5 font-mono text-xs font-bold tabular-nums ${isUp ? "text-green-400" : "text-red-400"}`}>
            {isUp ? <TrendingUp className="w-3 h-3" /> : <TrendingDown className="w-3 h-3" />}
            {isUp ? "+" : ""}{pnlPct.toFixed(2)}%
          </div>
        </div>

        {/* Stats grid */}
        <div className="grid grid-cols-2 gap-x-4 gap-y-2">
          <div>
            <p className="font-mono text-[9px] text-muted-foreground uppercase tracking-wider mb-0.5">Qty held</p>
            <p className="font-mono text-xs text-foreground tabular-nums font-semibold">
              {available.toLocaleString()}
              {asset.lockedQty > 0 && (
                <span className="text-orange-400/70 text-[9px] ml-1">+{asset.lockedQty.toLocaleString()} locked</span>
              )}
            </p>
          </div>
          <div>
            <p className="font-mono text-[9px] text-muted-foreground uppercase tracking-wider mb-0.5">Current price</p>
            <p className="font-mono text-xs text-foreground tabular-nums font-semibold">{fmt(asset.currentPrice)}</p>
          </div>
          <div>
            <p className="font-mono text-[9px] text-muted-foreground uppercase tracking-wider mb-0.5">Avg cost</p>
            <p className="font-mono text-xs text-muted-foreground tabular-nums">{fmt(asset.avgCost)}</p>
          </div>
          <div>
            <p className="font-mono text-[9px] text-muted-foreground uppercase tracking-wider mb-0.5">Market value</p>
            <p className="font-mono text-xs text-foreground tabular-nums font-semibold">
              {fmtCompact(asset.quantity * asset.currentPrice)}
            </p>
          </div>
        </div>

        {/* P&L bar */}
        <div className={`flex items-center justify-between pt-2 border-t border-border/30`}>
          <span className="font-mono text-[9px] text-muted-foreground uppercase tracking-wider">Unrealised P&amp;L</span>
          <span className={`font-mono text-xs font-bold tabular-nums ${isUp ? "text-green-400" : "text-red-400"}`}>
            {isUp ? "+" : ""}{fmtCompact(totalPnl)}
          </span>
        </div>
      </div>
    </Link>
  )
}

// ---------------------------------------------------------------------------

export default function WalletPage() {
  const [wallet, setWallet]           = useState<WalletInfo | null>(null)
  const [assets, setAssets]           = useState<AssetWithMarket[]>([])
  const [transactions, setTxns]       = useState<Transaction[]>([])
  const [offset, setOffset]           = useState(0)
  const [hasMore, setHasMore]         = useState(true)
  const [loading, setLoading]         = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)

  useEffect(() => {
    async function init() {
      setLoading(true)
      const [walletRes, assetsRes, txnRes] = await Promise.all([
        ApiCaller<unknown, WalletInfo>({
          requestType: RequestType.GET,
          paths: ["api", "rivon", "wallet", "me"],
          body: {},
        }),
        fetch(`${API_BASE}/api/rivon/wallet/assets`, { credentials: "include" }),
        fetch(`${API_BASE}/api/rivon/wallet/transactions?limit=${PAGE_SIZE}&offset=0`, { credentials: "include" }),
      ])

      if (walletRes.ok) setWallet(walletRes.response.data as WalletInfo)

      if (assetsRes.ok) {
        const body = await assetsRes.json()
        setAssets(body?.data ?? [])
      }

      if (txnRes.ok) {
        const body = await txnRes.json()
        const data: Transaction[] = body?.data ?? []
        setTxns(data)
        setHasMore(data.length === PAGE_SIZE)
        setOffset(data.length)
      }

      setLoading(false)
    }
    init()
  }, [])

  async function loadMore() {
    setLoadingMore(true)
    const res = await fetch(
      `${API_BASE}/api/rivon/wallet/transactions?limit=${PAGE_SIZE}&offset=${offset}`,
      { credentials: "include" }
    )
    if (res.ok) {
      const body = await res.json()
      const data: Transaction[] = body?.data ?? []
      setTxns((prev) => [...prev, ...data])
      setHasMore(data.length === PAGE_SIZE)
      setOffset((o) => o + data.length)
    }
    setLoadingMore(false)
  }

  if (loading) {
    return (
      <div className="flex min-h-[calc(100vh-3.5rem)] items-center justify-center">
        <Loading />
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-8 space-y-8">

      {/* ── Balance ── */}
      <div className="terminal-panel p-6 flex flex-col gap-1">
        <span className="font-mono text-[10px] text-muted-foreground uppercase tracking-widest">Available Balance</span>
        <span className="font-mono text-4xl font-bold text-orange-400 tabular-nums">
          {wallet ? fmt(wallet.balance) : "—"}
        </span>
        <span className="font-mono text-[10px] text-muted-foreground">
          Wallet · {wallet?.id?.slice(0, 8)}…
        </span>
      </div>

      {/* ── Portfolio Holdings ── */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <span className="w-1.5 h-1.5 rounded-full bg-orange-500 shadow-[0_0_8px_rgba(249,115,22,0.5)]" />
          <span className="font-mono text-[10px] text-orange-400 tracking-widest font-bold uppercase">Portfolio</span>
        </div>

        {assets.length === 0 ? (
          <div className="terminal-panel px-4 py-10 text-center">
            <span className="font-mono text-xs text-muted-foreground/50">No holdings yet — place a buy order to get started</span>
          </div>
        ) : (
          <>
            <PortfolioSummary assets={assets} />
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {assets.map((a) => <AssetCard key={a.marketId} asset={a} />)}
            </div>
          </>
        )}
      </section>

      {/* ── Transaction History ── */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <span className="w-1.5 h-1.5 rounded-full bg-orange-500 shadow-[0_0_8px_rgba(249,115,22,0.5)]" />
          <span className="font-mono text-[10px] text-orange-400 tracking-widest font-bold uppercase">Transactions</span>
        </div>

        <div className="terminal-panel overflow-hidden">
          <div className="px-4 py-2.5 border-b border-border/40 flex items-center justify-between">
            <span className="font-mono text-[10px] text-muted-foreground uppercase tracking-wider">History</span>
            <span className="font-mono text-[10px] text-muted-foreground tabular-nums">{transactions.length} entries</span>
          </div>

          {transactions.length === 0 ? (
            <div className="px-4 py-10 text-center">
              <span className="font-mono text-xs text-muted-foreground/50">No transactions yet</span>
            </div>
          ) : (
            <div className="divide-y divide-border/30">
              {transactions.map((tx) => (
                <div key={tx.id} className="flex items-center gap-3 px-4 py-3 hover:bg-muted/10 transition-colors">
                  <div className={`w-7 h-7 rounded-sm flex items-center justify-center shrink-0 ${
                    tx.type === "credit" ? "bg-green-500/10 text-green-400" : "bg-red-500/10 text-red-400"
                  }`}>
                    {tx.type === "credit"
                      ? <ArrowDownRight className="w-3.5 h-3.5" />
                      : <ArrowUpRight className="w-3.5 h-3.5" />}
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className={`font-mono text-xs font-semibold capitalize ${
                        tx.type === "credit" ? "text-green-400" : "text-red-400"
                      }`}>{tx.type}</span>
                      {tx.marketCode && (
                        <span className="font-mono text-[9px] px-1 py-0.5 bg-muted/30 border border-border rounded text-muted-foreground">
                          {tx.marketCode}
                        </span>
                      )}
                    </div>
                    <div className="font-mono text-[10px] text-muted-foreground mt-0.5">
                      {tx.marketName ?? "Balance"} · {timeAgo(tx.createdAt)}
                    </div>
                  </div>

                  <div className="text-right shrink-0">
                    <div className={`font-mono text-xs font-bold tabular-nums ${
                      tx.type === "credit" ? "text-green-400" : "text-red-400"
                    }`}>
                      {tx.type === "credit" ? "+" : "-"}{fmt(tx.amount)}
                    </div>
                    <div className="font-mono text-[9px] text-muted-foreground tabular-nums mt-0.5">
                      → {fmt(tx.balanceAfter)}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

          {hasMore && (
            <div className="px-4 py-3 border-t border-border/30">
              <button
                onClick={loadMore}
                disabled={loadingMore}
                className="w-full font-mono text-[10px] uppercase tracking-wider text-muted-foreground hover:text-foreground border border-border/50 rounded-sm py-2 transition-colors disabled:opacity-40"
              >
                {loadingMore ? "Loading…" : "Load more"}
              </button>
            </div>
          )}
        </div>
      </section>

    </div>
  )
}
