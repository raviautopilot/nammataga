import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/features/tagatowers/data/tagatowers_repository.dart';
import 'package:nammataga_mobile/features/tagatowers/domain/booking_model.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/cancellation_screen.dart';
import 'package:nammataga_mobile/features/tagatowers/providers/tagatowers_provider.dart';

class MockTagaTowersRepository extends TagaTowersRepository {
  bool shouldFail = false;
  int cancelCallCount = 0;
  String? lastCancelledId;
  List<BookingResponse> activeBookings = [];

  @override
  Future<bool> cancelBooking(String bookingId) async {
    cancelCallCount++;
    lastCancelledId = bookingId;
    if (shouldFail) {
      throw Exception('Server error: cancellation service unavailable');
    }
    return true;
  }

  @override
  Future<List<BookingResponse>> getUserBookings(String bookerId) async {
    return activeBookings;
  }
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  // Reference future date (always eligible)
  final futureCheckIn = DateTime.now().add(const Duration(days: 10));
  final futureCheckOut = DateTime.now().add(const Duration(days: 12));

  // Past/Today check-in date (ineligible)
  final todayCheckIn = DateTime.now();
  final tomorrowCheckOut = DateTime.now().add(const Duration(days: 1));

  final sampleValidBooking = BookingResponse(
    id: 'BK-VALID-1001',
    roomId: 'apex-suite',
    roomName: 'Apex Suite A/C',
    checkInDate: futureCheckIn.toIso8601String().substring(0, 10),
    checkOutDate: futureCheckOut.toIso8601String().substring(0, 10),
    bookerName: 'Dr. Ramesh Kumar',
    bookerId: 'TAGA-108',
    bookingFor: 'self',
    bedCount: 2,
    gender: 'male',
    paymentStatus: 'confirmed',
    advanceAmount: 400,
    bookingStatus: 'upcoming',
  );

  final sampleIneligibleBooking = BookingResponse(
    id: 'BK-LATE-2002',
    roomId: 'gents-dorm',
    roomName: 'Gents Dormitory',
    checkInDate: todayCheckIn.toIso8601String().substring(0, 10),
    checkOutDate: tomorrowCheckOut.toIso8601String().substring(0, 10),
    bookerName: 'Suresh Babu',
    bookerId: 'TAGA-205',
    bookingFor: 'self',
    bedCount: 1,
    gender: 'male',
    paymentStatus: 'confirmed',
    advanceAmount: 100,
    bookingStatus: 'active',
  );

  final sampleAlreadyCancelledBooking = BookingResponse(
    id: 'BK-CANCELLED-3003',
    roomId: 'kurinchi',
    roomName: 'Kurinchi Suite',
    checkInDate: futureCheckIn.toIso8601String().substring(0, 10),
    checkOutDate: futureCheckOut.toIso8601String().substring(0, 10),
    bookerName: 'Priya Sundar',
    bookerId: 'TAGA-309',
    bookingFor: 'guest',
    bedCount: 1,
    gender: 'female',
    paymentStatus: 'cancelled',
    advanceAmount: 200,
    bookingStatus: 'cancelled',
  );

  Widget createTestWidget({
    required BookingResponse? booking,
    String? bookingId,
    required MockTagaTowersRepository repository,
  }) {
    return ProviderScope(
      overrides: [
        tagaTowersRepositoryProvider.overrideWithValue(repository),
      ],
      child: MaterialApp(
        home: CancellationScreen(
          booking: booking,
          bookingId: bookingId,
        ),
      ),
    );
  }

