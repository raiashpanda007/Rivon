"use client"

import { Button } from "@workspace/ui/components/button"
import Link from "next/link"
import { BarChart3, ShieldCheck, Zap, Bot, GraduationCap, Globe2, ArrowLeftRight, Coins, Activity } from "lucide-react"
import { useEffect, useState, useCallback } from "react"
import { motion, AnimatePresence } from "framer-motion"
import LandingPageRefRouting from "@/components/Landing/LandingCard"

// ─── Animation variants ─────────────────────────────────────────────────────
const fadeInUp = {
  hidden: { opacity: 0, y: 16 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.5, ease: "easeOut" as const } }
}
const staggerContainer = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.08 } }
}

// ─── Mock data ───────────────────────────────────────────────────────────────
const TICKER_DATA = [
  { code: "RMFC", price: 4.23, change: +2.14 },
  { code: "BMFC", price: 1.87, change: -1.30 },
  { code: "MCFC", price: 3.52, change: +0.88 },
  { code: "LCFC", price: 0.94, change: -3.21 },
  { code: "MUFC", price: 2.71, change: +1.52 },
  { code: "ARFC", price: 3.18, change: +4.10 },
  { code: "LSFC", price: 1.44, change: -0.67 },
  { code: "ATMC", price: 2.06, change: +2.33 },
  { code: "PSGFC", price: 2.95, change: -0.44 },
  { code: "JVFC", price: 1.63, change: +1.77 },
  { code: "IMFC", price: 2.28, change: -2.05 },
  { code: "BFCB", price: 1.12, change: +0.55 },
]

const ACTIVITY_ITEMS = [
  { type: "trade", side: "BUY", team: "RMFC", price: 4.23, qty: 50, time: "14:32:07" },
  { type: "bet", team: "MCFC", market: "WIN", odds: "2.10", time: "14:32:04" },
  { type: "trade", side: "SELL", team: "BMFC", price: 1.87, qty: 100, time: "14:31:58" },
  { type: "trade", side: "BUY", team: "ARFC", price: 3.18, qty: 25, time: "14:31:51" },
  { type: "bet", team: "RMFC", market: "DRAW", odds: "3.40", time: "14:31:44" },
  { type: "trade", side: "BUY", team: "MUFC", price: 2.71, qty: 75, time: "14:31:39" },
]

const LEAGUES = [
  { name: "Premier League", country: "ENG" },
  { name: "La Liga", country: "ESP" },
  { name: "Bundesliga", country: "GER" },
  { name: "Serie A", country: "ITA" },
  { name: "Ligue 1", country: "FRA" },
  { name: "Champions League", country: "UEFA" },
]

const FEATURES = [
  { icon: <BarChart3 className="w-4 h-4" />, title: "Trade Teams Like Stocks", desc: "Buy low, sell high, or hold positions as matches evolve using virtual currency." },
  { icon: <Zap className="w-4 h-4" />, title: "Model + Market Odds", desc: "Betting odds set by advanced algorithms, adjusted in real-time by volume." },
  { icon: <ArrowLeftRight className="w-4 h-4" />, title: "Real-Time Order Books", desc: "Each market runs on an independent high-performance matching engine." },
  { icon: <Bot className="w-4 h-4" />, title: "Automated Market Makers", desc: "Bots ensure liquidity while real users drive price discovery." },
  { icon: <GraduationCap className="w-4 h-4" />, title: "Learn Risk Management", desc: "Master trading psychology and timing in a zero-risk environment." },
  { icon: <ShieldCheck className="w-4 h-4" />, title: "Fair & Transparent", desc: "No hidden margins. Pure peer-to-peer exchange mechanics." },
]

