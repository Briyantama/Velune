import 'package:flutter/material.dart';

import '../../core/theme/app_tokens.dart';

class LoadingSkeleton extends StatefulWidget {
  final double? width;
  final double? height;
  final BorderRadius? borderRadius;

  const LoadingSkeleton({
    super.key,
    this.width,
    this.height,
    this.borderRadius,
  });

  @override
  State<LoadingSkeleton> createState() => _LoadingSkeletonState();
}

class _LoadingSkeletonState extends State<LoadingSkeleton>
    with SingleTickerProviderStateMixin {
  @override
  Widget build(BuildContext context) {
    final radius = widget.borderRadius ?? AppTokens.radiusSmBorder;
    final baseColor = Theme.of(context).colorScheme.onSurface.withOpacity(0.08);

    return AnimatedOpacity(
      duration: const Duration(milliseconds: 900),
      opacity: 0.9,
      child: Container(
        width: widget.width,
        height: widget.height,
        decoration: BoxDecoration(
          color: baseColor,
          borderRadius: radius,
        ),
      ),
    );
  }
}

