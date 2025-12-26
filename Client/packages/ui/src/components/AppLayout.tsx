"use client"

import { useRef, useState, useEffect } from "react";
import Header from "@workspace/ui/components/Header/Header";
import { ScrollArea } from "@workspace/ui/components/scroll-area";
import { useScrollDirection } from "@workspace/ui/hooks/useScrollDirection";

interface AppLayoutProps {
    children: React.ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
    const viewportRef = useRef<HTMLDivElement>(null);
    const scrollDirection = useScrollDirection(viewportRef);
    const [isScrolled, setIsScrolled] = useState(false);

    useEffect(() => {
        const viewport = viewportRef.current;
        if (!viewport) return;

        const handleScroll = () => {
            setIsScrolled(viewport.scrollTop > 10);
        };

        viewport.addEventListener("scroll", handleScroll);
        return () => viewport.removeEventListener("scroll", handleScroll);
    }, []);

    return (
        <div className="h-screen w-full flex flex-col overflow-hidden">
            <Header scrollDirection={scrollDirection} isScrolled={isScrolled} />
            <ScrollArea className="flex-1" viewportRef={viewportRef}>
                <div className="pt-16 min-h-full w-full">
                    {children}
                </div>
            </ScrollArea>
        </div>
    );
}
