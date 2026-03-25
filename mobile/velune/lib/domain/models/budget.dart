import 'package:freezed_annotation/freezed_annotation.dart';

part 'budget.freezed.dart';
part 'budget.g.dart';

enum BudgetPeriodType {
  @JsonValue('monthly')
  monthly,
  @JsonValue('weekly')
  weekly,
  @JsonValue('custom')
  custom,
}

@freezed
class Budget with _$Budget {
  const factory Budget({
    required String id,
    required String userId,
    required String name,
    required BudgetPeriodType periodType,
    String? categoryId,
    required String startDate,
    required String endDate,
    required int limitAmountMinor,
    required String currency,
    required int version,
  }) = _Budget;

  factory Budget.fromJson(Map<String, dynamic> json) =>
      _$BudgetFromJson(json);
}

