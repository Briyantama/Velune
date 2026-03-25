import '../../core/http/dio_client.dart';
import '../../core/json/normalized_deserialize.dart';
import '../../domain/models/notification_ping.dart';

abstract class NotificationRepository {
  Future<NotificationPing> ping();
}

class DioNotificationRepository implements NotificationRepository {
  final DioClient _client;

  DioNotificationRepository({required DioClient client}) : _client = client;

  @override
  Future<NotificationPing> ping() async {
    final resp = await _client.dio.get('/notifications/ping');
    return fromJsonNormalized<NotificationPing>(
      resp.data,
      fromJson: NotificationPing.fromJson,
    );
  }
}

