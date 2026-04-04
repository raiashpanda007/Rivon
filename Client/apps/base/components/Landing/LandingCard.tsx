"use client"
import { useRouter } from "next/navigation";
import { Button } from "@workspace/ui/components/button";
import { motion } from "framer-motion";
export default function LandingPageRefRouting() {

  const router = useRouter();
  return (
    <>
      <motion.div whileHover={{ scale: 1.03 }} whileTap={{ scale: 0.97 }}>
        <Button
          size="lg"
          className="h-11 px-8 text-sm bg-orange-500 hover:bg-orange-600 text-white border-0 shadow-[0_0_20px_rgba(249,115,22,0.3)] cursor-pointer rounded-sm font-mono tracking-wide"
          onClick={() => router.push(process.env.NEXT_PUBLIC_BET_APP_URL ?? "")}

        >
          EXPLORE_BETTING
        </Button>
      </motion.div>
      <motion.div whileHover={{ scale: 1.03 }} whileTap={{ scale: 0.97 }}>
        <Button
          onClick={() => router.push(process.env.NEXT_PUBLIC_TRADE_APP_URL ?? "")}
          size="lg"
          variant="outline"
          className="h-11 px-8 text-sm hover:text-orange-500 hover:border-orange-500/50 cursor-pointer rounded-sm font-mono tracking-wide"
        >
          START_TRADING
        </Button>
      </motion.div>
    </>
  )
}
