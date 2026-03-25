import 'package:freezed_annotation/freezed_annotation.dart';

part 'monthly_report.freezed.dart';
part 'monthly_report.g.dart';

@freezed
class MonthlyCategoryBreakdown with _$MonthlyCategoryBreakdown {
  const factory MonthlyCategoryBreakdown({
    String? categoryId,
    required String categoryName,
    required int totalMinor,
    required String currency,
  }) = _MonthlyCategoryBreakdown;

  factory MonthlyCategoryBreakdown.fromJson(Map<String, dynamic> json) =>
      _$MonthlyCategoryBreakdownFromJson(json);
}

@freezed
class MonthlyReport with _$MonthlyReport {
  const factory MonthlyReport({
    required String userId,
    required int year,
    required int month,
    required int incomeMinor,
    required int expenseMinor,
    required String currency,
    required List<MonthlyCategoryBreakdown> byCategory,
    required String generatedAt,
  }) = _MonthlyReport;

  factory MonthlyReport.fromJson(Map<String, dynamic> json) =>
      _$MonthlyReportFromJson(json);
}

