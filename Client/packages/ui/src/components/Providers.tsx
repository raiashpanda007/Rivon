"use client"

import * as React from "react"
import { ThemeProvider as NextThemesProvider, useTheme } from "next-themes"
import { StoreProvider } from "@workspace/store"

const THEME_COOKIE = "rivon_theme"
const THEME_LS_KEY = "theme"

// Sync shared cookie → localStorage so NextThemesProvider picks up the right theme on mount
function syncCookieToLocalStorage() {
    if (typeof window === "undefined") return
    const match = document.cookie.match(/(?:^|;\s*)rivon_theme=([^;]*)/)
    const cookieTheme = match ? decodeURIComponent(match[1]!) : null
    if (cookieTheme) {
        localStorage.setItem(THEME_LS_KEY, cookieTheme)
    }
}

// Writes the current theme to a cookie so other apps can read it on mount
function ThemeCookieSync() {
    const { theme } = useTheme()

    React.useEffect(() => {
        if (!theme) return
        const maxAge = 365 * 24 * 60 * 60
        document.cookie = `${THEME_COOKIE}=${encodeURIComponent(theme)};max-age=${maxAge};path=/;SameSite=Lax`
    }, [theme])

    return null
}

export function Providers({ children }: { children: React.ReactNode }) {
    // Run synchronously on first client render so NextThemesProvider reads the shared theme
    syncCookieToLocalStorage()

    return (
        <StoreProvider>
            <NextThemesProvider
                attribute="class"
                defaultTheme="dark"
                enableSystem={false}
                disableTransitionOnChange
                enableColorScheme
            >
                <ThemeCookieSync />
                {children}
            </NextThemesProvider>
        </StoreProvider>
    )
}
