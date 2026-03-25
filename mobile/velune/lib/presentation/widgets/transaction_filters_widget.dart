import 'package:flutter/material.dart';

import '../../domain/models/transaction.dart';

class TransactionFiltersWidget extends StatelessWidget {
  final TransactionType? type;
  final String? from;
  final String? to;
  final String? currency;
  final void Function({
    TransactionType? type,
    String? from,
    String? to,
    String? currency,
  }) onChanged;

  const TransactionFiltersWidget({
    super.key,
    this.type,
    this.from,
    this.to,
    this.currency,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 0,
      margin: EdgeInsets.zero,
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            DropdownButtonFormField<TransactionType>(
              initialValue: type,
              decoration: const InputDecoration(
                labelText: 'Type',
              ),
              items: const [
                DropdownMenuItem(
                  value: TransactionType.income,
                  child: Text('Income'),
                ),
                DropdownMenuItem(
                  value: TransactionType.expense,
                  child: Text('Expense'),
                ),
                DropdownMenuItem(
                  value: TransactionType.transfer,
                  child: Text('Transfer'),
                ),
                DropdownMenuItem(
                  value: TransactionType.adjustment,
                  child: Text('Adjustment'),
                ),
              ],
              onChanged: (v) {
                onChanged(
                  type: v,
                  from: from,
                  to: to,
                  currency: currency,
                );
              },
            ),
            const SizedBox(height: 10),
            TextFormField(
              initialValue: from,
              decoration: const InputDecoration(labelText: 'From (YYYY-MM-DD)'),
              onChanged: (v) => onChanged(
                type: type,
                from: v.isEmpty ? null : v,
                to: to,
                currency: currency,
              ),
            ),
            const SizedBox(height: 10),
            TextFormField(
              initialValue: to,
              decoration: const InputDecoration(labelText: 'To (YYYY-MM-DD)'),
              onChanged: (v) => onChanged(
                type: type,
                from: from,
                to: v.isEmpty ? null : v,
                currency: currency,
              ),
            ),
            const SizedBox(height: 10),
            TextFormField(
              initialValue: currency,
              decoration: const InputDecoration(labelText: 'Currency'),
              onChanged: (v) => onChanged(
                type: type,
                from: from,
                to: to,
                currency: v.isEmpty ? null : v,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

