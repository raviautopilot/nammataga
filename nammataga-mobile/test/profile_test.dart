import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:nammataga_mobile/features/auth/domain/member_session.dart';
import 'package:nammataga_mobile/features/auth/providers/auth_provider.dart';
import 'package:nammataga_mobile/features/profile/domain/member_notification_model.dart';
import 'package:nammataga_mobile/features/profile/domain/member_profile.dart';
import 'package:nammataga_mobile/features/profile/domain/subscription_plan.dart';
import 'package:nammataga_mobile/features/profile/presentation/account_settings_screen.dart';
import 'package:nammataga_mobile/features/profile/presentation/address_details_screen.dart';
import 'package:nammataga_mobile/features/profile/presentation/announcements_screen.dart';
import 'package:nammataga_mobile/features/profile/presentation/contact_details_screen.dart';
import 'package:nammataga_mobile/features/profile/presentation/employment_details_screen.dart';
import 'package:nammataga_mobile/features/profile/presentation/membership_details_screen.dart';
import 'package:nammataga_mobile/features/profile/presentation/official_details_screen.dart';
import 'package:nammataga_mobile/features/profile/presentation/personal_info_screen.dart';
import 'package:nammataga_mobile/features/profile/presentation/profile_overview_screen.dart';
import 'package:nammataga_mobile/features/profile/presentation/request_profile_edit_screen.dart';
import 'package:nammataga_mobile/features/profile/presentation/subscription_plans_screen.dart';
import 'package:nammataga_mobile/features/profile/presentation/widgets/initials_avatar.dart';
import 'package:nammataga_mobile/features/profile/providers/profile_provider.dart';

final mockMemberProfile = MemberProfile(
  id: 'usr_001',
  tagaId: 'TAGA-2024-0842',
  username: 'sivakumar',
  name: 'R. Sivakumar',
  initial: '',
  gender: 'Male',
  dateOfBirth: '1985-05-15',
  fatherName: 'P. Ramanathan',
  motherName: 'R. Meenakshi',
  mobileNumber: '9876543210',
  emailId: 'sivakumar.taga@gmail.com',
  designation: 'Agricultural Officer',
  recruitmentBatch: '2012',
  seniorityNumber: 'SN-4021',
  workingDistrict: 'Thanjavur',
  nativeDistrict: 'Madurai',
  residentialAddress: 'No. 42, Green Avenue, New Housing Unit, Thanjavur - 613005',
  permanentAddress: '15, South Masi Street, Madurai - 625001',
  educationalQualification: 'B.Sc. (Agriculture), TNAU Coimbatore',
  tbfNumber: 'TBF-8921',
  cpsGpfNumber: 'CPS-548920',
  isPaid: true,
  subscriptionActive: true,
  paymentStatus: 'Paid',
);

class FakeAuthNotifier extends AuthNotifier {
  final MemberSession mockSession;
  FakeAuthNotifier(this.mockSession);

  @override
  Future<MemberSession> build() async => mockSession;
}

