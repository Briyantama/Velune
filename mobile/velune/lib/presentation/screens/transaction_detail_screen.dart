import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../application/providers/transactions_provider.dart';
import '../../domain/models/transaction.dart';
import '../widgets/stat_card.dart';

class TransactionDetailScreen extends ConsumerWidget {
  final String transactionId;

  const TransactionDetailScreen({
    super.key,
    required this.transactionId,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final txAsync = ref.watch(transactionDetailProvider(transactionId));

    return txAsync.when(
      data: (tx) => _TransactionDetailBody(transaction: tx),
      loading: () => const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      ),
      error: (err, st) => Scaffold(
        body: Center(
          child: Text('Failed to load transaction: $err'),
        ),
      ),
    );
  }
}

class _TransactionDetailBody extends StatelessWidget {
  final Transaction transaction;

  const _TransactionDetailBody({
    required this.transaction,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Transaction'),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          StatCard(
            title: 'Amount (${transaction.currency})',
            value: transaction.amountMinor.toString(),
            subtitle: 'Minor units',
            icon: Icons.receipt_long,
          ),
          const SizedBox(height: 12),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(14),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('Type: ${transaction.type.name}'),
                  const SizedBox(height: 8),
                  Text('Occurred: ${transaction.occurredAt}'),
                  const SizedBox(height: 8),
                  Text('Description: ${transaction.description}'),
                  const SizedBox(height: 8),
                  Text('Account: ${transaction.accountId}'),
                  const SizedBox(height: 8),
                  Text('Category: ${transaction.categoryId ?? '-'}'),
                  const SizedBox(height: 8),
                  Text(
                      'Counterparty: ${transaction.counterpartyAccountId ?? '-'}'),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

