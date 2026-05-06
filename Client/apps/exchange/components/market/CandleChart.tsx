"use client"

import { useEffect, useRef, useState, type MutableRefObject } from "react"
import {
  createChart,
  CandlestickSeries,
  type IChartApi,
  type ISeriesApi,
  type CandlestickData,
  type UTCTimestamp,
  type IRange,
  type Time,
} from "lightweight-charts"

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8000"
const HISTORY_WINDOW_SEC = 7 * 24 * 60 * 60

type Interval = "15m" | "1h" | "1d"

const INTERVAL_OPTIONS: { label: string; value: Interval }[] = [
  { label: "Day", value: "1d" },
  { label: "Hour", value: "1h" },
  { label: "15 Min", value: "15m" },
]

const INTERVAL_SECONDS: Record<Interval, number> = {
  "15m": 15 * 60,
  "1h": 60 * 60,
  "1d": 24 * 60 * 60,
}

interface CandlePayload {
  openTime: string
  open: number
  high: number
  low: number
  close: number
  volume: number
}

interface Props {
  marketId: string
}

export function CandleChart({ marketId }: Props) {
  const [interval, setInterval] = useState<Interval>("1h")
  const [chartReady, setChartReady] = useState(false)

  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const seriesRef = useRef<ISeriesApi<"Candlestick"> | null>(null)
  const candlesRef = useRef<CandlestickData[]>([])
  const earliestRef = useRef<number | null>(null)
  const latestRef = useRef<number | null>(null)
  const loadingHistoryRef = useRef(false)
  const historyEndRef = useRef(false)
  const autoFollowRef = useRef(true)
  const intervalRef = useRef<Interval>(interval)
  const historyAbortRef = useRef<AbortController | null>(null)
  const requestHistoryRef = useRef<(fromSec: number, toSec: number, mode: "replace" | "merge") => void>(
    () => undefined
  )

  useEffect(() => {
    intervalRef.current = interval
  }, [interval])

  useEffect(() => {
    if (!chartContainerRef.current) return

    const container = chartContainerRef.current
    const chart = createChart(container, {
      width: container.clientWidth,
      height: container.clientHeight || 300,
      layout: {
        background: { color: "#0a0a0a" },
        textColor: "#9ca3af",
      },
      grid: {
        vertLines: { color: "#1a1a1a" },
        horzLines: { color: "#1a1a1a" },
      },
      localization: {
        timeFormatter: (timestamp) => {
          const d = new Date((timestamp as number) * 1000)
          return d.toLocaleString(undefined, {
            month: "short",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit",
            hour12: false,
          })
        },
      },
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
      },
    })

    const series = chart.addSeries(CandlestickSeries, {
      upColor: "#22c55e",
      downColor: "#ef4444",
      borderUpColor: "#22c55e",
      borderDownColor: "#ef4444",
      wickUpColor: "#16a34a",
      wickDownColor: "#dc2626",
    })

    chartRef.current = chart
    seriesRef.current = series

    const timeScale = chart.timeScale()
    const handleRangeChange = (range: IRange<Time> | null) => {
      if (!range) return
      const from = Number(range.from)
      const to = Number(range.to)
      const earliest = earliestRef.current
      const latest = latestRef.current
      const intervalSec = INTERVAL_SECONDS[intervalRef.current]

      if (
        earliest != null &&
        !loadingHistoryRef.current &&
        !historyEndRef.current &&
        from <= earliest + intervalSec * 5
      ) {
        const fromSec = Math.max(earliest - HISTORY_WINDOW_SEC, 0)
        requestHistoryRef.current(fromSec, earliest, "merge")
      }

      if (latest != null) {
        if (to < latest - intervalSec * 2) {
          autoFollowRef.current = false
        } else if (to >= latest - intervalSec) {
          autoFollowRef.current = true
        }
      }
    }

    timeScale.subscribeVisibleTimeRangeChange(handleRangeChange)

    const ro = new ResizeObserver(() => {
      if (!chartRef.current || !chartContainerRef.current) return
      const width = chartContainerRef.current.clientWidth
      const height = chartContainerRef.current.clientHeight
      if (width > 0 && height > 0) {
        chartRef.current.resize(width, height)
      }
    })
    ro.observe(container)

    setChartReady(true)

    return () => {
      timeScale.unsubscribeVisibleTimeRangeChange(handleRangeChange)
      ro.disconnect()
      chart.remove()
      chartRef.current = null
      seriesRef.current = null
    }
  }, [])

  useEffect(() => {
    if (!chartReady || !seriesRef.current || !chartRef.current) return

    historyAbortRef.current?.abort()
    const controller = new AbortController()
    historyAbortRef.current = controller

    autoFollowRef.current = true
    historyEndRef.current = false
    loadingHistoryRef.current = false
    candlesRef.current = []
    earliestRef.current = null
    latestRef.current = null
    seriesRef.current.setData([])

    requestHistoryRef.current = (fromSec, toSec, mode) => {
      void loadHistoryWindow(fromSec, toSec, mode, controller.signal)
    }

    const nowSec = Math.floor(Date.now() / 1000)
    requestHistoryRef.current(nowSec - HISTORY_WINDOW_SEC, nowSec, "replace")

    const url = `${API_BASE}/api/rivon/candles/stream?market=${marketId}&interval=${interval}`
    const es = new EventSource(url, { withCredentials: true })

    es.addEventListener("candle", (e: MessageEvent) => {
      try {
        const payload: CandlePayload[] = JSON.parse(e.data)
        const incoming = mapPayloadToCandles(payload)
        if (incoming.length === 0) return

        const merged = mergeCandles(candlesRef.current, incoming)
        candlesRef.current = merged
        updateBounds(merged, earliestRef, latestRef)
        seriesRef.current?.setData(merged)

        if (autoFollowRef.current) {
          chartRef.current?.timeScale().scrollToRealTime()
        }
      } catch {
        // malformed message — ignore
      }
    })

    es.onerror = () => {
      // EventSource reconnects automatically — no action needed
    }

    return () => {
      controller.abort()
      es.close()
    }
  }, [chartReady, interval, marketId])

  async function loadHistoryWindow(
    fromSec: number,
    toSec: number,
    mode: "replace" | "merge",
    signal?: AbortSignal
  ) {
    if (loadingHistoryRef.current) return
    loadingHistoryRef.current = true

    try {
      const candles = await fetchHistory(fromSec, toSec, signal)
      if (candles.length === 0) {
        historyEndRef.current = true
        return
      }

      const merged = mode === "replace" ? sortCandles(candles) : mergeCandles(candlesRef.current, candles)
      candlesRef.current = merged
      updateBounds(merged, earliestRef, latestRef)
      seriesRef.current?.setData(merged)

      if (mode === "replace") {
        chartRef.current?.timeScale().fitContent()
      }
    } catch (err) {
      if (signal && signal.aborted) return
      console.warn("History fetch failed", err)
    } finally {
      loadingHistoryRef.current = false
    }
  }

  async function fetchHistory(fromSec: number, toSec: number, signal?: AbortSignal) {
    const url = new URL(`${API_BASE}/api/rivon/candles/history`)
    url.searchParams.set("market", marketId)
    url.searchParams.set("interval", interval)
    url.searchParams.set("from", String(fromSec))
    url.searchParams.set("to", String(toSec))

    const res = await fetch(url.toString(), { signal, credentials: "include" })
    if (!res.ok) {
      throw new Error(`History fetch failed: ${res.status}`)
    }
    const body = await res.json()
    const payload = Array.isArray(body?.data) ? (body.data as CandlePayload[]) : []
    return mapPayloadToCandles(payload)
  }

  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden rounded-md bg-[#0a0a0a]">
      <div className="flex items-center gap-2 border-b border-border/40 px-3 py-2">
        {INTERVAL_OPTIONS.map((option) => (
          <button
            key={option.value}
            type="button"
            onClick={() => setInterval(option.value)}
            className={`rounded-sm px-2 py-1 font-mono text-[10px] uppercase tracking-widest transition-colors ${
              interval === option.value
                ? "bg-emerald-500/15 text-emerald-300 border border-emerald-500/40"
                : "text-muted-foreground border border-transparent hover:text-foreground"
            }`}
            aria-pressed={interval === option.value}
          >
            {option.label}
          </button>
        ))}
      </div>
      <div ref={chartContainerRef} className="flex-1 min-h-0" />
    </div>
  )
}

function mapPayloadToCandles(payload: CandlePayload[]): CandlestickData[] {
  return payload.map((c) => ({
    time: Math.floor(new Date(c.openTime).getTime() / 1000) as UTCTimestamp,
    open: c.open / 100,
    high: c.high / 100,
    low: c.low / 100,
    close: c.close / 100,
  }))
}

function mergeCandles(existing: CandlestickData[], incoming: CandlestickData[]) {
  const map = new Map<number, CandlestickData>()
  for (const candle of existing) {
    map.set(Number(candle.time), candle)
  }
  for (const candle of incoming) {
    map.set(Number(candle.time), candle)
  }
  return sortCandles(Array.from(map.values()))
}

function sortCandles(candles: CandlestickData[]) {
  return candles.sort((a, b) => Number(a.time) - Number(b.time))
}

function updateBounds(
  candles: CandlestickData[],
  earliestRef: MutableRefObject<number | null>,
  latestRef: MutableRefObject<number | null>
) {
  if (candles.length === 0) {
    earliestRef.current = null
    latestRef.current = null
    return
  }
  earliestRef.current = Number(candles[0]!.time)
  latestRef.current = Number(candles[candles.length - 1]!.time)
}
