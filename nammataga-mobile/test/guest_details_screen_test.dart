import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/features/profile/domain/member_profile.dart';
import 'package:nammataga_mobile/features/profile/providers/profile_provider.dart';
import 'package:nammataga_mobile/features/tagatowers/domain/room_model.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/guest_details_screen.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/widgets/booking_step_progress_bar.dart';
import 'package:nammataga_mobile/features/tagatowers/providers/booking_flow_provider.dart';

const dummyProfile = MemberProfile(
  id: 'test-user-id',
  tagaId: 'TAGA-1001',
  username: 'johndoe',
  name: 'John Doe',
  initial: 'M',
  gender: 'male',
  dateOfBirth: '1992-06-15',
  fatherName: '',
  motherName: '',
  mobileNumber: '9876543210',
  emailId: 'john@nammataga.in',
  designation: 'Agriculture Officer',
  recruitmentBatch: '2018',
  seniorityNumber: '42',
  workingDistrict: 'Chennai',
  nativeDistrict: 'Madurai',
  residentialAddress: '',
  permanentAddress: '',
  educationalQualification: 'B.Sc. Agriculture',
  tbfNumber: 'TBF-99',
  cpsGpfNumber: 'CPS-1234',
  paymentStatus: 'paid',
  isPaid: true,
  subscriptionActive: true,
);

const testRoom = Room(
  id: 'apex-1',
  name: 'Apex Suite A/C',
  type: 'apex-suite',
  capacity: 3,
  allowSingleBed: true,
);

const dormRoom = Room(
  id: 'gents-dorm',
  name: 'Gents Dormitory',
  type: 'gents-dorm',
  capacity: 6,
  allowSingleBed: true,
);

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('GuestDetailsScreen Tests', () {
    testWidgets('Renders progress bar with currentStep: 2 and N guest cards',
        (tester) async {
      final container = ProviderContainer(
        overrides: [
          memberProfileProvider.overrideWith((ref) => dummyProfile),
        ],
      );

      // Setup state with 2 beds
      container.read(bookingFlowNotifierProvider.notifier).selectRoom(testRoom);
      container.read(bookingFlowNotifierProvider.notifier).setBookingFor('guest');
      container.read(bookingFlowNotifierProvider.notifier).setBedCount(2);

      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: const MaterialApp(
            home: GuestDetailsScreen(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      // Check BookingStepProgressBar(currentStep: 2)
      expect(find.byType(BookingStepProgressBar), findsOneWidget);
      expect(find.text('STEP 2 OF 4'), findsOneWidget);

      // Exactly 2 guest cards rendered
      expect(find.text('Guest 1'), findsOneWidget);
      expect(find.text('Guest 2'), findsOneWidget);
      expect(find.text('Guest 3'), findsNothing);

      // Check fields for each guest
      expect(find.text('Full Name *'), findsNWidgets(2));
      expect(find.text('Age *'), findsNWidgets(2));
      expect(find.text('Contact Number *'), findsNWidgets(2));
      expect(find.text('Gender *'), findsNWidgets(2));

      // Guest 1 has "Fill as Self" pill button
      expect(find.text('Fill as Self'), findsOneWidget);
    });

    testWidgets('Fill as Self toggle populates Guest 1 and toggling clears it',
        (tester) async {
      final container = ProviderContainer(
        overrides: [
          memberProfileProvider.overrideWith((ref) => dummyProfile),
        ],
      );

      container.read(bookingFlowNotifierProvider.notifier).selectRoom(testRoom);
      container.read(bookingFlowNotifierProvider.notifier).setBookingFor('guest');
      container.read(bookingFlowNotifierProvider.notifier).setBedCount(1);

      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: const MaterialApp(
            home: GuestDetailsScreen(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      // Tap "Fill as Self"
      final fillSelfBtn = find.text('Fill as Self');
      expect(fillSelfBtn, findsOneWidget);
      await tester.tap(fillSelfBtn);
      await tester.pumpAndSettle();

      // Guest 1 should now be auto-filled
      expect(find.text('✓ Self Applied'), findsOneWidget);
      expect(find.text('John Doe M'), findsOneWidget);
      expect(find.text('9876543210'), findsWidgets);

      // Tap again to clear Guest 1
      await tester.tap(find.text('✓ Self Applied'));
      await tester.pumpAndSettle();

      // Guest 1 should now be cleared
      expect(find.text('Fill as Self'), findsOneWidget);
      expect(find.text('John Doe M'), findsNothing);
    });

    testWidgets('Gents Dormitory locks gender to Male', (tester) async {
      final container = ProviderContainer(
        overrides: [
          memberProfileProvider.overrideWith((ref) => dummyProfile),
        ],
      );

      container.read(bookingFlowNotifierProvider.notifier).selectRoom(dormRoom);
      container.read(bookingFlowNotifierProvider.notifier).setBookingFor('guest');
      container.read(bookingFlowNotifierProvider.notifier).setBedCount(1);

      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: const MaterialApp(
            home: GuestDetailsScreen(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      // Should show Locked: MALE badge
      expect(find.text('Locked: MALE'), findsOneWidget);
      expect(find.text('Gents Dormitory Policy'), findsOneWidget);
    });

    testWidgets('Validation fails when required fields are empty',
        (tester) async {
      final container = ProviderContainer(
        overrides: [
          memberProfileProvider.overrideWith((ref) => dummyProfile),
        ],
      );

      container.read(bookingFlowNotifierProvider.notifier).selectRoom(testRoom);
      container.read(bookingFlowNotifierProvider.notifier).setBookingFor('guest');
      container.read(bookingFlowNotifierProvider.notifier).setBedCount(1);

      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: const MaterialApp(
            home: GuestDetailsScreen(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      // Tap Proceed to Review without filling form
      await tester.tap(find.text('Review Booking'));
      await tester.pumpAndSettle();

      // Expect validation errors
      expect(find.text('Full name is required'), findsOneWidget);
      expect(find.text('Age is required'), findsOneWidget);
      expect(find.text('Contact number is required'), findsOneWidget);
    });
  });
}
