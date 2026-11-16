import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/features/home/domain/association_info.dart';
import 'package:nammataga_mobile/features/home/providers/home_provider.dart';
import 'package:nammataga_mobile/features/profile/presentation/privacy_policy_screen.dart';
import 'package:nammataga_mobile/features/profile/presentation/terms_and_conditions_screen.dart';
import 'package:nammataga_mobile/features/profile/presentation/widgets/about_taga_modal.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('TermsAndConditionsScreen Tests', () {
    testWidgets('Renders all major legal sections from Word document', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: TermsAndConditionsScreen(),
        ),
      );

      expect(find.text('Terms & Conditions'), findsWidgets);
      expect(find.text('Legal Terms of Use'), findsOneWidget);
      expect(find.text('Last Updated: 14 November 2026'), findsOneWidget);
      expect(find.text('Eligibility and Official Use'), findsOneWidget);
      expect(find.text('License to Use'), findsOneWidget);
      expect(find.text('Nature of the App and User Conduct'), findsOneWidget);
      expect(find.text('Account Management'), findsOneWidget);
      expect(find.text('Payments and Razorpay Integration'), findsOneWidget);
      expect(find.text('Events and Annual Fair'), findsOneWidget);
      expect(find.text('Intellectual Property'), findsOneWidget);
      expect(find.text('Termination'), findsOneWidget);
      expect(find.text('Disclaimer of Warranties'), findsOneWidget);
      expect(find.text('Limitation of Liability'), findsOneWidget);
      expect(find.text('Governing Law'), findsOneWidget);
      expect(find.text('Changes to Terms'), findsOneWidget);
      expect(find.text('Contact Us'), findsOneWidget);

      // Verify contact placeholders are preserved
      expect(find.text('[Insert TAGA Email]'), findsOneWidget);
    });
  });

  group('PrivacyPolicyScreen Tests', () {
    testWidgets('Renders DPDP compliant privacy policy sections', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: PrivacyPolicyScreen(),
        ),
      );

      expect(find.text('Privacy Policy'), findsWidgets);
      expect(find.text('Data Privacy & DPDP Compliance'), findsOneWidget);
      expect(find.text('Last Updated: 14 November 2026'), findsOneWidget);
      expect(find.text('Information We Collect'), findsOneWidget);
      expect(find.text('How We Use Your Information'), findsOneWidget);
      expect(find.text('Data Sharing and Disclosure'), findsOneWidget);
      expect(find.text('Data Retention and Deletion'), findsOneWidget);
      expect(find.text('Data Security'), findsOneWidget);
      expect(find.text('Your Rights under the DPDP Act'), findsOneWidget);
      expect(find.text('Children\'s Privacy'), findsOneWidget);
      expect(find.text('Changes to This Privacy Policy'), findsOneWidget);
      expect(find.text('Contact Us'), findsOneWidget);

      // Verify Grievance officer placeholders are preserved
      expect(find.text('[Insert Grievance Officer Name]'), findsOneWidget);
      expect(find.text('[Insert TAGA Phone Number]'), findsOneWidget);
    });
  });

  group('AboutTagaModal Tests', () {
    testWidgets('Displays dynamic About Us data directly from backend provider', (tester) async {
      const mockAbout = AboutData(
        id: 1,
        name: 'Tamil Nadu Government Agricultural Graduates Association',
        acronym: 'TAGA',
        establishedYear: 1929,
        tagline: 'Empowering agricultural graduates',
        heroImageUrl: '',
        mission: 'To unite agricultural graduates',
        vision: 'To create a prosperous Tamil Nadu',
        description: 'The Tamil Nadu Agricultural Graduates’ Association (TAGA) is the apex body representing the professional interests of agricultural science graduates of the Tamil Nadu agriculture department. Since our inception, we have served as a bridge between scientific research and field implementation.We believe that the future of agriculture lies in the hands of trained professionals who can balance traditional wisdom with modern technology. Our association provides a platform for networking, continuous learning, and policy advocacy.',
      );

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            aboutInfoProvider.overrideWith((ref) async => mockAbout),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: AboutTagaSheet(),
            ),
          ),
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('About TAGA'), findsOneWidget);
      expect(find.text('ESTABLISHED 1929'), findsOneWidget);
      expect(find.text('About Us'), findsOneWidget);
      expect(
        find.text(mockAbout.description),
        findsOneWidget,
      );
      expect(find.text('Our Mission'), findsOneWidget);
      expect(find.text('To unite agricultural graduates'), findsOneWidget);
      expect(find.text('Our Vision'), findsOneWidget);
      expect(find.text('To create a prosperous Tamil Nadu'), findsOneWidget);
    });
  });
}
