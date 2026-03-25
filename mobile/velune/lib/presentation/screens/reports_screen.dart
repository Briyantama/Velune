import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../application/providers/auth_session_provider.dart';
import '../../application/providers/reports_provider.dart';
import '../widgets/empty_state.dart';
import '../widgets/error_banner.dart';
import '../widgets/loading_skeleton.dart';
import '../widgets/stat_card.dart';

class ReportsScreen extends ConsumerWidget {
  const ReportsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authAsync = ref.watch(authSessionProvider);
    final now = DateTime.now();
    final year = now.year;
    final month = now.month;

    final currency = authAsync.valueOrNull?.user?.baseCurrency;
    final query = MonthlyReportQuery(
      year: year,
      month: month,
      currency: currency ?? 'USD',
    );

    final reportAsync = ref.watch(monthlyReportProvider(query));

    return Scaffold(
      appBar: AppBar(
        title: const Text('Monthly Report'),
      ),
      body: authAsync.isLoading
          ? const Center(child: CircularProgressIndicator())
          : reportAsync.when(
              loading: () => const Center(
                child: LoadingSkeleton(width: 260, height: 140),
              ),
              error: (err, _) => Center(
                child: ErrorBanner(
                  title: 'Could not load report',
                  message: err.toString(),
                  onRetry: () {
                    ref.refresh(monthlyReportProvider(query));
                  },
                ),
              ),
              data: (report) {
                if (report.byCategory.isEmpty) {
                  return const EmptyState(
                    title: 'No data',
                    message: 'There are no transactions for this month.',
                  );
                }

                final groups = <BarChartGroupData>[];
                for (var i = 0; i < report.byCategory.length; i++) {
                  final c = report.byCategory[i];
                  final value = c.totalMinor.toDouble();
                  groups.add(
                    BarChartGroupData(
                      x: i,
                      barRods: [
                        BarChartRodData(
                          toY: value,
                          width: 14,
                          borderRadius: BorderRadius.circular(6),
                        ),
                      ],
                    ),
                  );
                }

                return ListView(
                  padding: const EdgeInsets.all(16),
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: StatCard(
                            title: 'Income (minor)',
                            value: report.incomeMinor.toString(),
                            subtitle: 'Total',
                            icon: Icons.trending_up,
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: StatCard(
                            title: 'Expense (minor)',
                            value: report.expenseMinor.toString(),
                            subtitle: 'Total',
                            icon: Icons.trending_down,
                            accentColor: Theme.of(context).colorScheme.error,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Text(
                      'Spending by category',
                      style: Theme.of(context).textTheme.titleLarge?.copyWith(
                            fontWeight: FontWeight.w700,
                          ),
                    ),
                    const SizedBox(height: 12),
                    SizedBox(
                      height: 260,
                      child: BarChart(
                        BarChartData(
                          alignment: BarChartAlignment.spaceAround,
                          maxY: report.byCategory
                              .map((e) => e.totalMinor)
                              .fold<int>(0, (a, b) => a > b ? a : b)
                              .toDouble() *
                              1.1,
                          barTouchData: BarTouchData(enabled: true),
                          titlesData: FlTitlesData(
                            leftTitles: const AxisTitles(
                              sideTitles: SideTitles(showTitles: false),
                            ),
                            bottomTitles: AxisTitles(
                              sideTitles: SideTitles(
                                showTitles: true,
                                getTitlesWidget: (value, meta) {
                                  final idx = value.toInt();
                                  if (idx < 0 || idx >= report.byCategory.length) {
                                    return const SizedBox.shrink();
                                  }
                                  return Padding(
                                    padding: const EdgeInsets.only(top: 8),
                                    child: Text(
                                      report.byCategory[idx].categoryName,
                                      style: const TextStyle(fontSize: 10),
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                  );
                                },
                              ),
                            ),
                            rightTitles: const AxisTitles(
                              sideTitles: SideTitles(showTitles: false),
                            ),
                            topTitles: const AxisTitles(
                              sideTitles: SideTitles(showTitles: false),
                            ),
                          ),
                          borderData: FlBorderData(show: false),
                          barGroups: groups,
                        ),
                      ),
                    ),
                    const SizedBox(height: 16),
                    ...report.byCategory.map(
                      (c) => ListTile(
                        title: Text(c.categoryName),
                        trailing: Text(c.totalMinor.toString()),
                        subtitle: Text('Currency: ${c.currency}'),
                      ),
                    ),
                  ],
                );
              },
            ),
    );
  }
}

