import 'package:freezed_annotation/freezed_annotation.dart';

part 'transaction.freezed.dart';
part 'transaction.g.dart';

enum TransactionType {
  @JsonValue('income')
  income,
  @JsonValue('expense')
  expense,
  @JsonValue('transfer')
  transfer,
  @JsonValue('adjustment')
  adjustment,
}

@freezed
class Transaction with _$Transaction {
  const factory Transaction({
    required String id,
    required String userId,
    required String accountId,
    String? categoryId,
    String? counterpartyAccountId,
    required int amountMinor,
    required String currency,
    required TransactionType type,
    required String description,
    required String occurredAt,
    required int version,
    required String createdAt,
    required String updatedAt,
    String? deletedAt,
  }) = _Transaction;

  factory Transaction.fromJson(Map<String, dynamic> json) =>
      _$TransactionFromJson(json);
}

