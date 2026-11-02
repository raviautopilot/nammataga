import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/payment_recovery_screen.dart';

void main() {
  Widget createTestWidget({
    PaymentRecoveryStatus? initialStatus,
    String? bookingId = 'BK-TAGA-2026-991',
    String? orderId = 'order_test_98765',
    String? paymentId = 'pay_test_12345',
    String? signature = 'sig_test_abcdef',
    int? amount = 200,
    String? roomName = 'Apex Suite A/C',
  }) {
    return ProviderScope(
      child: MaterialApp(
        home: PaymentRecoveryScreen(
          initialStatus: initialStatus,
          bookingId: bookingId,
          razorpayOrderId: orderId,
          razorpayPaymentId: paymentId,
          razorpaySignature: signature,
          amount: amount,
          roomName: roomName,
        ),
      ),
    );
  }

  testWidgets(
    'PaymentRecoveryScreen displays Verifying Payment status and transaction IDs',
    (WidgetTester tester) async {
      await tester.pumpWidget(
        createTestWidget(initialStatus: PaymentRecoveryStatus.verifying),
      );
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.text('Payment Recovery'), findsOneWidget);
      expect(find.text('Verifying Payment...'), findsOneWidget);
      expect(find.text('BK-TAGA-2026-991'), findsOneWidget);
      expect(find.text('order_test_98765'), findsOneWidget);
      expect(find.text('pay_test_12345'), findsOneWidget);
      expect(find.text('₹200 Advance'), findsOneWidget);
      expect(find.text('Apex Suite A/C'), findsOneWidget);
      expect(find.text('Check / Re-verify Payment Status'), findsOneWidget);
    },
  );

  testWidgets(
    'PaymentRecoveryScreen displays Pending Confirmation banner and member advisory notice',
    (WidgetTester tester) async {
      await tester.pumpWidget(
        createTestWidget(
          initialStatus: PaymentRecoveryStatus.pendingConfirmation,
        ),
      );
      await tester.pump(const Duration(milliseconds: 100));

      expect(
        find.text('Payment Pending Backend Confirmation'),
        findsOneWidget,
      );
      expect(
        find.text('Important Member Safety Advisory'),
        findsOneWidget,
      );
      expect(
        find.textContaining('Do not pay again. If money was deducted from your account'),
        findsOneWidget,
      );
      expect(
        find.textContaining('Duplicate Safety Active: No new payment can be triggered here'),
        findsOneWidget,
      );
    },
  );

  testWidgets(
    'PaymentRecoveryScreen displays Verification Issue status and caretaker support buttons',
    (WidgetTester tester) async {
      await tester.pumpWidget(
        createTestWidget(initialStatus: PaymentRecoveryStatus.issue),
      );
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.text('Payment Verification Issue'), findsOneWidget);
      expect(find.text('On-Site Caretaker & Support Hub'), findsOneWidget);

      // Verify on-site caretaker contacts
      expect(find.text('Mr. Muthu'), findsOneWidget);
      expect(find.text('96007 63744'), findsOneWidget);
      expect(find.text('Mr. Mariyappan'), findsOneWidget);
      expect(find.text('96770 10300'), findsOneWidget);
      expect(find.text('Reception & Helpdesk'), findsOneWidget);
      expect(find.text('044 34919949'), findsOneWidget);
      expect(find.text('Email Payment Support'), findsOneWidget);
      expect(find.text('nammataga@gmail.com'), findsOneWidget);
    },
  );

  testWidgets(
    'PaymentRecoveryScreen displays Verified status and View Booking Confirmation button',
    (WidgetTester tester) async {
      await tester.pumpWidget(
        createTestWidget(initialStatus: PaymentRecoveryStatus.verified),
      );
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.text('Payment Verified Successfully'), findsOneWidget);
      expect(find.text('View Booking Confirmation'), findsOneWidget);
      expect(find.text('Back to TAGA Towers Home'), findsOneWidget);

      // Warning advisory should not appear when already verified
      expect(find.text('Important Member Safety Advisory'), findsNothing);
    },
  );
}
