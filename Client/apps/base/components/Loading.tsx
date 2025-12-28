"use client"
import React from "react";
import { motion } from "framer-motion";
import Logo from "@workspace/ui/components/Logo/Logo";

interface LoadingProps {
    heading?: string;
    message?: string;
}

export default function Loading({
    heading = "Loading...",
    message = "Please wait while we prepare everything for you."
}: LoadingProps) {
    return (
        <div className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-white/80 dark:bg-black/80 backdrop-blur-md">
            <motion.div
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ duration: 0.5 }}
                className="flex flex-col items-center space-y-8"
            >
                <motion.div
                    animate={{
                        scale: [1, 1.1, 1],
                        filter: ["brightness(1)", "brightness(1.2)", "brightness(1)"]
                    }}
                    transition={{
                        duration: 2,
                        repeat: Infinity,
                        ease: "easeInOut"
                    }}
                    className="relative"
                >
                    {/* Add a glow effect behind the logo */}
                    <div className="absolute inset-0 bg-orange-500/20 blur-xl rounded-full scale-150 animate-pulse" />
                    <Logo className="relative z-10 scale-150" />
                </motion.div>

                <div className="text-center space-y-3 px-4">
                    <motion.h2
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: 0.2, duration: 0.5 }}
                        className="text-3xl font-bold tracking-tight text-foreground"
                    >
                        {heading}
                    </motion.h2>
                    <motion.p
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: 0.4, duration: 0.5 }}
                        className="text-muted-foreground text-base max-w-md mx-auto font-medium opacity-80"
                    >
                        {message}
                    </motion.p>
                </div>
            </motion.div>
        </div>
    );
}
