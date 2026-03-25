import 'package:freezed_annotation/freezed_annotation.dart';

part 'notification_ping.freezed.dart';
part 'notification_ping.g.dart';

@freezed
class NotificationPing with _$NotificationPing {
  const factory NotificationPing({
    required String status,
  }) = _NotificationPing;

  factory NotificationPing.fromJson(Map<String, dynamic> json) =>
      _$NotificationPingFromJson(json);
}

