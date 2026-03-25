import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../application/providers/auth_session_provider.dart';
import '../../application/providers/budgets_provider.dart';
import '../../domain/models/budget.dart';
import '../widgets/budget_progress_card.dart';
import '../widgets/confirm_dialog.dart';
import '../widgets/empty_state.dart';
import '../widgets/error_banner.dart';
import '../widgets/loading_skeleton.dart';

class BudgetsScreen extends ConsumerStatefulWidget {
  const BudgetsScreen({super.key});

  @override
  ConsumerState<BudgetsScreen> createState() => _BudgetsScreenState();
}

class _BudgetsScreenState extends ConsumerState<BudgetsScreen> {
  String? _selectedBudgetId;

  String _dateOnly(DateTime d) {
    final mm = d.month.toString().padLeft(2, '0');
    final dd = d.day.toString().padLeft(2, '0');
    return '${d.year}-$mm-$dd';
  }

  @override
  Widget build(BuildContext context) {
    final authAsync = ref.watch(authSessionProvider);
    final user = authAsync.valueOrNull?.user;

    final now = DateTime.now();
    final budgetsQuery = BudgetsQuery(
      page: 1,
      limit: 20,
      activeOn: _dateOnly(now),
    );

    final budgetsAsync = ref.watch(budgetsListProvider(budgetsQuery));

    final selectedUsageAsync = _selectedBudgetId == null
        ? null
        : ref.watch(budgetUsageProvider(_selectedBudgetId!));

    return Scaffold(
      appBar: AppBar(
        title: const Text('Budgets'),
      ),
      body: authAsync.isLoading
          ? const Center(child: CircularProgressIndicator())
          : budgetsAsync.when(
              loading: () => const Center(
                child: LoadingSkeleton(width: 260, height: 160),
              ),
              error: (err, _) => Center(
                child: ErrorBanner(
                  title: 'Could not load budgets',
                  message: err.toString(),
                  onRetry: () {
                    ref.refresh(budgetsListProvider(budgetsQuery));
                  },
                ),
              ),
              data: (paged) {
                final budgets = paged.items;
                if (budgets.isEmpty) {
                  return const EmptyState(
                    title: 'No budgets',
                    message: 'Create a budget to start tracking category spending.',
                  );
                }

                final selectedBudgetId =
                    _selectedBudgetId != null && budgets.any((b) => b.id == _selectedBudgetId)
                        ? _selectedBudgetId
                        : budgets.first.id;

                if (_selectedBudgetId == null || _selectedBudgetId != selectedBudgetId) {
                  WidgetsBinding.instance.addPostFrameCallback(
                    (_) => setState(() => _selectedBudgetId = selectedBudgetId),
                  );
                }

                return ListView(
                  padding: const EdgeInsets.all(16),
                  children: [
                    const Text(
                      'Budget usage',
                      style: TextStyle(fontWeight: FontWeight.w700),
                    ),
                    const SizedBox(height: 12),
                    if (selectedUsageAsync != null)
                      selectedUsageAsync.when(
                        loading: () => const LoadingSkeleton(
                          width: double.infinity,
                          height: 120,
                        ),
                        error: (err, _) => ErrorBanner(
                          title: 'Could not load usage',
                          message: err.toString(),
                        ),
                        data: (usage) {
                          final limit = usage.limitAmountMinor;
                          final spent = usage.spentMinor;
                          final progress = limit == 0 ? 0 : spent / limit;
                          return BudgetProgressCard(
                            title: 'Usage: ${usage.currency}',
                            spentText: usage.spentMinor.toString(),
                            remainingText: usage.remainingMinor.toString(),
                            progress: progress,
                            isOverspent: usage.isOverspent,
                          );
                        },
                      ),
                    const SizedBox(height: 16),
                    const Text(
                      'Your budgets',
                      style: TextStyle(fontWeight: FontWeight.w700),
                    ),
                    const SizedBox(height: 8),
                    for (final b in budgets) ...[
                      Card(
                        child: ListTile(
                          title: Text(b.name),
                          subtitle: Text(
                            '${b.periodType.name} • Limit: ${b.limitAmountMinor} ${b.currency}',
                          ),
                          trailing: Wrap(
                            spacing: 8,
                            children: [
                              IconButton(
                                tooltip: 'Edit',
                                icon: const Icon(Icons.edit),
                                onPressed: () => _openEditDialog(b),
                              ),
                              IconButton(
                                tooltip: 'Delete',
                                icon: const Icon(Icons.delete_outline),
                                onPressed: () => _confirmDelete(b),
                              ),
                            ],
                          ),
                          selected: b.id == selectedBudgetId,
                          onTap: () => setState(() => _selectedBudgetId = b.id),
                        ),
                      ),
                      const SizedBox(height: 8),
                    ],
                    const SizedBox(height: 12),
                    FilledButton.tonal(
                      onPressed: () => _openCreateDialog(user?.baseCurrency ?? 'USD'),
                      child: const Text('Create budget'),
                    ),
                  ],
                );
              },
            ),
    );
  }

  Future<void> _confirmDelete(Budget b) async {
    final actions = ref.read(budgetsActionsProvider);

    await ConfirmDialog.show(
      context: context,
      title: 'Delete budget?',
      message: 'This budget and its usage tracking will be removed.',
      confirmText: 'Delete',
      cancelText: 'Cancel',
      onConfirm: () async {
        await actions.delete(id: b.id, version: b.version);
      },
    );
  }

