"use client"
import { cn } from "@workspace/ui/lib/utils"
import { useRouter } from "next/navigation"

interface LogoProps {
  className?: string
}

function Logo({ className }: LogoProps) {
  const router = useRouter();
  return (
    <div
      onClick={() => router.push(process.env.NEXT_PUBLIC_BASE_APP_URL ?? "")}
      className={cn("flex items-center gap-2 cursor-pointer group", className)}
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="28"
        height="28"
        viewBox="0 0 24 24"
        fill="none"
        stroke="#f97316"
        strokeWidth="3.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="group-hover:drop-shadow-[0_0_6px_rgba(249,115,22,0.6)] transition-all duration-200"
      >
        <polyline points="23 6 13.5 15.5 8.5 10.5 1 18" />
        <polyline points="17 6 23 6 23 12" />
      </svg>
      <span className="font-bold text-lg tracking-tight text-foreground group-hover:text-orange-500 transition-colors duration-200">
        Rivon
      </span>
    </div>
  )
}

export default Logo;
