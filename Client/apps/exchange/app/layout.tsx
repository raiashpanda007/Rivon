import { Geist, Geist_Mono } from "next/font/google"

import "@workspace/ui/globals.css"

import { Providers } from "@/components/providers"

const fontSans = Geist({
  subsets: ["latin"],
  variable: "--font-sans",
})

const fontMono = Geist_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
})

export const metadata = {
  title: "Trade - Rivon",
  description: "Rivon Exchange Application",
}

import { AppLayout } from "@workspace/ui/components/AppLayout"
import { Toaster } from "@workspace/ui/components/sonner"

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${fontSans.variable} ${fontMono.variable} font-mono antialiased`}
        suppressHydrationWarning
      >
        <Providers>
          <AppLayout currentApp="trade">
            {children}
            <Toaster richColors position="top-right" />
          </AppLayout>
        </Providers>
      </body>
    </html>
  )
}
