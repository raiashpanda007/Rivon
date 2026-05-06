"use client"
import { useState, useEffect } from "react"
import { usePathname } from "next/navigation"
import ApiCaller, { RequestType } from "@workspace/api-caller"
import Loading from "@/components/Loading"
import { MarketHeader } from "@/components/market/MarketHeader"
import { CandleChart } from "@/components/market/CandleChart"
import { OrderBook } from "@/components/market/OrderBook"
import { OrderForm } from "@/components/market/OrderForm"
import { OpenOrders } from "@/components/market/OpenOrders"
import { useMarketSocket } from "@/hooks/useMarketSocket"
import type { MarketData, WalletData } from "./types"

function reverseString(str: string): string {
  return str.split("").reverse().join("")
}

function getIdOfPathName(pathName: string): string {
  const rvId = reverseString(pathName).split("/")
  return reverseString(rvId[0] || "")
}

function Page() {
  const currentPath = usePathname()
  const marketId = getIdOfPathName(currentPath)

  const [market, setMarket] = useState<MarketData | null>(null)
  const [wallet, setWallet] = useState<WalletData | null>(null)
  const [userId, setUserId] = useState<string | undefined>(undefined)
  const [isLoading, setIsLoading] = useState(true)

  const { orderBook, livePrice, wsStatus, openOrders, addOpenOrder, cancelOrder } = useMarketSocket(marketId, userId)

  async function fetchWallet() {
    const walletRes = await ApiCaller<unknown, WalletData>({
      requestType: RequestType.GET,
      paths: ["api", "rivon", "wallet", "me"],
      body: {},
    })
    if (walletRes.ok) {
      const w = walletRes.response.data as WalletData
      setWallet(w)
      setUserId(w.userId)
    }
  }

  useEffect(() => {
    async function fetchData() {
      setIsLoading(true)
      const [marketRes, walletRes] = await Promise.all([
        ApiCaller<unknown, MarketData>({
          requestType: RequestType.GET,
          paths: ["api", "rivon", "markets"],
          body: {},
          queryParams: { marketId, teamDetails: "true" },
        }),
        ApiCaller<unknown, WalletData>({
          requestType: RequestType.GET,
          paths: ["api", "rivon", "wallet", "me"],
          body: {},
        }),
      ])
      if (marketRes.ok) {
        const raw = marketRes.response.data as MarketData
        setMarket({ ...raw, lastPrice: raw.lastPrice / 100, openPrice: raw.openPrice / 100 })
      }
      if (walletRes.ok) {
        const w = walletRes.response.data as WalletData
        setWallet(w)
        setUserId(w.userId)
      }
      setIsLoading(false)
    }
    fetchData()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  if (isLoading) {
    return (
      <div className="flex min-h-[calc(100vh-3.5rem)] items-center justify-center">
        <Loading />
      </div>
    )
  }

  return (
    <div className="flex flex-col h-[calc(100vh-3.5rem)] w-full bg-background overflow-hidden">
      <MarketHeader market={market} livePrice={livePrice} />
      <div className="flex flex-1 min-h-0">
        <div className="flex flex-1 min-h-0 flex-col bg-terminal-grid border-b border-border/40 p-3">
          <div className="flex-1 min-h-0">
            <CandleChart marketId={marketId} />
          </div>
        </div>
        <div className="flex flex-col w-1/4 min-w-[260px] shrink-0 overflow-y-auto">
          <OrderForm market={market} balance={wallet?.balance} onOrderPlaced={addOpenOrder} onOrderSettled={fetchWallet} />
          <OpenOrders orders={openOrders} marketId={marketId} market={market} cancelOrder={cancelOrder} />
          <OrderBook market={market} orderBook={orderBook} livePrice={livePrice} wsStatus={wsStatus} />
        </div>
      </div>
    </div>
  )
}

export default Page
