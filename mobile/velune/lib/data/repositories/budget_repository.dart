import 'package:dio/dio.dart';

import '../../core/http/dio_client.dart';
import '../../core/json/normalize_json_keys.dart';
import '../../core/json/normalized_deserialize.dart';
import '../../domain/models/budget.dart';
import '../../domain/models/budget_usage.dart';
import '../../domain/models/paged.dart';

abstract class BudgetRepository {
  Future<Paged<Budget>> list({
    required int page,
    required int limit,
    String? activeOn,
  });

  Future<BudgetUsage> usage(String id);

  Future<Budget> create({
    required String name,
    required BudgetPeriodType periodType,
    String? categoryId,
    required String startDate,
    required String endDate,
    required int limitAmountMinor,
    required String currency,
  });

  Future<Budget> update({
    required String id,
    required int version,
    required String name,
    required BudgetPeriodType periodType,
    String? categoryId,
    required String startDate,
    required String endDate,
    required int limitAmountMinor,
    required String currency,
  });

  Future<void> delete({
    required String id,
    required int version,
  });
}

class DioBudgetRepository implements BudgetRepository {
  final DioClient _client;

  DioBudgetRepository({required DioClient client}) : _client = client;

  @override
  Future<Paged<Budget>> list({
    required int page,
    required int limit,
    String? activeOn,
  }) async {
    final resp = await _client.dio.get(
      '/budgets',
      queryParameters: {
        'page': page,
        'limit': limit,
        'activeOn': ?activeOn,
      },
    );

    final normalized = normalizeJsonKeys(resp.data as Map<String, dynamic>);
    return Paged.fromJson(
      normalized,
      itemFromJson: (m) => Budget.fromJson(m),
    );
  }

  @override
  Future<BudgetUsage> usage(String id) async {
    final resp = await _client.dio.get(
      '/budgets/$id/usage',
      queryParameters: {'id': id},
    );

    return fromJsonNormalized<BudgetUsage>(
      resp.data,
      fromJson: BudgetUsage.fromJson,
    );
  }

  @override
  Future<Budget> create({
    required String name,
    required BudgetPeriodType periodType,
    String? categoryId,
    required String startDate,
    required String endDate,
    required int limitAmountMinor,
    required String currency,
  }) async {
    final resp = await _client.dio.post(
      '/budgets',
      data: {
        'name': name,
        'periodType': periodType.name,
        'categoryId': ?categoryId,
        'startDate': startDate,
        'endDate': endDate,
        'limitAmountMinor': limitAmountMinor,
        'currency': currency,
      },
    );

    return fromJsonNormalized<Budget>(
      resp.data,
      fromJson: Budget.fromJson,
    );
  }

  @override
  Future<Budget> update({
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
    final resp = await _client.dio.put(
      '/budgets/$id',
      queryParameters: {'id': id, 'version': version},
      data: {
        'name': name,
        'periodType': periodType.name,
        'categoryId': ?categoryId,
        'startDate': startDate,
        'endDate': endDate,
        'limitAmountMinor': limitAmountMinor,
        'currency': currency,
      },
    );

    return fromJsonNormalized<Budget>(
      resp.data,
      fromJson: Budget.fromJson,
    );
  }

  @override
  Future<void> delete({
    required String id,
    required int version,
  }) async {
    await _client.dio.delete(
      '/budgets/$id',
      queryParameters: {'id': id, 'version': version},
      options: Options(extra: const {}),
    );
  }
}

