import 'package:flutter_dotenv/flutter_dotenv.dart';

class AppEnv {
  /// Default value exists only as a fallback; the intended source of truth
  /// is `.env` / `.env.example`.
  static const String _defaultGatewayBaseUrl = 'http://127.0.0.1:8080';

  static String get gatewayBaseUrl {
    final value = dotenv.env['GATEWAY_BASE_URL'];
    final trimmed = value?.trim();
    if (trimmed == null || trimmed.isEmpty) return _defaultGatewayBaseUrl;
    return trimmed;
  }
}

