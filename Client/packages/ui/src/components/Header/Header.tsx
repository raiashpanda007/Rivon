"use client"

import Logo from "@workspace/ui/components/Logo/Logo";
import AuthOptions from "@workspace/ui/components/Header/authOptions/Authoptions";
import Options from "@workspace/ui/components/Header/Options";
import SearchBar from "@workspace/ui/components/Header/searchBar/SearchBar";
import { useScrollDirection } from "@workspace/ui/hooks/useScrollDirection";
import { cn } from "@workspace/ui/lib/utils";
import { useState, useEffect } from "react";
import { ThemeToggle } from "@workspace/ui/components/ThemeToggle";

interface HeaderProps {
  scrollDirection?: "up" | "down" | null;
  isScrolled?: boolean;
  currentApp?: "trade" | "betting";
}

function Header({ scrollDirection: externalScrollDirection, isScrolled: externalIsScrolled, currentApp }: HeaderProps) {
  const internalScrollDirection = useScrollDirection();
  const [internalIsScrolled, setInternalIsScrolled] = useState(false);

  // Use external props if provided, otherwise fallback to internal logic (window scroll)
  const scrollDirection = externalScrollDirection !== undefined ? externalScrollDirection : internalScrollDirection;
  const isScrolled = externalIsScrolled !== undefined ? externalIsScrolled : internalIsScrolled;

  useEffect(() => {
    if (externalIsScrolled !== undefined) return; // Skip if controlled externally

    const handleScroll = () => {
      setInternalIsScrolled(window.scrollY > 10);
    };
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, [externalIsScrolled]);

  return (
    <>
      <header
        className={cn(
          "fixed top-0 left-0 right-0 z-50 w-full h-16 flex items-center justify-between px-6 transition-all duration-300 ease-in-out",
          "translate-y-0",
          isScrolled
            ? "bg-background/80 backdrop-blur-xl border-b border-border/40 shadow-sm supports-[backdrop-filter]:bg-background/60"
            : "bg-transparent border-b border-transparent"
        )}
      >
        <div className="flex items-center gap-8 w-full max-w-7xl mx-auto">
          <Logo className="h-10 w-auto shrink-0" />

          <div className="hidden md:flex items-center gap-6 flex-1 justify-center">
            <Options currentApp={currentApp} />
            <SearchBar />
          </div>

          <div className="flex items-center gap-4 shrink-0">
            <AuthOptions />
            <div className="h-6 w-px bg-border/50 mx-2" />
            <ThemeToggle />
          </div>
        </div>
      </header>
    </>
  )
}

export default Header;
