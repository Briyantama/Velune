import 'package:dio/dio.dart';

import '../../core/http/dio_client.dart';
import '../../core/http/app_error.dart';
import '../../core/json/normalize_json_keys.dart';
import '../../core/json/normalized_deserialize.dart';
import '../../core/storage/token_store.dart';
import '../../domain/models/user.dart';

abstract class AuthRepository {
  Future<void> login({
    required String email,
    required String password,
  });

  Future<void> register({
    required String email,
    required String password,
    required String baseCurrency,
  });

  /// Returns the authenticated user profile (`/api/v1/auth/me`).
  Future<User> me();

  /// Refreshes tokens using the stored refresh token.
  Future<void> refresh();

  /// Clears local tokens (mobile app does not call a server logout endpoint).
  Future<void> logout();
}

class DioAuthRepository implements AuthRepository {
  final DioClient _client;
  final TokenStore _tokenStore;

  DioAuthRepository({
    required DioClient client,
    required TokenStore tokenStore,
  })  : _client = client,
        _tokenStore = tokenStore;

  @override
  Future<void> login({
    required String email,
    required String password,
  }) async {
    await _loginOrRegister(
      '/api/v1/auth/login',
      body: {
        'email': email,
        'password': password,
      },
    );
  }

  @override
  Future<void> register({
    required String email,
    required String password,
    required String baseCurrency,
  }) async {
    await _loginOrRegister(
      '/api/v1/auth/register',
      body: {
        'email': email,
        'password': password,
        'baseCurrency': baseCurrency,
      },
    );
  }

  @override
  Future<User> me() async {
    final resp = await _client.dio.get('/api/v1/auth/me');
    return fromJsonNormalized<User>(
      resp.data,
      fromJson: User.fromJson,
    );
  }

  @override
  Future<void> refresh() async {
    final refreshToken = await _tokenStore.readRefreshToken();
    if (refreshToken == null || refreshToken.isEmpty) {
      throw const AppError(
        code: 'auth_expired',
        message: 'Missing refresh token.',
        statusCode: 401,
      );
    }

    final resp = await _client.dio.post(
      '/api/v1/auth/refresh',
      data: {'refresh_token': refreshToken},
      options: Options(
        extra: const {
          // These extra flags are used by the Dio interceptor.
          'skipAuth': true,
          'skipTokenRefresh': true,
        },
      ),
    );

    final accessToken = _extractAccessToken(resp.data);
    final newRefreshToken = _extractRefreshToken(resp.data);

    await _tokenStore.writeTokens(
      accessToken: accessToken,
      refreshToken: newRefreshToken,
    );
  }

  @override
  Future<void> logout() async {
    await _tokenStore.clearTokens();
  }

  Future<void> _loginOrRegister(
    String endpoint, {
    required Map<String, dynamic> body,
  }) async {
    final resp = await _client.dio.post(endpoint, data: body);

    final accessToken = _extractAccessToken(resp.data);
    final newRefreshToken = _extractRefreshToken(resp.data);

    await _tokenStore.writeTokens(
      accessToken: accessToken,
      refreshToken: newRefreshToken,
    );
  }

  String _extractAccessToken(dynamic data) {
    if (data is! Map) {
      throw const AppError(
        code: 'auth_error',
        message: 'Invalid auth response.',
      );
    }

    final normalized = normalizeJsonKeys(data.cast<String, dynamic>());
    final value = normalized['accessToken'] ?? normalized['access_token'];
    if (value is String && value.isNotEmpty) return value;

    throw const AppError(
      code: 'auth_error',
      message: 'Missing access token.',
    );
  }

  String _extractRefreshToken(dynamic data) {
    if (data is! Map) {
      throw const AppError(
        code: 'auth_error',
        message: 'Invalid auth response.',
      );
    }

    final normalized = normalizeJsonKeys(data.cast<String, dynamic>());
    final value = normalized['refreshToken'] ?? normalized['refresh_token'];
    if (value is String && value.isNotEmpty) return value;

    throw const AppError(
      code: 'auth_error',
      message: 'Missing refresh token.',
    );
  }
}

