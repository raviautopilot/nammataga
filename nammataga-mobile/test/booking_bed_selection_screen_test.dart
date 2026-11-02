import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/features/profile/domain/member_profile.dart';
import 'package:nammataga_mobile/features/profile/providers/profile_provider.dart';
import 'package:nammataga_mobile/features/tagatowers/domain/room_model.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/booking_bed_selection_screen.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/widgets/booking_step_progress_bar.dart';
import 'package:nammataga_mobile/features/tagatowers/providers/booking_flow_provider.dart';

const dummyProfile = MemberProfile(
  id: 'test-id',
  tagaId: 'TAGA-1001',
  username: 'testuser',
  name: 'Test Member',
  initial: 'T',
  gender: 'male',
  dateOfBirth: '1990-01-01',
  fatherName: '',
  motherName: '',
  mobileNumber: '9876543210',
  emailId: 'test@nammataga.in',
  designation: 'Staff',
  recruitmentBatch: '2020',
  seniorityNumber: '1',
  workingDistrict: 'Chennai',
  nativeDistrict: 'Madurai',
  residentialAddress: '',
  permanentAddress: '',
  educationalQualification: '',
  tbfNumber: '',
  cpsGpfNumber: '',
  paymentStatus: 'paid',
  isPaid: true,
  subscriptionActive: true,
);

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('BookingBedSelectionScreen Tests', () {
    testWidgets('Displays fallback empty state when no room is selected',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberProfileProvider.overrideWith((ref) => dummyProfile),
          ],
          child: const MaterialApp(
            home: BookingBedSelectionScreen(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('No Room Selected'), findsOneWidget);
      expect(find.text('Browse Available Rooms'), findsOneWidget);
    });

    testWidgets('Renders all requirements when room is selected',
        (tester) async {
      tester.view.physicalSize = const Size(1080, 2400);
      tester.view.devicePixelRatio = 2.0;
      addTearDown(() => tester.view.resetPhysicalSize());

      final container = ProviderContainer(
        overrides: [
          memberProfileProvider.overrideWith((ref) => dummyProfile),
        ],
      );

      final room = const Room(
        id: 'apex-1',
        name: 'Apex Suite 1',
        type: 'apex-suite',
        capacity: 3,
        allowSingleBed: true,
      );
      final checkIn = DateTime(2026, 11, 10);
      final checkOut = DateTime(2026, 11, 12);

      // Pre-configure flow state
      container
          .read(bookingFlowNotifierProvider.notifier)
          .setDates(checkIn, checkOut);
      container.read(bookingFlowNotifierProvider.notifier).selectRoom(
            room,
            RoomAvailability(
              room: room,
              available: true,
              availableBeds: 2,
            ),
          );

      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: const MaterialApp(
            home: BookingBedSelectionScreen(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      // 1. Progress Bar
      expect(find.byType(BookingStepProgressBar), findsOneWidget);
      expect(find.text('STEP 1 OF 4'), findsOneWidget);
      expect(find.text('Room & Beds'), findsOneWidget);

      // 2. Room & Stay summary
      expect(find.text('Apex Suite 1'), findsOneWidget);
      expect(find.text('Apex Suite A/C'), findsOneWidget);
      expect(find.textContaining('Stay Duration: 2 Day(s)'), findsOneWidget);

      // 3. Booker Mobile Number
      expect(find.text('Booker Mobile Number'), findsOneWidget);

      // 4. Booking Type options
      expect(find.text('Self Booking'), findsOneWidget);
      expect(find.text('Guest Booking'), findsOneWidget);

      // Self Booking defaults to 1 bed (₹200 advance for suite)
      expect(find.text('₹200'), findsWidgets);
      expect(find.text('Advance: ₹200 (1 Bed)'), findsOneWidget);

      // 5. Switching to Guest Booking shows Bed Quantity Selector
      await tester.tap(find.text('Guest Booking'));
      await tester.pumpAndSettle();

      expect(find.text('Number of Beds'), findsOneWidget);
      expect(find.text('1 Bed Selected'), findsOneWidget);

      // Check Review / Guest Details CTA
      expect(find.text('Guest Details'), findsOneWidget);
    });

    testWidgets('Dormitory room shows fixed gender restriction badge',
        (tester) async {
      tester.view.physicalSize = const Size(1080, 2400);
      tester.view.devicePixelRatio = 2.0;
      addTearDown(() => tester.view.resetPhysicalSize());

      final container = ProviderContainer(
        overrides: [
          memberProfileProvider.overrideWith((ref) => dummyProfile),
        ],
      );

      final room = const Room(
        id: 'gents-dorm',
        name: 'TAGA Gents Dormitory',
        type: 'gents-dorm',
        capacity: 4,
        allowSingleBed: true,
      );

      container.read(bookingFlowNotifierProvider.notifier).selectRoom(
            room,
            RoomAvailability(
              room: room,
              available: true,
              availableBeds: 4,
            ),
          );

      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: const MaterialApp(
            home: BookingBedSelectionScreen(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      // Check for Dorm fixed badge
      expect(find.text('Gents Dormitory — Male Only'), findsOneWidget);
      expect(find.text('MALE ONLY'), findsOneWidget);

      // Dormitory advance rate is ₹100
      expect(find.text('Advance: ₹100 (1 Bed)'), findsOneWidget);
    });
  });
}
