"use client"

import * as React from "react"
import { ThemeProvider as NextThemesProvider } from "next-themes"

import { StoreProvider } from "@workspace/store";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <StoreProvider>
      <NextThemesProvider
        attribute="class"
        defaultTheme="system"
        enableSystem
        disableTransitionOnChange
        enableColorScheme
      >
        {children}
      </NextThemesProvider>
    </StoreProvider>
  )
}