void main() {
  group('MemberProfile Domain Model & Initials', () {
    test('Correctly calculates monogram initials from name', () {
      expect(mockMemberProfile.avatarInitials, 'RS');

      final singleName = mockMemberProfile.copyWith(name: 'Kavitha');
      expect(singleName.avatarInitials, 'KA');

      final initialField = mockMemberProfile.copyWith(name: 'Sivakumar', initial: 'R');
      expect(initialField.avatarInitials, 'SR');

      final empty = mockMemberProfile.copyWith(name: '', initial: '');
      expect(empty.avatarInitials, 'T');
    });

    test('Correctly calculates age from dateOfBirth', () {
      expect(mockMemberProfile.calculatedAge, isNotNull);
      expect(mockMemberProfile.calculatedAge, contains('Years'));
    });

    test('Parses from backend JSON structure', () {
      final json = {
        'id': 'usr_99',
        'name': 'Dr. Anitha',
        'initial': 'M',
        'gender': 'Female',
        'date_of_birth': '1990-08-20',
        'father_name': 'M. Murugan',
        'mobile_number': '9443322110',
        'emailId': 'anitha@tn.gov.in',
        'designation': 'Assistant Director of Agriculture',
        'working_district': 'Trichy',
        'tagaId': 'TAGA-1001',
        'username': 'anitha',
        'isPaid': true,
        'subscription_active': true,
      };

      final profile = MemberProfile.fromJson(json);
      expect(profile.name, 'Dr. Anitha');
      expect(profile.initial, 'M');
      expect(profile.gender, 'Female');
      expect(profile.workingDistrict, 'Trichy');
      expect(profile.subscriptionActive, isTrue);
    });
  });

  group('InitialsAvatar Widget', () {
    testWidgets('Renders monogram initials with emerald styling', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: Center(
              child: InitialsAvatar(
                initials: 'RS',
                size: 80,
              ),
            ),
          ),
        ),
      );

      expect(find.text('RS'), findsOneWidget);
    });
  });

  group('Profile Feature Screens Rendering', () {
    testWidgets('ProfileOverviewScreen renders Member Header, TAGA ID, and navigation cards', (tester) async {
      tester.view.physicalSize = const Size(430, 1200);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberProfileProvider.overrideWith((ref) => Future.value(mockMemberProfile)),
            authProvider.overrideWith(() => FakeAuthNotifier(
              MemberSession(
                token: 'jwt_mock',
                role: 'member',
                isPaid: true,
                isGuest: false,
              ),
            )),
          ],
          child: const MaterialApp(home: ProfileOverviewScreen()),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('My Profile'), findsOneWidget);
      expect(find.text('R. Sivakumar'), findsOneWidget);
      expect(find.text('TAGA ID: TAGA-2024-0842'), findsOneWidget);
      expect(find.text('Subscriber Member'), findsOneWidget);
      expect(find.text('Subscription Plans & Dues'), findsOneWidget);
      expect(find.text('Membership Status & Benefits'), findsOneWidget);
      expect(find.text('Personal Information'), findsOneWidget);
      expect(find.text('Contact Details'), findsOneWidget);
      expect(find.text('Employment Details'), findsOneWidget);
      expect(find.text('Address Details'), findsOneWidget);
      expect(find.text('Official Details'), findsOneWidget);
      expect(find.text('Membership Status'), findsOneWidget);
      expect(find.text('Manage Membership & Dues'), findsOneWidget);
      expect(find.text('Request Profile Edit'), findsOneWidget);
      expect(find.byIcon(LucideIcons.settings), findsOneWidget);
    });

    testWidgets('PersonalInfoScreen renders member identity, DOB and parents info', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberProfileProvider.overrideWith((ref) => Future.value(mockMemberProfile)),
          ],
          child: const MaterialApp(home: PersonalInfoScreen()),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('Personal Information'), findsWidgets);
      expect(find.text('R. Sivakumar'), findsOneWidget);
      expect(find.text('Male'), findsOneWidget);
      expect(find.text('P. Ramanathan'), findsOneWidget);
      expect(find.text('R. Meenakshi'), findsOneWidget);
      expect(find.text('Request Profile Edit'), findsWidgets);
    });

    testWidgets('ContactDetailsScreen renders phone and email cards', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberProfileProvider.overrideWith((ref) => Future.value(mockMemberProfile)),
          ],
          child: const MaterialApp(home: ContactDetailsScreen()),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('Contact Details'), findsWidgets);
      expect(find.text('9876543210'), findsOneWidget);
      expect(find.text('sivakumar.taga@gmail.com'), findsOneWidget);
    });

    testWidgets('EmploymentDetailsScreen renders designation, batch, and districts', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberProfileProvider.overrideWith((ref) => Future.value(mockMemberProfile)),
          ],
          child: const MaterialApp(home: EmploymentDetailsScreen()),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('Employment Details'), findsWidgets);
      expect(find.text('Agricultural Officer'), findsOneWidget);
      expect(find.text('2012'), findsOneWidget);
      expect(find.text('SN-4021'), findsOneWidget);
      expect(find.text('Thanjavur'), findsOneWidget);
      expect(find.text('Madurai'), findsOneWidget);
    });

    testWidgets('AddressDetailsScreen renders residential and permanent address cards', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberProfileProvider.overrideWith((ref) => Future.value(mockMemberProfile)),
          ],
          child: const MaterialApp(home: AddressDetailsScreen()),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('Address Details'), findsWidgets);
      expect(find.text('Residential Address'), findsOneWidget);
      expect(find.text('Permanent Address'), findsOneWidget);
      expect(find.text('No. 42, Green Avenue, New Housing Unit, Thanjavur - 613005'), findsOneWidget);
    });

    testWidgets('OfficialDetailsScreen renders educational and financial record identifiers', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberProfileProvider.overrideWith((ref) => Future.value(mockMemberProfile)),
          ],
          child: const MaterialApp(home: OfficialDetailsScreen()),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('Official Details'), findsWidgets);
      expect(find.text('B.Sc. (Agriculture), TNAU Coimbatore'), findsOneWidget);
      expect(find.text('TBF-8921'), findsOneWidget);
      expect(find.text('CPS-548920'), findsOneWidget);
    });

    testWidgets('MembershipDetailsScreen renders crown status, tier, and member benefits', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberProfileProvider.overrideWith((ref) => Future.value(mockMemberProfile)),
          ],
          child: const MaterialApp(home: MembershipDetailsScreen()),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('Membership Details'), findsWidgets);
      expect(find.text('Active'), findsOneWidget);
      expect(find.text('Subscriber Member'), findsOneWidget);
      expect(find.text('Subscriber Benefits'), findsOneWidget);
    });

    testWidgets('RequestProfileEditScreen renders landing page and switches to edit form', (tester) async {
      tester.view.physicalSize = const Size(430, 1200);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberProfileProvider.overrideWith((ref) => Future.value(mockMemberProfile)),
            districtsListProvider.overrideWith((ref) => Future.value(['Chennai', 'Thanjavur', 'Madurai'])),
          ],
          child: const MaterialApp(home: RequestProfileEditScreen()),
        ),
      );

      await tester.pumpAndSettle();

      // Step 1: Landing screen
      expect(find.text('Request Profile Edit'), findsWidgets);
      expect(find.text('You can request changes for:'), findsOneWidget);
      expect(find.text('New Edit Request'), findsOneWidget);

      // Step 2: Open edit form
      await tester.tap(find.text('New Edit Request'));
      await tester.pumpAndSettle();

      expect(find.text('Edit Request Details'), findsOneWidget);
      expect(find.text('Mobile Number'), findsOneWidget);
      expect(find.text('Email Address'), findsOneWidget);
      expect(find.text('Designation'), findsOneWidget);
      expect(find.text('Working District'), findsOneWidget);
      expect(find.text('Submit Edit Request'), findsOneWidget);
    });

    testWidgets('AccountSettingsScreen renders security options, announcements and logout', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            unreadAnnouncementsCountProvider.overrideWith((ref) => Future.value(2)),
            authProvider.overrideWith(() => FakeAuthNotifier(
              MemberSession(
                token: 'jwt_mock',
                role: 'member',
                isPaid: true,
                isGuest: false,
                user: {'emailId': 'sivakumar.taga@gmail.com'},
              ),
            )),
          ],
          child: const MaterialApp(home: AccountSettingsScreen()),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('Account & Settings'), findsOneWidget);
      expect(find.text('Change Password'), findsOneWidget);
      expect(find.text('Announcements'), findsOneWidget);
      expect(find.text('Logout'), findsOneWidget);
    });

    testWidgets('SubscriptionPlansScreen renders backend plans, filters, and payment action', (tester) async {
      tester.view.physicalSize = const Size(430, 1600);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);

      final mockPlans = [
        const SubscriptionPlan(
          id: 'annual-subscription',
          name: 'TAGA Annual Subscription',
          amount: 3500,
          frequency: 'Annual',
          status: 'active',
          allowCustomAmount: false,
          oneTime: false,
          description: 'Official annual subscription supporting association activities',
        ),
        const SubscriptionPlan(
          id: 'legal-fund',
          name: 'TAGA Legal Fund Contribution',
          amount: 1000,
          frequency: 'One-time',
          status: 'active',
          allowCustomAmount: true,
          oneTime: true,
          description: 'Voluntary legal defence and advocacy support',
        ),
      ];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            subscriptionPlansProvider.overrideWith((ref) => Future.value(mockPlans)),
            memberSubscriptionStatusProvider.overrideWith((ref) => Future.value({
              'isPaid': true,
              'subscriptionActive': true,
              'active': true,
            })),
            memberProfileProvider.overrideWith((ref) => Future.value(mockMemberProfile)),
            authProvider.overrideWith(() => FakeAuthNotifier(
              MemberSession(
                token: 'jwt_mock',
                role: 'member',
                isPaid: true,
                isGuest: false,
                user: {'emailId': 'sivakumar.taga@gmail.com', 'name': 'R. Sivakumar', 'tagaId': 'TAGA-2024-0842'},
              ),
            )),
          ],
          child: const MaterialApp(home: SubscriptionPlansScreen()),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('Subscription Plans & Dues'), findsWidgets);
      expect(find.text('All Plans'), findsOneWidget);
      expect(find.text('Membership Fees'), findsOneWidget);
      expect(find.text('Voluntary Funds'), findsOneWidget);
      expect(find.text('TAGA Annual Subscription'), findsOneWidget);
      expect(find.text('₹3,500'), findsOneWidget);
      expect(find.text('TAGA Legal Fund Contribution'), findsOneWidget);
      expect(find.textContaining('Razorpay'), findsWidgets);
    });

    testWidgets('AnnouncementsScreen renders official association updates and priorities', (tester) async {
      final mockAnnouncements = [
        const MemberNotification(
          id: 'ann_1',
          title: 'Special General Body Meeting 2026',
          message: 'All members are cordially invited to TAGA Towers auditorium.',
          priority: 'urgent',
          isRead: false,
          createdAt: '2026-10-25',
        ),
        const MemberNotification(
          id: 'ann_2',
          title: 'Annual Subscription Renewal Notice',
          message: 'Please complete your annual dues before November 30.',
          priority: 'normal',
          isRead: true,
          createdAt: '2026-10-20',
        ),
      ];

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            memberAnnouncementsProvider.overrideWith((ref) => Future.value(mockAnnouncements)),
          ],
          child: const MaterialApp(home: AnnouncementsScreen()),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('Announcements'), findsWidgets);
      expect(find.text('Special General Body Meeting 2026'), findsOneWidget);
      expect(find.text('All members are cordially invited to TAGA Towers auditorium.'), findsOneWidget);
      expect(find.text('URGENT'), findsOneWidget);
      expect(find.text('Annual Subscription Renewal Notice'), findsOneWidget);
      expect(find.text('NORMAL'), findsOneWidget);
    });
  });
}
