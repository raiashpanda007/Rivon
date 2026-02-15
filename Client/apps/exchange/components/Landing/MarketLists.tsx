"use client"

import { Button } from "@workspace/ui/components/button"
import Loading from "../Loading"
import ApiCaller, { RequestType } from "@workspace/api-caller"
import { useEffect, useState } from "react"
import { ArrowUpRight, ArrowDownRight, MoreHorizontal } from "lucide-react"

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

function MarketTableRow({ market }: { market: MarketWithTeamDetails }) {
    const priceChange = market.lastPrice - market.openPrice;
    const priceChangePercent = market.openPrice == 0 ? 0 : (priceChange / market.openPrice) * 100;
    const isPositive = priceChange >= 0;

    return (
        <tr className="border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted group">
            <td className="p-4 align-middle">
                <div className="flex items-center gap-3">
                    <img
                        src={market.teamDetails.emblem}
                        alt={market.teamDetails.name}
                        className="w-8 h-8 rounded-full object-contain bg-background p-1 border"
                    />
                    <div className="flex flex-col">
                        <span className="font-medium text-sm text-foreground">{market.teamDetails.name}</span>
                        <span className="text-xs text-muted-foreground visible md:hidden">{market.marketCode}</span>
                    </div>
                </div>
            </td>
            <td className="p-4 align-middle hidden md:table-cell">
                <span className="inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-semibold text-foreground bg-muted">
                    {market.marketCode}
                </span>
            </td>
            <td className="p-4 align-middle hidden sm:table-cell">
                <div className="flex items-center gap-2">
                    <span className={`flex h-2 w-2 rounded-full ${market.status === 'open' ? 'bg-green-500' : market.status == 'closed' ? `bg-gray-500` : 'bg-red-500'}`} />
                    <span className="text-sm capitalize text-muted-foreground">{market.status.toLowerCase()}</span>
                </div>
            </td>
            <td className="p-4 align-middle text-right">
                <div className="flex flex-col items-end gap-1">
                    <span className="font-medium text-sm">
                        ${market.volume24h.toLocaleString(undefined, { maximumFractionDigits: 0 })}
                    </span>
                    <span className="text-[10px] text-muted-foreground">24h Vol</span>
                </div>
            </td>
            <td className="p-4 align-middle text-right hidden lg:table-cell">
                <span className="text-sm text-muted-foreground">
                    ${market.openPrice.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                </span>
            </td>
            <td className="p-4 align-middle text-right">
                <div className="flex flex-col items-end gap-1">
                    <span className="font-mono font-bold text-sm">
                        ${market.lastPrice.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                    </span>
                    <div className={`flex items-center text-xs font-medium ${isPositive ? 'text-green-500' : 'text-red-500'}`}>
                        {isPositive ? <ArrowUpRight className="mr-1 h-3 w-3" /> : <ArrowDownRight className="mr-1 h-3 w-3" />}
                        {Math.abs(priceChangePercent).toFixed(2)}%
                    </div>
                </div>
            </td>

        </tr>
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
                    queryParams: {
                        teamDetails: true,
                    }
                })
                if (response.ok) {
                    setMarkets(response.response.data.sort((a, b) => a.teamDetails.code.localeCompare(b.teamDetails.code)))
                }
            } catch (error) {
                console.error("Failed to fetch markets", error)
            } finally {
                setIsLoading(false)
            }
        }
        fetchMarkets()
    }, [])

    if (isLoading) {
        return <Loading />
    }

    return (
        <div className="w-full rounded-md border bg-card">
            <table className="w-full caption-bottom text-sm">
                <thead className="[&_tr]:border-b bg-muted/30 sticky top-0 backdrop-blur-sm z-10">
                    <tr className="border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted">
                        <th className="h-10 px-4 text-left align-middle font-medium text-muted-foreground [&:has([role=checkbox])]:pr-0 w-[200px]">Team</th>
                        <th className="h-10 px-4 text-left align-middle font-medium text-muted-foreground [&:has([role=checkbox])]:pr-0 hidden md:table-cell">Code</th>
                        <th className="h-10 px-4 text-left align-middle font-medium text-muted-foreground [&:has([role=checkbox])]:pr-0 hidden sm:table-cell">Status</th>
                        <th className="h-10 px-4 text-right align-middle font-medium text-muted-foreground [&:has([role=checkbox])]:pr-0">Volume</th>
                        <th className="h-10 px-4 text-right align-middle font-medium text-muted-foreground [&:has([role=checkbox])]:pr-0 hidden lg:table-cell">Open</th>
                        <th className="h-10 px-4 text-right align-middle font-medium text-muted-foreground [&:has([role=checkbox])]:pr-0">Last Price</th>
                    </tr>
                </thead>
                <tbody className="[&_tr:last-child]:border-0 text-card-foreground">
                    {markets.map((market) => (
                        <MarketTableRow key={market.id} market={market} />
                    ))}
                </tbody>
            </table>
        </div>
    )
}

export default MarketLists