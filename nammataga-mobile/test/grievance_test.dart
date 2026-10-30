import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:nammataga_mobile/core/widgets/bouncy_button.dart';
import 'package:nammataga_mobile/features/auth/domain/member_session.dart';
import 'package:nammataga_mobile/features/grievance/data/grievance_repository.dart';
import 'package:nammataga_mobile/features/grievance/domain/grievance_model.dart';
import 'package:nammataga_mobile/features/grievance/domain/grievance_priority_model.dart';
import 'package:nammataga_mobile/features/grievance/presentation/grievance_form_screen.dart';
import 'package:nammataga_mobile/features/grievance/presentation/grievance_intro_screen.dart';
import 'package:nammataga_mobile/features/grievance/presentation/grievance_success_screen.dart';
import 'package:nammataga_mobile/features/grievance/presentation/widgets/grievance_access_restricted_screen.dart';
import 'package:nammataga_mobile/features/grievance/providers/grievance_provider.dart';
import 'package:nammataga_mobile/features/profile/domain/member_profile.dart';

/// Fake repository for controlling network outcomes and concurrency in tests.
class FakeGrievanceRepository extends GrievanceRepository {
  bool shouldThrow = false;
  String errorMessage = 'Server unreachable';
  Completer<GrievanceModel>? completer;
  int submitCount = 0;
  GrievanceModel? cannedGrievance;

  @override
  Future<GrievanceModel> submitGrievance({
    required String subject,
    required String category,
    required String priority,
    required String description,
    String? contactPhone,
    required String preferredResponse,
  }) async {
    submitCount++;

    if (completer != null) {
      return completer!.future;
    }

    if (shouldThrow) {
      throw Exception(errorMessage);
    }

    return cannedGrievance ??
        GrievanceModel(
          id: 'GRV-001',
          subject: subject,
          category: category,
          priority: priority,
          description: description,
          contactPhone: contactPhone,
          preferredResponse: preferredResponse,
          status: 'Pending',
          submittedDate: DateTime(2026, 10, 30, 14, 30),
        );
  }
}

/// Helper function representing the client-side access control rule for the grievance portal.
bool canAccessGrievancePortal({
  MemberSession? session,
  MemberProfile? profile,
}) {
  if (profile != null && (profile.isPaid || profile.subscriptionActive)) {
    return true;
  }
  if (session != null && session.isPaid && !session.isGuest) {
    return true;
  }
  return false;
}

