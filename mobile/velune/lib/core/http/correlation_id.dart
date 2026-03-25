import 'dart:math';

String ensureCorrelationId(String? incoming) {
  final trimmed = incoming?.trim();
  if (trimmed != null && trimmed.isNotEmpty) return trimmed;

  final nowMicros = DateTime.now().microsecondsSinceEpoch;
  final rand = Random();
  final suffix = rand.nextInt(1 << 20); // Enough uniqueness for client-side traces.
  return 'velune-$nowMicros-$suffix';
}

