import LeagueSelection from "@/components/Dashboard/LeagueSelection"
import Image from "next/image"
import icon from "../icon.svg"

export default function Page() {
    return (
        <div className="relative flex min-h-screen w-full flex-col items-center pt-20 overflow-hidden bg-background">
            <div className="absolute inset-0 z-0 flex items-center justify-center opacity-5 pointer-events-none">
                <Image
                    src={icon}
                    alt="Background Icon"
                    className="h-[600px] w-[600px] object-contain"
                    priority
                />
            </div>
            <div className="relative z-10 w-full max-w-5xl px-4">
                <LeagueSelection />
            </div>
        </div>
    )
}
