import 'package:flutter_test/flutter_test.dart';

import 'package:velune/core/http/app_error.dart';

void main() {
  test('AppError.fromBackendJson extracts code and message', () {
    final err = AppError.fromBackendJson(
      {
        'code': 'bad_request',
        'message': 'Invalid input',
      },
      statusCode: 400,
    );

    expect(err.code, 'bad_request');
    expect(err.message, 'Invalid input');
    expect(err.statusCode, 400);
  });

  test('AppError.fromBackendJson falls back when fields are missing', () {
    final err = AppError.fromBackendJson(
      {
        'message': 'Something broke',
      },
      statusCode: 500,
      defaultCode: 'backend_error',
      defaultMessage: 'Request failed',
    );

    expect(err.code, 'backend_error');
    expect(err.message, 'Something broke');
    expect(err.statusCode, 500);
  });

  test('AppError.fromBackendJson uses defaults for non-Map payloads', () {
    final err = AppError.fromBackendJson(
      'not a map',
      statusCode: 400,
      defaultCode: 'backend_error',
      defaultMessage: 'Request failed',
    );

    expect(err.code, 'backend_error');
    expect(err.message, 'Request failed');
    expect(err.statusCode, 400);
  });
}

