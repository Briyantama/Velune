# Velune Design System QA Checklist (Web + Mobile)

## Prereqs

- Web: confirm `[frontend/src/app/globals.css](frontend/src/app/globals.css)` is loaded and dark-mode is enabled via `class="dark"`.
- Mobile: confirm `mobile/velune/lib/main.dart` uses `AppTheme.lightTheme()` and `AppTheme.darkTheme()` with `themeMode` wired.

## 1. Visual parity (light + dark)

Run through both web and mobile in light and dark:

- Page background, surfaces, and card borders look consistent.
- Primary action buttons and focus rings use the blue/teal primary token.
- Semantic status colors match meaning:
  - success: green/teal calm tone
  - warning: amber calm tone (overspend)
  - error: red/destructive tone
  - info: blue/teal informational tone

## 2. Overspend semantics (must be `warning`)

- Dashboard budget usage preview:
  - progress bar turns `warning` when `isOverspent=true`
  - badge label reads “Overspent”
- Budgets usage drawer/panel:
  - shows “Overspent” vs “On track” badge using `warning` semantics

## 3. State consistency

For every major screen (Dashboard, Transactions, Budgets, Reports, Notifications, Settings):

- Loading state uses the shared skeleton look.
- Empty state uses the shared soft-card look.
- Error state uses the shared readable error banner/card.
- Buttons follow the same hierarchy (primary vs secondary vs destructive).

## 4. Charts / data visualization

- Charts never rely on color alone:
  - labels/tooltips remain readable in dark mode
- Bar/line surfaces use the primary accent.
- Category breakdown labels do not overlap and remain legible at small widths.

## 5. Accessibility checks

- Ensure contrast between:
  - status text and status backgrounds (success/warning/error/info)
  - primary button text and primary background in both light and dark
- Do not rely on color alone:
  - overspend uses both label + badge/icon/shape
- Touch targets (mobile):
  - min comfortable tap targets for chips, filters, and destructive actions

## 6. Optional automated checks (if CI is set up)

- Web: unit tests for shared primitives rendering classNames for each status variant.
- Mobile: widget tests for `StatusBadge`, `AlertCard`, and `BudgetProgressCard` overspent vs on-track.
