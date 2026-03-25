import 'normalize_json_keys.dart';

/// Repository boundary helper: normalize keys before passing JSON into `fromJson`.
T fromJsonNormalized<T>(
  dynamic json, {
  required T Function(Map<String, dynamic> normalized) fromJson,
}) {
  if (json is Map) {
    final map = json.cast<String, dynamic>();
    final normalized = normalizeJsonKeys(map);
    return fromJson(normalized);
  }

  throw ArgumentError.value(json, 'json', 'Expected a JSON object (Map)');
}

