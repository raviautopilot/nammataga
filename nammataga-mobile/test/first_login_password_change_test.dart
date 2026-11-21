import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/features/auth/data/auth_repository.dart';
import 'package:nammataga_mobile/features/auth/domain/auth_exceptions.dart';
import 'package:nammataga_mobile/features/auth/domain/member_session.dart';
import 'package:nammataga_mobile/features/auth/presentation/login_screen.dart';
import 'package:nammataga_mobile/features/auth/presentation/widgets/change_password_sheet.dart';
import 'package:nammataga_mobile/features/auth/providers/auth_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

class MockLoginAuthRepository extends AuthRepository {
  final Future<MemberSession> Function(String identifier, String password, bool rememberMe)? onLogin;
  final Future<String> Function(String email, String oldPassword, String newPassword, String confirmPassword)? onChangePassword;

  MockLoginAuthRepository({
    this.onLogin,
    this.onChangePassword,
  });

  @override
  Future<MemberSession> login({
    required String identifier,
    required String password,
    bool rememberMe = false,
  }) async {
    if (onLogin != null) {
      return onLogin!(identifier, password, rememberMe);
    }
    return super.login(identifier: identifier, password: password, rememberMe: rememberMe);
  }

  @override
  Future<String> changePassword({
    required String email,
    required String oldPassword,
    required String newPassword,
    required String confirmPassword,
  }) async {
    if (onChangePassword != null) {
      return onChangePassword!(email, oldPassword, newPassword, confirmPassword);
    }
    return super.changePassword(
      email: email,
      oldPassword: oldPassword,
      newPassword: newPassword,
      confirmPassword: confirmPassword,
    );
  }
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    SharedPreferences.setMockInitialValues({});
    const channel = MethodChannel('plugins.it_nomads.com/flutter_secure_storage');
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(channel, (MethodCall methodCall) async {
      return null;
    });
  });

  tearDown(() {
    const channel = MethodChannel('plugins.it_nomads.com/flutter_secure_storage');
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(channel, null);
  });

  group('1. AuthRepository - First Login & Temporary Password Parsing', () {
    test('HTTP 200 with forceChangePassword: true returns MemberSession with forceChangePassword == true', () async {
      final dio = Dio();
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            handler.resolve(
              Response(
                requestOptions: options,
                statusCode: 200,
                data: {
                  'forceChangePassword': true,
                  'message': 'Password change required',
                  'email': 'first_timer@taga.com',
                },
              ),
            );
          },
        ),
      );

      final repo = AuthRepository(dio: dio);
      final session = await repo.login(
        identifier: 'first_timer@taga.com',
        password: 'TempPassword123!',
      );

      expect(session.forceChangePassword, isTrue);
      expect(session.isAuthenticated, isFalse);
      expect(session.tempEmail, 'first_timer@taga.com');
      expect(session.memberEmail, 'first_timer@taga.com');
    });

    test('HTTP 403 with forceChangePassword: true in DioException returns MemberSession instead of DeletionPendingException', () async {
      final dio = Dio();
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            handler.reject(
              DioException(
                requestOptions: options,
                response: Response(
                  requestOptions: options,
                  statusCode: 403,
                  data: {
                    'forceChangePassword': true,
                    'message': 'Password change required',
                    'email': 'temp_member@taga.com',
                  },
                ),
              ),
            );
          },
        ),
      );

      final repo = AuthRepository(dio: dio);
      final session = await repo.login(
        identifier: 'temp_member@taga.com',
        password: 'TempPassword123!',
      );

      expect(session.forceChangePassword, isTrue);
      expect(session.isAuthenticated, isFalse);
      expect(session.tempEmail, 'temp_member@taga.com');
    });

    test('HTTP 403 with "Password change required" message in DioException parses forceChangePassword: true', () async {
      final dio = Dio();
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            handler.reject(
              DioException(
                requestOptions: options,
                response: Response(
                  requestOptions: options,
                  statusCode: 403,
                  data: {
                    'message': 'Password change required',
                    'email': 'temp_msg@taga.com',
                  },
                ),
              ),
            );
          },
        ),
      );

      final repo = AuthRepository(dio: dio);
      final session = await repo.login(
        identifier: 'temp_msg@taga.com',
        password: 'TempPassword123!',
      );

      expect(session.forceChangePassword, isTrue);
      expect(session.tempEmail, 'temp_msg@taga.com');
    });

    test('Normal HTTP 403 without forceChangePassword still throws DeletionPendingException', () async {
      final dio = Dio();
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            handler.reject(
              DioException(
                requestOptions: options,
                response: Response(
                  requestOptions: options,
                  statusCode: 403,
                  data: {
                    'error': 'Your account deletion request is awaiting administrator review.',
                  },
                ),
              ),
            );
          },
        ),
      );

      final repo = AuthRepository(dio: dio);
      expect(
        () => repo.login(identifier: 'deleting@taga.com', password: 'password123'),
        throwsA(isA<DeletionPendingException>()),
      );
    });

    test('HTTP 200 with standard login success returns authenticated session', () async {
      final dio = Dio();
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            handler.resolve(
              Response(
                requestOptions: options,
                statusCode: 200,
                data: {
                  'token': 'jwt_auth_token_xyz',
                  'role': 'member',
                  'user': {
                    'id': 'MEM001',
                    'name': 'Permanent Member',
                    'email': 'perm@taga.com',
                    'isPaid': true,
                  },
                },
              ),
            );
          },
        ),
      );

      final repo = AuthRepository(dio: dio);
      final session = await repo.login(
        identifier: 'perm@taga.com',
        password: 'PermanentPassword123!',
      );

      expect(session.forceChangePassword, isFalse);
      expect(session.isAuthenticated, isTrue);
      expect(session.displayName, 'Permanent Member');
      expect(session.token, 'jwt_auth_token_xyz');
    });
  });

  group('2. Widget Flow - Temporary Password Login, ChangePasswordSheet & Auto-Login', () {
    testWidgets('Temporary password detection triggers sheet and logs in automatically with new password', (tester) async {
      tester.view.physicalSize = const Size(1080, 2400);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(() {
        tester.view.resetPhysicalSize();
        tester.view.resetDevicePixelRatio();
      });

      int loginCallCount = 0;
      int changePasswordCallCount = 0;
      bool loginSuccessTriggered = false;

      final testRepo = MockLoginAuthRepository(
        onLogin: (identifier, password, rememberMe) async {
          loginCallCount++;
          if (loginCallCount == 1) {
            // First attempt with temporary password
            expect(identifier, 'newuser@taga.com');
            expect(password, 'TempPass123!');
            return const MemberSession(
              forceChangePassword: true,
              tempEmail: 'newuser@taga.com',
            );
          } else {
            // Second automatic attempt with new permanent password
            expect(identifier, 'newuser@taga.com');
            expect(password, 'NewSecurePass@2026');
            return const MemberSession(
              token: 'new_session_token_999',
              role: 'member',
              isGuest: false,
              isPaid: true,
              user: {
                'id': 'MEM777',
                'name': 'New Member',
                'email': 'newuser@taga.com',
                'isPaid': true,
              },
            );
          }
        },
        onChangePassword: (email, oldPassword, newPassword, confirmPassword) async {
          changePasswordCallCount++;
          expect(email, 'newuser@taga.com');
          expect(oldPassword, 'TempPass123!');
          expect(newPassword, 'NewSecurePass@2026');
          expect(confirmPassword, 'NewSecurePass@2026');
          return 'Password changed successfully.';
        },
      );

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authRepositoryProvider.overrideWithValue(testRepo),
          ],
          child: MaterialApp(
            home: LoginScreen(
              onLoginSuccess: () {
                loginSuccessTriggered = true;
              },
            ),
          ),
        ),
      );

      await tester.pump(const Duration(milliseconds: 100));

      // 1. Enter temporary credentials in login form
      final textFields = find.byType(TextFormField);
      await tester.enterText(textFields.first, 'newuser@taga.com');
      await tester.enterText(textFields.last, 'TempPass123!');
      await tester.pump(const Duration(milliseconds: 100));

      // 2. Tap Sign In button
      final signInButton = find.text('Sign In');
      await tester.ensureVisible(signInButton);
      await tester.tap(signInButton);
      await tester.pump(const Duration(milliseconds: 300));
      await tester.pump(const Duration(milliseconds: 300));

      // Verify prompt/toast displayed: "Temporary password detected. Change password before login."
      expect(
        find.text('Temporary password detected. Change password before login.'),
        findsOneWidget,
      );

      // Verify ChangePasswordSheet has opened
      expect(find.byType(ChangePasswordSheet), findsOneWidget);
      expect(find.text('Set Permanent Password'), findsOneWidget);

      // 3. Fill in the new password and confirm password in the sheet
      final sheetFields = find.descendant(
        of: find.byType(ChangePasswordSheet),
        matching: find.byType(TextFormField),
      );
      expect(sheetFields, findsNWidgets(4));

      // Enter new password and confirmation
      await tester.enterText(sheetFields.at(2), 'NewSecurePass@2026');
      await tester.enterText(sheetFields.at(3), 'NewSecurePass@2026');
      await tester.pump(const Duration(milliseconds: 100));

      // 4. Tap the submit button inside ChangePasswordSheet ("Save & Proceed to Login")
      final submitButton = find.descendant(
        of: find.byType(ChangePasswordSheet),
        matching: find.text('Save & Proceed to Login'),
      );
      await tester.ensureVisible(submitButton);
      await tester.tap(submitButton);

      // Advance frames to process sheet pop and async login flow
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 300));
      await tester.pump(const Duration(milliseconds: 500));

      // Verify sheet popped and password change API was invoked
      expect(changePasswordCallCount, 1);
      expect(loginCallCount, 2);

      // Verify automatic login success callback was triggered
      expect(loginSuccessTriggered, isTrue);

      // Verify success toast title or message
      expect(find.text('Password Changed & Logged In'), findsOneWidget);
      expect(find.text('Welcome to TAGA, New Member!'), findsOneWidget);
    });
  });
}
