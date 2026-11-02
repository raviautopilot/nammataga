import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/features/auth/domain/member_session.dart';
import 'package:nammataga_mobile/features/auth/providers/auth_provider.dart';
import 'package:nammataga_mobile/features/profile/domain/member_profile.dart';
import 'package:nammataga_mobile/features/profile/providers/profile_provider.dart';
import 'package:nammataga_mobile/features/tagatowers/domain/booking_model.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/my_bookings_screen.dart';
import 'package:nammataga_mobile/features/tagatowers/providers/tagatowers_provider.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  const mockProfile = MemberProfile(
    id: 'mem_123',
    tagaId: 'TAGA_0010',
    username: 'member10',
    name: 'Member Test',
    initial: 'M',
    gender: 'Male',
    dateOfBirth: '1990-01-01',
    fatherName: 'Father',
    motherName: 'Mother',
    mobileNumber: '9944637254',
    emailId: 'member@test.com',
    designation: 'Engineer',
    recruitmentBatch: '2020',
    seniorityNumber: '100',
    workingDistrict: 'Chennai',
    nativeDistrict: 'Chennai',
    residentialAddress: 'Egmore, Chennai',
    permanentAddress: 'Egmore, Chennai',
    educationalQualification: 'B.E',
    tbfNumber: 'TBF100',
    cpsGpfNumber: 'GPF100',
    paymentStatus: 'paid',
    isPaid: true,
    subscriptionActive: true,
  );

  final mockSession = MemberSession(
    token: 'valid_jwt_token',
    role: 'member',
    isPaid: true,
    user: const {
      'id': 'mem_123',
      'tagaId': 'TAGA_0010',
      'name': 'Member Test',
      'email': 'member@test.com',
    },
  );

  // Future booking: check-in is tomorrow
  final tomorrow = DateTime.now().add(const Duration(days: 1));
  final dayAfter = DateTime.now().add(const Duration(days: 2));
  final tomorrowStr =
      '${tomorrow.year}-${tomorrow.month.toString().padLeft(2, '0')}-${tomorrow.day.toString().padLeft(2, '0')}';
  final dayAfterStr =
      '${dayAfter.year}-${dayAfter.month.toString().padLeft(2, '0')}-${dayAfter.day.toString().padLeft(2, '0')}';

  // Past booking: check-in 10 days ago
  final pastStart = DateTime.now().subtract(const Duration(days: 10));
  final pastEnd = DateTime.now().subtract(const Duration(days: 9));
  final pastStartStr =
      '${pastStart.year}-${pastStart.month.toString().padLeft(2, '0')}-${pastStart.day.toString().padLeft(2, '0')}';
  final pastEndStr =
      '${pastEnd.year}-${pastEnd.month.toString().padLeft(2, '0')}-${pastEnd.day.toString().padLeft(2, '0')}';

  final upcomingBooking = BookingResponse(
    id: 'BK_UPCOMING_001',
    roomId: 'apex-1',
    roomName: 'Apex Suite A/C',
    checkInDate: tomorrowStr,
    checkOutDate: dayAfterStr,
    bookerName: 'Member Test',
    bookerId: 'TAGA_0010',
    bookingFor: 'self',
    bedCount: 1,
    paymentStatus: 'confirmed',
    advanceAmount: 200,
    bookingStatus: 'upcoming',
  );

  final activeBooking = BookingResponse(
    id: 'BK_ACTIVE_002',
    roomId: 'gents-dorm',
    roomName: 'Gents Dormitory',
    checkInDate:
        '${DateTime.now().year}-${DateTime.now().month.toString().padLeft(2, '0')}-${DateTime.now().day.toString().padLeft(2, '0')}',
    checkOutDate: tomorrowStr,
    bookerName: 'Member Test',
    bookerId: 'TAGA_0010',
    bookingFor: 'guest',
    bedCount: 2,
    guestDetails: const [
      GuestDetail(name: 'Guest 1', age: 30, contact: '9876543210', gender: 'male'),
      GuestDetail(name: 'Guest 2', age: 32, contact: '9876543211', gender: 'male'),
    ],
    paymentStatus: 'confirmed',
    advanceAmount: 200,
    bookingStatus: 'active',
  );

  final completedBooking = BookingResponse(
    id: 'BK_PAST_003',
    roomId: 'ac-room',
    roomName: 'A/C Room',
    checkInDate: pastStartStr,
    checkOutDate: pastEndStr,
    bookerName: 'Member Test',
    bookerId: 'TAGA_0010',
    bookingFor: 'self',
    bedCount: 1,
    paymentStatus: 'confirmed',
    advanceAmount: 200,
    bookingStatus: 'completed',
  );

  final cancelledBooking = BookingResponse(
    id: 'BK_CANCEL_004',
    roomId: 'kurinchi',
    roomName: 'Kurinchi Suite',
    checkInDate: pastStartStr,
    checkOutDate: pastEndStr,
    bookerName: 'Member Test',
    bookerId: 'TAGA_0010',
    bookingFor: 'self',
    bedCount: 1,
    paymentStatus: 'cancelled',
    advanceAmount: 200,
    bookingStatus: 'cancelled',
  );

  Widget createWidget({
    List<BookingResponse> userBookings = const [],
    List<BookingResponse> pastBookings = const [],
    int initialTab = 0,
  }) {
    return ProviderScope(
      overrides: [
        authProvider.overrideWith(() => _MockAuthNotifier(mockSession)),
        memberProfileProvider.overrideWith((ref) => Future.value(mockProfile)),
        userBookingsProvider('TAGA_0010')
            .overrideWith((ref) => Future.value(userBookings)),
        pastBookingsProvider(null)
            .overrideWith((ref) => Future.value(pastBookings)),
        pastBookingsProvider('2026')
            .overrideWith((ref) => Future.value(pastBookings)),
      ],
      child: MaterialApp(
        home: MyBookingsScreen(initialTabIndex: initialTab),
      ),
    );
  }

  group('MyBookingsScreen UI Structure Tests', () {
    testWidgets('Renders Title and Two Distinct Tabs', (WidgetTester tester) async {
      await tester.pumpWidget(createWidget());
      await tester.pumpAndSettle();

      expect(find.text('My Bookings'), findsOneWidget);
      expect(find.text('Active & Upcoming'), findsOneWidget);
      expect(find.text('Past Bookings'), findsOneWidget);
    });

    testWidgets('Active & Upcoming Tab shows bookings and status badges',
        (WidgetTester tester) async {
      await tester.pumpWidget(createWidget(
        userBookings: [upcomingBooking, activeBooking],
      ));
      await tester.pumpAndSettle();

      // Booking References
      expect(find.textContaining('BK_UPCOMING_'), findsOneWidget);
      expect(find.textContaining('BK_ACTIVE_00'), findsOneWidget);

      // Status badges
      expect(find.text('Upcoming'), findsOneWidget);
      expect(find.text('Active Stay'), findsOneWidget);

      // Room names
      expect(find.text('Apex Suite A/C'), findsWidgets);
      expect(find.text('Gents Dormitory'), findsWidgets);

      // Advance amounts
      expect(find.text('₹200'), findsWidgets);

      // Details buttons
      expect(find.text('View Details'), findsNWidgets(2));

      // Cancellation button ONLY for upcoming booking (isCancellable == true)
      expect(find.text('Cancel Booking'), findsOneWidget);

      // Active stay has started, so cancel button is hidden with policy notice
      expect(find.textContaining('Cannot cancel after check-in'), findsOneWidget);
    });

    testWidgets('Tapping Cancel Booking opens confirmation bottom sheet',
        (WidgetTester tester) async {
      await tester.pumpWidget(createWidget(
        userBookings: [upcomingBooking],
      ));
      await tester.pumpAndSettle();

      expect(find.text('Cancel Booking'), findsOneWidget);
      await tester.tap(find.text('Cancel Booking'));
      await tester.pumpAndSettle();

      // Bottom sheet confirmation
      expect(find.text('Cancel Booking?'), findsOneWidget);
      expect(find.text('This action cannot be undone.'), findsOneWidget);
      expect(find.textContaining('Association Advance Policy'), findsOneWidget);
      expect(find.text('Keep Booking'), findsOneWidget);
      expect(find.text('Yes, Cancel'), findsOneWidget);

      // Dismiss dialog
      await tester.tap(find.text('Keep Booking'));
      await tester.pumpAndSettle();
      expect(find.text('Cancel Booking?'), findsNothing);
    });

    testWidgets('Past Bookings tab displays completed and cancelled stays',
        (WidgetTester tester) async {
      await tester.pumpWidget(createWidget(
        userBookings: [completedBooking, cancelledBooking],
        pastBookings: [completedBooking],
        initialTab: 1,
      ));
      await tester.pumpAndSettle();

      // Year filter chips - exactly 3 years (e.g. 2026, 2025, 2024)
      expect(find.text('Filter by Year:'), findsOneWidget);
      expect(find.text(DateTime.now().year.toString()), findsOneWidget);

      // Badges
      expect(find.text('Completed'), findsOneWidget);
      expect(find.text('Cancelled'), findsOneWidget);

      // Room names
      expect(find.text('A/C Room'), findsWidgets);
      expect(find.text('Kurinchi Suite'), findsWidgets);
    });

    testWidgets('Empty states render when no bookings exist',
        (WidgetTester tester) async {
      await tester.pumpWidget(createWidget(
        userBookings: [],
        pastBookings: [],
        initialTab: 0,
      ));
      await tester.pumpAndSettle();

      expect(find.text('No Active or Upcoming Bookings'), findsOneWidget);
      expect(find.text('Book a Room'), findsOneWidget);

      // Switch to Past tab
      await tester.tap(find.text('Past Bookings'));
      await tester.pumpAndSettle();

      expect(find.text('No Past Bookings'), findsOneWidget);
      expect(find.text('Book a Room'), findsOneWidget);
    });
  });
}

class _MockAuthNotifier extends AuthNotifier {
  final MemberSession session;
  _MockAuthNotifier(this.session);

  @override
  Future<MemberSession> build() async => session;
}
