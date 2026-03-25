/// Generic pagination wrapper used by list endpoints.
class Paged<T> {
  final List<T> items;
  final int total;
  final int page;
  final int limit;

  const Paged({
    required this.items,
    required this.total,
    required this.page,
    required this.limit,
  });

  factory Paged.fromJson(
    Map<String, dynamic> json, {
    required T Function(Map<String, dynamic>) itemFromJson,
  }) {
    final itemsJson = (json['items'] as List?) ?? const [];
    return Paged<T>(
      items: itemsJson
          .map((e) => itemFromJson((e as Map).cast<String, dynamic>()))
          .toList(),
      total: (json['total'] as num?)?.toInt() ?? 0,
      page: (json['page'] as num?)?.toInt() ?? 1,
      limit: (json['limit'] as num?)?.toInt() ?? 0,
    );
  }
}

