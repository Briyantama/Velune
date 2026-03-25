# Cross-Platform Component Mapping (Web <-> Mobile)

This mapping guide keeps both platforms feeling like the same product. Prefer these primitives over introducing new one-off styling.

## Navigation / Shell

- Web `frontend/src/components/layout/app-shell.tsx` -> Mobile: app scaffold with bottom navigation + persistent navigation shell
- Page header Web `frontend/src/components/common/page-header.tsx` -> Mobile: section header at the top of each screen

## Cards / Summary

- Web `frontend/src/components/common/stat-card.tsx` -> Mobile `mobile/velune/lib/presentation/widgets/stat_card.dart`
- Web `frontend/src/components/common/empty-state.tsx` -> Mobile `EmptyState`
- Web `frontend/src/components/common/loading-skeleton.tsx` -> Mobile `LoadingSkeleton`
- Web `frontend/src/components/common/confirm-dialog.tsx` -> Mobile `ConfirmDialog`

## Finance status + warnings

- Overspend / budget warning badge Web: `StatusBadge variant="warning"` -> Mobile: `StatusBadge(StatusBadgeVariant.warning)`
- Success states (paid/within budget) Web: `StatusBadge variant="success"` -> Mobile: `StatusBadge(StatusBadgeVariant.success)`
- Error/info states Web: `StatusBadge variant="error"|"info"` -> Mobile: `StatusBadge(StatusBadgeVariant.error|"info")`

## Overspend/Alerts

- Web `AlertCard` -> Mobile `AlertCard`
- Use `AlertCard` for important, user-actionable conditions (for example: “overspent”, “service unavailable”, “retry recommended”).

## Filters & inputs

- Web `frontend/src/components/common/filter-bar.tsx` -> Mobile: filter section widgets placed near the top of the list
- Web form fields: use the same layout and spacing as existing shadcn fields -> Mobile: labeled `TextFormField`/input widgets with inline validation

## Charts

- Web: Recharts chart cards should use the same primary accent color -> Mobile: chart surfaces should use theme primary color and compact legends

## Rule of thumb

- If a UI element is repeated across multiple screens (card headers, status labels, empty/error blocks), it belongs in this design system vocabulary.
