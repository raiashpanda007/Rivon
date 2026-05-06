export type OrderSide = "BUY" | "SELL"
export type WsStatus = "connecting" | "live" | "disconnected"
export type DepthMap = Record<string, number>

export interface WsOrderbookPayload {
  bidDepth: DepthMap
  askDepth: DepthMap
  currentPrice: number
  fills?: Array<{ price: number; quantity: number; tradeId: string }>
}

export interface WsOrderCancelledPayload {
  orderId: string
  success: boolean
  cancelledQty?: number
}

export interface WsMessage {
  type: "SUBSCRIBED" | "ORDERBOOK_DATA" | "ORDERBOOK_UPDATE" | "DEPTH_DATA" | "ERROR" | "ORDER_CANCELLED"
  payload: WsOrderbookPayload | WsOrderCancelledPayload | Record<string, unknown>
  connectionId?: string
  userId?: string
}

export interface OpenOrder {
  orderId: string
  side: OrderSide
  price: number
  quantity: number
  filled: number
  status: "open" | "cancelling"
}

export interface OrderLevel {
  price: number
  quantity: number
  total: number
}

export interface TeamDetails {
  name: string
  shortName: string
  tla: string
  emblem: string
}

export interface MarketData {
  id: string
  marketName: string
  marketCode: string
  lastPrice: number
  openPrice: number
  volume24h: number
  status: string
  teamDetails?: TeamDetails
}

export interface OrderBookData {
  asks: OrderLevel[]
  bids: OrderLevel[]
}

export interface WalletData {
  id: string
  userId: string
  balance: number
}
