import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/core/widgets/sign_out_confirm_dialog.dart';

void main() {
  group('SignOutConfirmDialog Tests', () {
    testWidgets('Renders sign-out dialog elements and Stay Signed In pops false',
        (tester) async {
      bool? dialogResult;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (context) => ElevatedButton(
                onPressed: () async {
                  dialogResult = await showDialog<bool>(
                    context: context,
                    builder: (_) => const SignOutConfirmDialog(),
                  );
                },
                child: const Text('Open Sign Out Dialog'),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Open Sign Out Dialog'));
      await tester.pumpAndSettle();

      expect(find.text('Sign Out of TAGA?'), findsOneWidget);
      expect(
        find.text(
          'Are you sure you want to sign out of your account? You will need your credentials to sign back in.',
        ),
        findsOneWidget,
      );
      expect(find.text('Stay Signed In'), findsOneWidget);
      expect(find.text('Sign Out'), findsOneWidget);

      await tester.tap(find.text('Stay Signed In'));
      await tester.pumpAndSettle();

      expect(dialogResult, isFalse);
    });

    testWidgets('Tapping Sign Out pops true', (tester) async {
      bool? dialogResult;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (context) => ElevatedButton(
                onPressed: () async {
                  dialogResult = await showDialog<bool>(
                    context: context,
                    builder: (_) => const SignOutConfirmDialog(),
                  );
                },
                child: const Text('Open Sign Out Dialog'),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Open Sign Out Dialog'));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Sign Out'));
      await tester.pumpAndSettle();

      expect(dialogResult, isTrue);
    });
  });
}
