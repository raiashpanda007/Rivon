import Image from "next/image"
import icon from "../icon.svg"

export default function AuthLayout({
    children,
}: {
    children: React.ReactNode
}) {
    return (
        <div className="relative flex min-h-screen w-full items-center justify-center overflow-hidden bg-background">
            <div className="absolute inset-0 z-0 flex items-center justify-center opacity-5">
                <Image
                    src={icon}
                    alt="Background Icon"
                    className="h-[500px] w-[500px]"
                    priority
                />
            </div>
            <div className="relative z-10 w-full max-w-md p-4">{children}</div>
        </div>
    )
}
