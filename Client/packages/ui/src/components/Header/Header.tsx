"use client"

import Logo from "@workspace/ui/components/Logo/Logo";
import AuthOptions from "@workspace/ui/components/Header/authOptions/Authoptions";
import Options from "@workspace/ui/components/Header/Options";
import SearchBar from "@workspace/ui/components/Header/searchBar/SearchBar";
import { cn } from "@workspace/ui/lib/utils";
import { useState, useEffect } from "react";
import { ThemeToggle } from "@workspace/ui/components/ThemeToggle";

interface HeaderProps {
  scrollDirection?: "up" | "down" | null;
  isScrolled?: boolean;
  currentApp?: "trade" | "betting";
}

function Header({ isScrolled: externalIsScrolled, currentApp }: HeaderProps) {
  const [internalIsScrolled, setInternalIsScrolled] = useState(false);
  const isScrolled = externalIsScrolled !== undefined ? externalIsScrolled : internalIsScrolled;

  useEffect(() => {
    if (externalIsScrolled !== undefined) return;
    const handleScroll = () => setInternalIsScrolled(window.scrollY > 10);
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, [externalIsScrolled]);

  return (
    <header
      className={cn(
        "fixed top-0 left-0 right-0 z-50 w-full h-14 flex items-center transition-all duration-200",
        "bg-background/95 border-b border-border backdrop-blur-md",
        isScrolled && "shadow-[0_1px_0_0_rgba(249,115,22,0.1)]"
      )}
    >
      {/* Orange left accent */}
      <div className="absolute left-0 top-0 bottom-0 w-[2px] bg-orange-500/70" />

      <div className="flex items-center gap-6 w-full px-5">
        <Logo className="shrink-0" />

        <div className="hidden md:flex w-full items-center gap-4 flex-1">
          <Options currentApp={currentApp} />
          <SearchBar />
        </div>

        <div className="flex items-center gap-3 shrink-0 ml-auto">
          <div className="hidden sm:flex items-center gap-1.5 px-2 py-0.5 border border-border rounded-sm bg-muted/30">
            <span className="dot-live" />
            <span className="text-[10px] font-mono text-muted-foreground">LIVE</span>
          </div>
          <div className="h-4 w-px bg-border" />
          <AuthOptions />
          <div className="h-4 w-px bg-border" />
          <ThemeToggle />
        </div>
      </div>
    </header>
  )
}

export default Header;
