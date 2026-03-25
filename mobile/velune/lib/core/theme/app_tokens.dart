import 'package:flutter/material.dart';

/// Shared visual tokens for the Velune mobile UI.
///
/// These constants are meant to keep the product visually consistent,
/// regardless of which feature screen is currently being displayed.
class AppTokens {
  // Geometry
  static const double radiusSoft = 16;
  static const double radiusMd = 14;
  static const double radiusSm = 12;

  static const BorderRadius radiusSoftBorder =
      BorderRadius.all(Radius.circular(radiusSoft));
  static const BorderRadius radiusSmBorder =
      BorderRadius.all(Radius.circular(radiusSm));

  // Elevation (Material3 provides most shadows automatically, but we keep
  // explicit values for any custom elevated surfaces we add later).
  static const double elevationSurface = 0;
  static const double elevationModal = 6;

  // Spacing
  static const double space2 = 8;
  static const double space3 = 12;
  static const double space4 = 16;
  static const double space5 = 20;
}