  group('CancellationScreen Widget & Verification Tests', () {
    testWidgets('1. Displays booking summary, room name, dates, beds, and non-refundable advance fee',
        (WidgetTester tester) async {
      tester.view.physicalSize = const Size(800, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final mockRepo = MockTagaTowersRepository();

      await tester.pumpWidget(
        createTestWidget(
          booking: sampleValidBooking,
          repository: mockRepo,
        ),
      );
      await tester.pumpAndSettle();

      // Header
      expect(find.text('Cancel Reservation'), findsOneWidget);
      expect(find.text('TAGA Towers Egmore'), findsOneWidget);

      // Booking Reference ID
      expect(find.text('BK-VALID-1001'), findsOneWidget);

      // Room name & Bed count
      expect(find.text('Apex Suite A/C'), findsWidgets);
      expect(find.text('2 Beds'), findsOneWidget);

      // Booker details
      expect(find.text('Dr. Ramesh Kumar (SELF)'), findsOneWidget);

      // Advance Paid
      expect(find.text('₹400'), findsWidgets);

      // No-refund alert notice
      expect(find.text('No Refund Will Be Provided'), findsOneWidget);
      expect(
        find.textContaining('advance payment of ₹400 is strictly non-refundable upon cancellation'),
        findsOneWidget,
      );

      // Checkbox is displayed
      expect(
        find.text('I understand that cancelling this reservation is irreversible and that no refund will be provided.'),
        findsOneWidget,
      );

      // "Yes, Cancel Booking" and "No, Keep Booking"
      expect(find.text('Yes, Cancel Booking'), findsOneWidget);
      expect(find.text('No, Keep Booking'), findsOneWidget);
    });

    testWidgets('2. Check-in date on or after today displays warning and disables cancellation',
        (WidgetTester tester) async {
      tester.view.physicalSize = const Size(800, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final mockRepo = MockTagaTowersRepository();

      await tester.pumpWidget(
        createTestWidget(
          booking: sampleIneligibleBooking,
          repository: mockRepo,
        ),
      );
      await tester.pumpAndSettle();

      // Verification warning must be prominently displayed
      expect(
        find.text('Cannot cancel booking on or after check-in date.'),
        findsOneWidget,
      );

      // Terms acknowledgement checkbox should NOT be available
      expect(
        find.text('I understand that cancelling this reservation is irreversible and that no refund will be provided.'),
        findsNothing,
      );

      // Tapping "Yes, Cancel Booking" does nothing (button disabled)
      await tester.tap(find.text('Yes, Cancel Booking'));
      await tester.pumpAndSettle();

      expect(mockRepo.cancelCallCount, 0);
      expect(find.text('Confirm Cancellation'), findsNothing);
    });

    testWidgets('3. Already cancelled booking displays notice and disables cancellation',
        (WidgetTester tester) async {
      tester.view.physicalSize = const Size(800, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final mockRepo = MockTagaTowersRepository();

      await tester.pumpWidget(
        createTestWidget(
          booking: sampleAlreadyCancelledBooking,
          repository: mockRepo,
        ),
      );
      await tester.pumpAndSettle();

      // Notice displayed
      expect(find.text('Booking Already Cancelled'), findsOneWidget);

      // Checkbox should NOT be available
      expect(
        find.text('I understand that cancelling this reservation is irreversible and that no refund will be provided.'),
        findsNothing,
      );

      // Tapping "Yes, Cancel Booking" does nothing
      await tester.tap(find.text('Yes, Cancel Booking'));
      await tester.pumpAndSettle();

      expect(mockRepo.cancelCallCount, 0);
    });

    testWidgets('4. Checkbox toggles "Yes, Cancel Booking" button enabled/disabled state',
        (WidgetTester tester) async {
      tester.view.physicalSize = const Size(800, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final mockRepo = MockTagaTowersRepository();

      await tester.pumpWidget(
        createTestWidget(
          booking: sampleValidBooking,
          repository: mockRepo,
        ),
      );
      await tester.pumpAndSettle();

      // Initially checkbox is unchecked: tapping button should NOT open confirm modal
      await tester.tap(find.text('Yes, Cancel Booking'));
      await tester.pumpAndSettle();
      expect(find.text('Confirm Cancellation'), findsNothing);

      // Tap on the terms checkbox / text
      await tester.tap(
        find.text('I understand that cancelling this reservation is irreversible and that no refund will be provided.'),
      );
      await tester.pumpAndSettle();

      // Now tapping "Yes, Cancel Booking" opens confirmation modal
      await tester.tap(find.text('Yes, Cancel Booking'));
      await tester.pumpAndSettle();

      expect(find.text('Confirm Cancellation'), findsOneWidget);
      expect(find.textContaining('Reminder: The advance payment of ₹400 is strictly non-refundable'), findsOneWidget);

      // Dismiss dialog
      await tester.tap(find.text('Keep Booking'));
      await tester.pumpAndSettle();
      expect(find.text('Confirm Cancellation'), findsNothing);
    });

    testWidgets('5. Successful cancellation calls repository and shows success card',
        (WidgetTester tester) async {
      tester.view.physicalSize = const Size(800, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final mockRepo = MockTagaTowersRepository();

      await tester.pumpWidget(
        createTestWidget(
          booking: sampleValidBooking,
          repository: mockRepo,
        ),
      );
      await tester.pumpAndSettle();

      // Check terms checkbox
      await tester.tap(find.byType(Checkbox));
      await tester.pumpAndSettle();

      // Tap "Yes, Cancel Booking"
      await tester.tap(find.text('Yes, Cancel Booking'));
      await tester.pumpAndSettle();

      // Confirm in dialog
      expect(find.text('Confirm Cancellation'), findsOneWidget);
      final confirmButtons = find.widgetWithText(ElevatedButton, 'Yes, Cancel Booking');
      await tester.tap(confirmButtons);
      await tester.pumpAndSettle();

      // Repository cancelBooking should have been called with booking ID
      expect(mockRepo.cancelCallCount, 1);
      expect(mockRepo.lastCancelledId, 'BK-VALID-1001');

      // Success view displayed
      expect(find.text('Booking Cancelled'), findsOneWidget);
      expect(find.text('Your reservation has been cancelled.'), findsOneWidget);
      expect(find.text('Return to My Bookings'), findsOneWidget);
    });

    testWidgets('6. Failed cancellation displays error message and prominent "Retry Cancellation" button',
        (WidgetTester tester) async {
      tester.view.physicalSize = const Size(800, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final mockRepo = MockTagaTowersRepository()..shouldFail = true;

      await tester.pumpWidget(
        createTestWidget(
          booking: sampleValidBooking,
          repository: mockRepo,
        ),
      );
      await tester.pumpAndSettle();

      // Check checkbox
      await tester.tap(find.byType(Checkbox));
      await tester.pumpAndSettle();

      // Tap "Yes, Cancel Booking" and confirm
      await tester.tap(find.text('Yes, Cancel Booking'));
      await tester.pumpAndSettle();

      final confirmButton = find.widgetWithText(ElevatedButton, 'Yes, Cancel Booking');
      await tester.tap(confirmButton);
      await tester.pumpAndSettle();

      // Failure error card should appear
      expect(find.text('Cancellation Failed'), findsOneWidget);
      expect(find.textContaining('Server error: cancellation service unavailable'), findsOneWidget);

      // Prominent "Retry Cancellation" button must be present
      expect(find.text('Retry Cancellation'), findsOneWidget);

      // Make repository succeed and tap Retry
      mockRepo.shouldFail = false;
      await tester.tap(find.text('Retry Cancellation'));
      await tester.pumpAndSettle();

      // Confirmation dialog opens for retry
      expect(find.text('Confirm Cancellation'), findsOneWidget);
      await tester.tap(find.widgetWithText(ElevatedButton, 'Yes, Cancel Booking'));
      await tester.pumpAndSettle();

      expect(mockRepo.cancelCallCount, 2);
      expect(find.text('Booking Cancelled'), findsOneWidget);
    });
  });
}
