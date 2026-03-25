/// Normalizes inconsistent backend JSON key casing into a consistent lowerCamel form.
///
/// Backend variations seen across services:
/// - snake_case: `budget_id` -> `budgetId`
/// - PascalCase: `MonthlyReport` -> `monthlyReport`
/// - special suffix rule: `*ID` -> `*Id` (e.g. `AccountID` -> `accountId`)
class NormalizeJsonKeys {
  const NormalizeJsonKeys();

  Map<String, dynamic> normalizeJsonKeys(Map<String, dynamic> input) {
    final result = <String, dynamic>{};
    for (final entry in input.entries) {
      final normalizedKey = _normalizeKey(entry.key);
      result[normalizedKey] = _normalizeValue(entry.value);
    }
    return result;
  }

  dynamic _normalizeValue(dynamic value) {
    if (value is Map) {
      final asMap = value.map((key, v) => MapEntry(key.toString(), v));
      return normalizeJsonKeys(asMap.cast<String, dynamic>());
    }

    if (value is List) {
      return value.map(_normalizeValue).toList();
    }

    return value;
  }

  String _normalizeKey(String key) {
    if (key.contains('_')) return _normalizeSnakeCase(key);

    // PascalCase/camelCase: convert the `*ID` suffix to `*Id`.
    // This covers cases like `UserID`, `AccountID`, and `CorrelationID`.
    final withIdRule = key.replaceAllMapped(
      RegExp(r'ID(?=$|[A-Z])'),
      (_) => 'Id',
    );

    return _lowerFirst(withIdRule);
  }

  String _normalizeSnakeCase(String key) {
    final parts = key.split('_').where((p) => p.isNotEmpty).toList();
    if (parts.isEmpty) return key;

    final first = parts.first.toLowerCase();
    final rest = parts.skip(1).map((p) {
      final lower = p.toLowerCase();
      if (lower.isEmpty) return '';
      return '${lower[0].toUpperCase()}${lower.substring(1)}';
    }).join();

    return '$first$rest';
  }

  String _lowerFirst(String s) {
    if (s.isEmpty) return s;
    return s[0].toLowerCase() + s.substring(1);
  }
}

Map<String, dynamic> normalizeJsonKeys(Map<String, dynamic> input) =>
    NormalizeJsonKeys().normalizeJsonKeys(input);