// ─── Main Page ───────────────────────────────────────────────────────────────
export default function Page() {
  const [showIntro, setShowIntro] = useState(true)

  useEffect(() => {
    if (sessionStorage.getItem("rivon_intro_shown")) {
      setShowIntro(false)
    }
  }, [])

  const handleIntroComplete = useCallback(() => {
    sessionStorage.setItem("rivon_intro_shown", "1")
    setShowIntro(false)
  }, [])

  return (
    <div className="w-full">
      <AnimatePresence mode="wait">
        {showIntro && <IntroAnimation onComplete={handleIntroComplete} />}
      </AnimatePresence>

      {!showIntro && (
        <motion.div className="w-full" initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.6 }}>

          {/* ── Hero ──────────────────────────────────────────────────────── */}
          <section className="relative flex flex-col items-center justify-center px-4 py-24 md:py-32 text-center overflow-hidden">
            {/* Grid overlay */}
            <div className="absolute inset-0 -z-10 bg-terminal-grid" />
            {/* Orange radial glow */}
            <motion.div
              animate={{ scale: [1, 1.15, 1], opacity: [0.15, 0.25, 0.15] }}
              transition={{ duration: 10, repeat: Infinity, ease: "easeInOut" }}
              className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 -z-10 w-[480px] h-[480px] rounded-full bg-orange-500/20 blur-[100px]"
            />

            {/* System tag */}
            <motion.div
              initial={{ opacity: 0, y: -8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 }}
              className="mb-6 flex items-center gap-2 px-3 py-1 border border-orange-500/30 bg-orange-500/5 rounded-sm"
            >
              <span className="dot-live" />
              <span className="font-mono text-[10px] text-orange-400 tracking-widest">SYS:RIVON_EXCHANGE · v1.0 · LIVE</span>
            </motion.div>

            <motion.div
              variants={staggerContainer}
              initial="hidden"
              animate="visible"
              className="max-w-4xl space-y-5"
            >
              <motion.h1
                variants={fadeInUp}
                className="text-4xl font-extrabold tracking-tight sm:text-5xl md:text-6xl lg:text-7xl"
              >
                Trade Teams.{" "}
                <motion.span
                  className="text-orange-500 inline-block"
                  animate={{ textShadow: ["0 0 0px #f97316", "0 0 20px #f97316", "0 0 0px #f97316"] }}
                  transition={{ duration: 4, repeat: Infinity }}
                >
                  Bet Smarter.
                </motion.span>
                <br className="hidden sm:inline" />
                One Ecosystem.
              </motion.h1>

              <motion.p variants={fadeInUp} className="mx-auto max-w-2xl text-base text-muted-foreground sm:text-lg">
                Rivon combines a <span className="text-foreground font-medium">virtual exchange</span> with{" "}
                <span className="text-foreground font-medium">intelligent betting</span>. Trade teams like stocks,
                then use your portfolio value to place bets driven by{" "}
                <span className="text-orange-500">predictive models</span> and{" "}
                <span className="text-orange-500">real-time market action</span>.
              </motion.p>

              <motion.div
                variants={fadeInUp}
                className="flex flex-col items-center justify-center gap-3 sm:flex-row pt-3"
              >
                <LandingPageRefRouting />

              </motion.div>
            </motion.div>
          </section>

          {/* ── Scrolling Ticker ──────────────────────────────────────────── */}
          <MarketTicker />

          {/* ── Two Engines ──────────────────────────────────────────────── */}
          <section className="py-20 border-y border-border overflow-hidden">
            <div className="container px-4 mx-auto max-w-6xl">
              <div className="grid gap-10 lg:grid-cols-2 lg:gap-12 items-start">
                <motion.div
                  initial="hidden"
                  whileInView="visible"
                  viewport={{ once: true, margin: "-80px" }}
                  variants={staggerContainer}
                  className="space-y-6"
                >
                  <motion.div variants={fadeInUp}>
                    <span className="font-mono text-[10px] text-orange-400 tracking-widest">ARCHITECTURE</span>
                    <h2 className="text-2xl font-bold tracking-tight sm:text-3xl mt-1">
                      Two Engines. <span className="text-orange-500">One Platform.</span>
                    </h2>
                  </motion.div>

                  <motion.div variants={fadeInUp} className="terminal-panel">
                    <div className="terminal-panel-header">
                      <div className="w-2 h-2 rounded-full bg-blue-400" />
                      <span className="font-mono text-[10px] text-muted-foreground">MODULE_01 · EXCHANGE</span>
                    </div>
                    <div className="flex gap-3 px-3 py-3"> <div className="mt-0.5 p-1.5 bg-orange-500/10 text-orange-500 rounded-sm shrink-0">
                      <ArrowLeftRight className="w-4 h-4" />
                    </div>
                      <div>
                        <h3 className="text-sm font-semibold mb-1">Virtual Stock Exchange</h3>
                        <p className="text-xs text-muted-foreground leading-relaxed">
                          Use virtual money to trade football teams like stocks.
                          Buy and sell based on real-time demand and match events.
                          Master market behavior and risk management.
                        </p>
                      </div>
                    </div>
                  </motion.div>

                  <motion.div variants={fadeInUp} className="terminal-panel">
                    <div className="terminal-panel-header">
                      <div className="w-2 h-2 rounded-full bg-orange-400" />
                      <span className="font-mono text-[10px] text-muted-foreground">MODULE_02 · BETTING</span>
                    </div>
                    <div className="flex gap-3 px-3 py-3">
                      <div className="mt-0.5 p-1.5 bg-orange-500/10 text-orange-500 rounded-sm shrink-0">
                        <Coins className="w-4 h-4" />
                      </div>
                      <div>
                        <h3 className="text-sm font-semibold mb-1">Intelligent Betting Layer</h3>
                        <p className="text-xs text-muted-foreground leading-relaxed">
                          Use exchange portfolio value to place bets. Odds fuse{" "}
                          <span className="text-foreground">proprietary predictive models</span> with{" "}
                          <span className="text-foreground">live market sentiment</span> — the perfect
                          balance of data accuracy and market liquidity.
                        </p>
                      </div>
                    </div>
                  </motion.div>

                  <motion.p
                    variants={fadeInUp}
                    className="text-xs italic border-l-2 border-orange-500 pl-3 py-1 text-muted-foreground"
                  >
                    "A financial ecosystem where you trade, hedge, and bet using the same underlying markets."
                  </motion.p>
                </motion.div>

                {/* Right: Live panels */}
                <motion.div
                  initial={{ opacity: 0, x: 40 }}
                  whileInView={{ opacity: 1, x: 0 }}
                  viewport={{ once: true }}
                  transition={{ duration: 0.7, ease: "easeOut" }}
                  className="grid gap-4"
                >
                  <MiniOrderBook />
                  <ActivityFeed />
                </motion.div>
              </div>
            </div>
          </section>

          {/* ── Active Markets ────────────────────────────────────────────── */}
          <section className="py-20 px-4 border-b border-border overflow-hidden">
            <div className="container mx-auto max-w-5xl">
              <motion.div
                initial="hidden"
                whileInView="visible"
                viewport={{ once: true }}
                variants={staggerContainer}
                className="mb-6"
              >
                <motion.div variants={fadeInUp} className="flex items-center gap-3 mb-1">
                  <Globe2 className="w-4 h-4 text-orange-500" />
                  <span className="font-mono text-[10px] text-orange-400 tracking-widest">ACTIVE_MARKETS</span>
                </motion.div>
                <motion.h2 variants={fadeInUp} className="text-2xl font-bold tracking-tight">
                  Global Market Coverage
                </motion.h2>
                <motion.p variants={fadeInUp} className="text-sm text-muted-foreground mt-1 max-w-xl">
                  Covering the biggest stages in football — from Premier League intensity to Champions League glory.
                </motion.p>
              </motion.div>

              <motion.div
                initial="hidden"
                whileInView="visible"
                viewport={{ once: true }}
                variants={staggerContainer}
                className="terminal-panel"
              >
                <div className="terminal-panel-header">
                  <span className="dot-live" />
                  <span className="font-mono text-[10px] text-muted-foreground">LEAGUES · {LEAGUES.length} ACTIVE</span>
                </div>
                <div className="grid grid-cols-2 sm:grid-cols-3 divide-border">
                  {LEAGUES.map((league, i) => (
                    <motion.div
                      key={league.name}
                      variants={{
                        hidden: { opacity: 0 },
                        visible: { opacity: 1, transition: { delay: i * 0.05 } }
                      }}
                      className="terminal-data-row border-r border-border last:border-r-0 [&:nth-child(2n)]:border-r-0 sm:[&:nth-child(2n)]:border-r sm:[&:nth-child(3n)]:border-r-0"
                    >
                      <div className="flex items-center gap-2">
                        <span className="w-1.5 h-1.5 rounded-full bg-green-500" />
                        <span className="text-sm font-medium">{league.name}</span>
                      </div>
                      <span className="font-mono text-[10px] text-muted-foreground">{league.country}</span>
                    </motion.div>
                  ))}
                </div>
                <div className="px-3 py-2 text-[10px] text-muted-foreground/60 font-mono border-t border-border">
                  ROADMAP: expanding to all major team sports globally
                </div>
              </motion.div>
            </div>
          </section>

          {/* ── System Capabilities ──────────────────────────────────────── */}
          <section className="py-20 px-4 overflow-hidden">
            <div className="container mx-auto max-w-5xl">
              <motion.div
                initial="hidden"
                whileInView="visible"
                viewport={{ once: true }}
                variants={fadeInUp}
                className="mb-6"
              >
                <span className="font-mono text-[10px] text-orange-400 tracking-widest">SYSTEM_CAPABILITIES</span>
                <h2 className="text-2xl font-bold tracking-tight mt-1">What You Can Do</h2>
              </motion.div>

              <motion.div
                initial="hidden"
                whileInView="visible"
                viewport={{ once: true, margin: "-40px" }}
                variants={staggerContainer}
                className="terminal-panel"
              >
                <div className="terminal-panel-header">
                  <span className="font-mono text-[10px] text-muted-foreground">6 MODULES LOADED</span>
                </div>
                {FEATURES.map((feature, i) => (
                  <motion.div
                    key={feature.title}
                    variants={fadeInUp}
                    className="group terminal-data-row"
                  >
                    <div className="flex items-center gap-3 flex-1 min-w-0">
                      <div className="p-1.5 rounded-sm bg-orange-500/10 text-orange-500 group-hover:bg-orange-500 group-hover:text-white transition-colors duration-200 shrink-0">
                        {feature.icon}
                      </div>
                      <div className="min-w-0">
                        <span className="text-sm font-medium">{feature.title}</span>
                      </div>
                    </div>
                    <span className="text-xs text-muted-foreground ml-4 text-right hidden md:block max-w-xs">
                      {feature.desc}
                    </span>
                  </motion.div>
                ))}
              </motion.div>
            </div>
          </section>

          {/* ── Builder Section ──────────────────────────────────────────── */}
          <section className="py-16 border-t border-border">
            <div className="container mx-auto px-4 max-w-3xl">
              <motion.div
                initial={{ opacity: 0, y: 16 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5 }}
                className="terminal-panel"
              >
                <div className="terminal-panel-header">
                  <Activity className="w-3 h-3 text-muted-foreground" />
                  <span className="font-mono text-[10px] text-muted-foreground">BUILD_INFO</span>
                </div>
                <div className="px-4 py-5 text-center space-y-3">
                  <p className="text-sm text-muted-foreground">
                    Rivon is engineered for speed and reliability — not just a website, but a high-performance trading engine.
                  </p>
                  <div className="flex items-center justify-center gap-1.5 text-xs text-muted-foreground font-mono">
                    <span>BUILT BY</span>
                    <span className="text-foreground font-bold">ASHWIN RAI</span>
                  </div>
                  <div className="flex gap-6 justify-center">
                    <Link href="https://ashprojects.tech" className="text-orange-500 text-xs hover:underline underline-offset-4 font-mono">
                      ashprojects.tech
                    </Link>
                    <Link href="https://github.com/raiashpanda007" className="text-orange-500 text-xs hover:underline underline-offset-4 font-mono">
                      github
                    </Link>
                  </div>
                </div>
              </motion.div>
            </div>
          </section>

          {/* ── Footer ───────────────────────────────────────────────────── */}
          <footer className="py-8 border-t border-border bg-card/50">
            <div className="container mx-auto px-4 flex flex-col md:flex-row justify-between items-center gap-4">
              <span className="font-mono text-[10px] text-muted-foreground">
                © {new Date().getFullYear()} RIVON · ALL RIGHTS RESERVED
              </span>
              <div className="flex gap-5 text-[10px] font-mono text-muted-foreground">
                <Link href="#" className="hover:text-orange-500 transition-colors">SUPPORT</Link>
                <Link href="#" className="hover:text-orange-500 transition-colors">TERMS</Link>
                <Link href="#" className="hover:text-orange-500 transition-colors">PRIVACY</Link>
              </div>
            </div>
            <div className="container mx-auto px-4 mt-4 text-[10px] text-muted-foreground/50 text-center font-mono">
              DISCLAIMER: SIMULATED TRADING PLATFORM · NO REAL MONEY INVOLVED
            </div>
          </footer>
        </motion.div>
      )}
    </div>
  )
}

