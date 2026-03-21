export default function AuthLayout({
    children,
}: {
    children: React.ReactNode
}) {
    return (
        <div className="relative flex min-h-screen w-full items-center justify-center overflow-hidden bg-background bg-terminal-grid">
            {/* Radial orange glow */}
            <div className="absolute inset-0 pointer-events-none">
                <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] rounded-full bg-orange-500/5 blur-[120px]" />
            </div>

            {/* Corner decorations */}
            <div className="absolute top-0 left-0 w-16 h-16 border-t-2 border-l-2 border-orange-500/20" />
            <div className="absolute top-0 right-0 w-16 h-16 border-t-2 border-r-2 border-orange-500/20" />
            <div className="absolute bottom-0 left-0 w-16 h-16 border-b-2 border-l-2 border-orange-500/20" />
            <div className="absolute bottom-0 right-0 w-16 h-16 border-b-2 border-r-2 border-orange-500/20" />

            {/* System label */}
            <div className="absolute top-6 left-1/2 -translate-x-1/2">
                <span className="font-mono text-[9px] text-muted-foreground/40 tracking-widest">
                    RIVON · AUTH_TERMINAL
                </span>
            </div>

            <div className="relative z-10 w-full max-w-md px-4">{children}</div>
        </div>
    )
}
