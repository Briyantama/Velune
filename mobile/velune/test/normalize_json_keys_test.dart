import 'package:flutter_test/flutter_test.dart';

import 'package:velune/core/json/normalize_json_keys.dart';

void main() {
  test('normalizes snake_case to lowerCamel', () {
    final input = <String, dynamic>{
      'budget_id': 'b1',
      'limit_amount_minor': 100,
    };

    final normalized = normalizeJsonKeys(input);

    expect(normalized['budgetId'], 'b1');
    expect(normalized['limitAmountMinor'], 100);
  });

  test('normalizes PascalCase and *ID suffix', () {
    final input = <String, dynamic>{
      'AccountID': 'a1',
      'MonthlyReport': {
        'GeneratedAt': '2026-01-01T00:00:00Z',
      },
    };

    final normalized = normalizeJsonKeys(input);

    expect(normalized['accountId'], 'a1');
    expect(normalized['monthlyReport'], isA<Map>());
    expect(
      (normalized['monthlyReport'] as Map)['generatedAt'],
      '2026-01-01T00:00:00Z',
    );
  });

  test('recursively normalizes nested values', () {
    final input = <String, dynamic>{
      'items': [
        {'UserID': 'u1'},
        {'access_token': 't1'},
      ],
    };

    final normalized = normalizeJsonKeys(input);

    expect(normalized['items'], isA<List>());
    final items = normalized['items'] as List;

    expect((items[0] as Map)['userId'], 'u1');
    expect((items[1] as Map)['accessToken'], 't1');
  });
}