// ─── Intro Animation (unchanged) ─────────────────────────────────────────────
function IntroAnimation({ onComplete }: { onComplete: () => void }) {
  const [step, setStep] = useState(0)

  useEffect(() => {
    let timer: NodeJS.Timeout
    if (step === 0) timer = setTimeout(() => setStep(1), 1500)
    else if (step === 1) timer = setTimeout(() => setStep(2), 1500)
    else if (step === 2) timer = setTimeout(() => setStep(3), 2500)
    else if (step === 3) timer = setTimeout(() => onComplete(), 2000)
    return () => clearTimeout(timer)
  }, [step, onComplete])

  return (
    <motion.div
      className="fixed inset-0 z-50 flex items-center justify-center bg-background"
      exit={{ opacity: 0, y: -50, transition: { duration: 0.8, ease: "easeInOut" as const } }}
    >
      <AnimatePresence mode="wait">
        {step === 0 && (
          <motion.h1
            key="invest"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.4, ease: "easeOut" as const }}
            className="text-5xl md:text-7xl font-black tracking-tighter"
          >
            INVEST
          </motion.h1>
        )}
        {step === 1 && (
          <motion.h1
            key="trade"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.4, ease: "easeOut" as const }}
            className="text-5xl md:text-7xl font-black tracking-tighter"
          >
            TRADE
          </motion.h1>
        )}
        {step === 2 && (
          <motion.h1
            key="bet"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.4, ease: "easeOut" as const }}
            className="text-3xl md:text-5xl font-bold tracking-tight text-center px-4"
          >
            BET <span className="text-orange-500">ALL IN</span> <br /> YOUR FAV TEAM
          </motion.h1>
        )}
        {step === 3 && (
          <motion.div
            key="logo"
            initial={{ scale: 0.8, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            transition={{ duration: 0.5, ease: "backOut" as const }}
            className="flex items-center gap-4"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#f97316" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="23 6 13.5 15.5 8.5 10.5 1 18" />
              <polyline points="17 6 23 6 23 12" />
            </svg>
            <motion.span
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.3, duration: 0.4 }}
              className="text-5xl font-bold tracking-tight"
            >
              Rivon
            </motion.span>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  )
}

