import 'package:dio/dio.dart';

import '../../core/http/dio_client.dart';
import '../../core/json/normalize_json_keys.dart';
import '../../core/json/normalized_deserialize.dart';
import '../../domain/models/paged.dart';
import '../../domain/models/transaction.dart';
import '../../domain/models/transaction_summary.dart';

abstract class TransactionRepository {
  Future<Paged<Transaction>> list({
    required int page,
    required int limit,
    String? accountId,
    String? categoryId,
    TransactionType? type,
    String? from,
    String? to,
    String? currency,
  });

  Future<Transaction> getById(String id);

  Future<Transaction> create({
    required String accountId,
    String? categoryId,
    String? counterpartyAccountId,
    required int amountMinor,
    required String currency,
    required TransactionType type,
    String? description,
    required String occurredAt,
  });

  Future<Transaction> update({
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
  });

  Future<void> delete({
    required String id,
    required int version,
  });

  Future<TransactionSummary> summary({
    required String from,
    required String to,
    required String currency,
  });
}

class DioTransactionRepository implements TransactionRepository {
  final DioClient _client;

  DioTransactionRepository({required DioClient client}) : _client = client;

  @override
  Future<Paged<Transaction>> list({
    required int page,
    required int limit,
    String? accountId,
    String? categoryId,
    TransactionType? type,
    String? from,
    String? to,
    String? currency,
  }) async {
    final resp = await _client.dio.get(
      '/transactions',
      queryParameters: {
        'page': page,
        'limit': limit,
        'accountId': ?accountId,
        'categoryId': ?categoryId,
        if (type != null) 'type': type.name,
        'from': ?from,
        'to': ?to,
        'currency': ?currency,
      },
    );

    final normalized = normalizeJsonKeys(resp.data as Map<String, dynamic>);
    return Paged.fromJson(
      normalized,
      itemFromJson: (m) => Transaction.fromJson(m),
    );
  }

  @override
  Future<Transaction> getById(String id) async {
    final resp = await _client.dio.get('/transactions/$id');
    return fromJsonNormalized<Transaction>(
      resp.data,
      fromJson: Transaction.fromJson,
    );
  }

  @override
  Future<Transaction> create({
    required String accountId,
    String? categoryId,
    String? counterpartyAccountId,
    required int amountMinor,
    required String currency,
    required TransactionType type,
    String? description,
    required String occurredAt,
  }) async {
    final resp = await _client.dio.post(
      '/transactions',
      data: {
        'accountId': accountId,
        'categoryId': ?categoryId,
        'counterpartyAccountId': ?counterpartyAccountId,
        'amountMinor': amountMinor,
        'currency': currency,
        'type': type.name,
        'description': ?description,
        'occurredAt': occurredAt,
      },
    );

    return fromJsonNormalized<Transaction>(
      resp.data,
      fromJson: Transaction.fromJson,
    );
  }

  @override
  Future<Transaction> update({
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
    final resp = await _client.dio.patch(
      '/transactions/$id',
      queryParameters: {'id': id, 'version': version},
      data: {
        'accountId': accountId,
        'categoryId': ?categoryId,
        'counterpartyAccountId': ?counterpartyAccountId,
        'amountMinor': amountMinor,
        'currency': currency,
        'type': type.name,
        'description': ?description,
        'occurredAt': occurredAt,
      },
    );

    return fromJsonNormalized<Transaction>(
      resp.data,
      fromJson: Transaction.fromJson,
    );
  }

  @override
  Future<void> delete({
    required String id,
    required int version,
  }) async {
    await _client.dio.delete(
      '/transactions/$id',
      queryParameters: {'id': id, 'version': version},
      options: Options(extra: const {}),
    );
  }

  @override
  Future<TransactionSummary> summary({
    required String from,
    required String to,
    required String currency,
  }) async {
    final resp = await _client.dio.get(
      '/transactions/summary',
      queryParameters: {
        'from': from,
        'to': to,
        'currency': currency,
      },
    );

    final normalized = normalizeJsonKeys(resp.data as Map<String, dynamic>);
    return TransactionSummary.fromJson(normalized);
  }
}