  void _openCreateDialog(String currencyDefault) {
    final actions = ref.read(budgetsActionsProvider);

    String name = '';
    BudgetPeriodType periodType = BudgetPeriodType.monthly;
    String? categoryId;
    String startDate = DateTime.now().toIso8601String().substring(0, 10);
    String endDate = DateTime.now().toIso8601String().substring(0, 10);
    int limitAmountMinor = 0;
    String currency = currencyDefault;

    showDialog<void>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Create budget'),
          content: StatefulBuilder(
            builder: (context, setState) {
              return SingleChildScrollView(
                child: Column(
                  children: [
                    TextFormField(
                      decoration: const InputDecoration(labelText: 'Name'),
                      onChanged: (v) => setState(() => name = v),
                    ),
                    const SizedBox(height: 8),
                    DropdownButtonFormField<BudgetPeriodType>(
                      initialValue: periodType,
                      decoration: const InputDecoration(labelText: 'Period'),
                      items: BudgetPeriodType.values
                          .map(
                            (p) => DropdownMenuItem(
                              value: p,
                              child: Text(p.name),
                            ),
                          )
                          .toList(),
                      onChanged: (v) =>
                          setState(() => periodType = v ?? BudgetPeriodType.monthly),
                    ),
                    const SizedBox(height: 8),
                    TextFormField(
                      decoration: const InputDecoration(
                        labelText: 'Category ID (optional)',
                      ),
                      onChanged: (v) => setState(() => categoryId = v.isEmpty ? null : v),
                    ),
                    const SizedBox(height: 8),
                    TextFormField(
                      decoration: const InputDecoration(
                        labelText: 'Limit amount (minor units)',
                      ),
                      keyboardType: TextInputType.number,
                      onChanged: (v) =>
                          setState(() => limitAmountMinor = int.tryParse(v) ?? 0),
                    ),
                  ],
                ),
              );
            },
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () async {
                if (name.trim().isEmpty) return;
                await actions.create(
                  name: name.trim(),
                  periodType: periodType,
                  categoryId: categoryId,
                  startDate: startDate,
                  endDate: endDate,
                  limitAmountMinor: limitAmountMinor,
                  currency: currency,
                );
                if (context.mounted) Navigator.of(context).pop();
              },
              child: const Text('Create'),
            ),
          ],
        );
      },
    );
  }

  void _openEditDialog(Budget b) {
    final actions = ref.read(budgetsActionsProvider);

    String name = b.name;
    BudgetPeriodType periodType = b.periodType;
    String? categoryId = b.categoryId;
    String startDate = b.startDate;
    String endDate = b.endDate;
    int limitAmountMinor = b.limitAmountMinor;
    String currency = b.currency;

    showDialog<void>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Edit budget'),
          content: StatefulBuilder(
            builder: (context, setState) {
              return SingleChildScrollView(
                child: Column(
                  children: [
                    TextFormField(
                      decoration: const InputDecoration(labelText: 'Name'),
                      initialValue: name,
                      onChanged: (v) => setState(() => name = v),
                    ),
                    const SizedBox(height: 8),
                    DropdownButtonFormField<BudgetPeriodType>(
                      initialValue: periodType,
                      decoration: const InputDecoration(labelText: 'Period'),
                      items: BudgetPeriodType.values
                          .map(
                            (p) => DropdownMenuItem(
                              value: p,
                              child: Text(p.name),
                            ),
                          )
                          .toList(),
                      onChanged: (v) =>
                          setState(() => periodType = v ?? BudgetPeriodType.monthly),
                    ),
                    const SizedBox(height: 8),
                    TextFormField(
                      decoration: const InputDecoration(
                        labelText: 'Category ID (optional)',
                      ),
                      initialValue: categoryId,
                      onChanged: (v) => setState(() => categoryId = v.isEmpty ? null : v),
                    ),
                    const SizedBox(height: 8),
                    TextFormField(
                      decoration: const InputDecoration(
                        labelText: 'Limit amount (minor units)',
                      ),
                      keyboardType: TextInputType.number,
                      initialValue: limitAmountMinor.toString(),
                      onChanged: (v) =>
                          setState(() => limitAmountMinor = int.tryParse(v) ?? limitAmountMinor),
                    ),
                    const SizedBox(height: 8),
                    TextFormField(
                      decoration: const InputDecoration(labelText: 'Start date (YYYY-MM-DD)'),
                      initialValue: startDate,
                      onChanged: (v) => setState(() => startDate = v),
                    ),
                    const SizedBox(height: 8),
                    TextFormField(
                      decoration: const InputDecoration(labelText: 'End date (YYYY-MM-DD)'),
                      initialValue: endDate,
                      onChanged: (v) => setState(() => endDate = v),
                    ),
                  ],
                ),
              );
            },
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () async {
                if (name.trim().isEmpty) return;
                await actions.update(
                  id: b.id,
                  version: b.version,
                  name: name.trim(),
                  periodType: periodType,
                  categoryId: categoryId,
                  startDate: startDate,
                  endDate: endDate,
                  limitAmountMinor: limitAmountMinor,
                  currency: currency,
                );
                if (context.mounted) Navigator.of(context).pop();
              },
              child: const Text('Save'),
            ),
          ],
        );
      },
    );
  }
}

