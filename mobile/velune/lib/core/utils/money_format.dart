/// Formats fixed-point minor-unit amounts into a human-readable decimal string.
///
/// Examples (fractionDigits=2):
/// - 12345 -> "123.45"
/// - -123  -> "-1.23"
String formatMoneyMinor(
  int amountMinor, {
  int fractionDigits = 2,
  String decimalSeparator = '.',
}) {
  if (fractionDigits < 0) {
    throw ArgumentError.value(fractionDigits, 'fractionDigits', 'Must be >= 0');
  }

  final sign = amountMinor < 0 ? '-' : '';
  final abs = amountMinor.abs();

  var factor = 1;
  for (var i = 0; i < fractionDigits; i++) {
    factor *= 10;
  }

  final major = abs ~/ factor;
  final minor = abs % factor;

  if (fractionDigits == 0) return '$sign$major';

  final minorStr = minor.toString().padLeft(fractionDigits, '0');
  return '$sign$major$decimalSeparator$minorStr';
}

