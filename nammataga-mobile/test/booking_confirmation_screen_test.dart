import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nammataga_mobile/features/tagatowers/domain/booking_model.dart';
import 'package:nammataga_mobile/features/tagatowers/presentation/booking_confirmation_screen.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  final confirmedSelfBooking = BookingResponse(
    id: 'TT-CONF-9841',
    roomId: 'apex-1',
    roomName: 'Apex Suite A/C',
    checkInDate: '2026-10-15',
    checkOutDate: '2026-10-18',
    bookerName: 'Karthik Raja',
    bookerId: 'TAGA-1082',
    bookerPhone: '+91 98401 23456',
    bookingFor: 'self',
    bedCount: 1,
    gender: 'male',
    paymentStatus: 'confirmed',
    advanceAmount: 200,
    bookingStatus: 'upcoming',
    razorpayPaymentId: 'pay_ABC123456789',
    razorpayOrderId: 'order_XYZ987654321',
  );

  final confirmedGuestBooking = BookingResponse(
    id: 'TT-GUEST-5521',
    roomId: 'gents-dorm',
    roomName: 'Gents Dormitory',
    checkInDate: '2026-11-01',
    checkOutDate: '2026-11-03',
    bookerName: 'Suresh Kumar',
    bookerId: 'TAGA-2041',
    bookingFor: 'guest',
    bedCount: 2,
    gender: 'male',
    guestDetails: const [
      GuestDetail(
        name: 'Ravi Teja',
        age: 29,
        contact: '+91 98765 11111',
        gender: 'male',
      ),
      GuestDetail(
        name: 'Manoj Pillai',
        age: 33,
        contact: '+91 98765 22222',
        gender: 'male',
      ),
    ],
    paymentStatus: 'confirmed',
    advanceAmount: 200,
    bookingStatus: 'upcoming',
  );

  final pendingBooking = BookingResponse(
    id: 'TT-PEND-1001',
    roomId: 'pavalam',
    roomName: 'Pavalam',
    checkInDate: '2026-10-20',
    checkOutDate: '2026-10-21',
    bookerName: 'Anand Kumar',
    bookerId: 'TAGA-3099',
    bookingFor: 'self',
    bedCount: 1,
    paymentStatus: 'pending',
    advanceAmount: 200,
    bookingStatus: 'upcoming',
  );

  Widget createSubject(BookingResponse booking, {String? paymentId, String? orderId}) {
    return ProviderScope(
      child: MaterialApp(
        home: BookingConfirmationScreen(
          booking: booking,
          razorpayPaymentId: paymentId,
          razorpayOrderId: orderId,
        ),
      ),
    );
  }

  group('BookingConfirmationScreen Tests', () {
    testWidgets('Strict Rule: Displays celebratory confirmation when booking is confirmed',
        (tester) async {
      await tester.pumpWidget(createSubject(confirmedSelfBooking));
      await tester.pumpAndSettle();

      // Celebratory header
      expect(find.text('Booking Confirmed!'), findsOneWidget);
      expect(find.textContaining('Karthik Raja'), findsOneWidget);

      // Official Booking Reference ID
      expect(find.text('OFFICIAL BOOKING REFERENCE ID'), findsOneWidget);
      expect(find.text('TT-CONF-9841'), findsOneWidget);
      expect(find.text('Copy'), findsWidgets);

      // Room name & type badge
      expect(find.text('Apex Suite A/C'), findsWidgets);
      expect(find.text('Self Booking'), findsOneWidget);
      expect(find.text('1 Bed Booked'), findsOneWidget);

      // Check-in & Check-out dates (12:00 AM Midnight)
      expect(find.text('CHECK-IN'), findsOneWidget);
      expect(find.text('CHECK-OUT'), findsOneWidget);
      expect(find.text('12:00 AM (Midnight)'), findsNWidgets(2));

      // Stay duration
      expect(find.textContaining('Stay Duration: 3 Days'), findsOneWidget);
      expect(find.textContaining('72 Hours'), findsOneWidget);

      // Payment Details card
      expect(find.text('PAYMENT DETAILS'), findsOneWidget);
      expect(find.text('Advance Amount Paid'), findsOneWidget);
      expect(find.text('₹200'), findsOneWidget);
      expect(find.text('Confirmed'), findsOneWidget);
      expect(find.text('pay_ABC123456789'), findsOneWidget);
      expect(find.text('order_XYZ987654321'), findsOneWidget);
      expect(find.textContaining('Remaining Balance:'), findsOneWidget);

      // Caretaker contacts & Egmore Chennai
      expect(find.text('CHECK-IN INSTRUCTIONS & CONTACTS'), findsOneWidget);
      expect(find.textContaining('Egmore, Chennai'), findsNWidgets(2));
      expect(find.text('Mr. Muthu'), findsOneWidget);
      expect(find.text('96007 63744'), findsOneWidget);
      expect(find.text('Mr. Mariyappan'), findsOneWidget);
      expect(find.text('96770 10300'), findsOneWidget);

      // Non-Refundable Cancellation notice
      expect(find.textContaining('Cancellation Policy (Non-Refundable Advance)'), findsOneWidget);
      expect(find.textContaining('advance payment of ₹200 is strictly non-refundable'), findsOneWidget);

      // Navigation Action Buttons
      expect(find.text('View in My Bookings'), findsOneWidget);
      expect(find.text('Back to Home'), findsWidgets);
    });

    testWidgets('Guest booking renders Guest List summary with details', (tester) async {
      await tester.pumpWidget(createSubject(confirmedGuestBooking));
      await tester.pumpAndSettle();

      expect(find.text('Guest Booking'), findsOneWidget);
      expect(find.text('GUEST INFORMATION (2)'), findsOneWidget);
      expect(find.text('Ravi Teja'), findsOneWidget);
      expect(find.text('Age: 29'), findsOneWidget);
      expect(find.text('+91 98765 11111'), findsOneWidget);
      expect(find.text('Manoj Pillai'), findsOneWidget);
      expect(find.text('Age: 33'), findsOneWidget);
      expect(find.text('+91 98765 22222'), findsOneWidget);
    });

    testWidgets('Strict Rule: Displays unconfirmed safety screen when paymentStatus != confirmed',
        (tester) async {
      await tester.pumpWidget(createSubject(pendingBooking));
      await tester.pumpAndSettle();

      // Should show safety unconfirmed state instead of celebratory confirmation
      expect(find.text('🎉 Booking Confirmed!'), findsNothing);
      expect(find.text('Awaiting Confirmation'), findsOneWidget);
      expect(find.textContaining('Current payment status: "pending"'), findsOneWidget);
      expect(find.text('Check Status in My Bookings'), findsOneWidget);
      expect(find.text('Back to Home'), findsWidgets);
    });

    testWidgets('Tapping copy button on Booking Reference ID copies text', (tester) async {
      await tester.pumpWidget(createSubject(confirmedSelfBooking));
      await tester.pumpAndSettle();

      // Find first copy button
      final copyBtn = find.text('Copy').first;
      await tester.tap(copyBtn);
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      // Verify "Copied" text appears
      expect(find.text('Copied'), findsOneWidget);
    });
  });
}
