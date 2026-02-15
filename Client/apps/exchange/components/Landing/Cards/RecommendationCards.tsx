
interface RecommendationCardsProps {
    by: "FAV" | "TOP" | "POPULAR";
}

function HeadingCards(by: string): string {
    switch (by) {
        case "FAV":
            return "Creator Fav"
        case "TOP":
            return "Top Movers"
        case "POPULAR":
            return "Popular"
        default:
            return ""
    }
}

const MOCK_DATA = {
    FAV: [
        { name: "STRK-PERP", price: "$0.05126", change: "+2.68%", isPositive: true },
        { name: "CC-PERP", price: "$0.16373", change: "-2.13%", isPositive: false },
        { name: "XMR-PERP", price: "$347.26", change: "-2.42%", isPositive: false },
        { name: "ZAMA-PERP", price: "$0.02117", change: "-0.47%", isPositive: false },
        { name: "SKR-PERP", price: "$0.02308", change: "-2.37%", isPositive: false },
    ],
    TOP: [
        { name: "kPEPE-PERP", price: "$0.004926", change: "+29.97%", isPositive: true },
        { name: "DOGE-PERP", price: "$0.11622", change: "+20.05%", isPositive: true },
        { name: "ZEC-PERP", price: "$321.40", change: "+14.02%", isPositive: true },
        { name: "PENGU-PERP", price: "$0.007748", change: "+13.24%", isPositive: true },
        { name: "FARTCOIN-PERP", price: "$0.2142", change: "+12.26%", isPositive: true },
    ],
    POPULAR: [
        { name: "BTC-PERP", price: "$70,292.10", change: "+2.14%", isPositive: true },
        { name: "ETH-PERP", price: "$2,088.34", change: "+1.73%", isPositive: true },
        { name: "SOL-PERP", price: "$89.32", change: "+5.24%", isPositive: true },
        { name: "BNB-PERP", price: "$636.86", change: "+3.10%", isPositive: true },
        { name: "HYPE-PERP", price: "$31.69", change: "+0.40%", isPositive: true },
    ]
}

function RecommendationCards({ by }: RecommendationCardsProps) {

    const data = MOCK_DATA[by as keyof typeof MOCK_DATA] || []

    return (
        <div className="flex flex-col gap-3 p-4 rounded-xl dark:bg-zinc-950/50 border dark:border-zinc-900 w-full min-w-[250px] ">
            <div className="flex justify-between items-center mb-1">
                <h2 className="text-base font-semibold">{HeadingCards(by)}</h2>
                <span className="text-xs">24h Change</span>
            </div>

            <div className="flex flex-col gap-1">
                {data.map((item, i) => (
                    <div key={i} className="flex justify-between items-center py-1.5 dark:hover:bg-zinc-900/50  hover:bg-zinc-100 rounded-md px-2 transition-colors cursor-pointer group">
                        <div className="flex items-center gap-2">
                            <div className="w-6 h-6 rounded-full dark:bg-zinc-800 flex items-center justify-center overflow-hidden shrink-0 border dark:border-zinc-700/50">
                                <span className="text-[9px] dark:text-zinc-500 font-bold">{item.name.substring(0, 1)}</span>
                            </div>
                            <span className="font-medium text-xs dark:text-zinc-200 dark:group-hover:text-white  transition-colors">
                                {item.name}
                            </span>
                        </div>

                        <div className="flex items-center gap-4 justify-end flex-1">
                            <span className="text-xs font-medium dark:text-zinc-200 tabular-nums">
                                {item.price}
                            </span>
                            <span className={`text-[10px] font-bold w-14 text-right tabular-nums ${item.isPositive ? 'text-green-500' : 'text-red-500'}`}>
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