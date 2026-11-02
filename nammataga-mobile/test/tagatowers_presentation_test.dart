import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/tagatowers_access_restricted_screen.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/theme/tagatowers_design_tokens.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/widgets/booking_step_progress_bar.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/widgets/price_breakdown_card.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/widgets/tagatowers_hero_card.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('TAGA Towers Design Tokens', () {
    test('Advance rate helper returns correct bed rates', () {
      expect(TagaTowersDesignTokens.dormitoryAdvance, 100);
      expect(TagaTowersDesignTokens.standardAdvance, 200);

      expect(TagaTowersDesignTokens.getAdvanceRatePerBed('gents-dorm'), 100);
      expect(TagaTowersDesignTokens.getAdvanceRatePerBed('ladies-dorm'), 100);
      expect(TagaTowersDesignTokens.getAdvanceRatePerBed('Gents Dormitory'), 100);
      expect(TagaTowersDesignTokens.getAdvanceRatePerBed('Dorm Bed 4'), 100);

      expect(TagaTowersDesignTokens.getAdvanceRatePerBed('apex-suite'), 200);
      expect(TagaTowersDesignTokens.getAdvanceRatePerBed('ac-room'), 200);
      expect(TagaTowersDesignTokens.getAdvanceRatePerBed('pavalam'), 200);
      expect(TagaTowersDesignTokens.getAdvanceRatePerBed('kurinchi'), 200);
    });

    test('Room display names mapped correctly', () {
      expect(TagaTowersDesignTokens.getRoomDisplayName('apex-suite'), 'Apex Suite A/C');
      expect(TagaTowersDesignTokens.getRoomDisplayName('ac-room'), 'A/C Room');
      expect(TagaTowersDesignTokens.getRoomDisplayName('gents-dorm'), 'Gents Dormitory');
      expect(TagaTowersDesignTokens.getRoomDisplayName('ladies-dorm'), 'Ladies Dormitory');
      expect(TagaTowersDesignTokens.getRoomDisplayName('kurinchi'), 'Kurinchi Suite');
    });

    test('Decorations and shadows generate proper styles', () {
      final cardDeco = TagaTowersDesignTokens.cardDecoration();
      expect(cardDeco.borderRadius, BorderRadius.circular(16));
      expect(cardDeco.border, isNotNull);

      final badgeDeco = TagaTowersDesignTokens.badgeDecoration();
      expect(badgeDeco.borderRadius, BorderRadius.circular(999));
    });
  });

  group('BookingStepProgressBar Widget Tests', () {
    testWidgets('Renders all 4 steps and highlights active step', (WidgetTester tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: BookingStepProgressBar(currentStep: 2),
          ),
        ),
      );

      // Verify header indicators
      expect(find.text('STEP 2 OF 4'), findsOneWidget);
      expect(find.text('Guest Info'), findsOneWidget);
      expect(find.text('33%'), findsOneWidget); // (2-1)/(4-1) = 33%

      // Verify step short titles
      expect(find.text('Rooms'), findsOneWidget);
      expect(find.text('Guests'), findsOneWidget);
      expect(find.text('Review'), findsOneWidget);
      expect(find.text('Payment'), findsOneWidget);
    });

    testWidgets('Tapping completed step triggers onStepTapped callback', (WidgetTester tester) async {
      int? tappedIndex;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: BookingStepProgressBar(
              currentStep: 3,
              onStepTapped: (index) => tappedIndex = index,
            ),
          ),
        ),
      );

      // Tap on step 1 (completed)
      await tester.tap(find.text('Rooms'));
      await tester.pump();

      expect(tappedIndex, 1);
    });
  });

  group('TagaTowersHeroCard Widget Tests', () {
    testWidgets('Renders title, subtitle, and location pill', (WidgetTester tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SingleChildScrollView(
              child: TagaTowersHeroCard(),
            ),
          ),
        ),
      );

      expect(find.text('TAGA Towers'), findsOneWidget);
      expect(find.text('Comfortable accommodation facility for TAGA members in Chennai'), findsOneWidget);
      expect(find.text('Egmore, Chennai'), findsOneWidget);
    });

    testWidgets('Tapping hero card fires onTap callback', (WidgetTester tester) async {
      bool tapped = false;

      tester.view.physicalSize = const Size(1080, 2400);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(() {
        tester.view.resetPhysicalSize();
        tester.view.resetDevicePixelRatio();
      });

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SingleChildScrollView(
              child: TagaTowersHeroCard(
                onTap: () => tapped = true,
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('TAGA Towers'));
      await tester.pump();

      expect(tapped, isTrue);
    });
  });

  group('PriceBreakdownCard Widget Tests', () {
    testWidgets('Calculates advance payment accurately and shows all line items', (WidgetTester tester) async {
      bool payClicked = false;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SingleChildScrollView(
              child: PriceBreakdownCard(
                roomType: 'apex-suite',
                roomNumber: 'Apex 1',
                stayDurationDays: 2,
                bedCount: 2,
                bookingFor: 'guest',
                onPayPressed: () => payClicked = true,
              ),
            ),
          ),
        ),
      );

      // Line items
      expect(find.text('Booking Summary'), findsOneWidget);
      expect(find.text('Apex Suite A/C (Apex 1)'), findsOneWidget);
      expect(find.text('Guest Booking'), findsOneWidget);
      expect(find.text('2 Day(s) • 48 Hours'), findsOneWidget);
      expect(find.text('2 Beds'), findsOneWidget);
      expect(find.text('₹200'), findsOneWidget); // rate per bed

      // Total advance calculation: ₹200 * 2 = ₹400
      expect(find.text('₹400'), findsOneWidget);
      expect(find.text('₹200 × 2 Beds'), findsOneWidget);

      // Disclaimers
      expect(
        find.text('Remaining balance to be settled after stay upon checkout.'),
        findsOneWidget,
      );
      expect(
        find.textContaining('Cancellation Policy: Please note that the advance booking payment is strictly non-refundable'),
        findsOneWidget,
      );

      // Tap CTA
      expect(find.text('Proceed to Pay ₹400'), findsOneWidget);
      await tester.tap(find.text('Proceed to Pay ₹400'));
      await tester.pump();
      expect(payClicked, isTrue);
    });

    testWidgets('Dormitory rate calculation ₹100 per bed', (WidgetTester tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SingleChildScrollView(
              child: PriceBreakdownCard(
                roomType: 'gents-dorm',
                stayDurationDays: 1,
                bedCount: 3,
                bookingFor: 'self',
              ),
            ),
          ),
        ),
      );

      // Advance: ₹100 * 3 = ₹300
      expect(find.text('₹300'), findsOneWidget);
      expect(find.text('₹100 × 3 Beds'), findsOneWidget);
    });
  });

  group('TagatowersAccessRestrictedScreen Widget Tests', () {
    testWidgets('Renders restricted screen, explanation, and action buttons', (WidgetTester tester) async {
      bool manageSubscriptionPressed = false;
      bool backHomePressed = false;

      await tester.pumpWidget(
        MaterialApp(
          home: TagatowersAccessRestrictedScreen(
            onManageSubscription: () => manageSubscriptionPressed = true,
            onBackToHome: () => backHomePressed = true,
          ),
        ),
      );

      expect(find.text('TAGA Towers Accommodation'), findsOneWidget);
      expect(find.text('Paid Members Only'), findsOneWidget);
      expect(find.text('Subscriber Access Only'), findsOneWidget);
      expect(find.textContaining('reserved for active annual subscribers'), findsOneWidget);

      // Benefits list
      expect(find.text('Subsidized accommodation in prime Egmore, Chennai'), findsOneWidget);
      expect(find.text('24-Hour stay flexibility with A/C suites & dormitories'), findsOneWidget);

      // Buttons
      await tester.scrollUntilVisible(find.text('Manage Subscription'), 100);
      await tester.tap(find.text('Manage Subscription'));
      await tester.pump();
      expect(manageSubscriptionPressed, isTrue);

      await tester.scrollUntilVisible(find.text('Back to Home'), 100);
      await tester.tap(find.text('Back to Home'));
      await tester.pump();
      expect(backHomePressed, isTrue);
    });
  });
}
