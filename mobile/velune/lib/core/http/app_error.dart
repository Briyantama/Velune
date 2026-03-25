import 'package:dio/dio.dart';

/// UI-friendly error that normalizes backend `{ code, message }` responses.
class AppError implements Exception {
  final String code;
  final String message;
  final int? statusCode;

  const AppError({
    required this.code,
    required this.message,
    this.statusCode,
  });

  static AppError fromBackendJson(
    dynamic data, {
    int? statusCode,
    String defaultCode = 'backend_error',
    String defaultMessage = 'Request failed',
  }) {
    if (data is Map) {
      final codeValue = data['code'];
      final messageValue = data['message'];

      final code = codeValue?.toString();
      final message = messageValue?.toString();

      return AppError(
        code: code?.trim().isNotEmpty == true ? code!.trim() : defaultCode,
        message: message?.trim().isNotEmpty == true ? message!.trim() : defaultMessage,
        statusCode: statusCode,
      );
    }

    return AppError(
      code: defaultCode,
      message: defaultMessage,
      statusCode: statusCode,
    );
  }

  static AppError fromDioError(DioException err) {
    final statusCode = err.response?.statusCode;
    final defaultStatusMessage = err.message ?? 'Request failed';

    // Prefer explicit backend payload if present.
    final data = err.response?.data;
    if (data != null) {
      return AppError.fromBackendJson(
        data,
        statusCode: statusCode,
        defaultMessage: defaultStatusMessage,
      );
    }

    return AppError(
      code: 'network_error',
      message: defaultStatusMessage,
      statusCode: statusCode,
    );
  }
}

