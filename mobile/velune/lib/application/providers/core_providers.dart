import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/http/dio_client.dart';
import '../../core/storage/token_store.dart';
import '../../data/repositories/admin_repository.dart';
import '../../data/repositories/auth_repository.dart';
import '../../data/repositories/budget_repository.dart';
import '../../data/repositories/notification_repository.dart';
import '../../data/repositories/report_repository.dart';
import '../../data/repositories/transaction_repository.dart';

final tokenStoreProvider = Provider<TokenStore>((ref) {
  return TokenStore();
});

final dioClientProvider = Provider<DioClient>((ref) {
  final tokenStore = ref.watch(tokenStoreProvider);
  return DioClient.create(tokenStore: tokenStore);
});

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return DioAuthRepository(
    client: ref.watch(dioClientProvider),
    tokenStore: ref.watch(tokenStoreProvider),
  );
});

final transactionRepositoryProvider = Provider<TransactionRepository>((ref) {
  return DioTransactionRepository(
    client: ref.watch(dioClientProvider),
  );
});

final budgetRepositoryProvider = Provider<BudgetRepository>((ref) {
  return DioBudgetRepository(
    client: ref.watch(dioClientProvider),
  );
});

final reportRepositoryProvider = Provider<ReportRepository>((ref) {
  return DioReportRepository(
    client: ref.watch(dioClientProvider),
  );
});

final notificationRepositoryProvider =
    Provider<NotificationRepository>((ref) {
  return DioNotificationRepository(
    client: ref.watch(dioClientProvider),
  );
});

final adminRepositoryProvider = Provider<AdminRepository>((ref) {
  return MobileAdminRepository();
});