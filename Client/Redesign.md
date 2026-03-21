Redesign the entire client-side UI of my app (Rivon), a trading + betting exchange platform.

Context:
The app is already built using React, Tailwind, and Framer Motion. It currently feels like a generic SaaS UI with card-based layouts. I want to transform the ENTIRE app (not just the landing page) into a cohesive, product-level trading system UI.

Core Direction:
- This is NOT an AI tool or simple dashboard
- It should feel like a real-time trading terminal
- Inspired by TradingView, Linear, and modern developer tools
- Dense, alive, and slightly intense — not minimal SaaS

Global Design System:
- Replace card-heavy UI with panel-based layouts (glass panels, layered components)
- Use a dark theme with Tailwind orange-500 as the primary accent (brand consistency)
- Introduce a "neo-retro terminal" aesthetic:
  - subtle grid backgrounds
  - soft glow effects (especially orange accents)
  - depth using blur, shadows, and layering
  - minimal retro influence (no pixel fonts for main UI)
- Maintain strong readability and hierarchy

Layout (Across All Pages):
- Avoid centered SaaS layouts
- Use structured, multi-panel layouts where relevant:
  - left navigation / markets
  - central main content (charts, data)
  - right panels (activity, order book, bets)
- Use asymmetry and dynamic layouts instead of static sections
- Ensure layout consistency across pages

Landing Page:
- Keep the existing intro animation EXACTLY as it is
- Persist intro animation using sessionStorage (run once per session)
- Replace static sections (features, leagues, etc.) with “live system previews”:
  - scrolling market ticker
  - mini order book
  - activity feed
  - enhanced live market preview
- Make landing feel like a preview of the product, not a marketing page

Core App Pages (Important):
Apply the new design system consistently across:
- Dashboard / Home
- Market / Trading pages
- Betting interface
- Portfolio / holdings
- Any existing pages

For these pages:
- Focus on information density and clarity
- Highlight real-time changes (prices, odds, trades)
- Add micro-interactions (hover, updates, highlights)
- Prioritize speed and usability over decoration

Component Refactor:
- Refactor existing components (FeatureCard, LeagueCard, etc.)
- Remove “rounded card + border” feel
- Convert into terminal-style panels
- Introduce reusable layout primitives (panel, section, header, data row)

Animations:
- Use Framer Motion for transitions and interactions
- Add micro-interactions:
  - hover glow
  - price updates (green/red flashes)
  - subtle pulses or flickers for live data
- Keep animations fast and responsive

Background & Identity:
- Enhance backgrounds globally:
  - grid + radial glow + subtle noise
- Use orange-500 as accent with glow (not flat usage everywhere)
- Maintain a consistent visual language across all pages

Tech Constraints:
- Keep all existing business logic intact
- Use React + Tailwind best practices
- Component-driven architecture
- Avoid unnecessary re-renders

Output:
- Updated global layout system
- Refactored components
- New UI elements (ticker, activity feed, etc.)
- Code changes across relevant pages (not just landing)

Do NOT:
- Do not limit redesign to the landing page
- Do not keep card-based layouts
- Do not make small cosmetic changes only
- Do not reduce readability for style
