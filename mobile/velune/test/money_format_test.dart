import 'package:flutter_test/flutter_test.dart';

import 'package:velune/core/utils/money_format.dart';

void main() {
  test('formatMoneyMinor formats major/minor with 2 decimals', () {
    expect(formatMoneyMinor(12345), '123.45');
    expect(formatMoneyMinor(-123), '-1.23');
    expect(formatMoneyMinor(0), '0.00');
  });

  test('formatMoneyMinor supports custom fraction digits', () {
    expect(formatMoneyMinor(1234, fractionDigits: 3), '1.234');
    expect(formatMoneyMinor(1234, fractionDigits: 0), '1234');
  });
}

