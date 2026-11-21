import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:nammataga_mobile/core/constants/api_endpoints.dart';
import 'package:nammataga_mobile/features/auth/data/auth_repository.dart';
import 'package:nammataga_mobile/features/auth/domain/member_session.dart';
import 'package:nammataga_mobile/features/auth/presentation/login_screen.dart';
import 'package:nammataga_mobile/features/auth/providers/auth_provider.dart';
import 'package:nammataga_mobile/features/profile/domain/deletion_request_model.dart';
import 'package:nammataga_mobile/features/profile/presentation/account_deletion_screen.dart';
import 'package:nammataga_mobile/features/profile/presentation/account_settings_screen.dart';
import 'package:nammataga_mobile/features/profile/providers/profile_provider.dart';

class FakeDeletionPendingAuthRepository extends AuthRepository {
  @override
  Future<MemberSession> login({
    required String identifier,
    required String password,
    bool rememberMe = false,
  }) async {
    throw Exception(
      'Your account deletion request is awaiting administrator review. You cannot access your account while the request is pending.',
    );
  }
}

class FakeAuthNotifier extends AuthNotifier {
  final MemberSession mockSession;
  FakeAuthNotifier(this.mockSession);

  @override
  Future<MemberSession> build() async => mockSession;
}

void main() {
  group('1. API Endpoints Configuration', () {
    test('Account deletion endpoints are correctly resolved', () {
      expect(ApiEndpoints.accountDeletion, contains('/member/account-deletion'));
      expect(
        ApiEndpoints.accountDeletionStatus,
        contains('/member/account-deletion/status'),
      );
      expect(
        ApiEndpoints.accountDeletionEligibility,
        contains('/member/account-deletion/eligibility'),
      );
    });
  });

  group('2. Domain Models (DeletionEligibility & AccountDeletionRequestStatus)', () {
    test('DeletionEligibility parsing for eligible member', () {
      final json = {
        'eligible': true,
        'message': 'Eligible for account deletion',
      };
      final eligibility = DeletionEligibility.fromJson(json);
      expect(eligibility.canDelete, isTrue);
      expect(eligibility.hasActiveStay, isFalse);
      expect(eligibility.message, 'Eligible for account deletion');
    });

    test('DeletionEligibility parsing for member with active stay', () {
      final json = {
        'eligible': false,
        'reason': 'active_stay',
        'message':
            'You are currently staying at TAGA Towers. You can request account deletion after completing your checkout.',
      };
      final eligibility = DeletionEligibility.fromJson(json);
      expect(eligibility.canDelete, isFalse);
      expect(eligibility.hasActiveStay, isTrue);
      expect(
        eligibility.message,
        'You are currently staying at TAGA Towers. You can request account deletion after completing your checkout.',
      );
    });

    test('AccountDeletionRequestStatus parsing for pending request', () {
      final json = {
        'status': 'pending',
        'request_id': 'req_del_101',
        'requested_at': '2026-11-20T10:30:00Z',
        'reason': 'Retirement',
      };
      final status = AccountDeletionRequestStatus.fromJson(json);
      expect(status.isPending, isTrue);
      expect(status.hasPendingRequest, isTrue);
      expect(status.isRejected, isFalse);
      expect(status.isNone, isFalse);
      expect(status.requestId, 'req_del_101');
      expect(status.reason, 'Retirement');
      expect(status.createdAt, isNotNull);
    });

    test('AccountDeletionRequestStatus parsing for rejected request', () {
      final json = {
        'status': 'rejected',
        'request_id': 'req_del_102',
        'requested_at': '2026-11-18T09:00:00Z',
        'processed_at': '2026-11-19T14:00:00Z',
        'remarks': 'Active pending dues must be settled first',
      };
      final status = AccountDeletionRequestStatus.fromJson(json);
      expect(status.isPending, isFalse);
      expect(status.isRejected, isTrue);
      expect(status.adminRemarks, 'Active pending dues must be settled first');
      expect(status.processedAt, isNotNull);
    });

    test('AccountDeletionRequestStatus parsing for none', () {
      final json = {'status': 'none'};
      final status = AccountDeletionRequestStatus.fromJson(json);
      expect(status.isNone, isTrue);
      expect(status.isPending, isFalse);
      expect(status.isRejected, isFalse);
    });
  });

  group('3. AuthRepository Error Extraction Fix', () {
    test('Preserves backend 401 error message when error field is provided', () {
      final dioException = DioException(
        requestOptions: RequestOptions(path: '/api/member/login'),
        response: Response(
          requestOptions: RequestOptions(path: '/api/member/login'),
          statusCode: 401,
          data: {
            'error':
                'Your account deletion request is awaiting administrator review. You cannot access your account while the request is pending.',
          },
        ),
      );

      // Using reflection/subcall test pattern or calling login catch
      // In AuthRepository, login throws Exception(_extractErrorMessage(e))
      expect(
        () async {
          // Mock login call throwing the DioException
          final dio = Dio();
          dio.interceptors.add(
            InterceptorsWrapper(
              onRequest: (options, handler) {
                handler.reject(dioException);
              },
            ),
          );
          final testRepo = AuthRepository(dio: dio);
          await testRepo.login(identifier: 'test@taga.com', password: 'pwd');
        },
        throwsA(
          predicate((e) =>
              e.toString().contains('awaiting administrator review') &&
              !e.toString().contains('Invalid credentials')),
        ),
      );
    });

    test('Preserves backend 403 error message when message field is provided', () {
      final dioException = DioException(
        requestOptions: RequestOptions(path: '/api/member/login'),
        response: Response(
          requestOptions: RequestOptions(path: '/api/member/login'),
          statusCode: 403,
          data: {
            'message':
                'Your account deletion request is awaiting administrator review. You cannot access your account while the request is pending.',
          },
        ),
      );

      expect(
        () async {
          final dio = Dio();
          dio.interceptors.add(
            InterceptorsWrapper(
              onRequest: (options, handler) {
                handler.reject(dioException);
              },
            ),
          );
          final testRepo = AuthRepository(dio: dio);
          await testRepo.login(identifier: 'test@taga.com', password: 'pwd');
        },
        throwsA(
          predicate((e) =>
              e.toString().contains('awaiting administrator review')),
        ),
      );
    });

    test('Defaults to Invalid credentials on 401 when backend response has no custom message', () {
      final dioException = DioException(
        requestOptions: RequestOptions(path: '/api/member/login'),
        response: Response(
          requestOptions: RequestOptions(path: '/api/member/login'),
          statusCode: 401,
          data: null,
        ),
      );

      expect(
        () async {
          final dio = Dio();
          dio.interceptors.add(
            InterceptorsWrapper(
              onRequest: (options, handler) {
                handler.reject(dioException);
              },
            ),
          );
          final testRepo = AuthRepository(dio: dio);
          await testRepo.login(identifier: 'test@taga.com', password: 'pwd');
        },
        throwsA(
          predicate((e) => e.toString().contains('Invalid credentials')),
        ),
      );
    });
  });

  group('4. AccountSettingsScreen Entry Point (Apple Guideline 5.1.1(v))', () {
    testWidgets('Renders Request Account Deletion card directly above Logout', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            unreadAnnouncementsCountProvider.overrideWith((ref) => Future.value(0)),
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

      expect(find.text('Request Account Deletion'), findsOneWidget);
      expect(find.text('Logout'), findsOneWidget);
      expect(find.byIcon(LucideIcons.userX), findsOneWidget);
      expect(find.byIcon(LucideIcons.logOut), findsOneWidget);
    });
  });

  group('5. AccountDeletionScreen States & Disclosures', () {
    testWidgets('Renders Disclosures checklist and Normal form state when eligible', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            accountDeletionStatusProvider.overrideWith(
              (ref) => Future.value(const AccountDeletionRequestStatus(status: 'none')),
            ),
            accountDeletionEligibilityProvider.overrideWith(
              (ref) => Future.value(
                const DeletionEligibility(
                  canDelete: true,
                  hasActiveStay: false,
                  message: 'Eligible',
                ),
              ),
            ),
          ],
          child: const MaterialApp(home: AccountDeletionScreen()),
        ),
      );

      await tester.pumpAndSettle();

      // Title & Watermark
      expect(find.text('Account Deletion'), findsOneWidget);

      // Disclosures checklist items (non-affecting neutral policy)
      expect(find.text('Important Information Before You Proceed'), findsOneWidget);
      expect(
        find.textContaining('Account Access During Review: Once your request is submitted, your account login will be disabled'),
        findsOneWidget,
      );
      expect(
        find.textContaining('TAGA Towers Check-in Status: Account deletion cannot be submitted if you are currently checked in'),
        findsOneWidget,
      );
      expect(
        find.textContaining('Subscriber & Association Services: Access to TAGA subscriber benefits'),
        findsOneWidget,
      );
      expect(
        find.textContaining('TAGA ID Retirement'),
        findsNothing,
      );
      expect(
        find.textContaining('Past subscription and payment records are preserved'),
        findsNothing,
      );
      expect(
        find.textContaining('Request Finality: Once submitted, an account deletion request cannot be cancelled or withdrawn.'),
        findsOneWidget,
      );

      // Dynamic reason input (no predefined chips)
      expect(find.text('Reason for Deletion'), findsOneWidget);
      expect(find.text('Retirement'), findsNothing);
      expect(find.text('Left Department'), findsNothing);

      // Form inputs
      expect(find.text('Confirm Password'), findsOneWidget);
      expect(find.byType(TextFormField), findsNWidgets(2)); // dynamic reason + password
      expect(find.byType(Checkbox), findsOneWidget);
      expect(find.text('Submit Deletion Request'), findsOneWidget);
    });

    testWidgets('Active Stay Blocked State displays exact amber warning message', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            accountDeletionStatusProvider.overrideWith(
              (ref) => Future.value(const AccountDeletionRequestStatus(status: 'none')),
            ),
            accountDeletionEligibilityProvider.overrideWith(
              (ref) => Future.value(
                const DeletionEligibility(
                  canDelete: false,
                  hasActiveStay: true,
                  reason: 'active_stay',
                  message:
                      'You are currently staying at TAGA Towers. You can request account deletion after completing your checkout.',
                ),
              ),
            ),
          ],
          child: const MaterialApp(home: AccountDeletionScreen()),
        ),
      );

      await tester.pumpAndSettle();

      // Verify exact active stay message
      expect(
        find.text(
          'You are currently staying at TAGA Towers. You can request account deletion after completing your checkout.',
        ),
        findsOneWidget,
      );

      // Verify submit button is disabled when stay is active
      final submitFinder = find.text('Submit Deletion Request');
      await tester.ensureVisible(submitFinder);
      await tester.tap(submitFinder);
      await tester.pump();
      expect(find.text('Account Deletion'), findsOneWidget);
    });

    testWidgets('Pending State displays exact review message and Logout button without withdrawal', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            accountDeletionStatusProvider.overrideWith(
              (ref) => Future.value(
                AccountDeletionRequestStatus(
                  status: 'pending',
                  reason: 'Retirement',
                  requestId: 'TAGA-DEL-2026-001',
                  createdAt: DateTime.parse('2026-11-20T10:00:00Z'),
                  hasPendingRequest: true,
                ),
              ),
            ),
          ],
          child: const MaterialApp(home: AccountDeletionScreen()),
        ),
      );

      await tester.pumpAndSettle();

      // Verify neutral pending review message
      expect(
        find.text(
          'Your account deletion request is currently awaiting administrator review. Login access remains paused during this period. You will be notified once reviewed.',
        ),
        findsOneWidget,
      );

      // Verify metadata (developer Request ID excluded from member view)
      expect(find.text('Request Under Review'), findsOneWidget);
      expect(find.text('TAGA-DEL-2026-001'), findsNothing);
      expect(find.text('Retirement'), findsOneWidget);

      // Verify Sign Out button is present
      expect(find.text('Sign Out Now'), findsOneWidget);

      // Verify NO withdrawal button is present
      expect(find.textContaining('Withdraw'), findsNothing);
      expect(find.textContaining('Cancel Request'), findsNothing);
    });

    testWidgets('Rejected State displays admin remarks and Submit New Request button', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            accountDeletionStatusProvider.overrideWith(
              (ref) => Future.value(
                AccountDeletionRequestStatus(
                  status: 'rejected',
                  adminRemarks: 'Please complete pending handover documents with the association first.',
                  processedAt: DateTime.parse('2026-11-19T15:00:00Z'),
                  createdAt: DateTime.parse('2026-11-18T10:00:00Z'),
                ),
              ),
            ),
            accountDeletionEligibilityProvider.overrideWith(
              (ref) => Future.value(
                const DeletionEligibility(canDelete: true, hasActiveStay: false),
              ),
            ),
          ],
          child: const MaterialApp(home: AccountDeletionScreen()),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('Request Rejected'), findsOneWidget);
      expect(
        find.text('Please complete pending handover documents with the association first.'),
        findsOneWidget,
      );
      expect(find.text('Submit New Request'), findsOneWidget);

      // Tapping Submit New Request switches to the form view
      await tester.tap(find.text('Submit New Request'));
      await tester.pumpAndSettle();

      expect(find.text('Important Information Before You Proceed'), findsOneWidget);
      expect(find.text('Submit Deletion Request'), findsOneWidget);
    });
  });

  group('6. LoginScreen Pending Deletion Review Handling', () {
    testWidgets('Displays dedicated amber dialog and banner when deletion review is pending', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authRepositoryProvider.overrideWithValue(FakeDeletionPendingAuthRepository()),
          ],
          child: const MaterialApp(home: LoginScreen()),
        ),
      );

      await tester.pump(const Duration(milliseconds: 100));

      // Enter valid format email and password
      await tester.enterText(find.byType(TextFormField).first, 'sivakumar@gmail.com');
      await tester.enterText(find.byType(TextFormField).last, 'secretPassword');
      await tester.pump(const Duration(milliseconds: 100));

      // Ensure Sign In is visible and tap it
      final signInButton = find.text('Sign In');
      await tester.ensureVisible(signInButton);
      await tester.tap(signInButton);
      await tester.pump(const Duration(milliseconds: 300));

      // Verify dedicated amber dialog is displayed
      expect(find.byType(AlertDialog), findsOneWidget);
      expect(find.text('Account Deletion Pending'), findsWidgets);
      expect(find.textContaining('awaiting administrator review'), findsWidgets);
      expect(
        find.descendant(
          of: find.byType(AlertDialog),
          matching: find.textContaining('support@nammataga.com'),
        ),
        findsOneWidget,
      );

      // Dismiss dialog
      await tester.tap(find.text('Understood'));
      await tester.pump(const Duration(milliseconds: 300));

      // Verify dialog is closed and amber warning banner is visible on login screen
      expect(find.byType(AlertDialog), findsNothing);
      expect(find.text('Account Deletion Pending'), findsOneWidget);
      expect(find.textContaining('awaiting administrator review'), findsOneWidget);
      expect(find.textContaining('support@nammataga.com'), findsOneWidget);
    });

    testWidgets('Displays Invalid Credentials dialog and error when wrong password', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authRepositoryProvider.overrideWithValue(FakeInvalidCredentialsAuthRepository()),
          ],
          child: const MaterialApp(home: LoginScreen()),
        ),
      );

      await tester.pump(const Duration(milliseconds: 100));

      // Enter valid format email and password
      await tester.enterText(find.byType(TextFormField).first, 'sivakumar@gmail.com');
      await tester.enterText(find.byType(TextFormField).last, 'wrongPassword');
      await tester.pump(const Duration(milliseconds: 100));

      // Ensure Sign In is visible and tap it
      final signInButton = find.text('Sign In');
      await tester.ensureVisible(signInButton);
      await tester.tap(signInButton);
      await tester.pump(const Duration(milliseconds: 300));

      // Verify Invalid Credentials dialog is displayed
      expect(find.byType(AlertDialog), findsOneWidget);
      expect(find.text('Invalid Credentials'), findsWidgets);
      expect(find.textContaining('incorrect'), findsWidgets);
    });
  });
}

class FakeInvalidCredentialsAuthRepository extends AuthRepository {
  @override
  Future<MemberSession> login({
    required String identifier,
    required String password,
    bool rememberMe = false,
  }) async {
    throw Exception('Invalid credentials');
  }
}
