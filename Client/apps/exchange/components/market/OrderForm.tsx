"use client"
import { useState, useMemo } from "react"
import { Button } from "@workspace/ui/components/button"
import { Input } from "@workspace/ui/components/input"
import { Label } from "@workspace/ui/components/label"
import { toast } from "@workspace/ui/components/sonner"
import ApiCaller, { RequestType } from "@workspace/api-caller"
import type { MarketData, OrderSide, OpenOrder } from "@/app/markets/[market]/types"

interface PlaceOrderResponseData {
  orderId: string
  executedQty: number
  status: string
  message: string
}

interface Props {
  market: MarketData | null
  balance?: number
  onOrderPlaced?: (order: OpenOrder) => void
  onOrderSettled?: () => void
}

export function OrderForm({ market, balance, onOrderPlaced, onOrderSettled }: Props) {
  const [side, setSide] = useState<OrderSide>("BUY")
  const [price, setPrice] = useState("")
  const [quantity, setQuantity] = useState("")
  const [isPlacing, setIsPlacing] = useState(false)

  const orderValue = useMemo(() => {
    const p = parseFloat(price)
    const q = parseFloat(quantity)
    if (isNaN(p) || isNaN(q)) return 0
    return p * q
  }, [price, quantity])

  async function placeOrder() {
    if (!market) return
    const qty = parseInt(quantity)
    const priceDollars = parseFloat(price)
    if (isNaN(qty) || qty <= 0) {
      toast.error("Invalid quantity")
      return
    }
    if (isNaN(priceDollars) || priceDollars <= 0) {
      toast.error("Invalid price")
      return
    }
    const priceCents = Math.round(priceDollars * 100)
    setIsPlacing(true)
    const res = await ApiCaller<unknown, PlaceOrderResponseData>({
      requestType: RequestType.POST,
      paths: ["api", "rivon", "markets", "create-order"],
      body: { marketId: market.id, price: priceCents, quantity: qty, orderType: side },
    })
    if (res.ok) {
      const data = res.response.data as PlaceOrderResponseData
      toast.success(data.message || data.status)
      if (data.status !== "filled" && data.orderId && onOrderPlaced) {
        onOrderPlaced({
          orderId: data.orderId,
          side,
          price: priceDollars,
          quantity: qty,
          filled: data.executedQty ?? 0,
          status: "open",
        })
      }
      setQuantity("")
      setPrice("")
      onOrderSettled?.()
    } else {
      const errData = res.response.data
      toast.error(typeof errData === "string" ? errData : "Order failed")
    }
    setIsPlacing(false)
  }

  return (
    <div className=" rounded-none border-x-0 border-t-0 w-full shrink-0 flex flex-col">

      {/* Buy / Sell toggle */}
      <div className="grid grid-cols-2 border-b border-border/60">
        <button
          onClick={() => setSide("BUY")}
          className={`py-2.5 font-mono text-xs font-semibold tracking-widest uppercase transition-colors ${side === "BUY"
            ? "bg-green-500/10 text-green-400 border-b-2 border-green-500"
            : "text-muted-foreground hover:text-foreground"
            }`}
        >
          Buy
        </button>
        <button
          onClick={() => setSide("SELL")}
          className={`py-2.5 font-mono text-xs font-semibold tracking-widest uppercase transition-colors ${side === "SELL"
            ? "bg-red-500/10 text-red-400 border-b-2 border-red-500"
            : "text-muted-foreground hover:text-foreground"
            }`}
        >
          Sell
        </button>
      </div>

      <div className="flex flex-col gap-4 p-4">

        {/* Price */}
        <div className="flex flex-col gap-1.5">
          <Label className="font-mono text-[10px] text-muted-foreground uppercase tracking-wider">
            Price
          </Label>
          <Input
            type="number"
            placeholder="0.00"
            min={0}
            step="0.01"
            value={price}
            onChange={(e) => setPrice(e.target.value)}
            className="font-mono text-sm h-9"
          />
        </div>

        {/* Quantity */}
        <div className="flex flex-col gap-1.5">
          <Label className="font-mono text-[10px] text-muted-foreground uppercase tracking-wider">
            Quantity
          </Label>
          <Input
            type="number"
            placeholder="0"
            min={1}
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
            className="font-mono text-sm h-9"
          />
        </div>

        {/* Order value */}
        <div className="flex justify-between items-center py-2 border-t border-border/40">
          <span className="font-mono text-[10px] text-muted-foreground uppercase tracking-wider">
            Order Value
          </span>
          <span className="font-mono text-xs text-foreground tabular-nums">
            $ {orderValue.toFixed(2)}
          </span>
        </div>

        {/* Place order */}
        <Button
          onClick={placeOrder}
          disabled={isPlacing}
          className={`w-full font-mono text-xs uppercase tracking-widest h-9 border-0 transition-colors ${side === "BUY"
            ? "bg-orange-600 hover:bg-orange-500 text-white"
            : "bg-orange-600 hover:bg-orange-500 text-white"
            }`}
        >
          {isPlacing
            ? "Placing…"
            : `${side === "BUY" ? "Buy" : "Sell"} ${market?.marketCode ?? ""}`}
        </Button>

        {/* Available balance */}
        {balance !== undefined && (
          <div className="flex justify-between items-center py-1 border-t border-border/40">
            <span className="font-mono text-[10px] text-muted-foreground uppercase tracking-wider">
              Available
            </span>
            <span className="font-mono text-xs tabular-nums">
              $ {(balance / 100).toFixed(2)}
            </span>
          </div>
        )}
      </div>
    </div>
  )
}
