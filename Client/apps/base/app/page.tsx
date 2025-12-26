"use client"

import { Button } from "@workspace/ui/components/button"
import Link from "next/link"
import { ArrowRight, BarChart3, ShieldCheck, Zap, Bot, GraduationCap, Globe2, Trophy, ArrowLeftRight, Coins } from "lucide-react"
import { useEffect, useState, useCallback } from "react"
import { motion, AnimatePresence } from "framer-motion"

// Animation Variants
const fadeInUp = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.6, ease: "easeOut" as const } }
}

const staggerContainer = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1
    }
  }
}

export default function Page() {
  const [showIntro, setShowIntro] = useState(true)



  const handleIntroComplete = useCallback(() => {
    setShowIntro(false)
  }, [])

  return (
    <div className="flex flex-col min-h-screen bg-background text-foreground overflow-x-hidden">
      <AnimatePresence mode="wait">
        {showIntro && <IntroAnimation onComplete={handleIntroComplete} />}
      </AnimatePresence>

      {!showIntro && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.8 }}
        >
          {/* Hero Section */}
          <section className="relative flex flex-col items-center justify-center px-4 py-24 text-center md:py-32 lg:py-40 overflow-hidden">
            {/* Animated Background Grid */}
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 1.5 }}
              className="absolute inset-0 -z-10 bg-[linear-gradient(to_right,#80808012_1px,transparent_1px),linear-gradient(to_bottom,#80808012_1px,transparent_1px)] bg-[size:24px_24px]"
            />

            {/* Animated Glow Blob */}
            <motion.div
              animate={{
                scale: [1, 1.2, 1],
                opacity: [0.2, 0.3, 0.2],
              }}
              transition={{
                duration: 8,
                repeat: Infinity,
                ease: "easeInOut" as const
              }}
              className="absolute left-0 right-0 top-0 -z-10 m-auto h-[310px] w-[310px] rounded-full bg-orange-500/20 blur-[100px]"
            />

            <motion.div
              variants={staggerContainer}
              initial="hidden"
              animate="visible"
              className="max-w-4xl space-y-6"
            >
              <motion.h1
                variants={fadeInUp}
                className="text-4xl font-extrabold tracking-tight sm:text-5xl md:text-6xl lg:text-7xl"
              >
                Trade Teams. <motion.span
                  className="text-orange-500 inline-block"
                  animate={{ color: ["#f97316", "#fb923c", "#f97316"] }}
                  transition={{ duration: 4, repeat: Infinity }}
                >Bet Smarter.</motion.span> <br className="hidden sm:inline" />
                One Ecosystem.
              </motion.h1>

              <motion.p
                variants={fadeInUp}
                className="mx-auto max-w-2xl text-lg text-muted-foreground sm:text-xl"
              >
                Rivon combines a <span className="text-foreground font-medium">virtual exchange</span> with <span className="text-foreground font-medium">intelligent betting</span>.
                Trade teams like stocks to build value, then leverage your portfolio to place bets—with odds driven by <span className="text-orange-500">predictive models</span> and <span className="text-orange-500">real-time market action</span>.
              </motion.p>

              <motion.div
                variants={fadeInUp}
                className="flex flex-col items-center justify-center gap-4 sm:flex-row pt-4"
              >
                <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
                  <Button size="lg" className="h-12 px-8 text-base bg-orange-500 hover:bg-orange-600 text-white border-0 shadow-lg shadow-orange-500/20">
                    Start Trading
                  </Button>
                </motion.div>
                <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
                  <Button size="lg" variant="outline" className="h-12 px-8 text-base hover:text-orange-500 hover:border-orange-500/50">
                    Explore Markets
                  </Button>
                </motion.div>
              </motion.div>

              <motion.p
                variants={fadeInUp}
                className="pt-8 text-sm text-muted-foreground/80 font-medium"
              >
                Built by an engineer obsessed with performance, fairness, and <span className="text-orange-500/80">real market mechanics</span>.
              </motion.p>
            </motion.div>
          </section>

          {/* What is Rivon Section */}
          <section className="py-20 bg-muted/30 border-y overflow-hidden">
            <div className="container px-4 mx-auto max-w-5xl">
              <div className="grid gap-12 lg:grid-cols-2 lg:gap-8 items-center">
                <motion.div
                  initial="hidden"
                  whileInView="visible"
                  viewport={{ once: true, margin: "-100px" }}
                  variants={staggerContainer}
                  className="space-y-8"
                >
                  <motion.h2 variants={fadeInUp} className="text-3xl font-bold tracking-tight sm:text-4xl">
                    Two Engines. <span className="text-orange-500">One Platform.</span>
                  </motion.h2>

                  <div className="space-y-6">
                    <motion.div variants={fadeInUp} className="flex gap-4 group">
                      <div className="mt-1 p-2 w-fit h-fit rounded-lg bg-orange-500/10 text-orange-500 group-hover:bg-orange-500 group-hover:text-white transition-colors duration-300">
                        <ArrowLeftRight className="w-5 h-5" />
                      </div>
                      <div>
                        <h3 className="text-xl font-semibold mb-2">1. The Exchange</h3>
                        <p className="text-muted-foreground">
                          Use <span className="text-foreground font-medium">virtual money</span> to trade teams like stocks.
                          Buy and sell based on real-time demand and match events.
                          Master market behavior and risk management without the financial loss.
                        </p>
                      </div>
                    </motion.div>

                    <motion.div variants={fadeInUp} className="flex gap-4 group">
                      <div className="mt-1 p-2 w-fit h-fit rounded-lg bg-orange-500/10 text-orange-500 group-hover:bg-orange-500 group-hover:text-white transition-colors duration-300">
                        <Coins className="w-5 h-5" />
                      </div>
                      <div>
                        <h3 className="text-xl font-semibold mb-2">2. The Betting</h3>
                        <p className="text-muted-foreground">
                          Use the value you've built on the exchange to place bets.
                          Odds are dynamically generated by fusing our <span className="text-foreground font-medium">proprietary predictive models</span> with <span className="text-foreground font-medium">live market sentiment</span>.
                          It's the perfect balance of data-driven accuracy and market-driven liquidity.
                        </p>
                      </div>
                    </motion.div>

                    <motion.p
                      variants={fadeInUp}
                      className="text-sm italic border-l-2 border-orange-500 pl-4 py-1 text-muted-foreground"
                    >
                      "It's not just betting; it's a financial ecosystem where you trade, hedge, and bet using the same underlying markets."
                    </motion.p>
                  </div>
                </motion.div>

                <motion.div
                  initial={{ opacity: 0, x: 50 }}
                  whileInView={{ opacity: 1, x: 0 }}
                  viewport={{ once: true }}
                  transition={{ duration: 0.8, ease: "easeOut" as const }}
                >
                  <LiveMarketPreview />
                </motion.div>
              </div>
            </div>
          </section>

          {/* Global Coverage Section */}
          <section className="py-24 px-4 bg-background border-b overflow-hidden">
            <div className="container mx-auto max-w-6xl text-center">
              <motion.div
                initial="hidden"
                whileInView="visible"
                viewport={{ once: true }}
                variants={staggerContainer}
                className="mb-12 space-y-4"
              >
                <motion.h2 variants={fadeInUp} className="text-3xl font-bold tracking-tight sm:text-4xl flex items-center justify-center gap-3">
                  <Globe2 className="w-8 h-8 text-orange-500" />
                  Global Market Coverage
                </motion.h2>
                <motion.p variants={fadeInUp} className="text-lg text-muted-foreground max-w-2xl mx-auto">
                  We cover the biggest stages in football. From the intensity of the Premier League to the glory of the Champions League.
                </motion.p>
              </motion.div>

              <motion.div
                initial="hidden"
                whileInView="visible"
                viewport={{ once: true }}
                variants={{
                  hidden: { opacity: 0 },
                  visible: {
                    opacity: 1,
                    transition: { staggerChildren: 0.1 }
                  }
                }}
                className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-6"
              >
                <LeagueCard name="Premier League" country="England" delay={0} />
                <LeagueCard name="La Liga" country="Spain" delay={0.1} />
                <LeagueCard name="Bundesliga" country="Germany" delay={0.2} />
                <LeagueCard name="Serie A" country="Italy" delay={0.3} />
                <LeagueCard name="Ligue 1" country="France" delay={0.4} />
                <LeagueCard name="Champions League" country="Europe" highlight delay={0.5} />
              </motion.div>

              <motion.p
                initial={{ opacity: 0 }}
                whileInView={{ opacity: 1 }}
                viewport={{ once: true }}
                transition={{ delay: 0.8 }}
                className="mt-12 text-sm font-medium text-muted-foreground"
              >
                Our roadmap includes expanding to every major team sport globally.
              </motion.p>
            </div>
          </section>

          {/* Features Section */}
          <section className="py-24 px-4 overflow-hidden">
            <div className="container mx-auto max-w-6xl">
              <motion.div
                initial="hidden"
                whileInView="visible"
                viewport={{ once: true }}
                variants={fadeInUp}
                className="text-center mb-16"
              >
                <h2 className="text-3xl font-bold tracking-tight sm:text-4xl mb-4">What You Can Do</h2>
                <p className="text-muted-foreground text-lg">Real trading mechanics applied to sports.</p>
              </motion.div>

              <motion.div
                initial="hidden"
                whileInView="visible"
                viewport={{ once: true, margin: "-50px" }}
                variants={staggerContainer}
                className="grid gap-6 md:grid-cols-2 lg:grid-cols-3"
              >
                <FeatureCard
                  icon={<BarChart3 className="w-6 h-6" />}
                  title="Trade Teams Like Stocks"
                  description="Buy low, sell high, or hold positions as matches evolve using virtual currency."
                />
                <FeatureCard
                  icon={<Zap className="w-6 h-6" />}
                  title="Model + Market Odds"
                  description="Betting odds are set by advanced algorithms and adjusted in real-time by market volume."
                />
                <FeatureCard
                  icon={<ArrowRight className="w-6 h-6" />}
                  title="Real-Time Order Books"
                  description="Each market runs independently for speed and isolation."
                />
                <FeatureCard
                  icon={<Bot className="w-6 h-6" />}
                  title="Automated Market Makers"
                  description="Bots ensure liquidity while real users shape prices."
                />
                <FeatureCard
                  icon={<GraduationCap className="w-6 h-6" />}
                  title="Learn Risk Management"
                  description="Master trading psychology and timing in a risk-free environment."
                />
                <FeatureCard
                  icon={<ShieldCheck className="w-6 h-6" />}
                  title="Fair & Transparent"
                  description="No hidden margins. Pure peer-to-peer exchange dynamics."
                />
              </motion.div>
            </div>
          </section>

          {/* Builder / Trust Section */}
          <section className="py-20 bg-muted/30 border-t">
            <div className="container mx-auto px-4 text-center max-w-3xl">
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.6 }}
              >
                <h2 className="text-2xl font-bold mb-6">How it's built</h2>
                <p className="text-lg text-muted-foreground mb-8">
                  Rivon is engineered for speed and reliability. It's not just a website; it's a high-performance trading engine.
                </p>

                <div className="flex flex-col items-center justify-center space-y-4">
                  <p className="font-medium">Built by Ashwin Rai</p>
                  <div className="flex gap-6">
                    <Link href="https://ashprojects.tech" className="text-orange-500 hover:underline underline-offset-4 hover:text-orange-600 transition-colors">
                      ashprojects.tech
                    </Link>
                    <Link href="https://github.com/raiashpanda007" className="text-orange-500 hover:underline underline-offset-4 hover:text-orange-600 transition-colors">
                      GitHub
                    </Link>
                  </div>
                </div>
              </motion.div>
            </div>
          </section>

          {/* Footer */}
          <footer className="py-12 border-t bg-background">
            <div className="container mx-auto px-4 flex flex-col md:flex-row justify-between items-center gap-6">
              <div className="text-sm text-muted-foreground">
                © {new Date().getFullYear()} Rivon. All rights reserved.
              </div>
              <div className="flex gap-6 text-sm text-muted-foreground">
                <Link href="#" className="hover:text-foreground transition-colors">Support</Link>
                <Link href="#" className="hover:text-foreground transition-colors">Terms</Link>
                <Link href="#" className="hover:text-foreground transition-colors">Privacy</Link>
              </div>
            </div>
            <div className="container mx-auto px-4 mt-8 text-xs text-muted-foreground text-center max-w-2xl">
              <p>
                Disclaimer: Rivon is a simulated trading platform. No real money gambling is involved in the current version.
                Trading involves risk.
              </p>
            </div>
          </footer>
        </motion.div>
      )}
    </div>
  )
}

