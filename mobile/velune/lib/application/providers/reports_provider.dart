import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../domain/models/monthly_report.dart';
import '../providers/core_providers.dart';

@immutable
class MonthlyReportQuery {
  final int year;
  final int month;
  final String currency;

  const MonthlyReportQuery({
    required this.year,
    required this.month,
    required this.currency,
  });
}

final monthlyReportProvider = FutureProvider.autoDispose
    .family<MonthlyReport, MonthlyReportQuery>((ref, q) async {
  final repo = ref.read(reportRepositoryProvider);
  return repo.monthly(year: q.year, month: q.month, currency: q.currency);
});

