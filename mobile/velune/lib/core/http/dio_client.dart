import 'dart:async';

import 'package:dio/dio.dart';

import '../env/app_env.dart';
import '../storage/token_store.dart';
import 'app_error.dart';
import 'correlation_id.dart';

const String _extraSkipAuth = 'skipAuth';
const String _extraSkipTokenRefresh = 'skipTokenRefresh';
const String _extraTokenRefreshRetried = 'tokenRefreshRetried';

class DioClient {
  final Dio dio;
  final TokenStore tokenStore;

  DioClient._({
    required this.dio,
    required this.tokenStore,
  });

  static DioClient create({required TokenStore tokenStore}) {
    final dio = Dio(
      BaseOptions(
        baseUrl: AppEnv.gatewayBaseUrl,
        connectTimeout: const Duration(seconds: 15),
        receiveTimeout: const Duration(seconds: 30),
      ),
    );

    final client = DioClient._(dio: dio, tokenStore: tokenStore);
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: client._onRequest,
        onError: client._onError,
      ),
    );

    return client;
  }

  Future<void> _onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final extra = options.extra;
    final skipAuth = extra[_extraSkipAuth] == true;

    final headers = options.headers;
    final existingCid =
        headers['X-Correlation-ID']?.toString() ?? headers['x-correlation-id']?.toString();
    headers['X-Correlation-ID'] = ensureCorrelationId(existingCid);

    if (!skipAuth) {
      final accessToken = await tokenStore.readAccessToken();
      if (accessToken != null && accessToken.isNotEmpty) {
        headers['Authorization'] = 'Bearer $accessToken';
      }
    }

    handler.next(options);
  }

  Future<void> _onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    final requestOptions = err.requestOptions;
    final extra = requestOptions.extra;

    final statusCode = err.response?.statusCode;
    final skipTokenRefresh = extra[_extraSkipTokenRefresh] == true;
    final alreadyRetried = extra[_extraTokenRefreshRetried] == true;

    final shouldRefresh =
        statusCode == 401 && !skipTokenRefresh && !alreadyRetried;

    if (shouldRefresh) {
      extra[_extraTokenRefreshRetried] = true;

      try {
        await _refreshTokens();

        // Retry the original request once, with the updated Authorization header.
        final retryOptions = Options(
          method: requestOptions.method,
          headers: requestOptions.headers,
          extra: extra,
          responseType: requestOptions.responseType,
          contentType: requestOptions.contentType,
          followRedirects: requestOptions.followRedirects,
          validateStatus: requestOptions.validateStatus,
        );

        final retryResponse = await dio.request(
          requestOptions.path,
          data: requestOptions.data,
          queryParameters: requestOptions.queryParameters,
          cancelToken: requestOptions.cancelToken,
          options: retryOptions,
        );

        handler.resolve(retryResponse);
        return;
      } catch (_) {
        await tokenStore.clearTokens();

        final authExpired = AppError(
          code: 'auth_expired',
          message: 'Session expired. Please log in again.',
          statusCode: statusCode,
        );

        handler.next(err.copyWith(error: authExpired));
        return;
      }
    }

    // For all other failures, normalize backend `{ code, message }` errors into `AppError`.
    final appError = AppError.fromDioError(err);
    handler.next(err.copyWith(error: appError));
  }

  Future<void> _refreshTokens() async {
    final refreshToken = await tokenStore.readRefreshToken();
    if (refreshToken == null || refreshToken.isEmpty) {
      throw AppError(
        code: 'auth_expired',
        message: 'Missing refresh token.',
        statusCode: 401,
      );
    }

    // Refresh token flow should not itself trigger auth header attachment or refresh recursion.
    final refreshResponse = await dio.post(
      '/api/v1/auth/refresh',
      data: {'refresh_token': refreshToken},
      options: Options(
        extra: const {
          _extraSkipAuth: true,
          _extraSkipTokenRefresh: true,
        },
      ),
    );

    final data = refreshResponse.data;
    if (data is! Map) {
      throw AppError(
        code: 'auth_expired',
        message: 'Invalid refresh response.',
        statusCode: 401,
      );
    }

    final accessToken = _extractString(data, const [
      'access_token',
      'accessToken',
    ]);
    final newRefreshToken = _extractString(data, const [
      'refresh_token',
      'refreshToken',
    ]);

    if (accessToken == null || newRefreshToken == null) {
      throw AppError(
        code: 'auth_expired',
        message: 'Invalid refresh response.',
        statusCode: 401,
      );
    }

    await tokenStore.writeTokens(
      accessToken: accessToken,
      refreshToken: newRefreshToken,
    );
  }

  String? _extractString(Map<String, dynamic> map, List<String> keys) {
    for (final key in keys) {
      final value = map[key];
      if (value is String && value.trim().isNotEmpty) return value.trim();
    }
    return null;
  }
}

