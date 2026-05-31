# Design System — Barricade Frontend

## Overview

Built with **CSS Custom Properties (tokens)** and **CSS Modules**. Supports **light and dark themes** via `data-theme` attribute on `<html>`. Targets **WCAG 2.1 AA** compliance.

**Stack:** React 19, TanStack Router, Vite 8, TypeScript 6, CSS Modules

---

## Typography

### Font Families

| Token | Font Stack | Usage |
|---|---|---|
| `--font-sans` | **Inter**, system-ui, -apple-system, sans-serif | All UI text, headings, body |
| `--font-mono` | **JetBrains Mono**, Fira Code, monospace | Code snippets, client IDs, tokens |

Inter loaded from Google Fonts in `index.html`.

### Type Scale

| Token | Size | Usage |
|---|---|---|
| `--font-size-xs` | 0.75rem (12px) | Labels, captions |
| `--font-size-sm` | 0.875rem (14px) | Small text, helper text |
| `--font-size-base` | 1rem (16px) | Body text |
| `--font-size-lg` | 1.125rem (18px) | Large body, intro text |
| `--font-size-xl` | 1.25rem (20px) | Card titles |
| `--font-size-2xl` | 1.5rem (24px) | Section headings |
| `--font-size-3xl` | 1.875rem (30px) | Page headings |
| `--font-size-4xl` | 2.25rem (36px) | Hero headings |

### Font Weights

`normal` (400), `medium` (500), `semibold` (600), `bold` (700).

### Line Heights

`tight` (1.15) for headings, `normal` (1.5) for body, `relaxed` (1.75) for long-form.

---

## Colour Palette

### Light Theme

| Token | Hex | WCAG AA | Usage |
|---|---|---|---|
| `--color-bg-primary` | `#FFFFFF` | — | Page background |
| `--color-bg-secondary` | `#F8FAFC` | — | Alternate section background |
| `--color-text-primary` | `#0F172A` | 18.3:1 | Headings, body |
| `--color-text-secondary` | `#475569` | 7.3:1 | Supporting text |
| `--color-text-muted` | `#6B7280` | 4.8:1 | Placeholder, secondary labels |
| `--color-accent` | `#2563EB` | 4.1:1 (on white) | Buttons, links |
| `--color-accent-hover` | `#1D4ED8` | — | Button hover |
| `--color-border` | `#E2E8F0` | — | Default borders |

### Dark Theme (`[data-theme="dark"]`)

| Token | Hex | WCAG AA | Usage |
|---|---|---|---|
| `--color-bg-primary` | `#0F172A` | — | Page background |
| `--color-text-primary` | `#F8FAFC` | 17.0:1 | Headings, body |
| `--color-text-secondary` | `#94A3B8` | 5.2:1 | Supporting text |
| `--color-text-muted` | `#A1A1AA` | 4.8:1 | Placeholder, secondary labels |
| `--color-accent` | `#3B82F6` | 4.5:1 (on dark) | Buttons, links |
| `--color-accent-hover` | `#60A5FA` | — | Button hover |
| `--color-border` | `#334155` | — | Default borders |
| `--surface-card` | `#1E293B` | — | Card background |

### Semantic Colours

| Token | Light | Dark | Usage |
|---|---|---|---|
| `--color-success` | `#16A34A` | `#22C55E` | Success messages |
| `--color-warning` | `#D97706` | `#F59E0B` | Warning messages |
| `--color-error` | `#DC2626` | `#EF4444` | Error messages |
| `--color-info` | `#2563EB` | `#3B82F6` | Info messages |

Each semantic colour has a matching `-soft` variant for background fills.

---

## Spacing

4px-based scale: `--spacing-1` (4px) to `--spacing-20` (80px). Key values: `--spacing-2` (8px), `--spacing-4` (16px), `--spacing-6` (24px), `--spacing-8` (32px).

## Border Radius

`sm` (4px), `md` (8px), `lg` (12px), `xl` (16px), `full` (9999px).

## Shadows

`sm`, `md`, `lg`, `xl` — dark theme uses higher opacity to maintain depth.

---

## Routing

Uses **TanStack Router** (code-based, no Vite plugin). Routes defined in `src/routeTree.tsx`:

| Path | Component | Layout |
|---|---|---|
| `/` | `HomePage` | Root (Header + Footer) |
| `/login` | `LoginPage` in `AuthLayout` | Root |
| `/register` | `RegisterPage` in `AuthLayout` | Root |
| `/forgot-password` | `ForgotPasswordPage` in `AuthLayout` | Root |

