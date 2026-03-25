import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../application/providers/dashboard_provider.dart';
import '../widgets/empty_state.dart';
import '../widgets/error_banner.dart';
import '../widgets/loading_skeleton.dart';
import '../widgets/stat_card.dart';

class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final dashboardAsync = ref.watch(dashboardProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Dashboard'),
      ),
      body: dashboardAsync.when(
        loading: () => const Center(
          child: LoadingSkeleton(width: 260, height: 160),
        ),
        error: (err, st) => Center(
          child: ErrorBanner(
            title: 'Could not load dashboard',
            message: err.toString(),
            onRetry: () {
              ref.refresh(dashboardProvider);
            },
          ),
        ),
        data: (vm) {
          if (vm.recentTransactions.isEmpty && vm.budgets.isEmpty) {
            return const EmptyState(
              title: 'Nothing yet',
              message: 'Add a transaction or create a budget to see insights.',
            );
          }

          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Row(
                children: [
                  Expanded(
                    child: StatCard(
                      title: 'Net (minor units)',
                      value: vm.summary.netMinor.toString(),
                      subtitle: 'Income - Expense',
                      icon: Icons.savings,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: StatCard(
                      title: 'Income (minor units)',
                      value: vm.summary.incomeMinor.toString(),
                      subtitle: 'For this period',
                      icon: Icons.trending_up,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              StatCard(
                title: 'Expense (minor units)',
                value: vm.summary.expenseMinor.toString(),
                subtitle: 'For this period',
                icon: Icons.trending_down,
                accentColor: Theme.of(context).colorScheme.error,
              ),
              const SizedBox(height: 16),
              const Text(
                'Recent transactions',
                style: TextStyle(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              ...vm.recentTransactions.map(
                (tx) => ListTile(
                  title: Text('${tx.type.name.toUpperCase()} - ${tx.description}'),
                  subtitle: Text('Occurred: ${tx.occurredAt}'),
                  trailing: Text(tx.amountMinor.toString()),
                ),
              ),
              const SizedBox(height: 16),
              const Text(
                'Budgets',
                style: TextStyle(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              ...vm.budgets.map(
                (b) => ListTile(
                  title: Text(b.name),
                  subtitle: Text('Limit: ${b.limitAmountMinor} (${b.currency})'),
                  trailing: Text(b.periodType.name),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

