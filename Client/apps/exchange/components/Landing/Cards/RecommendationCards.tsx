interface RecommendationCardsProps {
    by: "FAV" | "TOP" | "POPULAR";
}

function HeadingLabel(by: string): string {
    switch (by) {
        case "FAV": return "CREATOR_FAV"
        case "TOP": return "TOP_MOVERS"
        case "POPULAR": return "POPULAR"
        default: return ""
    }
}

function AccentColor(by: string): string {
    switch (by) {
        case "FAV": return "bg-blue-400"
        case "TOP": return "bg-green-400"
        case "POPULAR": return "bg-orange-400"
        default: return "bg-muted-foreground"
    }
}

const MOCK_DATA = {
    FAV: [
        { name: "RMFC", price: "$4.23", change: "+2.14%", isPositive: true },
        { name: "BMFC", price: "$1.87", change: "-1.30%", isPositive: false },
        { name: "MCFC", price: "$3.52", change: "+0.88%", isPositive: true },
        { name: "ARFC", price: "$3.18", change: "+4.10%", isPositive: true },
        { name: "LCFC", price: "$0.94", change: "-3.21%", isPositive: false },
    ],
    TOP: [
        { name: "ARFC", price: "$3.18", change: "+4.10%", isPositive: true },
        { name: "RMFC", price: "$4.23", change: "+2.14%", isPositive: true },
        { name: "MUFC", price: "$2.71", change: "+1.52%", isPositive: true },
        { name: "ATMC", price: "$2.06", change: "+2.33%", isPositive: true },
        { name: "JVFC", price: "$1.63", change: "+1.77%", isPositive: true },
    ],
    POPULAR: [
        { name: "RMFC", price: "$4.23", change: "+2.14%", isPositive: true },
        { name: "MCFC", price: "$3.52", change: "+0.88%", isPositive: true },
        { name: "BMFC", price: "$1.87", change: "-1.30%", isPositive: false },
        { name: "MUFC", price: "$2.71", change: "+1.52%", isPositive: true },
        { name: "PSGFC", price: "$2.95", change: "-0.44%", isPositive: false },
    ]
}

function RecommendationCards({ by }: RecommendationCardsProps) {
    const data = MOCK_DATA[by as keyof typeof MOCK_DATA] || []

    return (
        <div className="terminal-panel overflow-hidden w-full">
            <div className="terminal-panel-header">
                <span className={`w-1.5 h-1.5 rounded-full ${AccentColor(by)}`} />
                <span className="font-mono text-[10px] text-muted-foreground">{HeadingLabel(by)}</span>
                <span className="font-mono text-[10px] text-muted-foreground/40 ml-auto">24H CHG</span>
            </div>

            <div className="divide-y divide-border/50">
                {data.map((item, i) => (
                    <div
                        key={i}
                        className="flex items-center justify-between px-3 py-2 hover:bg-muted/20 transition-colors cursor-pointer group"
                    >
                        <div className="flex items-center gap-2">
                            <div className="w-5 h-5 rounded-sm bg-muted flex items-center justify-center shrink-0">
                                <span className="font-mono text-[9px] text-muted-foreground font-bold">
                                    {item.name.substring(0, 1)}
                                </span>
                            </div>
                            <span className="font-mono text-xs font-semibold text-foreground group-hover:text-orange-400 transition-colors">
                                {item.name}
                            </span>
                        </div>

                        <div className="flex items-center gap-4">
                            <span className="font-mono text-xs text-muted-foreground tabular-nums">
                                {item.price}
                            </span>
                            <span className={`font-mono text-[10px] font-bold w-14 text-right tabular-nums ${item.isPositive ? 'text-green-500' : 'text-red-500'}`}>
                                {item.change}
                            </span>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    )
}

export default RecommendationCards