// ─── Market Ticker ────────────────────────────────────────────────────────────
function MarketTicker() {
  const doubled = [...TICKER_DATA, ...TICKER_DATA]
  return (
    <div className="w-[100vw] border-y border-border bg-card overflow-hidden select-none relative left-1/2 -ml-[50vw]">
      <div className="animate-ticker flex w-max">
        {doubled.map((t, i) => (
          <div
            key={i}
            className="flex items-center gap-2 px-5 py-2 border-r border-border/60 shrink-0"
          >
            <span className="font-mono text-xs font-bold text-foreground">{t.code}</span>
            <span className="font-mono text-xs text-muted-foreground">${t.price.toFixed(2)}</span>
            <span className={`font-mono text-[10px] font-bold ${t.change >= 0 ? "text-green-500" : "text-red-500"}`}>
              {t.change >= 0 ? "+" : ""}{t.change.toFixed(2)}%
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

// ─── Mini Order Book ──────────────────────────────────────────────────────────
function MiniOrderBook() {
  const [lastPrice, setLastPrice] = useState(4.2300)
  const [priceUp, setPriceUp] = useState(true)
  const asks = [
    { price: 4.2600, qty: 150 },
    { price: 4.2500, qty: 200 },
    { price: 4.2400, qty: 75 },
  ]
  const bids = [
    { price: 4.2300, qty: 300 },
    { price: 4.2200, qty: 125 },
    { price: 4.2100, qty: 90 },
  ]

  useEffect(() => {
    const interval = setInterval(() => {
      const move = (Math.random() - 0.5) * 0.015
      setLastPrice(p => {
        const n = Math.max(0.01, parseFloat((p + move).toFixed(4)))
        setPriceUp(move >= 0)
        return n
      })
    }, 1800)
    return () => clearInterval(interval)
  }, [])

  return (
    <div className="terminal-panel overflow-hidden">
      <div className="terminal-panel-header">
        <span className="dot-live" />
        <span className="font-mono text-[10px] text-muted-foreground">ORDER_BOOK</span>
        <span className="font-mono text-[10px] text-orange-400 ml-auto">RMFC/USD</span>
      </div>

      <div className="font-mono text-xs">
        {/* Header row */}
        <div className="flex justify-between px-3 py-1 text-[10px] text-muted-foreground/60 border-b border-border/40">
          <span>PRICE</span>
          <span>QTY</span>
        </div>

        {/* Asks */}
        {asks.map((ask, i) => (
          <div key={i} className="flex justify-between px-3 py-1.5 hover:bg-red-500/5 transition-colors">
            <span className="text-red-400">{ask.price.toFixed(2)}</span>
            <span className="text-muted-foreground">{ask.qty}</span>
          </div>
        ))}

        {/* Last price */}
        <div className={`flex items-center justify-between px-3 py-2 border-y border-border ${priceUp ? "bg-green-500/5" : "bg-red-500/5"}`}>
          <motion.span
            key={lastPrice}
            initial={{ opacity: 0.5 }}
            animate={{ opacity: 1 }}
            className={`font-bold text-sm ${priceUp ? "text-green-400" : "text-red-400"}`}
          >
            {priceUp ? "▲" : "▼"} ${lastPrice.toFixed(2)}
          </motion.span>
          <span className="text-[10px] text-muted-foreground">LAST PRICE</span>
        </div>

        {/* Bids */}
        {bids.map((bid, i) => (
          <div key={i} className="flex justify-between px-3 py-1.5 hover:bg-green-500/5 transition-colors">
            <span className="text-green-400">{bid.price.toFixed(2)}</span>
            <span className="text-muted-foreground">{bid.qty}</span>
          </div>
        ))}

        <div className="px-3 py-1.5 text-[10px] text-muted-foreground/50 border-t border-border/40 text-center">
          * simulated data
        </div>
      </div>
    </div>
  )
}

// ─── Activity Feed ────────────────────────────────────────────────────────────
function ActivityFeed() {
  return (
    <div className="terminal-panel overflow-hidden">
      <div className="terminal-panel-header">
        <span className="dot-live" />
        <span className="font-mono text-[10px] text-muted-foreground">LIVE_ACTIVITY</span>
        <span className="font-mono text-[10px] text-muted-foreground/40 ml-auto">GLOBAL FEED</span>
      </div>

      <div className="font-mono text-xs divide-y divide-border/60">
        {ACTIVITY_ITEMS.map((item, idx) => (
          <motion.div
            key={idx}
            initial={{ opacity: 0, x: -8 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ delay: idx * 0.05 }}
            className="flex items-center gap-2 px-3 py-2 hover:bg-muted/20 transition-colors"
          >
            <span className="text-[10px] text-muted-foreground/60 w-14 shrink-0">{item.time}</span>
            {item.type === "trade" ? (
              <>
                <span className={`w-8 shrink-0 font-bold ${item.side === "BUY" ? "text-green-400" : "text-red-400"}`}>
                  {item.side}
                </span>
                <span className="text-foreground font-bold">{item.team}</span>
                <span className="text-muted-foreground ml-auto">
                  ${item.price} × {item.qty}
                </span>
              </>
            ) : (
              <>
                <span className="w-8 shrink-0 font-bold text-orange-400">BET</span>
                <span className="text-foreground font-bold">{item.team}</span>
                <span className="text-muted-foreground">{item.market}</span>
                <span className="text-orange-400 ml-auto">@{item.odds}</span>
              </>
            )}
          </motion.div>
        ))}
      </div>
    </div>
  )
}
