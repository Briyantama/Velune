import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/http/app_error.dart';
import '../../domain/models/budget.dart';
import '../../domain/models/transaction.dart';
import '../../domain/models/transaction_summary.dart';
import 'auth_session_provider.dart';
import 'core_providers.dart';

class DashboardViewModel {
  final TransactionSummary summary;
  final List<Transaction> recentTransactions;
  final List<Budget> budgets;

  const DashboardViewModel({
    required this.summary,
    required this.recentTransactions,
    required this.budgets,
  });
}

class DashboardController extends AsyncNotifier<DashboardViewModel> {
  @override
  Future<DashboardViewModel> build() async {
    final auth = await ref.watch(authSessionProvider.future);
    if (auth.status != AuthSessionStatus.authenticated || auth.user == null) {
      throw const AppError(
        code: 'unauthenticated',
        message: 'Please log in.',
        statusCode: 401,
      );
    }

    final user = auth.user!;
    final now = DateTime.now();

    final from = DateTime(now.year, now.month, 1);
    final fromStr = _dateOnly(from);
    final toStr = _dateOnly(now);

    final transactionRepo = ref.read(transactionRepositoryProvider);
    final budgetRepo = ref.read(budgetRepositoryProvider);

    final summary = await transactionRepo.summary(
      from: fromStr,
      to: toStr,
      currency: user.baseCurrency,
    );

    final recentPaged = await transactionRepo.list(
      page: 1,
      limit: 5,
      currency: user.baseCurrency,
    );

    final budgetsPaged = await budgetRepo.list(
      page: 1,
      limit: 5,
      activeOn: _dateOnly(now),
    );

    return DashboardViewModel(
      summary: summary,
      recentTransactions: recentPaged.items,
      budgets: budgetsPaged.items,
    );
  }

  static String _dateOnly(DateTime d) {
    final mm = d.month.toString().padLeft(2, '0');
    final dd = d.day.toString().padLeft(2, '0');
    return '${d.year}-$mm-$dd';
  }
}

final dashboardProvider = AsyncNotifierProvider<DashboardController,
    DashboardViewModel>(DashboardController.new);

