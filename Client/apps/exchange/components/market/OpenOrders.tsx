"use client"
import { useState, useRef, useEffect } from "react"
import type { OpenOrder, MarketData } from "@/app/markets/[market]/types"

interface Props {
  orders: OpenOrder[]
  marketId: string
  market: MarketData | null
  cancelOrder: (orderId: string, cancelQty?: number) => void
}

interface DialogState {
  order: OpenOrder
  cancelQty: number
}

export function OpenOrders({ orders, cancelOrder }: Props) {
  const [dialog, setDialog] = useState<DialogState | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (dialog) inputRef.current?.select()
  }, [dialog])

  function openDialog(order: OpenOrder) {
    const remaining = order.quantity - order.filled
    setDialog({ order, cancelQty: remaining })
  }

  function confirmCancel() {
    if (!dialog) return
    const remaining = dialog.order.quantity - dialog.order.filled
    const qty = Math.min(Math.max(1, dialog.cancelQty), remaining)
    cancelOrder(dialog.order.orderId, qty < remaining ? qty : undefined)
    setDialog(null)
  }

  return (
    <>
      <div className="flex flex-col border-b border-border/40 shrink-0">
        <div className="flex items-center justify-between px-3 py-2 border-b border-border/40">
          <span className="font-mono text-[10px] text-muted-foreground uppercase tracking-wider">
            Open Orders
          </span>
          {orders.length > 0 && (
            <span className="font-mono text-[10px] text-muted-foreground tabular-nums">
              {orders.length}
            </span>
          )}
        </div>

        <div className="max-h-[160px] overflow-y-auto">
          {orders.length === 0 ? (
            <div className="px-3 py-4 text-center">
              <span className="font-mono text-[10px] text-muted-foreground/50">
                No open orders
              </span>
            </div>
          ) : (
            <table className="w-full">
              <thead>
                <tr className="border-b border-border/20">
                  <th className="px-3 py-1.5 font-mono text-[9px] text-muted-foreground uppercase tracking-wider text-left">Side</th>
                  <th className="px-2 py-1.5 font-mono text-[9px] text-muted-foreground uppercase tracking-wider text-right">Price</th>
                  <th className="px-2 py-1.5 font-mono text-[9px] text-muted-foreground uppercase tracking-wider text-right">Qty</th>
                  <th className="px-2 py-1.5 font-mono text-[9px] text-muted-foreground uppercase tracking-wider text-right">Filled</th>
                  <th className="px-2 py-1.5 w-10"></th>
                </tr>
              </thead>
              <tbody>
                {orders.map((order) => (
                  <tr key={order.orderId} className="border-b border-border/10 hover:bg-muted/5">
                    <td className="px-3 py-1.5">
                      <span className={`font-mono text-[10px] font-semibold ${
                        order.side === "BUY" ? "text-green-400" : "text-red-400"
                      }`}>
                        {order.side}
                      </span>
                    </td>
                    <td className="px-2 py-1.5 text-right font-mono text-[10px] tabular-nums text-foreground">
                      {order.price.toLocaleString()}
                    </td>
                    <td className="px-2 py-1.5 text-right font-mono text-[10px] tabular-nums text-foreground">
                      {order.quantity.toLocaleString()}
                    </td>
                    <td className="px-2 py-1.5 text-right font-mono text-[10px] tabular-nums text-muted-foreground">
                      {order.filled.toLocaleString()}
                    </td>
                    <td className="px-2 py-1.5 text-right">
                      <button
                        onClick={() => openDialog(order)}
                        disabled={order.status === "cancelling"}
                        className="font-mono text-[9px] uppercase tracking-wider text-orange-400 hover:text-orange-300 disabled:text-muted-foreground/40 disabled:cursor-not-allowed transition-colors"
                      >
                        {order.status === "cancelling" ? "…" : "✕"}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {/* Cancel dialog */}
      {dialog && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/60"
          onClick={() => setDialog(null)}
        >
          <div
            className="w-72 bg-[#0f0f0f] border border-border rounded-md p-4 flex flex-col gap-4"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between">
              <span className="font-mono text-xs font-bold text-foreground uppercase tracking-wider">
                Cancel Order
              </span>
              <button
                onClick={() => setDialog(null)}
                className="font-mono text-[10px] text-muted-foreground hover:text-foreground"
              >
                ✕
              </button>
            </div>

            <div className="grid grid-cols-2 gap-x-3 gap-y-1.5 text-[10px] font-mono">
              <span className="text-muted-foreground">Side</span>
              <span className={dialog.order.side === "BUY" ? "text-green-400 font-semibold" : "text-red-400 font-semibold"}>
                {dialog.order.side}
              </span>
              <span className="text-muted-foreground">Price</span>
              <span className="text-foreground tabular-nums">{dialog.order.price.toLocaleString()}</span>
              <span className="text-muted-foreground">Total qty</span>
              <span className="text-foreground tabular-nums">{dialog.order.quantity.toLocaleString()}</span>
              <span className="text-muted-foreground">Filled</span>
              <span className="text-foreground tabular-nums">{dialog.order.filled.toLocaleString()}</span>
              <span className="text-muted-foreground">Remaining</span>
              <span className="text-orange-400 tabular-nums font-semibold">
                {(dialog.order.quantity - dialog.order.filled).toLocaleString()}
              </span>
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="font-mono text-[9px] text-muted-foreground uppercase tracking-wider">
                Qty to cancel
              </label>
              <input
                ref={inputRef}
                type="number"
                min={1}
                max={dialog.order.quantity - dialog.order.filled}
                value={dialog.cancelQty}
                onChange={(e) =>
                  setDialog((d) => d ? { ...d, cancelQty: Math.max(1, parseInt(e.target.value) || 1) } : d)
                }
                className="w-full bg-background border border-border rounded-sm px-2 py-1.5 font-mono text-xs text-foreground focus:outline-none focus:border-orange-500/60"
              />
              <span className="font-mono text-[9px] text-muted-foreground">
                {dialog.cancelQty >= dialog.order.quantity - dialog.order.filled
                  ? "Full cancel"
                  : `Partial — ${dialog.order.quantity - dialog.order.filled - dialog.cancelQty} remaining`}
              </span>
            </div>

            <div className="flex gap-2">
              <button
                onClick={() => setDialog(null)}
                className="flex-1 font-mono text-[10px] uppercase tracking-wider border border-border rounded-sm py-1.5 text-muted-foreground hover:text-foreground hover:border-border/80 transition-colors"
              >
                Keep
              </button>
              <button
                onClick={confirmCancel}
                className="flex-1 font-mono text-[10px] uppercase tracking-wider bg-red-500/10 border border-red-500/40 rounded-sm py-1.5 text-red-400 hover:bg-red-500/20 transition-colors"
              >
                Cancel {dialog.cancelQty.toLocaleString()} qty
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
