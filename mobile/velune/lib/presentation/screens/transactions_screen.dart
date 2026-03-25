import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../application/providers/auth_session_provider.dart';
import '../../application/providers/settings_provider.dart';
import '../../application/providers/transactions_provider.dart';
import '../../domain/models/paged.dart';
import '../../domain/models/transaction.dart';
import '../widgets/budget_progress_card.dart';
import '../widgets/confirm_dialog.dart';
import '../widgets/empty_state.dart';
import '../widgets/loading_skeleton.dart';
import '../widgets/transaction_filters_widget.dart';

class TransactionsScreen extends ConsumerStatefulWidget {
  const TransactionsScreen({super.key});

  @override
  ConsumerState<TransactionsScreen> createState() => _TransactionsScreenState();
}

class _TransactionsScreenState extends ConsumerState<TransactionsScreen> {
  TransactionType? _type;
  String? _from;
  String? _to;
  String? _currency;

  late final int _page;
  late final int _limit;

  @override
  void initState() {
    super.initState();
    _page = 1;
    _limit = 20;
  }

  @override
  Widget build(BuildContext context) {
    final authAsync = ref.watch(authSessionProvider);
    final currency = authAsync.valueOrNull?.user?.baseCurrency ?? _currency;

    final txQuery = TransactionsQuery(
      page: _page,
      limit: _limit,
      type: _type,
      from: _from,
      to: _to,
      currency: currency,
    );

    final txAsync = ref.watch(transactionsListProvider(txQuery));

    return Scaffold(
      appBar: AppBar(
        title: const Text('Transactions'),
      ),
      body: authAsync.isLoading
          ? const Center(child: CircularProgressIndicator())
          : txAsync.when(
              data: (paged) => _TransactionsBody(
                items: paged.items,
                onOpenDetail: (id) => context.go('/transactions/$id'),
                onEdit: (tx) => _openEditDialog(tx),
                onDelete: (tx) => _confirmDelete(tx),
              ),
              loading: () => const Center(
                child: LoadingSkeleton(width: 220, height: 80),
              ),
              error: (err, _) => Center(
                child: Text('Failed to load transactions: $err'),
              ),
              ),
      floatingActionButton: FloatingActionButton(
        onPressed: _openCreateDialog,
        child: const Icon(Icons.add),
      ),
    );
  }

  Future<void> _confirmDelete(Transaction tx) async {
    final actions = ref.read(transactionsActionsProvider);

    await ConfirmDialog.show(
      context: context,
      title: 'Delete transaction?',
      message: 'This will remove the transaction from your ledger.',
      confirmText: 'Delete',
      cancelText: 'Cancel',
      onConfirm: () async {
        await actions.delete(id: tx.id, version: tx.version);
      },
    );
  }

