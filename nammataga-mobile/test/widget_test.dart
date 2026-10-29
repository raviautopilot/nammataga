import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/core/router/app_router.dart';
import 'package:nammataga_mobile/features/auth/data/auth_repository.dart';
import 'package:nammataga_mobile/features/auth/domain/member_session.dart';
import 'package:nammataga_mobile/features/auth/presentation/login_screen.dart';
import 'package:nammataga_mobile/features/auth/providers/auth_provider.dart';
import 'package:nammataga_mobile/features/home/presentation/home_screen.dart';
import 'package:nammataga_mobile/features/splash/presentation/splash_screen.dart';

class FakeFailingAuthRepository extends AuthRepository {
  @override
  Future<MemberSession> login({
    required String identifier,
    required String password,
    bool rememberMe = false,
  }) async {
    throw Exception('Invalid credentials');
  }
}

void main() {
  testWidgets(
    'Splash Screen renders cinematic TAGA Tower visual and branding',
    (WidgetTester tester) async {
      await tester.pumpWidget(
        const ProviderScope(child: MaterialApp(home: SplashScreen())),
      );

      // Initial frame
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.text('TAGA TOWERS'), findsOneWidget);
      expect(find.text('தாகா டவர்ஸ்'), findsOneWidget);
      expect(
        find.text('தமிழ்நாடு அரசு வேளாண்மை பட்டதாரிகள் சங்கம்'),
        findsOneWidget,
      );
      expect(
        find.text('TAMIL NADU GOVT. AGRI. GRADUATES ASSOCIATION'),
        findsOneWidget,
      );
    },
  );

  testWidgets(
    'Splash Screen handles small and tall viewports without overflow',
    (WidgetTester tester) async {
      // 1. Test on a small viewport (e.g. 360 x 640)
      tester.view.physicalSize = const Size(360, 640);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);

      await tester.pumpWidget(
        const ProviderScope(child: MaterialApp(home: SplashScreen())),
      );

      await tester.pump(const Duration(milliseconds: 100));
      expect(tester.takeException(), isNull);

      // 2. Test on a tall flagship viewport (e.g. 430 x 932)
      tester.view.physicalSize = const Size(430, 932);
      await tester.pump(const Duration(milliseconds: 100));
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'App router navigates from /splash to /login after timer duration',
    (WidgetTester tester) async {
      await tester.pumpWidget(
        ProviderScope(child: MaterialApp.router(routerConfig: appRouter)),
      );

      // Initially on Splash
      await tester.pump(const Duration(milliseconds: 100));
      expect(find.text('TAGA TOWERS'), findsOneWidget);

      // Fast-forward past the 2800ms splash duration
      await tester.pump(const Duration(milliseconds: 2900));
      await tester.pump(
        const Duration(milliseconds: 400),
      ); // Page transition fade

      // Should now be on LoginScreen
      expect(find.text('Member Login'), findsOneWidget);
      expect(find.text('Email Address'), findsOneWidget);
      expect(find.text('Sign In'), findsOneWidget);
    },
  );

  testWidgets(
    'Member Login Screen renders animated logo, credentials inputs, and has Browse Public Pages removed',
    (WidgetTester tester) async {
      await tester.pumpWidget(
        const ProviderScope(child: MaterialApp(home: LoginScreen())),
      );

      await tester.pump(const Duration(milliseconds: 200));

      expect(find.text('Member Login'), findsOneWidget);
      expect(find.text('Email Address'), findsOneWidget);
      expect(find.text('Password'), findsOneWidget);
      expect(find.text('Sign In'), findsOneWidget);
      // Browse Public Pages must be removed completely
      expect(find.text('Browse Public Pages'), findsNothing);
      expect(find.text('OR'), findsNothing);

      // Trigger empty submission to verify credentials validation
      await tester.tap(find.text('Sign In'));
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.text('Please enter your registered email address'), findsOneWidget);
    },
  );

  testWidgets(
    'Member Login Screen validates credentials and displays Invalid credentials banner on failure',
    (WidgetTester tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authRepositoryProvider.overrideWithValue(FakeFailingAuthRepository()),
          ],
          child: const MaterialApp(home: LoginScreen()),
        ),
      );

      await tester.pump(const Duration(milliseconds: 200));

      // Enter valid format email and password
      await tester.enterText(
        find.byType(TextFormField).first,
        'sivakumar@gmail.com',
      );
      await tester.enterText(
        find.byType(TextFormField).last,
        'wrongpassword',
      );

      // Tap Sign In
      await tester.tap(find.text('Sign In'));
      await tester.pump(const Duration(milliseconds: 300));

      expect(find.text('Invalid credentials'), findsOneWidget);
      expect(
        find.textContaining('The email address or password you entered is incorrect'),
        findsOneWidget,
      );
    },
  );

  testWidgets(
    'Home Screen renders dashboard, services, stats and contact hub',
    (WidgetTester tester) async {
      await tester.pumpWidget(
        const ProviderScope(child: MaterialApp(home: HomeScreen())),
      );

      await tester.pump(const Duration(milliseconds: 200));

      expect(find.text('Namma TAGA'), findsWidgets);
      expect(find.text('Association Services'), findsOneWidget);
      expect(find.text('Office Bearers'), findsOneWidget);
      expect(find.text('About Us'), findsOneWidget);
      expect(find.text('Our Objectives'), findsOneWidget);
      expect(find.text('Contact Us'), findsOneWidget);
    },
  );
}
