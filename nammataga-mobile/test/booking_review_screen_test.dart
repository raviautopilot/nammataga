import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/features/tagatowers/domain/booking_model.dart';
import 'package:nammataga_mobile/features/tagatowers/domain/room_model.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/booking_review_screen.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/widgets/booking_step_progress_bar.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/widgets/price_breakdown_card.dart';
import 'package:nammataga_mobile/features/tagatowers/providers/tagatowers_provider.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('BookingReviewScreen Widget Tests', () {
    testWidgets('Renders empty fallback state if no room is selected', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: BookingReviewScreen(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('No Room Selected'), findsOneWidget);
      expect(find.text('Browse Available Rooms'), findsOneWidget);
    });

    testWidgets('Renders full BookingReviewScreen for Self Booking', (tester) async {
      final container = ProviderContainer();
      final now = DateTime(2026, 11, 10);
      final nextDay = DateTime(2026, 11, 11);

      const room = Room(
        id: 'apex-1',
        name: 'Apex Suite 1',
        type: 'apex-suite',
        capacity: 3,
        allowSingleBed: true,
      );

      final notifier = container.read(bookingFlowNotifierProvider.notifier);
      notifier.setDates(now, nextDay);
      notifier.selectRoom(room);
      notifier.setBookingFor('self');
      notifier.setBookerPhone('+91 9876543210');

      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: const MaterialApp(
            home: BookingReviewScreen(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      // 1. Progress Bar
      expect(find.byType(BookingStepProgressBar), findsOneWidget);
      expect(find.text('STEP 3 OF 4'), findsOneWidget);

      // 2. Booking Summary Card Details
      expect(find.text('Apex Suite 1'), findsWidgets);
      expect(find.text('Self Booking'), findsWidgets);
      expect(find.textContaining('10 Nov 2026, 12:00 AM (Midnight)'), findsOneWidget);
      expect(find.textContaining('11 Nov 2026, 12:00 AM (Midnight)'), findsOneWidget);
      expect(find.textContaining('1 Night • 24-Hour Stay Schedule'), findsOneWidget);
      expect(find.text('Total Beds Booked'), findsOneWidget);
      expect(find.text('1 Bed'), findsWidgets);
      expect(find.text('+91 9876543210'), findsOneWidget);

      // 3. Price Breakdown Card
      expect(find.byType(PriceBreakdownCard), findsOneWidget);
      expect(find.text('Advance Payment'), findsWidgets);
      expect(find.text('₹200'), findsWidgets);
      expect(find.textContaining('non-refundable'), findsWidgets);
    });

    testWidgets('Renders Guest Details Roster for Guest Booking', (tester) async {
      final container = ProviderContainer();
      final now = DateTime(2026, 11, 10);
      final nextDay = DateTime(2026, 11, 12); // 2 nights

      const room = Room(
        id: 'gents-dorm',
        name: 'Gents Dormitory',
        type: 'gents-dorm',
        capacity: 4,
        allowSingleBed: true,
      );

      final guests = [
        const GuestDetail(name: 'Ramesh Kumar', age: 29, contact: '9840123456', gender: 'male'),
        const GuestDetail(name: 'Suresh Raina', age: 34, contact: '9840654321', gender: 'male'),
      ];

      final notifier = container.read(bookingFlowNotifierProvider.notifier);
      notifier.setDates(now, nextDay);
      notifier.selectRoom(room);
      notifier.setBookingFor('guest');
      notifier.setBedCount(2);
      notifier.setGuestDetails(guests);
      notifier.setBookerPhone('+91 9123456789');

      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: const MaterialApp(
            home: BookingReviewScreen(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      // Guest Booking Badge
      expect(find.text('Guest Booking'), findsWidgets);

      // Guest Details Roster
      expect(find.text('Guest Details Roster'), findsOneWidget);
      expect(find.text('2 Guests Registered'), findsOneWidget);
      expect(find.text('Ramesh Kumar'), findsOneWidget);
      expect(find.text('Suresh Raina'), findsOneWidget);
      expect(find.text('29 yrs'), findsOneWidget);
      expect(find.text('34 yrs'), findsOneWidget);
      expect(find.text('9840123456'), findsOneWidget);
      expect(find.text('9840654321'), findsOneWidget);

      // Dormitory Advance Rate (₹100 x 2 = ₹200)
      expect(find.text('₹200'), findsWidgets);
      expect(find.text('2 Beds'), findsWidgets);
    });
  });
}
