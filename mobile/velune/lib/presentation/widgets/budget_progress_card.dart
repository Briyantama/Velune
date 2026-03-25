import 'package:flutter/material.dart';

import '../../core/theme/app_tokens.dart';
import '../../core/theme/app_theme.dart';
import 'status_badge.dart';

class BudgetProgressCard extends StatelessWidget {
  final String title;
  final String spentText;
  final String remainingText;
  final double progress; // 0..1
  final bool isOverspent;

  const BudgetProgressCard({
    super.key,
    required this.title,
    required this.spentText,
    required this.remainingText,
    required this.progress,
    required this.isOverspent,
  });

  @override
  Widget build(BuildContext context) {
    final pct = progress.clamp(0.0, 1.0);
    final isOver = isOverspent;
    final color =
        isOver ? AppStatusColors.warning : Theme.of(context).colorScheme.primary;

    return Card(
      elevation: 0,
      shape:
          RoundedRectangleBorder(borderRadius: AppTokens.radiusSoftBorder),
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                Expanded(
                  child: Text(
                    title,
                    style: Theme.of(context)
                        .textTheme
                        .labelLarge
                        ?.copyWith(fontWeight: FontWeight.w700),
                  ),
                ),
                if (isOver)
                  const StatusBadge(
                    variant: StatusBadgeVariant.warning,
                    label: 'Overspent',
                  ),
              ],
            ),
            const SizedBox(height: 10),
            LinearProgressIndicator(
              value: pct,
              color: color,
              backgroundColor: color.withOpacity(0.14),
              minHeight: 8,
              borderRadius: BorderRadius.circular(999),
            ),
            const SizedBox(height: 10),
            Row(
              children: [
                Expanded(
                  child: Text(
                    'Spent: $spentText',
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),
                ),
                Expanded(
                  child: Text(
                    'Remaining: $remainingText',
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

