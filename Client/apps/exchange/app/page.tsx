import RecommendationCards from "@/components/Landing/Cards/RecommendationCards"
import MarketLists from "@/components/Landing/MarketLists"

export default function Page() {
  return (
    <div className="relative flex min-h-screen w-full flex-col pt-10 bg-background px-4 md:px-8">
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-bold mb-1 py-2 text-orange-500">Exchange</h1>
        <h2 className="text-xl font-semibold mb-8 opacity-60 border-b-2 border-foreground/20">Exchange with your favourite football teams</h2>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <RecommendationCards by="FAV" />
        <RecommendationCards by="TOP" />
        <RecommendationCards by="POPULAR" />
      </div>

      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold mb-2 py-2">All Markets</h1>
        <MarketLists />
      </div>



    </div>
  )
}
