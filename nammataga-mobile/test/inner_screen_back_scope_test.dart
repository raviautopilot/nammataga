import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nammataga_mobile/core/widgets/inner_screen_back_scope.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('InnerScreenBackScope Tests', () {
    testWidgets('Pressing back on inner screen with no history navigates to /home without exiting',
        (tester) async {
      final List<MethodCall> systemLogs = <MethodCall>[];
      TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
          .setMockMethodCallHandler(SystemChannels.platform, (MethodCall call) async {
        systemLogs.add(call);
        return null;
      });

      final router = GoRouter(
        initialLocation: '/inner',
        routes: [
          GoRoute(
            path: '/home',
            builder: (ctx, state) => const Scaffold(
              body: Text('Home Page'),
            ),
          ),
          GoRoute(
            path: '/inner',
            builder: (ctx, state) => const InnerScreenBackScope(
              child: Scaffold(
                body: Text('Inner Screen Page'),
              ),
            ),
          ),
        ],
      );

      await tester.pumpWidget(
        MaterialApp.router(
          routerConfig: router,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Inner Screen Page'), findsOneWidget);

      // Trigger back navigation via router delegate (as system back does)
      await router.routerDelegate.popRoute();
      await tester.pumpAndSettle();

      // Should safely navigate to Home Page and NOT call SystemNavigator.pop
      expect(find.text('Home Page'), findsOneWidget);
      expect(
        systemLogs.any((call) => call.method == 'SystemNavigator.pop'),
        isFalse,
      );

      TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
          .setMockMethodCallHandler(SystemChannels.platform, null);
    });

    testWidgets('Pressing back on pushed inner screen pops back to previous screen smoothly',
        (tester) async {
      final router = GoRouter(
        initialLocation: '/home',
        routes: [
          GoRoute(
            path: '/home',
            builder: (ctx, state) => const Scaffold(body: Text('Home Page')),
          ),
          GoRoute(
            path: '/inner',
            builder: (ctx, state) => const InnerScreenBackScope(
              child: Scaffold(body: Text('Pushed Inner Page')),
            ),
          ),
        ],
      );

      await tester.pumpWidget(MaterialApp.router(routerConfig: router));
      await tester.pumpAndSettle();
      expect(find.text('Home Page'), findsOneWidget);

      // Push inner screen
      router.push('/inner');
      await tester.pumpAndSettle();
      expect(find.text('Pushed Inner Page'), findsOneWidget);

      // Trigger back navigation (simulating back swipe / gesture)
      await router.routerDelegate.popRoute();
      await tester.pumpAndSettle();

      // Cleanly pops back to Home Page without loop or freeze
      expect(find.text('Home Page'), findsOneWidget);
      expect(find.text('Pushed Inner Page'), findsNothing);
    });

    testWidgets('Inner modal or dialog pops first before navigating to home',
        (tester) async {
      final router = GoRouter(
        initialLocation: '/inner',
        routes: [
          GoRoute(
            path: '/home',
            builder: (ctx, state) => const Scaffold(body: Text('Home Page')),
          ),
          GoRoute(
            path: '/inner',
            builder: (ctx, state) => InnerScreenBackScope(
              child: Scaffold(
                body: Builder(
                  builder: (innerCtx) => ElevatedButton(
                    onPressed: () {
                      showDialog(
                        context: innerCtx,
                        builder: (_) => const AlertDialog(title: Text('Modal Dialog')),
                      );
                    },
                    child: const Text('Open Modal'),
                  ),
                ),
              ),
            ),
          ),
        ],
      );

      await tester.pumpWidget(
        MaterialApp.router(
          routerConfig: router,
        ),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.text('Open Modal'));
      await tester.pumpAndSettle();
      expect(find.text('Modal Dialog'), findsOneWidget);

      // Press back -> pops the dialog
      await router.routerDelegate.popRoute();
      await tester.pumpAndSettle();

      // Dialog pops, inner screen still visible, Home Page not reached yet
      expect(find.text('Modal Dialog'), findsNothing);
      expect(find.text('Open Modal'), findsOneWidget);
      expect(find.text('Home Page'), findsNothing);

      // Press back again -> now navigates to Home Page
      await router.routerDelegate.popRoute();
      await tester.pumpAndSettle();
      expect(find.text('Home Page'), findsOneWidget);
    });
  });
}
