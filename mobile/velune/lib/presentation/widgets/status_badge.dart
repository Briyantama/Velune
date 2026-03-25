import 'package:flutter/material.dart';

import '../../core/theme/app_theme.dart';
import '../../core/theme/app_tokens.dart';

enum StatusBadgeVariant { neutral, success, warning, error, info }

class StatusBadge extends StatelessWidget {
  final StatusBadgeVariant variant;
  final String label;

  const StatusBadge({
    super.key,
    required this.variant,
    required this.label,
  });

  @override
  Widget build(BuildContext context) {
    final (bg, fg, border) = _colors();

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: bg,
        border: Border.all(color: border),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: Theme.of(context).textTheme.labelSmall?.copyWith(
              color: fg,
              fontWeight: FontWeight.w700,
            ),
      ),
    );
  }

  (Color bg, Color fg, Color border) _colors() {
    switch (variant) {
      case StatusBadgeVariant.success:
        return (
          AppStatusColors.success.withOpacity(0.12),
          AppStatusColors.successForeground,
          AppStatusColors.success.withOpacity(0.35),
        );
      case StatusBadgeVariant.warning:
        return (
          AppStatusColors.warning.withOpacity(0.12),
          AppStatusColors.warningForeground,
          AppStatusColors.warning.withOpacity(0.35),
        );
      case StatusBadgeVariant.error:
        return (
          AppStatusColors.error.withOpacity(0.12),
          AppStatusColors.errorForeground,
          AppStatusColors.error.withOpacity(0.35),
        );
      case StatusBadgeVariant.info:
        return (
          AppStatusColors.info.withOpacity(0.12),
          AppStatusColors.infoForeground,
          AppStatusColors.info.withOpacity(0.35),
        );
      case StatusBadgeVariant.neutral:
      default:
        final borderColor = Theme.of(context).colorScheme.outlineVariant;
        final bg = Theme.of(context)
            .colorScheme
            .surfaceContainerHighest
            .withOpacity(0.65);
        final fg = Theme.of(context).colorScheme.onSurface;
        return (bg, fg, borderColor);
    }
  }
}

