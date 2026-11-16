import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/core/widgets/taga_toast.dart';

void main() {
  group('TagaToast Option A Tests', () {
    testWidgets('TagaToast.success displays floating capsule with check icon',
        (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (ctx) => Center(
                child: ElevatedButton(
                  onPressed: () {
                    TagaToast.success(ctx, 'Saved successfully', title: 'Success');
                  },
                  child: const Text('Show Success'),
                ),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Show Success'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 300));

      expect(find.text('Success'), findsOneWidget);
      expect(find.text('Saved successfully'), findsOneWidget);
    });

    testWidgets('TagaToast.error displays crimson capsule with Error title',
        (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (ctx) => Center(
                child: ElevatedButton(
                  onPressed: () {
                    TagaToast.error(ctx, 'Failed to update booking');
                  },
                  child: const Text('Show Error'),
                ),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Show Error'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 300));

      expect(find.text('Error'), findsOneWidget);
      expect(find.text('Failed to update booking'), findsOneWidget);
    });

    testWidgets('TagaToast.exitNotice displays exit notice with EXIT action button',
        (tester) async {
      bool exitTapped = false;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (ctx) => Center(
                child: ElevatedButton(
                  onPressed: () {
                    TagaToast.exitNotice(
                      ctx,
                      duration: const Duration(seconds: 5),
                      onExitTapped: () => exitTapped = true,
                    );
                  },
                  child: const Text('Show Exit Toast'),
                ),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Show Exit Toast'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 300));

      expect(find.text('Exit Application?'), findsOneWidget);
      expect(find.text('Press back or swipe again to exit'), findsOneWidget);
      expect(find.text('EXIT'), findsOneWidget);

      await tester.tap(find.text('EXIT'));
      await tester.pump();
      expect(exitTapped, isTrue);
    });
  });
}
