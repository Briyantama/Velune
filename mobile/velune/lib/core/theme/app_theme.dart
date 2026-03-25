import 'package:flutter/material.dart';

import 'app_tokens.dart';

class AppStatusColors {
  static const Color primary = Color(0xFF14B8A6); // blue/teal

  static const Color success = Color(0xFF10B981);
  static const Color successForeground = Color(0xFF052E1A);

  static const Color warning = Color(0xFFF59E0B);
  static const Color warningForeground = Color(0xFF2A1B00);

  static const Color error = Color(0xFFEF4444);
  static const Color errorForeground = Color(0xFFFFF1F1);

  static const Color info = Color(0xFF38BDF8);
  static const Color infoForeground = Color(0xFF06263A);

  static const Color onPrimary = Color(0xFFECFFFB);
  static const Color onPrimaryDark = Color(0xFF0B1D1A);
}

class AppTheme {
  static ThemeData lightTheme() {
    const background = Color(0xFFFAFAFF);
    const surface = Color(0xFFFFFFFF);

    final cs = ColorScheme.fromSeed(
      seedColor: AppStatusColors.primary,
      brightness: Brightness.light,
    ).copyWith(
      primary: AppStatusColors.primary,
      secondary: const Color(0xFFEFF6FF),
      error: AppStatusColors.error,
      background: background,
      surface: surface,
    );

    return ThemeData(
      useMaterial3: true,
      colorScheme: cs,
      scaffoldBackgroundColor: background,
      cardTheme: CardTheme(
        shape: RoundedRectangleBorder(borderRadius: AppTokens.radiusSmBorder),
        elevation: 0,
      ),
      appBarTheme: const AppBarTheme(centerTitle: false),
    );
  }

  static ThemeData darkTheme() {
    const background = Color(0xFF0D0F1A);
    const surface = Color(0xFF12162A);

    final cs = ColorScheme.fromSeed(
      seedColor: AppStatusColors.primary,
      brightness: Brightness.dark,
    ).copyWith(
      primary: AppStatusColors.primary,
      secondary: const Color(0xFF1D2B4A),
      error: AppStatusColors.error,
      background: background,
      surface: surface,
    );

    return ThemeData(
      useMaterial3: true,
      colorScheme: cs,
      scaffoldBackgroundColor: background,
      cardTheme: CardTheme(
        shape: RoundedRectangleBorder(borderRadius: AppTokens.radiusSmBorder),
        elevation: 0,
      ),
      appBarTheme: const AppBarTheme(centerTitle: false),
    );
  }
}