function IntroAnimation({ onComplete }: { onComplete: () => void }) {
  const [step, setStep] = useState(0)

  useEffect(() => {
    let timer: NodeJS.Timeout

    if (step === 0) {
      // INVEST
      timer = setTimeout(() => setStep(1), 1500)
    } else if (step === 1) {
      // TRADE
      timer = setTimeout(() => setStep(2), 1500)
    } else if (step === 2) {
      // BET
      timer = setTimeout(() => setStep(3), 2500)
    } else if (step === 3) {
      // Logo
      timer = setTimeout(() => {
        onComplete()
      }, 2000)
    }

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

function FeatureCard({ icon, title, description }: { icon: React.ReactNode, title: string, description: string }) {
  return (
    <motion.div
      variants={fadeInUp}
      whileHover={{ y: -5, transition: { duration: 0.2 } }}
      className="p-6 rounded-xl border bg-card text-card-foreground shadow-sm hover:shadow-lg hover:border-orange-500/30 transition-all group"
    >
      <div className="mb-4 p-3 w-fit rounded-lg bg-orange-500/10 text-orange-500 group-hover:bg-orange-500 group-hover:text-white transition-colors duration-300">
        {icon}
      </div>
      <h3 className="text-xl font-semibold mb-2">{title}</h3>
      <p className="text-muted-foreground">{description}</p>
    </motion.div>
  )
}

function LeagueCard({ name, country, highlight, delay }: { name: string, country: string, highlight?: boolean, delay?: number }) {
  return (
    <motion.div
      variants={{
        hidden: { opacity: 0, scale: 0.8 },
        visible: { opacity: 1, scale: 1, transition: { duration: 0.4, delay } }
      }}
      whileHover={{ scale: 1.05 }}
      className={`p-4 rounded-lg border flex flex-col items-center justify-center text-center gap-2 cursor-default ${highlight ? 'bg-orange-500/5 border-orange-500/20' : 'bg-card hover:border-orange-500/30'}`}
    >
      <Trophy className={`w-5 h-5 ${highlight ? 'text-orange-500' : 'text-muted-foreground'}`} />
      <div>
        <div className="font-semibold text-sm">{name}</div>
        <div className="text-xs text-muted-foreground">{country}</div>
      </div>
    </motion.div>
  )
}

function LiveMarketPreview() {
  const [rmPrice, setRmPrice] = useState(4.20)
  const [bmPrice, setBmPrice] = useState(1.80)
  const [rmTrend, setRmTrend] = useState(1) // 1 up, -1 down
  const [bmTrend, setBmTrend] = useState(-1)

  useEffect(() => {
    const interval = setInterval(() => {
      // Simulate random price movement
      const moveA = (Math.random() - 0.4) * 0.1 // Slight upward bias
      const moveB = (Math.random() - 0.6) * 0.1 // Slight downward bias

      setRmPrice(p => Math.max(0.1, p + moveA))
      setBmPrice(p => Math.max(0.1, p + moveB))

      setRmTrend(moveA > 0 ? 1 : -1)
      setBmTrend(moveB > 0 ? 1 : -1)
    }, 2000)

    return () => clearInterval(interval)
  }, [])

  return (
    <motion.div
      whileHover={{ scale: 1.02 }}
      className="relative p-8 bg-card border rounded-xl shadow-sm overflow-hidden"
    >
      <motion.div
        animate={{ opacity: [0.3, 0.6, 0.3], scale: [1, 1.1, 1] }}
        transition={{ duration: 3, repeat: Infinity }}
        className="absolute -top-4 -right-4 w-24 h-24 bg-orange-500/10 rounded-full blur-2xl"
      />
      <div className="space-y-4">
        <div className="flex items-center justify-between p-4 bg-background border rounded-lg transition-colors duration-500 hover:border-orange-500/30">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-gray-100 flex items-center justify-center text-xs font-bold text-black">RM</div>
            <span className="font-medium">Real Madrid</span>
          </div>
          <motion.div
            key={rmPrice}
            initial={{ scale: 1.2, color: rmTrend > 0 ? "#22c55e" : "#ef4444" }}
            animate={{ scale: 1, color: rmTrend > 0 ? "#22c55e" : "#ef4444" }}
            className="font-mono"
          >
            {rmTrend > 0 ? '▲' : '▼'} ${rmPrice.toFixed(2)}
          </motion.div>
        </div>
        <div className="flex items-center justify-between p-4 bg-background border rounded-lg transition-colors duration-500 hover:border-orange-500/30">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-red-100 flex items-center justify-center text-xs font-bold text-red-800">BM</div>
            <span className="font-medium">Bayern Munich</span>
          </div>
          <motion.div
            key={bmPrice}
            initial={{ scale: 1.2, color: bmTrend > 0 ? "#22c55e" : "#ef4444" }}
            animate={{ scale: 1, color: bmTrend > 0 ? "#22c55e" : "#ef4444" }}
            className="font-mono"
          >
            {bmTrend > 0 ? '▲' : '▼'} ${bmPrice.toFixed(2)}
          </motion.div>
        </div>
        <div className="flex items-center justify-between p-4 bg-background border rounded-lg opacity-50">
          <span className="font-medium">Market Volume</span>
          <span className="font-mono">12,450 Trades</span>
        </div>
      </div>
      <div className="absolute bottom-2 right-4 text-[10px] text-muted-foreground/50">
        * Live simulated data
      </div>
    </motion.div>
  )
}
