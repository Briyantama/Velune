import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../domain/models/paged.dart';
import '../../domain/models/transaction.dart';
import '../../data/repositories/transaction_repository.dart';
import '../providers/core_providers.dart';

@immutable
class TransactionsQuery {
  final int page;
  final int limit;
  final String? accountId;
  final String? categoryId;
  final TransactionType? type;
  final String? from;
  final String? to;
  final String? currency;

  const TransactionsQuery({
    required this.page,
    required this.limit,
    this.accountId,
    this.categoryId,
    this.type,
    this.from,
    this.to,
    this.currency,
  });
}

final transactionsListProvider = FutureProvider.autoDispose
    .family<Paged<Transaction>, TransactionsQuery>((ref, q) async {
  final repo = ref.read(transactionRepositoryProvider);
  return repo.list(
    page: q.page,
    limit: q.limit,
    accountId: q.accountId,
    categoryId: q.categoryId,
    type: q.type,
    from: q.from,
    to: q.to,
    currency: q.currency,
  );
});

final transactionDetailProvider =
    FutureProvider.autoDispose.family<Transaction, String>((ref, id) async {
  final repo = ref.read(transactionRepositoryProvider);
  return repo.getById(id);
});

class TransactionsActions {
  final Ref ref;
  final TransactionRepository repo;

  TransactionsActions({
    required this.ref,
    required this.repo,
  });

  Future<void> create({
    required String accountId,
    String? categoryId,
    String? counterpartyAccountId,
    required int amountMinor,
    required String currency,
    required TransactionType type,
    String? description,
    required String occurredAt,
  }) async {
    await repo.create(
      accountId: accountId,
      categoryId: categoryId,
      counterpartyAccountId: counterpartyAccountId,
      amountMinor: amountMinor,
      currency: currency,
      type: type,
      description: description,
      occurredAt: occurredAt,
    );
    ref.invalidate(transactionsListProvider);
  }

  Future<void> update({
    required String id,
    required int version,
    required String accountId,
    String? categoryId,
    String? counterpartyAccountId,
    required int amountMinor,
    required String currency,
    required TransactionType type,
    String? description,
    required String occurredAt,
  }) async {
    await repo.update(
      id: id,
      version: version,
      accountId: accountId,
      categoryId: categoryId,
      counterpartyAccountId: counterpartyAccountId,
      amountMinor: amountMinor,
      currency: currency,
      type: type,
      description: description,
      occurredAt: occurredAt,
    );
    ref.invalidate(transactionsListProvider);
  }

  Future<void> delete({
    required String id,
    required int version,
  }) async {
    await repo.delete(id: id, version: version);
    ref.invalidate(transactionsListProvider);
  }
}

final transactionsActionsProvider = Provider<TransactionsActions>((ref) {
  return TransactionsActions(
    ref: ref,
    repo: ref.watch(transactionRepositoryProvider),
  );
});

