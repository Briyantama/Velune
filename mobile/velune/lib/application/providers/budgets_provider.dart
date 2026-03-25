import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/repositories/budget_repository.dart';
import '../../domain/models/budget.dart';
import '../../domain/models/budget_usage.dart';
import '../../domain/models/paged.dart';
import '../providers/core_providers.dart';

@immutable
class BudgetsQuery {
  final int page;
  final int limit;
  final String? activeOn;

  const BudgetsQuery({
    required this.page,
    required this.limit,
    this.activeOn,
  });
}

final budgetsListProvider = FutureProvider.autoDispose
    .family<Paged<Budget>, BudgetsQuery>((ref, q) async {
  final repo = ref.read(budgetRepositoryProvider);
  return repo.list(page: q.page, limit: q.limit, activeOn: q.activeOn);
});

final budgetUsageProvider = FutureProvider.autoDispose
    .family<BudgetUsage, String>((ref, budgetId) async {
  final repo = ref.read(budgetRepositoryProvider);
  return repo.usage(budgetId);
});

class BudgetsActions {
  final Ref ref;
  final BudgetRepository repo;

  BudgetsActions({required this.ref, required this.repo});

  Future<void> create({
    required String name,
    required BudgetPeriodType periodType,
    String? categoryId,
    required String startDate,
    required String endDate,
    required int limitAmountMinor,
    required String currency,
  }) async {
    await repo.create(
      name: name,
      periodType: periodType,
      categoryId: categoryId,
      startDate: startDate,
      endDate: endDate,
      limitAmountMinor: limitAmountMinor,
      currency: currency,
    );
    ref.invalidate(budgetsListProvider);
  }

  Future<void> update({
    required String id,
    required int version,
    required String name,
    required BudgetPeriodType periodType,
    String? categoryId,
    required String startDate,
    required String endDate,
    required int limitAmountMinor,
    required String currency,
  }) async {
    await repo.update(
      id: id,
      version: version,
      name: name,
      periodType: periodType,
      categoryId: categoryId,
      startDate: startDate,
      endDate: endDate,
      limitAmountMinor: limitAmountMinor,
      currency: currency,
    );
    ref.invalidate(budgetsListProvider);
  }

  Future<void> delete({
    required String id,
    required int version,
  }) async {
    await repo.delete(id: id, version: version);
    ref.invalidate(budgetsListProvider);
  }
}

final budgetsActionsProvider = Provider<BudgetsActions>((ref) {
  return BudgetsActions(ref: ref, repo: ref.watch(budgetRepositoryProvider));
});

