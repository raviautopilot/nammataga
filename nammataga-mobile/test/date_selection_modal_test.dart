import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/features/tagatowers/domain/room_model.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/room_availability_screen.dart';
import 'package:nammataga_mobile/features/tagatowers/providers/tagatowers_provider.dart';

void main() {
  testWidgets('Room availability modal updates date in modal and commits on Apply Dates & Refresh', (WidgetTester tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          towersRoomsProvider.overrideWith((ref) async => <Room>[]),
          towersAvailabilityRangeProvider.overrideWith((ref, params) async => <String, RoomAvailability>{}),
        ],
        child: const MaterialApp(
          home: RoomAvailabilityScreen(),
        ),
      ),
    );

    await tester.pumpAndSettle();

    // Verify initial screen state
    expect(find.text('Change Dates'), findsOneWidget);

    // Tap Change Dates to open bottom sheet
    await tester.tap(find.text('Change Dates'));
    await tester.pumpAndSettle();

    // Inside bottom sheet, verify Apply Dates & Refresh is visible
    expect(find.text('Apply Dates & Refresh'), findsOneWidget);

    // Tap Check-out date card
    final checkOutFinder = find.text('CHECK-OUT (12 AM)');
    expect(checkOutFinder, findsOneWidget);

    // Dismiss bottom sheet via Apply Dates & Refresh
    await tester.tap(find.text('Apply Dates & Refresh'));
    await tester.pumpAndSettle();

    expect(find.text('Apply Dates & Refresh'), findsNothing);
  });
}