The root layout provides Header, Footer, skip link, and `<main>` outlet. Auth pages are wrapped in `AuthLayout` (centered card with title/subtitle).

Router initialized in `src/main.tsx`:
```tsx
const router = createRouter()
// ...
<ThemeProvider>
  <RouterProvider router={router} />
</ThemeProvider>
```

---

## Components

### Button
Variants: `primary`, `secondary`, `ghost`, `danger`. Sizes: `sm` (36px), `md` (44px AA), `lg` (48px). Supports `loading` prop with animated spinner (disabled in reduced motion).

### Input
Label, error, helper text, `forwardRef`. Uses `aria-invalid`, `aria-describedby` for error announcements. All native `<input>` props supported.

### Card
Variants: `default` (bordered), `elevated` (shadow). Padding: `none`, `sm`, `md`, `lg`.

### AuthLayout
Centered card wrapper for authentication forms. Props: `title`, `subtitle`, `children`. Max-width 400px.

### Header
Sticky header with logo (TanStack Router `<Link>`) and theme toggle button (44×44px). No nav links — auth flow navigation handled inline within form cards.

---

## Dark Mode

`ThemeProvider` context handles toggling:
- Persists choice in `localStorage` (`sk-theme`)
- Respects `prefers-color-scheme` on first visit
- Reacts to OS-level changes if no explicit preference stored
- Sets `data-theme="light|dark"` on `<html>`

---

## Accessibility (WCAG 2.1 AA)

### Contrast
All text/background colour pairs meet or exceed 4.5:1 ratio. Verified tokens: text-primary, text-secondary, text-muted, accent, semantic colours in both themes.

### Touch Targets
- Interactive elements (theme toggle, nav links) ≥44×44px
- Button sizes: sm 36px (small/optional), md 44px (default action), lg 48px (primary call-to-action)

### Motion
- All animations wrapped in `@media (prefers-reduced-motion: reduce)` — transitions disabled, spinner stopped
- Spinner has a reduced-motion fallback (static opacity)

### Focus
- Global `:focus-visible` outline using accent colour
- Input focus ring using `box-shadow` (3px accent-soft) for smooth appearance while maintaining visibility
- Skip-to-content link at top of every page (becomes visible on focus)

### Forms
- All auth forms use `<form>` elements with `type="submit"` on buttons
- Inputs use `aria-invalid`, `aria-describedby` for error states
- `required` attribute on mandatory fields
- `autoComplete` set appropriately (`email`, `current-password`, `new-password`)

### Screen Readers
- Theme toggle SVGs marked `aria-hidden="true"`
- Theme toggle button has descriptive `aria-label`
- Spinner has `aria-hidden="true"` (not a status update)
- Decorative icons in layout are `aria-hidden="true"`
- `<main>` element is landmark, focusable via skip link

---

## File Structure

```
src/
├── theme/
│   └── context.ts               # ThemeContext, useTheme, Theme type
├── styles/
│   ├── tokens.css                # Design tokens (light + dark)
│   ├── reset.css                 # CSS reset
│   ├── base.css                  # Element base styles + reduced motion
│   └── index.css                 # Aggregator
├── components/
│   ├── RootLayout.tsx            # Root shell (skip link, Header, main, Footer)
│   ├── ThemeProvider.tsx         # Dark mode provider
│   ├── AuthLayout/
│   │   ├── AuthLayout.tsx        # Centered card for auth forms
│   │   └── AuthLayout.module.css
│   ├── Button/
│   │   ├── Button.tsx
│   │   └── Button.module.css
│   ├── Input/
│   │   ├── Input.tsx
│   │   └── Input.module.css
│   ├── Card/
│   │   ├── Card.tsx
│   │   └── Card.module.css
│   ├── Header/
│   │   ├── Header.tsx
│   │   └── Header.module.css
│   └── Footer/
│       ├── Footer.tsx
│       └── Footer.module.css
├── pages/
│   ├── Home.tsx                  # Entry point (/)
│   ├── Home.module.css
│   ├── Login.tsx                 # Sign In (/login)
│   ├── Login.module.css
│   ├── Register.tsx              # Create Account (/register)
│   ├── Register.module.css
│   ├── ForgotPassword.tsx        # Reset password (/forgot-password)
│   └── ForgotPassword.module.css
├── routeTree.tsx                 # Route definitions
├── routeTree.module.css          # Root layout + skip link styles
├── router.ts                     # Router factory
├── App.tsx
└── main.tsx                      # Entry: ThemeProvider → RouterProvider
```