  void _openCreateDialog() {
    final actions = ref.read(transactionsActionsProvider);

    String? accountId;
    String? categoryId;
    String? counterpartyAccountId;
    String? description;
    String currency = authSessionCurrency() ?? '';
    TransactionType type = _type ?? TransactionType.expense;
    int amountMinor = 0;
    String occurredAt = DateTime.now().toIso8601String();

    showDialog<void>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Create transaction'),
          content: StatefulBuilder(
            builder: (context, setState) {
              return SingleChildScrollView(
                child: Column(
                  children: [
                    TextFormField(
                      decoration: const InputDecoration(labelText: 'Account ID'),
                      onChanged: (v) => setState(() => accountId = v),
                    ),
                    TextFormField(
                      decoration: const InputDecoration(labelText: 'Category ID (optional)'),
                      onChanged: (v) => setState(() => categoryId = v.isEmpty ? null : v),
                    ),
                    TextFormField(
                      decoration: const InputDecoration(labelText: 'Amount (minor units)'),
                      keyboardType: TextInputType.number,
                      onChanged: (v) => setState(() => amountMinor = int.tryParse(v) ?? 0),
                    ),
                    DropdownButtonFormField<TransactionType>(
                      initialValue: type,
                      decoration: const InputDecoration(labelText: 'Type'),
                      items: TransactionType.values
                          .map(
                            (t) => DropdownMenuItem(
                              value: t,
                              child: Text(t.name),
                            ),
                          )
                          .toList(),
                      onChanged: (v) => setState(() => type = v ?? TransactionType.expense),
                    ),
                    TextFormField(
                      decoration: const InputDecoration(labelText: 'Occurred At (ISO)'),
                      initialValue: occurredAt,
                      onChanged: (v) => setState(() => occurredAt = v),
                    ),
                    TextFormField(
                      decoration: const InputDecoration(labelText: 'Description (optional)'),
                      onChanged: (v) => setState(() => description = v.isEmpty ? null : v),
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
                if (accountId == null || accountId!.isEmpty) return;
                await actions.create(
                  accountId: accountId!,
                  categoryId: categoryId,
                  counterpartyAccountId: counterpartyAccountId,
                  amountMinor: amountMinor,
                  currency: currency,
                  type: type,
                  description: description,
                  occurredAt: occurredAt,
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

  void _openEditDialog(Transaction tx) {
    final actions = ref.read(transactionsActionsProvider);

    String? accountId = tx.accountId;
    String? categoryId = tx.categoryId;
    String? counterpartyAccountId = tx.counterpartyAccountId;
    String? description = tx.description;
    String currency = tx.currency;
    TransactionType type = tx.type;
    int amountMinor = tx.amountMinor;
    String occurredAt = tx.occurredAt;

    showDialog<void>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Edit transaction'),
          content: StatefulBuilder(
            builder: (context, setState) {
              return SingleChildScrollView(
                child: Column(
                  children: [
                    TextFormField(
                      decoration: const InputDecoration(labelText: 'Account ID'),
                      initialValue: accountId,
                      onChanged: (v) => setState(() => accountId = v),
                    ),
                    TextFormField(
                      decoration: const InputDecoration(labelText: 'Category ID (optional)'),
                      initialValue: categoryId,
                      onChanged: (v) => setState(() => categoryId = v.isEmpty ? null : v),
                    ),
                    TextFormField(
                      decoration: const InputDecoration(labelText: 'Amount (minor units)'),
                      keyboardType: TextInputType.number,
                      initialValue: amountMinor.toString(),
                      onChanged: (v) => setState(() => amountMinor = int.tryParse(v) ?? amountMinor),
                    ),
                    DropdownButtonFormField<TransactionType>(
                      initialValue: type,
                      decoration: const InputDecoration(labelText: 'Type'),
                      items: TransactionType.values
                          .map(
                            (t) => DropdownMenuItem(
                              value: t,
                              child: Text(t.name),
                            ),
                          )
                          .toList(),
                      onChanged: (v) => setState(() => type = v ?? tx.type),
                    ),
                    TextFormField(
                      decoration: const InputDecoration(labelText: 'Occurred At (ISO)'),
                      initialValue: occurredAt,
                      onChanged: (v) => setState(() => occurredAt = v),
                    ),
                    TextFormField(
                      decoration: const InputDecoration(labelText: 'Description (optional)'),
                      initialValue: description,
                      onChanged: (v) => setState(() => description = v.isEmpty ? null : v),
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
                if (accountId == null || accountId.isEmpty) return;
                await actions.update(
                  id: tx.id,
                  version: tx.version,
                  accountId: accountId!,
                  categoryId: categoryId,
                  counterpartyAccountId: counterpartyAccountId,
                  amountMinor: amountMinor,
                  currency: currency,
                  type: type,
                  description: description,
                  occurredAt: occurredAt,
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

  String? authSessionCurrency() => ref.read(authSessionProvider).valueOrNull?.user?.baseCurrency;
}

class _TransactionsBody extends StatelessWidget {
  final List<Transaction> items;
  final void Function(String id) onOpenDetail;
  final void Function(Transaction tx) onEdit;
  final Future<void> Function(Transaction tx) onDelete;

  const _TransactionsBody({
    required this.items,
    required this.onOpenDetail,
    required this.onEdit,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return const EmptyState(title: 'No transactions', message: 'Add your first transaction to get started.');
    }

    return ListView.separated(
      padding: const EdgeInsets.all(16),
      itemCount: items.length,
      separatorBuilder: (_, _) => const SizedBox(height: 12),
      itemBuilder: (context, i) {
        final tx = items[i];
        return Card(
          child: ListTile(
            title: Text('${tx.type.name.toUpperCase()} - ${tx.description}'),
            subtitle: Text('Occurred: ${tx.occurredAt}'),
            trailing: Wrap(
              spacing: 8,
              children: [
                IconButton(
                  icon: const Icon(Icons.edit),
                  onPressed: () => onEdit(tx),
                ),
                IconButton(
                  icon: const Icon(Icons.delete_outline),
                  onPressed: () => onDelete(tx),
                ),
              ],
            ),
            onTap: () => onOpenDetail(tx.id),
          ),
        );
      },
    );
  }
}

