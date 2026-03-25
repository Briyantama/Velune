import 'package:flutter/material.dart';

import '../../core/theme/app_theme.dart';

enum CategoryChipTone { neutral, primary }

class CategoryChip extends StatelessWidget {
  final String label;
  final CategoryChipTone tone;

  const CategoryChip({
    super.key,
    required this.label,
    this.tone = CategoryChipTone.neutral,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final bg = tone == CategoryChipTone.primary
        ? AppStatusColors.primary.withOpacity(0.12)
        : colorScheme.surfaceContainerHighest.withOpacity(0.6);
    final border = tone == CategoryChipTone.primary
        ? AppStatusColors.primary.withOpacity(0.35)
        : colorScheme.outlineVariant;
    final fg =
        tone == CategoryChipTone.primary ? AppStatusColors.primary : colorScheme.onSurface;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: bg,
        border: Border.all(color: border),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: Theme.of(context).textTheme.labelSmall?.copyWith(
              color: fg,
              fontWeight: FontWeight.w600,
            ),
      ),
    );
  }
}

