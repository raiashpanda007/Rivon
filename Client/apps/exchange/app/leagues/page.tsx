import LeagueSelection from "@/components/Dashboard/LeagueSelection"
import Image from "next/image"
import icon from "../icon.svg"

export default function Page() {
    return (
        <div className="relative flex min-h-[calc(100vh-3.5rem)] w-full flex-col items-center pt-20 overflow-hidden">
            {/* Ambient Glows */}
            <div className="absolute top-[20%] right-[20%] -z-10 w-[400px] h-[400px] rounded-full bg-orange-500/10 blur-[120px] pointer-events-none" />
            <div className="absolute bottom-[10%] left-[20%] -z-10 w-[400px] h-[400px] rounded-full bg-blue-500/5 blur-[120px] pointer-events-none" />
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
