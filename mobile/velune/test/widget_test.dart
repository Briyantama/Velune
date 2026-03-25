// Basic smoke test to ensure the widget tree boots.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:velune/presentation/screens/boot_screen.dart';

void main() {
  testWidgets('BootScreen shows a progress indicator', (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: BootScreen()));

    expect(find.byType(CircularProgressIndicator), findsOneWidget);
  });
}
