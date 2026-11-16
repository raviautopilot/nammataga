import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/core/widgets/app_exit_confirm_dialog.dart';
import 'package:nammataga_mobile/core/widgets/app_exit_scope.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('AppExitConfirmDialog Tests', () {
    testWidgets('Renders exit dialog elements and Stay in App pops false', (tester) async {
      bool? dialogResult;

      await tester.pumpWidget(
        MaterialApp(
          home: Builder(
            builder: (context) {
              return Scaffold(
                body: ElevatedButton(
                  onPressed: () async {
                    dialogResult = await showDialog<bool>(
                      context: context,
                      builder: (ctx) => const AppExitConfirmDialog(),
                    );
                  },
                  child: const Text('Open Dialog'),
                ),
              );
            },
          ),
        ),
      );

      // Open dialog
      await tester.tap(find.text('Open Dialog'));
      await tester.pumpAndSettle();

      expect(find.text('Exit Application?'), findsOneWidget);
      expect(
        find.text(
          'Are you sure you want to close TAGA? Any unsaved entries may be lost.',
        ),
        findsOneWidget,
      );
      expect(find.text('Stay in App'), findsOneWidget);
      expect(find.text('Exit Now'), findsOneWidget);

      // Tap Stay in App
      await tester.tap(find.text('Stay in App'));
      await tester.pumpAndSettle();

      expect(find.text('Exit Application?'), findsNothing);
      expect(dialogResult, isFalse);
    });

    testWidgets('Tapping Exit Now pops true', (tester) async {
      bool? dialogResult;

      await tester.pumpWidget(
        MaterialApp(
          home: Builder(
            builder: (context) {
              return Scaffold(
                body: ElevatedButton(
                  onPressed: () async {
                    dialogResult = await showDialog<bool>(
                      context: context,
                      builder: (ctx) => const AppExitConfirmDialog(),
                    );
                  },
                  child: const Text('Open Dialog'),
                ),
              );
            },
          ),
        ),
      );

      await tester.tap(find.text('Open Dialog'));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Exit Now'));
      await tester.pumpAndSettle();

      expect(dialogResult, isTrue);
    });
  });

  group('AppExitScope Tests', () {
    testWidgets('First back press shows Exit Toast and triggering exit confirmation dialog works', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: AppExitScope(
              child: Builder(
                builder: (context) => Center(
                  child: ElevatedButton(
                    onPressed: () => AppExitScope.showExitConfirmation(context),
                    child: const Text('Direct Popup'),
                  ),
                ),
              ),
            ),
          ),
        ),
      );

      expect(find.text('Direct Popup'), findsOneWidget);

      // Simulate system back button invocation (1st press)
      final dynamic widgetsAppState = tester.state(find.byType(WidgetsApp));
      await widgetsAppState.didPopRoute();
      await tester.pump(const Duration(milliseconds: 300));

      // Verify the toast message is shown with Option A styling
      expect(find.text('Exit Application?'), findsOneWidget);
      expect(find.text('Press back or swipe again to exit'), findsOneWidget);
      expect(find.text('EXIT'), findsOneWidget);

      // Direct popup via helper
      await tester.tap(find.text('Direct Popup'));
      await tester.pumpAndSettle();

      expect(find.text('Stay in App'), findsOneWidget);
      expect(find.text('Exit Now'), findsOneWidget);

      // Dismiss dialog
      await tester.tap(find.text('Stay in App'));
      await tester.pumpAndSettle();
      expect(find.text('Stay in App'), findsNothing);
    });

    testWidgets('Second back press within 5s cooldown shows Exit Confirmation dialog and exits', (tester) async {
      final List<MethodCall> log = <MethodCall>[];
      TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
          .setMockMethodCallHandler(SystemChannels.platform, (MethodCall methodCall) async {
        log.add(methodCall);
        return null;
      });

      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: AppExitScope(
              doubleBackCooldown: Duration(seconds: 5),
              child: Center(child: Text('Home Content')),
            ),
          ),
        ),
      );

      final dynamic widgetsAppState = tester.state(find.byType(WidgetsApp));

      // First back press -> Toast shown
      await widgetsAppState.didPopRoute();
      await tester.pump(const Duration(milliseconds: 300));
      expect(find.text('Press back or swipe again to exit'), findsOneWidget);

      // Second back press within 5s -> Shows Exit Confirmation dialog!
      await widgetsAppState.didPopRoute();
      await tester.pumpAndSettle();

      expect(find.text('Exit Application?'), findsOneWidget);
      expect(find.text('Exit Now'), findsOneWidget);
      expect(find.text('Stay in App'), findsOneWidget);

      // Confirm Exit
      await tester.tap(find.text('Exit Now'));
      await tester.pumpAndSettle();

      // Verify SystemNavigator.pop was called
      expect(
        log.any((call) => call.method == 'SystemNavigator.pop'),
        isTrue,
      );

      // Clean up mock
      TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
          .setMockMethodCallHandler(SystemChannels.platform, null);
    });

    testWidgets('When inner subroute can pop, inner route pops first', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: AppExitScope(
              child: Builder(
                builder: (innerContext) {
                  return ElevatedButton(
                    onPressed: () {
                      Navigator.of(innerContext).push(
                        MaterialPageRoute(
                          builder: (_) => const Scaffold(
                            body: Text('Subpage Content'),
                          ),
                        ),
                      );
                    },
                    child: const Text('Go to Subpage'),
                  );
                },
              ),
            ),
          ),
        ),
      );

      // Go to subpage
      await tester.tap(find.text('Go to Subpage'));
      await tester.pumpAndSettle();
      expect(find.text('Subpage Content'), findsOneWidget);

      // Trigger back button
      final dynamic widgetsAppState = tester.state(find.byType(WidgetsApp));
      await widgetsAppState.didPopRoute();
      await tester.pumpAndSettle();

      // Subpage is popped back to root page, and no exit toast was triggered
      expect(find.text('Subpage Content'), findsNothing);
      expect(find.text('Go to Subpage'), findsOneWidget);
      expect(find.text('Press back or swipe again to exit'), findsNothing);
    });
  });
}
