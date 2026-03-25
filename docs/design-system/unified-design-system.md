# Velune Unified Design System (Web + Mobile)

## Purpose

Velune’s web (Next.js + Tailwind/shadcn) and mobile (Flutter + Material) should feel like one coherent product. This design system defines a shared visual language (calm, trustworthy, premium, finance-focused) and maps it into both platform implementations.

## Brand Feel (product tone)

- Calm, organized like a financial dashboard
- Premium but approachable for daily personal finance use
- High readability for numbers, budgets, and alerts
- Friendly but professional microcopy

## Core Design Principles

- Use semantic tokens (not raw colors) for status: `success`, `warning`, `error`, `info`
- Keep visuals “soft but decisive”: subtle surfaces, subtle elevation, readable contrast
- Avoid dense layouts; prioritize roomy spacing and clear grouping
- “Meaning is not only color”: icons/labels accompany status treatment
- Dark-mode parity: every semantic token must have a light and dark value

## Shared Token Model (conceptual)

Both platforms should implement the same conceptual set of tokens:

- Color tokens
  - `background`, `surface`, `elevatedSurface`
  - `primary`, `secondary`, `muted`
  - `border`, `ring`
- Semantic tokens (with foreground)
  - `success` + `successForeground`
  - `warning` + `warningForeground`
  - `error` + `errorForeground`
  - `info` + `infoForeground`
- Geometry
  - `radiusSoft` (rounded corners for cards/surfaces)
  - `radiusSm` (inputs/badges)
  - `elevationSurface` (cards), `elevationModal` (dialogs/sheets)
- Typography
  - Type scale used by headers, section titles, and summary numbers
- Motion
  - Subtle transitions for loading and state feedback

## Chosen Semantics (from product direction)

- Primary accent family: blue/teal
- Overspend: `warning` visuals with explicit “overspent” / “over limit” labeling

## Web Token Values (Tailwind via CSS variables)

The web theme is implemented through CSS variables in:

- `[frontend/src/app/globals.css](frontend/src/app/globals.css)`
  Tailwind reads these via:
- `[frontend/tailwind.config.ts](frontend/tailwind.config.ts)`

Semantic tokens must be available as:

- `--success`, `--success-foreground`
- `--warning`, `--warning-foreground`
- `--info`, `--info-foreground`
- `--error`, `--error-foreground`

> Note: `destructive` already exists in the current web palette. In the unified system, it maps to `error` semantics for consistency.

## Mobile Token Values (Flutter Material)

Mobile should implement a theme layer in:

- `mobile/velune/lib/core/theme/app_tokens.dart`
- `mobile/velune/lib/core/theme/app_theme.dart`

It must provide:

- Light and dark `ColorScheme` values for primary + semantic statuses
- Radius/elevation language shared across widgets

## Component Vocabulary Mapping (web -> mobile)

These are the primitives that should be reused across both platforms:

- App shell / navigation
  - Web: `frontend/src/components/layout/app-shell.tsx`
  - Mobile: Bottom navigation + scaffold shell
- Page header
  - Web: `frontend/src/components/common/page-header.tsx`
  - Mobile: section header widget in presentation layer
- Summary/stat card
  - Web: `frontend/src/components/common/stat-card.tsx`
  - Mobile: `mobile/velune/lib/presentation/widgets/stat_card.dart`
- Empty/loading/error
  - Web: `frontend/src/components/common/empty-state.tsx`, `loading-skeleton.tsx`
  - Mobile: `EmptyState`, `LoadingSkeleton`, `ErrorBanner`
- Confirm dialog
  - Web: `frontend/src/components/common/confirm-dialog.tsx`
  - Mobile: `ConfirmDialog`
- Status treatment
  - Web: `StatusBadge`, `AlertCard`
  - Mobile: `StatusBadge`, `AlertCard` (warning/error/info/success)
- Overspend visuals
  - Must use `warning` semantics (amber-like) plus explicit overspent labeling

## Finance-Specific UI Rules

- Amounts should always be scannable first (typography + spacing)
- Income vs expense should be visually distinguishable via icon + sign + semantics
- Budget usage must communicate: `limit`, `spent`, `remaining`, and `overspent` status clearly
- Charts must remain compact and legible; labels should not clutter

## Status Semantics Guidelines

- `success`: achieved/paid/within budget (green/teal)
- `warning`: approaching/overspent/near limit (amber)
- `error`: failed operations or hard errors (red/destructive)
- `info`: informational states (blue/teal)

## Accessibility Requirements

- Do not rely on color alone for meaning (pair with labels/icons)
- Ensure adequate contrast in both light and dark mode
- Touch targets and focus states must remain visible and consistent

## Development Notes

When adding new UI:

- Prefer semantic tokens over raw colors
- Reuse component primitives from the vocabulary above
- Keep status visuals consistent across charts/cards/tables
