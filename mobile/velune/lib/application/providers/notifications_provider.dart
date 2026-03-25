import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../domain/models/notification_ping.dart';
import '../providers/core_providers.dart';

final notificationsPingProvider =
    FutureProvider.autoDispose<NotificationPing>((ref) async {
  final repo = ref.read(notificationRepositoryProvider);
  return repo.ping();
});

