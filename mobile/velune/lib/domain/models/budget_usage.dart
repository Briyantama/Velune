import 'package:freezed_annotation/freezed_annotation.dart';

part 'budget_usage.freezed.dart';
part 'budget_usage.g.dart';

@freezed
class BudgetUsage with _$BudgetUsage {
  const factory BudgetUsage({
    required String budgetId,
    required String from,
    required String to,
    required String currency,
    required int limitAmountMinor,
    required int spentMinor,
    required int remainingMinor,
    required int overspentMinor,
    required bool isOverspent,
  }) = _BudgetUsage;

  factory BudgetUsage.fromJson(Map<String, dynamic> json) =>
      _$BudgetUsageFromJson(json);
}

