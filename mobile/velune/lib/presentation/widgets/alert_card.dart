import 'package:flutter/material.dart';

import '../../core/theme/app_theme.dart';

enum AlertCardTone { success, warning, error, info }

class AlertCard extends StatelessWidget {
  final AlertCardTone tone;
  final String title;
  final String? message;
  final Widget? actions;

  const AlertCard({
    super.key,
    required this.tone,
    required this.title,
    this.message,
    this.actions,
  });

  @override
  Widget build(BuildContext context) {
    final (bg, border, fg) = _colors();

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: Theme.of(context).textTheme.labelLarge?.copyWith(
                  color: fg,
                  fontWeight: FontWeight.w800,
                ),
          ),
          if (message != null) ...[
            const SizedBox(height: 6),
            Text(
              message!,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: fg,
                  ),
            ),
          ],
          if (actions != null) ...[
            const SizedBox(height: 10),
            actions!,
          ],
        ],
      ),
    );
  }

  (Color bg, Color border, Color fg) _colors() {
    switch (tone) {
      case AlertCardTone.success:
        return (
          AppStatusColors.success.withOpacity(0.12),
          AppStatusColors.success.withOpacity(0.35),
          AppStatusColors.successForeground,
        );
      case AlertCardTone.warning:
        return (
          AppStatusColors.warning.withOpacity(0.12),
          AppStatusColors.warning.withOpacity(0.35),
          AppStatusColors.warningForeground,
        );
      case AlertCardTone.error:
        return (
          AppStatusColors.error.withOpacity(0.12),
          AppStatusColors.error.withOpacity(0.35),
          AppStatusColors.errorForeground,
        );
      case AlertCardTone.info:
      default:
        return (
          AppStatusColors.info.withOpacity(0.12),
          AppStatusColors.info.withOpacity(0.35),
          AppStatusColors.infoForeground,
        );
    }
  }
}

