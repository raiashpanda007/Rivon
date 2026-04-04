"use client"

import Loading from "../Loading"
import ApiCaller, { RequestType } from "@workspace/api-caller"
import { useEffect, useState } from "react"
import { ArrowUpRight, ArrowDownRight } from "lucide-react"
import Link from "next/link"

interface TeamDetails {
  id: string,
  name: string,
  shortName: string,
  code: string,
  tla: string,
  emblem: string,
  footballOrgId: number
}

interface MarketWithTeamDetails {
  id: string,
  teamId: string,
  marketName: string,
  marketCode: string,
  lastPrice: number,
  status: string,
  volume24h: number,
  totalVolume: number,
  openPrice: number,
  teamDetails: TeamDetails,
  createdAt: string,
  updatedAt: string
}

function StatusDot({ status }: { status: string }) {
  const color =
    status === "open" ? "bg-green-500" :
      status === "closed" ? "bg-zinc-500" :
        "bg-red-500"
  return <span className={`w-1.5 h-1.5 rounded-full ${color} shrink-0`} />
}

function MarketTableRow({ market }: { market: MarketWithTeamDetails }) {
  const priceChange = market.lastPrice - market.openPrice
  const priceChangePercent = market.openPrice === 0 ? 0 : (priceChange / market.openPrice) * 100
  const isPositive = priceChange >= 0

  return (
    <Link href={`/markets/${market.id}`} className="contents">
      <tr className="border-b border-border/60 hover:bg-muted/25 transition-colors cursor-pointer group">
        <td className="px-3 py-2.5 align-middle">
          <div className="flex items-center gap-2.5">
            <img
              src={market.teamDetails.emblem}
              alt={market.teamDetails.name}
              className="w-6 h-6 object-contain rounded-sm bg-background p-0.5 border border-border"
            />
            <div className="flex flex-col">
              <span className="text-xs font-medium text-foreground group-hover:text-orange-400 transition-colors">
                {market.teamDetails.name}
              </span>
              <span className="text-[10px] text-muted-foreground font-mono md:hidden">
                {market.marketCode}
              </span>
            </div>
          </div>
        </td>
        <td className="px-3 py-2.5 align-middle hidden md:table-cell">
          <span className="font-mono text-[10px] px-1.5 py-0.5 border border-border bg-muted/30 text-muted-foreground rounded-sm">
            {market.marketCode}
          </span>
        </td>
        <td className="px-3 py-2.5 align-middle hidden sm:table-cell">
          <div className="flex items-center gap-1.5">
            <StatusDot status={market.status} />
            <span className="text-[10px] font-mono capitalize text-muted-foreground">
              {market.status}
            </span>
          </div>
        </td>
        <td className="px-3 py-2.5 align-middle text-right">
          <span className="font-mono text-xs text-muted-foreground">
            ${market.volume24h.toLocaleString(undefined, { maximumFractionDigits: 0 })}
          </span>
        </td>
        <td className="px-3 py-2.5 align-middle text-right hidden lg:table-cell">
          <span className="font-mono text-xs text-muted-foreground">
            ${market.openPrice.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
          </span>
        </td>
        <td className="px-3 py-2.5 align-middle text-right">
          <div className="flex flex-col items-end">
            <span className="font-mono text-xs font-bold">
              ${market.lastPrice.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
            </span>
            <div className={`flex items-center text-[10px] font-mono font-bold ${isPositive ? 'text-green-500' : 'text-red-500'}`}>
              {isPositive
                ? <ArrowUpRight className="w-2.5 h-2.5 mr-0.5" />
                : <ArrowDownRight className="w-2.5 h-2.5 mr-0.5" />
              }
              {Math.abs(priceChangePercent).toFixed(2)}%
            </div>
          </div>
        </td>
      </tr>
    </Link>
  )
}

function MarketLists() {
  const [markets, setMarkets] = useState<MarketWithTeamDetails[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const fetchMarkets = async () => {
      try {
        const response = await ApiCaller<undefined, MarketWithTeamDetails[]>({
          requestType: RequestType.GET,
          paths: ["api", "rivon", "markets"],
          queryParams: { teamDetails: true }
        })
        if (response.ok) {
          setMarkets(response.response.data.sort((a, b) =>
            a.teamDetails.code.localeCompare(b.teamDetails.code)
          ))
        }
      } catch {
        // silent
      } finally {
        setIsLoading(false)
      }
    }
    fetchMarkets()
  }, [])

  if (isLoading) return <Loading />

  return (
    <div className="terminal-panel overflow-hidden">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-border bg-muted/20">
            <th className="px-3 py-2 text-left font-mono text-[10px] text-muted-foreground font-normal w-[200px]">
              TEAM
            </th>
            <th className="px-3 py-2 text-left font-mono text-[10px] text-muted-foreground font-normal hidden md:table-cell">
              CODE
            </th>
            <th className="px-3 py-2 text-left font-mono text-[10px] text-muted-foreground font-normal hidden sm:table-cell">
              STATUS
            </th>
            <th className="px-3 py-2 text-right font-mono text-[10px] text-muted-foreground font-normal">
              24H VOL
            </th>
            <th className="px-3 py-2 text-right font-mono text-[10px] text-muted-foreground font-normal hidden lg:table-cell">
              OPEN
            </th>
            <th className="px-3 py-2 text-right font-mono text-[10px] text-muted-foreground font-normal">
              LAST
            </th>
          </tr>
        </thead>
        <tbody>
          {markets.map((market) => (
            <MarketTableRow key={market.id} market={market} />
          ))}
        </tbody>
      </table>
    </div>
  )
}

export default MarketLists
