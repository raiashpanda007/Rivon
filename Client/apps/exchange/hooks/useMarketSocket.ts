"use client"
import { useState, useEffect, useRef, useCallback } from "react"
import { toast } from "@workspace/ui/components/sonner"
import type {
  OrderBookData,
  OrderLevel,
  DepthMap,
  WsMessage,
  WsOrderbookPayload,
  WsOrderCancelledPayload,
  WsStatus,
  OpenOrder,
} from "@/app/markets/[market]/types"

const WS_URL = process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8003"
const RECONNECT_DELAY_MS = 3000
const MAX_LEVELS = 15

function depthToLevels(depth: DepthMap, side: "bid" | "ask"): OrderLevel[] {
  const entries = Object.entries(depth)
    .map(([priceStr, qty]) => ({ price: Number(priceStr) / 100, quantity: qty }))
    .filter((e) => e.quantity > 0)

  if (side === "ask") {
    entries.sort((a, b) => a.price - b.price)
  } else {
    entries.sort((a, b) => b.price - a.price)
  }

  const levels: OrderLevel[] = []
  let running = 0
  for (const e of entries.slice(0, MAX_LEVELS)) {
    running += e.quantity
    levels.push({ price: e.price, quantity: e.quantity, total: running })
  }
  return levels
}

function applyDepth(payload: WsOrderbookPayload): OrderBookData {
  return {
    asks: depthToLevels(payload.askDepth, "ask"),
    bids: depthToLevels(payload.bidDepth, "bid"),
  }
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8000"

export function useMarketSocket(
  marketId: string,
  userId?: string
): {
  orderBook: OrderBookData
  livePrice: number | null
  wsStatus: WsStatus
  openOrders: OpenOrder[]
  addOpenOrder: (order: OpenOrder) => void
  cancelOrder: (orderId: string, cancelQty?: number) => void
} {
  const [orderBook, setOrderBook] = useState<OrderBookData>({ asks: [], bids: [] })
  const [livePrice, setLivePrice] = useState<number | null>(null)
  const [wsStatus, setWsStatus] = useState<WsStatus>("connecting")
  const [openOrders, setOpenOrders] = useState<OpenOrder[]>([])

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const mountedRef = useRef(true)

  const addOpenOrder = useCallback((order: OpenOrder) => {
    setOpenOrders((prev) =>
      prev.some((o) => o.orderId === order.orderId) ? prev : [...prev, order]
    )
  }, [])

  const cancelOrder = useCallback(
    (orderId: string, cancelQty?: number) => {
      const ws = wsRef.current
      if (!ws || ws.readyState !== WebSocket.OPEN) {
        toast.error("WebSocket not connected")
        return
      }
      setOpenOrders((prev) =>
        prev.map((o) => (o.orderId === orderId ? { ...o, status: "cancelling" as const } : o))
      )
      ws.send(
        JSON.stringify({
          type: "CANCEL_ORDER",
          payload: { marketID: marketId, orderId, ...(cancelQty ? { cancelQty } : {}) },
        })
      )
    },
    [marketId]
  )

  const subscribe = useCallback(
    (ws: WebSocket) => {
      ws.send(
        JSON.stringify({
          type: "SUBSCRIBE_MARKET",
          payload: { marketID: marketId, ...(userId ? { userID: userId } : {}) },
        })
      )
    },
    [marketId, userId]
  )

  const connect = useCallback(() => {
    if (!mountedRef.current) return

    const ws = new WebSocket(WS_URL)
    wsRef.current = ws
    setWsStatus("connecting")

    ws.onopen = () => {
      if (!mountedRef.current) { ws.close(); return }
      subscribe(ws)
    }

    ws.onmessage = (event) => {
      if (!mountedRef.current) return
      try {
        const msg: WsMessage = JSON.parse(event.data as string)

        if (msg.type === "ORDERBOOK_DATA" || msg.type === "ORDERBOOK_UPDATE") {
          const payload = msg.payload as WsOrderbookPayload
          setOrderBook(applyDepth(payload))
          if (payload.currentPrice > 0) setLivePrice(payload.currentPrice / 100)
          setWsStatus("live")
          return
        }

        if (msg.type === "ORDER_CANCELLED") {
          const payload = msg.payload as WsOrderCancelledPayload
          if (payload.success) {
            setOpenOrders((prev) => {
              const order = prev.find((o) => o.orderId === payload.orderId)
              if (!order) return prev.filter((o) => o.orderId !== payload.orderId)
              const remaining = order.quantity - order.filled
              if (!payload.cancelledQty || payload.cancelledQty >= remaining) {
                return prev.filter((o) => o.orderId !== payload.orderId)
              }
              // Partial cancel — reduce quantity, keep order open
              return prev.map((o) =>
                o.orderId === payload.orderId
                  ? { ...o, quantity: o.quantity - payload.cancelledQty!, status: "open" as const }
                  : o
              )
            })
          } else {
            setOpenOrders((prev) =>
              prev.map((o) =>
                o.orderId === payload.orderId ? { ...o, status: "open" as const } : o
              )
            )
            toast.error("Order not found — it may have already been filled")
          }
        }
      } catch {
        // malformed message — ignore
      }
    }

    ws.onclose = () => {
      if (!mountedRef.current) return
      setWsStatus("disconnected")
      reconnectTimer.current = setTimeout(connect, RECONNECT_DELAY_MS)
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [subscribe])

  // Fetch persisted open orders from DB when userId is known
  useEffect(() => {
    if (!userId) return
    const controller = new AbortController()
    fetch(
      `${API_BASE}/api/rivon/markets/open-orders?marketId=${marketId}`,
      { credentials: "include", signal: controller.signal }
    )
      .then((r) => r.ok ? r.json() : null)
      .then((body) => {
        if (!body?.data) return
        const dbOrders: OpenOrder[] = (body.data as Array<{
          orderId: string; side: string; price: number;
          quantity: number; filled: number; status: string
        }>).map((o) => ({
          orderId: o.orderId,
          side: o.side as "BUY" | "SELL",
          price: o.price / 100,
          quantity: o.quantity,
          filled: o.filled,
          status: "open" as const,
        }))
        setOpenOrders(dbOrders)
      })
      .catch(() => {/* ignore */})
    return () => controller.abort()
  }, [userId, marketId])

  useEffect(() => {
    mountedRef.current = true
    connect()

    return () => {
      mountedRef.current = false
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [connect])

  return { orderBook, livePrice, wsStatus, openOrders, addOpenOrder, cancelOrder }
}
