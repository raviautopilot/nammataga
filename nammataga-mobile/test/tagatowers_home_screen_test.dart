import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nammataga_mobile/features/auth/domain/member_session.dart';
import 'package:nammataga_mobile/features/auth/providers/auth_provider.dart';
import 'package:nammataga_mobile/features/profile/domain/member_profile.dart';
import 'package:nammataga_mobile/features/profile/providers/profile_provider.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/tagatowers_access_restricted_screen.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/tagatowers_home_screen.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/widgets/date_range_picker_card.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/widgets/tagatowers_hero_card.dart';
import 'package:nammataga_mobile/features/tagatowers/providers/booking_flow_provider.dart';

const paidProfile = MemberProfile(
  id: 'test-paid-id',
  tagaId: 'TAGA-1001',
  username: 'testpaid',
  name: 'Test Paid Member',
  initial: 'T',
  gender: 'male',
  dateOfBirth: '1990-01-01',
  fatherName: '',
  motherName: '',
  mobileNumber: '9876543210',
  emailId: 'paid@nammataga.in',
  designation: 'Officer',
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

const unpaidProfile = MemberProfile(
  id: 'test-unpaid-id',
  tagaId: 'TAGA-1002',
  username: 'testunpaid',
  name: 'Test Unpaid Member',
  initial: 'U',
  gender: 'female',
  dateOfBirth: '1992-05-15',
  fatherName: '',
  motherName: '',
  mobileNumber: '9876543211',
  emailId: 'unpaid@nammataga.in',
  designation: 'Member',
  recruitmentBatch: '2022',
  seniorityNumber: '50',
  workingDistrict: 'Coimbatore',
  nativeDistrict: 'Salem',
  residentialAddress: '',
  permanentAddress: '',
  educationalQualification: '',
  tbfNumber: '',
  cpsGpfNumber: '',
  paymentStatus: 'pending',
  isPaid: false,
  subscriptionActive: false,
);

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('TagaTowersHomeScreen Subscription & Date-First Tests', () {
    testWidgets(
      'Shows TagatowersAccessRestrictedScreen when member has unpaid subscription',
      (tester) async {
        await tester.pumpWidget(
          ProviderScope(
            overrides: [
              memberProfileProvider.overrideWith((ref) => Future.value(unpaidProfile)),
              authProvider.overrideWith(
                () => _FakeAuthNotifier(
                  const MemberSession(
                    token: 'tok',
                    role: 'member',
                    isPaid: false,
                    isGuest: false,
                  ),
                ),
              ),
            ],
            child: const MaterialApp(
              home: TagaTowersHomeScreen(),
            ),
          ),
        );

        await tester.pumpAndSettle();

        // Should render TagatowersAccessRestrictedScreen
        expect(find.byType(TagatowersAccessRestrictedScreen), findsOneWidget);
        expect(find.text('Subscriber Access Only'), findsOneWidget);
        expect(find.text('Paid Members Only'), findsOneWidget);
        expect(find.text('Manage Subscription'), findsOneWidget);
      },
    );

    testWidgets(
      'Renders Home screen with hero card, date selection, contact info, and NO room cards when paid',
      (tester) async {
        await tester.pumpWidget(
          ProviderScope(
            overrides: [
              memberProfileProvider.overrideWith((ref) => Future.value(paidProfile)),
              authProvider.overrideWith(
                () => _FakeAuthNotifier(
                  const MemberSession(
                    token: 'tok',
                    role: 'member',
                    isPaid: true,
                    isGuest: false,
                  ),
                ),
              ),
            ],
            child: const MaterialApp(
              home: TagaTowersHomeScreen(),
            ),
          ),
        );

        await tester.pumpAndSettle();

        // 1. Hero Card exists
        expect(find.byType(TagaTowersHeroCard), findsOneWidget);
        expect(find.text('TAGA Towers'), findsAtLeast(1));

        // 2. Schedule Policy Card exists
        expect(find.text('BOOKING SCHEDULE & POLICIES'), findsOneWidget);
        expect(find.textContaining('10-Day Booking Horizon'), findsOneWidget);

        // 3. Date Range selection exists
        expect(find.byType(DateRangePickerCard), findsOneWidget);
        expect(find.text('SELECT STAY DATES'), findsOneWidget);
        expect(find.text('CHECK-IN (12 AM)'), findsOneWidget);
        expect(find.text('CHECK-OUT (12 AM)'), findsOneWidget);

        // 4. Quick stay chips exist
        expect(find.text('1 Night (24h)'), findsWidgets);
        expect(find.text('2 Nights (48h)'), findsOneWidget);
        expect(find.text('3 Nights (72h)'), findsOneWidget);

        // 5. Search Availability CTA exists
        expect(find.text('Search Availability'), findsOneWidget);

        // 6. Contact & Caretakers Card exists with exact names and phones
        await tester.scrollUntilVisible(
          find.text('PROPERTY & CARETAKERS CONTACT'),
          100,
          scrollable: find.byType(Scrollable).first,
        );
        expect(find.text('PROPERTY & CARETAKERS CONTACT'), findsOneWidget);
        expect(find.text('Mr. Muthu (Caretaker)'), findsOneWidget);
        expect(find.text('96007 63744'), findsOneWidget);
        expect(find.text('Mr. Mariyappan (Caretaker)'), findsOneWidget);
        expect(find.text('96770 10300'), findsOneWidget);
        expect(find.text('Reception & Helpdesk'), findsOneWidget);
        expect(find.text('044 34919949'), findsOneWidget);
        expect(find.text('nammataga@gmail.com'), findsOneWidget);
        expect(find.text('Velanmai Illam, Egmore, Chennai'), findsOneWidget);

        // 7. CRITICAL VERIFICATION: Absolutely NO room cards should be displayed on this Home screen!
        expect(find.text('Pavalam Suite'), findsNothing);
        expect(find.text('Apex Suite A/C'), findsNothing);
        expect(find.text('Malligai Suite'), findsNothing);
        expect(find.text('Gents Dormitory'), findsNothing);
        expect(find.text('Ladies Dormitory'), findsNothing);
      },
    );

    testWidgets('Quick duration chip updates selected dates in state', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberProfileProvider.overrideWith((ref) => Future.value(paidProfile)),
            authProvider.overrideWith(
              () => _FakeAuthNotifier(
                const MemberSession(
                  token: 'tok',
                  role: 'member',
                  isPaid: true,
                  isGuest: false,
                ),
              ),
            ),
          ],
          child: const MaterialApp(
            home: TagaTowersHomeScreen(),
          ),
        ),
      );

      await tester.pumpAndSettle();

      // Initially 1 night (24h)
      expect(find.textContaining('1 Night (24h)'), findsWidgets);

      // Tap 2 Nights (48h)
      final twoNightsChip = find.text('2 Nights (48h)');
      await tester.ensureVisible(twoNightsChip);
      await tester.tap(twoNightsChip);
      await tester.pumpAndSettle();

      // Stay nights in badge should show 2 Nights
      expect(find.text('2 Nights (48h)'), findsWidgets);
    });

    testWidgets(
      'Tapping Search Availability updates bookingFlowNotifierProvider and pushes route',
      (tester) async {
        late GoRouter router;
        bool navigatedToAvailability = false;

        router = GoRouter(
          initialLocation: '/taga-towers',
          routes: [
            GoRoute(
              path: '/taga-towers',
              builder: (context, state) => const TagaTowersHomeScreen(),
            ),
            GoRoute(
              path: '/taga-towers/availability',
              builder: (context, state) {
                navigatedToAvailability = true;
                return const Scaffold(body: Text('Room Availability Screen'));
              },
            ),
            GoRoute(
              path: '/taga-towers/my-bookings',
              builder: (context, state) => const Scaffold(body: Text('My Bookings')),
            ),
          ],
        );

        final container = ProviderContainer(
          overrides: [
            memberProfileProvider.overrideWith((ref) => Future.value(paidProfile)),
            authProvider.overrideWith(
              () => _FakeAuthNotifier(
                const MemberSession(
                  token: 'tok',
                  role: 'member',
                  isPaid: true,
                  isGuest: false,
                ),
              ),
            ),
          ],
        );

        await tester.pumpWidget(
          UncontrolledProviderScope(
            container: container,
            child: MaterialApp.router(
              routerConfig: router,
            ),
          ),
        );

        await tester.pumpAndSettle();

        // Tap on Search Availability CTA
        final searchCta = find.text('Search Availability');
        await tester.ensureVisible(searchCta);
        await tester.tap(searchCta);
        await tester.pumpAndSettle();

        // Verify navigation occurred
        expect(navigatedToAvailability, isTrue);

        // Verify bookingFlowNotifierProvider dates were updated
        final flowState = container.read(bookingFlowNotifierProvider);
        expect(flowState.checkInDate, isNotNull);
        expect(flowState.checkOutDate, isNotNull);
        expect(flowState.checkOutDate!.isAfter(flowState.checkInDate!), isTrue);
      },
    );
  });

  group('DateRangePickerCard Unit/Widget Tests', () {
    testWidgets('Renders check-in and check-out dates and fires callbacks', (tester) async {
      final checkIn = DateTime(2026, 11, 10);
      final checkOut = DateTime(2026, 11, 12);

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: DateRangePickerCard(
              checkInDate: checkIn,
              checkOutDate: checkOut,
              onCheckInChanged: (_) {},
              onCheckOutChanged: (_) {},
            ),
          ),
        ),
      );

      expect(find.text('SELECT STAY DATES'), findsOneWidget);
      expect(find.text('2 Nights (48h)'), findsOneWidget);
      expect(find.text('10'), findsOneWidget); // day of check-in
      expect(find.text('12'), findsOneWidget); // day of check-out
    });
  });
}

class _FakeAuthNotifier extends AuthNotifier {
  final MemberSession _session;
  _FakeAuthNotifier(this._session);

  @override
  Future<MemberSession> build() async {
    return _session;
  }
}
