"use client"
import { useRouter } from "next/navigation"
interface LogoProps {
  className?: string
}

function Logo({ className }: LogoProps) {
  const router = useRouter();
  return (
    <div onClick={() => router.push(process.env.NEXT_PUBLIC_BASE_APP_URL ?? "")} className={"flex space-x-3 px-2 items-center cursor-pointer border" + className}>
      <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#f97316" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="23 6 13.5 15.5 8.5 10.5 1 18" />
        <polyline points="17 6 23 6 23 12" />
      </svg>
      <h1 className="font-bold text-4xl dark:text-white text-black">Rivon</h1>
    </div>
  )
}



export default Logo;