void main() {
  group('1. GrievanceModel Serialization & Deserialization', () {
    test('fromJson accurately deserializes full JSON payload', () {
      final json = {
        'id': 'GRV-98765',
        'subject': 'Fertilizer Subsidy Discrepancy',
        'category': 'Subsidies & Grants',
        'priority': 'High',
        'description': 'Discrepancy in quarterly fertilizer disbursement.',
        'contactPhone': '+91 9876543210',
        'preferredResponse': 'Phone',
        'status': 'Pending',
        'submittedDate': '2026-10-30T10:15:30.000Z',
        'memberName': 'R. Sivakumar',
        'memberEmail': 'sivakumar.taga@gmail.com',
        'assignedTo': 'Directorate Liason Officer',
        'lastUpdate': '2026-10-30T11:00:00.000Z',
      };

      final model = GrievanceModel.fromJson(json);

      expect(model.id, 'GRV-98765');
      expect(model.subject, 'Fertilizer Subsidy Discrepancy');
      expect(model.category, 'Subsidies & Grants');
      expect(model.priority, 'High');
      expect(model.description, 'Discrepancy in quarterly fertilizer disbursement.');
      expect(model.contactPhone, '+91 9876543210');
      expect(model.preferredResponse, 'Phone');
      expect(model.status, 'Pending');
      expect(model.submittedDate, isNotNull);
      expect(model.memberName, 'R. Sivakumar');
      expect(model.memberEmail, 'sivakumar.taga@gmail.com');
      expect(model.assignedTo, 'Directorate Liason Officer');
      expect(model.lastUpdate, isNotNull);
    });

    test('fromJson provides safe fallbacks for missing or null fields', () {
      final json = <String, dynamic>{};

      final model = GrievanceModel.fromJson(json);

      expect(model.id, '');
      expect(model.subject, '');
      expect(model.category, '');
      expect(model.priority, '');
      expect(model.description, '');
      expect(model.contactPhone, isNull);
      expect(model.preferredResponse, 'Email');
      expect(model.status, 'Pending');
      expect(model.submittedDate, isNull);
      expect(model.memberName, isNull);
      expect(model.memberEmail, isNull);
      expect(model.assignedTo, isNull);
      expect(model.lastUpdate, isNull);
    });

    test('Parses various backend date formats correctly', () {
      final formats = [
        {'input': '2026-10-30T14:30:00Z', 'expectedYear': 2026, 'expectedMonth': 10},
        {'input': '2026-10-30 14:30:00', 'expectedYear': 2026, 'expectedMonth': 10},
        {'input': '2026-10-30 14:30', 'expectedYear': 2026, 'expectedMonth': 10},
        {'input': '2026-10-30', 'expectedYear': 2026, 'expectedMonth': 10},
        {'input': '30-10-2026 14:30', 'expectedYear': 2026, 'expectedMonth': 10},
        {'input': '30-10-2026', 'expectedYear': 2026, 'expectedMonth': 10},
        {'input': '30/10/2026 14:30', 'expectedYear': 2026, 'expectedMonth': 10},
        {'input': '30/10/2026', 'expectedYear': 2026, 'expectedMonth': 10},
        {'input': '30 Oct 2026, 02:30 PM', 'expectedYear': 2026, 'expectedMonth': 10},
        {'input': '30 Oct 2026', 'expectedYear': 2026, 'expectedMonth': 10},
      ];

      for (final item in formats) {
        final model = GrievanceModel.fromJson({'submittedDate': item['input']});
        expect(model.submittedDate, isNotNull, reason: 'Failed for pattern: ${item['input']}');
        expect(model.submittedDate!.year, item['expectedYear']);
        expect(model.submittedDate!.month, item['expectedMonth']);
      }

      // Null or invalid dates safely produce null
      expect(GrievanceModel.fromJson({'submittedDate': null}).submittedDate, isNull);
      expect(GrievanceModel.fromJson({'submittedDate': ''}).submittedDate, isNull);
      expect(GrievanceModel.fromJson({'submittedDate': 'invalid-date'}).submittedDate, isNull);
    });

    test('toJson produces expected structure and round-trips cleanly', () {
      final original = GrievanceModel(
        id: 'GRV-002',
        subject: 'Irrigation Scheme Delay',
        category: 'Field Operations',
        priority: 'Urgent',
        description: 'Borewell clearance delayed past statutory deadline.',
        contactPhone: '9443322110',
        preferredResponse: 'Email',
        status: 'Pending',
        submittedDate: DateTime(2026, 10, 30, 9, 30),
        memberName: 'Dr. Anitha',
        memberEmail: 'anitha@tn.gov.in',
      );

      final map = original.toJson();
      expect(map['id'], 'GRV-002');
      expect(map['subject'], 'Irrigation Scheme Delay');
      expect(map['contactPhone'], '9443322110');
      expect(map['submittedDate'], contains('2026-10-30'));

      final restored = GrievanceModel.fromJson(map);
      expect(restored.id, original.id);
      expect(restored.subject, original.subject);
      expect(restored.category, original.category);
      expect(restored.priority, original.priority);
      expect(restored.description, original.description);
    });

    test('copyWith modifies selected fields while retaining existing values', () {
      const original = GrievanceModel(
        id: 'GRV-003',
        subject: 'Promotion Seniority Inquiry',
        category: 'Service Matters',
        priority: 'Medium',
        description: 'Query regarding 2012 batch seniority listing.',
        preferredResponse: 'Email',
        status: 'Pending',
      );

      final updated = original.copyWith(
        status: 'Resolved',
        assignedTo: 'Executive Secretary',
      );

      expect(updated.id, original.id);
      expect(updated.subject, original.subject);
      expect(updated.status, 'Resolved');
      expect(updated.assignedTo, 'Executive Secretary');
      expect(updated.category, original.category);
    });

    test('Equality and hashCode based on ID', () {
      const g1 = GrievanceModel(
        id: 'GRV-SAME',
        subject: 'Topic A',
        category: 'Cat A',
        priority: 'Low',
        description: 'Desc',
        preferredResponse: 'Email',
        status: 'Pending',
      );

      const g2 = GrievanceModel(
        id: 'GRV-SAME',
        subject: 'Topic B',
        category: 'Cat B',
        priority: 'High',
        description: 'Different desc',
        preferredResponse: 'Phone',
        status: 'Resolved',
      );

      const g3 = GrievanceModel(
        id: 'GRV-DIFF',
        subject: 'Topic A',
        category: 'Cat A',
        priority: 'Low',
        description: 'Desc',
        preferredResponse: 'Email',
        status: 'Pending',
      );

      expect(g1, equals(g2));
      expect(g1.hashCode, equals(g2.hashCode));
      expect(g1, isNot(equals(g3)));
      expect(g1.toString(), contains('GRV-SAME'));
    });
  });

  group('2. GrievancePriorityModel & Priority Helpers', () {
    test('GrievancePriorityModel serialization, equality, and case-insensitivity', () {
      final json = {'label': 'High Priority', 'value': 'high'};
      final priority = GrievancePriorityModel.fromJson(json);

      expect(priority.label, 'High Priority');
      expect(priority.value, 'high');
      expect(priority.toJson(), equals(json));
      expect(priority.toString(), contains('High Priority'));

      const priorityA = GrievancePriorityModel(label: 'High', value: 'high');
      const priorityB = GrievancePriorityModel(label: 'HIGH', value: 'HIGH');
      const priorityC = GrievancePriorityModel(label: 'Low', value: 'low');

      // Case-insensitive value equality
      expect(priorityA, equals(priorityB));
      expect(priorityA.hashCode, equals(priorityB.hashCode));
      expect(priorityA, isNot(equals(priorityC)));
    });

    test('Priority and date formatting getters on GrievanceModel', () {
      final date = DateTime(2026, 10, 30, 14, 30);
      final model = GrievanceModel(
        id: 'GRV-004',
        subject: 'Seed Distribution',
        category: 'Distribution',
        priority: 'High',
        description: 'Shortage of paddy seeds.',
        preferredResponse: 'Email',
        status: 'Pending',
        submittedDate: date,
      );

      expect(model.formattedDate, isNotEmpty);
      expect(model.formattedDate, contains('30 Oct 2026'));
      expect(model.formattedDateOnly, '30 Oct 2026');
      expect(model.isPending, isTrue);
      expect(model.isResolved, isFalse);

      final resolvedModel = model.copyWith(status: 'Resolved');
      expect(resolvedModel.isPending, isFalse);
      expect(resolvedModel.isResolved, isTrue);

      final closedModel = model.copyWith(status: 'Closed');
      expect(closedModel.isResolved, isTrue);

      final readModel = model.copyWith(status: 'Read');
      expect(readModel.isResolved, isTrue);

      const noDateModel = GrievanceModel(
        id: 'GRV-NO-DATE',
        subject: 'Sub',
        category: 'Cat',
        priority: 'Low',
        description: 'Desc',
        preferredResponse: 'Email',
        status: 'Pending',
      );
      expect(noDateModel.formattedDate, '');
      expect(noDateModel.formattedDateOnly, '');
    });
  });

  group('3. GrievanceSubmissionNotifier Anti-Duplicate Lock', () {
    test('Prevents concurrent duplicate submissions while isSubmitting is true', () async {
      final fakeRepo = FakeGrievanceRepository();
      final slowCompleter = Completer<GrievanceModel>();
      fakeRepo.completer = slowCompleter;

      final container = ProviderContainer(
        overrides: [
          grievanceRepositoryProvider.overrideWithValue(fakeRepo),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(grievanceSubmissionProvider.notifier);

      expect(container.read(grievanceSubmissionProvider).isSubmitting, isFalse);

      // 1. First submission initiates asynchronous request
      final firstSubmissionFuture = notifier.submitGrievance(
        subject: 'Water canal blockage',
        category: 'Infrastructure',
        priority: 'High',
        description: 'Canal silt removal requested.',
      );

      // State is immediately marked as submitting
      expect(container.read(grievanceSubmissionProvider).isSubmitting, isTrue);
      expect(fakeRepo.submitCount, 1);

      // 2. Second rapid attempt while first is still pending
      final secondSubmissionResult = await notifier.submitGrievance(
        subject: 'Water canal blockage - duplicate click',
        category: 'Infrastructure',
        priority: 'High',
        description: 'Duplicate click simulation.',
      );

      // Second attempt must be rejected immediately by anti-duplicate guard
      expect(secondSubmissionResult, isFalse);
      expect(fakeRepo.submitCount, 1, reason: 'Repository must not be called again');

      // 3. Complete the first slow submission
      slowCompleter.complete(
        GrievanceModel(
          id: 'GRV-SUCCESS-01',
          subject: 'Water canal blockage',
          category: 'Infrastructure',
          priority: 'High',
          description: 'Canal silt removal requested.',
          preferredResponse: 'Email',
          status: 'Pending',
          submittedDate: DateTime.now(),
        ),
      );

      final firstSubmissionResult = await firstSubmissionFuture;
      expect(firstSubmissionResult, isTrue);
      expect(container.read(grievanceSubmissionProvider).isSubmitting, isFalse);
      expect(container.read(grievanceSubmissionProvider).isSuccess, isTrue);
    });
  });

  group('4. Form Data Preservation on Submission Failure', () {
    test('Preserves user entered form values when backend call fails', () async {
      final fakeRepo = FakeGrievanceRepository()..shouldThrow = true;
      fakeRepo.errorMessage = 'Network connection timed out';

      final container = ProviderContainer(
        overrides: [
          grievanceRepositoryProvider.overrideWithValue(fakeRepo),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(grievanceSubmissionProvider.notifier);

      // Pre-fill form fields
      notifier.updateFormData(
        subject: 'Pension revision pending',
        category: 'Service Matters',
        priority: 'Urgent',
        description: 'Pension arrears computation pending with treasury office.',
        contactPhone: '9840012345',
        preferredResponse: 'Phone',
      );

      expect(container.read(grievanceSubmissionProvider).subject, 'Pension revision pending');

      // Submit grievance which fails
      final success = await notifier.submitGrievance();

      expect(success, isFalse);
      final state = container.read(grievanceSubmissionProvider);

      // Ensure submitting flag is cleared and error message populated
      expect(state.isSubmitting, isFalse);
      expect(state.hasError, isTrue);
      expect(state.errorMessage, contains('Network connection timed out'));
      expect(state.isSuccess, isFalse);

      // Critical: entered form data must NOT be cleared on failure
      expect(state.formData.subject, 'Pension revision pending');
      expect(state.formData.category, 'Service Matters');
      expect(state.formData.priority, 'Urgent');
      expect(state.formData.description, 'Pension arrears computation pending with treasury office.');
      expect(state.formData.contactPhone, '9840012345');
      expect(state.formData.preferredResponse, 'Phone');

      // Clearing error still retains user inputs
      notifier.clearError();
      final clearedState = container.read(grievanceSubmissionProvider);
      expect(clearedState.hasError, isFalse);
      expect(clearedState.formData.subject, 'Pension revision pending');

      // Reset completely resets state
      notifier.reset();
      final resetState = container.read(grievanceSubmissionProvider);
      expect(resetState.formData.subject, '');
      expect(resetState.hasError, isFalse);
    });

    test('Successful submission resets form cache and populates submitted model', () async {
      final fakeRepo = FakeGrievanceRepository();
      final container = ProviderContainer(
        overrides: [
          grievanceRepositoryProvider.overrideWithValue(fakeRepo),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(grievanceSubmissionProvider.notifier);

      final success = await notifier.submitGrievance(
        subject: 'Farm equipment hire rates',
        category: 'Agricultural Schemes',
        priority: 'Medium',
        description: 'Standardized hiring rates request.',
      );

      expect(success, isTrue);
      final state = container.read(grievanceSubmissionProvider);
      expect(state.isSuccess, isTrue);
      expect(state.submittedGrievance, isNotNull);
      expect(state.submittedGrievance!.id, 'GRV-001');
      // On success, form fields are reset to clean empty state
      expect(state.formData.subject, '');
      expect(state.formData.description, '');
    });
  });

  group('5. Access Control: Paid vs Unpaid Member Verification', () {
    test('Unauthenticated or guest session cannot access grievance portal', () {
      final guestSession = MemberSession.guest();
      expect(canAccessGrievancePortal(session: guestSession), isFalse);

      final unauthSession = MemberSession.unauthenticated();
      expect(canAccessGrievancePortal(session: unauthSession), isFalse);
    });

    test('Unpaid member session cannot access grievance portal', () {
      const unpaidSession = MemberSession(
        token: 'valid_jwt_token',
        role: 'member',
        isPaid: false,
        isGuest: false,
      );
      expect(canAccessGrievancePortal(session: unpaidSession), isFalse);
    });

    test('Paid member session is granted access to grievance portal', () {
      const paidSession = MemberSession(
        token: 'valid_jwt_token',
        role: 'member',
        isPaid: true,
        isGuest: false,
      );
      expect(canAccessGrievancePortal(session: paidSession), isTrue);
    });

    test('Member profile verifies subscription status correctly', () {
      const unpaidProfile = MemberProfile(
        id: 'usr_01',
        tagaId: 'TAGA-101',
        username: 'unpaid_user',
        name: 'M. Selvan',
        initial: '',
        gender: 'Male',
        dateOfBirth: '1988-04-10',
        fatherName: 'Murugan',
        motherName: 'Lakshmi',
        mobileNumber: '9988776655',
        emailId: 'selvan@example.com',
        designation: 'Agricultural Officer',
        recruitmentBatch: '2015',
        seniorityNumber: 'SN-102',
        workingDistrict: 'Salem',
        nativeDistrict: 'Salem',
        residentialAddress: 'Address',
        permanentAddress: 'Address',
        educationalQualification: 'B.Sc Agri',
        tbfNumber: 'TBF-1',
        cpsGpfNumber: 'CPS-1',
        isPaid: false,
        subscriptionActive: false,
        paymentStatus: 'Unpaid',
      );

      expect(canAccessGrievancePortal(profile: unpaidProfile), isFalse);

      final paidProfile = unpaidProfile.copyWith(
        isPaid: true,
        subscriptionActive: true,
        paymentStatus: 'Paid',
      );

      expect(canAccessGrievancePortal(profile: paidProfile), isTrue);
    });
  });

  group('6. GrievanceAccessRestrictedScreen Widget Tests', () {
    void setDeviceSize(WidgetTester tester) {
      tester.view.physicalSize = const Size(430, 1200);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
    }

    Widget buildTestScreen({
      VoidCallback? onUpgrade,
      VoidCallback? onReturn,
      String secondaryActionLabel = 'Maybe Later',
    }) {
      return MaterialApp(
        home: GrievanceAccessRestrictedScreen(
          onUpgrade: onUpgrade,
          onReturn: onReturn,
          secondaryActionLabel: secondaryActionLabel,
        ),
      );
    }

    testWidgets('Renders all required restricted card UI elements and benefits list', (
      tester,
    ) async {
      setDeviceSize(tester);
      await tester.pumpWidget(buildTestScreen());
      await tester.pumpAndSettle();

      // 1. Header with back button and title
      expect(find.text('Grievance Redressal'), findsOneWidget);
      expect(find.byIcon(LucideIcons.arrowLeft), findsOneWidget);

      // 2. Lock badge icon
      expect(find.byIcon(LucideIcons.lock), findsOneWidget);

      // 3. Title & description
      expect(find.text('Paid Member Exclusive'), findsOneWidget);
      expect(
        find.text(
          'The Grievance Redressal portal is an executive support service reserved for subscribed NammaTAGA members. Active membership ensures direct representation before state agricultural authorities.',
        ),
        findsOneWidget,
      );

      // 4. Three benefit reminder pills
      expect(
        find.text('Direct representation before agricultural directorates'),
        findsOneWidget,
      );
      expect(
        find.text('Priority petition tracking and escalation'),
        findsOneWidget,
      );
      expect(
        find.text('Official association intervention'),
        findsOneWidget,
      );
      expect(find.byIcon(LucideIcons.check), findsNWidgets(3));

      // 5. Primary action button
      expect(find.text('Upgrade to Paid Membership'), findsOneWidget);
      expect(find.byIcon(LucideIcons.crown), findsOneWidget);

      // 6. Secondary action button
      expect(find.text('Maybe Later'), findsOneWidget);
    });

    testWidgets('Tapping Upgrade button fires callback or navigation', (tester) async {
      setDeviceSize(tester);
      bool upgraded = false;

      await tester.pumpWidget(
        buildTestScreen(onUpgrade: () => upgraded = true),
      );
      await tester.pumpAndSettle();

      await tester.ensureVisible(find.text('Upgrade to Paid Membership'));
      await tester.tap(find.text('Upgrade to Paid Membership'));
      await tester.pumpAndSettle();

      expect(upgraded, isTrue);
    });

    testWidgets('Tapping Maybe Later button fires return callback', (tester) async {
      setDeviceSize(tester);
      bool returned = false;

      await tester.pumpWidget(
        buildTestScreen(onReturn: () => returned = true),
      );
      await tester.pumpAndSettle();

      await tester.ensureVisible(find.text('Maybe Later'));
      await tester.tap(find.text('Maybe Later'));
      await tester.pumpAndSettle();

      expect(returned, isTrue);
    });

    testWidgets('Tapping Header Back button fires return callback', (tester) async {
      setDeviceSize(tester);
      bool backPressed = false;

      await tester.pumpWidget(
        buildTestScreen(onReturn: () => backPressed = true),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(LucideIcons.arrowLeft));
      await tester.pumpAndSettle();

      expect(backPressed, isTrue);
    });

    testWidgets('Navigates to /profile/subscriptions using GoRouter', (tester) async {
      setDeviceSize(tester);
      final router = GoRouter(
        initialLocation: '/grievance-restricted',
        routes: [
          GoRoute(
            path: '/grievance-restricted',
            builder: (context, state) => const GrievanceAccessRestrictedScreen(),
          ),
          GoRoute(
            path: '/profile/subscriptions',
            builder: (context, state) => const Scaffold(
              body: Text('Profile Subscriptions Page'),
            ),
          ),
        ],
      );

      await tester.pumpWidget(MaterialApp.router(routerConfig: router));
      await tester.pumpAndSettle();

      expect(find.text('Paid Member Exclusive'), findsOneWidget);

      // Tap Upgrade button
      await tester.ensureVisible(find.text('Upgrade to Paid Membership'));
      await tester.tap(find.text('Upgrade to Paid Membership'));
      await tester.pumpAndSettle();

      // Verified arrival at /profile/subscriptions
      expect(find.text('Profile Subscriptions Page'), findsOneWidget);
    });

    testWidgets('Allows custom secondary action label (e.g. Return)', (tester) async {
      setDeviceSize(tester);
      await tester.pumpWidget(
        buildTestScreen(secondaryActionLabel: 'Return'),
      );
      await tester.pumpAndSettle();

      expect(find.text('Return'), findsOneWidget);
      expect(find.text('Maybe Later'), findsNothing);
    });
  });

  group('7. GrievanceIntroScreen Widget Tests', () {
    testWidgets('Renders header, benefit cards, and Raise Grievance button', (tester) async {
      tester.view.physicalSize = const Size(430, 1200);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: GrievanceIntroScreen(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Grievance Redressal'), findsWidgets);
      expect(find.text('Raise your concern. We are here to listen.'), findsOneWidget);
      expect(find.text('Share Your Issue'), findsOneWidget);
      expect(find.text('For a Better Community'), findsOneWidget);
      expect(find.text('Available for Paid Members'), findsOneWidget);
      expect(find.text('Raise Grievance'), findsOneWidget);
    });
  });

  group('8. GrievanceSuccessScreen Widget Tests', () {
    testWidgets('Renders success checkmark, heading, message, and Back to Home button', (tester) async {
      tester.view.physicalSize = const Size(430, 1200);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: GrievanceSuccessScreen(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Grievance Submitted Successfully!'), findsOneWidget);
      expect(find.text('Back to Home'), findsOneWidget);
      expect(find.byIcon(LucideIcons.check), findsWidgets);
    });
  });

  group('9. GrievanceFormScreen Form Validation Widget Tests', () {
    testWidgets('Renders all form fields and verifies required validations', (tester) async {
      tester.view.physicalSize = const Size(430, 1400);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(
            home: GrievanceFormScreen(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Submit Grievance'), findsNWidgets(2));
      expect(find.byType(TextFormField), findsNWidgets(3)); // Subject, Description, Phone
      expect(find.text('Email'), findsOneWidget);
      expect(find.text('Phone'), findsOneWidget);
      expect(find.text('Postal Letter'), findsOneWidget);
      expect(find.widgetWithText(BouncyButton, 'Submit Grievance'), findsOneWidget);

      // Tap submit with empty form to trigger validation errors
      await tester.tap(find.widgetWithText(BouncyButton, 'Submit Grievance'));
      await tester.pumpAndSettle();

      expect(find.text('Please enter a subject'), findsOneWidget);
      expect(find.text('Please select a category'), findsOneWidget);
      expect(find.text('Description must be at least 50 characters'), findsOneWidget);
    });
  });
}
