import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../../core/constants/app_colors.dart';
import '../../../../core/payment/razorpay_checkout_manager.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../auth/providers/auth_provider.dart';
import '../domain/booking_model.dart';
import '../domain/payment_model.dart';
import '../providers/tagatowers_provider.dart';
import 'theme/tagatowers_design_tokens.dart';
import 'widgets/booking_step_progress_bar.dart';

/// Screen representing Step 4 of the TAGA Towers booking flow:
/// Pre-Payment Review & Razorpay Gateway Integration.
///
/// Features:
/// - Step progress header: [BookingStepProgressBar(currentStep: 4)]
/// - Full accommodation summary card (Room details, stay dates, bed count, booking type)
/// - Backend-validated advance amount calculation & prominent display
/// - Non-refundable cancellation policy notice & explicit terms checkbox
/// - Order creation & Razorpay checkout launch
/// - Robust callbacks: verification, recovery redirection on error, and automatic hold release on dismissal
class PaymentReviewScreen extends ConsumerStatefulWidget {
  /// The pending booking instance to review and pay for.
  /// If null, falls back to [bookingFlowNotifierProvider.pendingBooking].
  final BookingResponse? booking;

  const PaymentReviewScreen({
    super.key,
    this.booking,
  });

  @override
  ConsumerState<PaymentReviewScreen> createState() => _PaymentReviewScreenState();
}

class _PaymentReviewScreenState extends ConsumerState<PaymentReviewScreen> {
  late final RazorpayCheckoutManager _checkoutManager;

  bool _isProcessing = false;
  bool _termsAccepted = false;
  String? _errorMessage;
  bool _isBookingCancelled = false;

  final DateFormat _dateFormat = DateFormat('EEE, dd MMM yyyy');

  @override
  void initState() {
    super.initState();
    _checkoutManager = RazorpayCheckoutManager();
  }

  @override
  void dispose() {
    _checkoutManager.dispose();
    super.dispose();
  }

  /// Resolves the active booking instance to pay for.
  BookingResponse? _resolveBooking(BookingFlowState flowState) {
    if (widget.booking != null) {
      return widget.booking;
    }
    if (flowState.pendingBooking != null) {
      return flowState.pendingBooking;
    }
    // Reconstruct a fallback instance if pendingBookingId is present
    if (flowState.pendingBookingId != null && flowState.selectedRoom != null) {
      final checkInStr = flowState.checkInDate != null
          ? DateFormat('yyyy-MM-dd').format(flowState.checkInDate!)
          : '';
      final checkOutStr = flowState.checkOutDate != null
          ? DateFormat('yyyy-MM-dd').format(flowState.checkOutDate!)
          : '';

      return BookingResponse(
        id: flowState.pendingBookingId!,
        roomId: flowState.selectedRoom!.id,
        roomName: flowState.selectedRoom!.name,
        checkInDate: checkInStr,
        checkOutDate: checkOutStr,
        bookerName: 'TAGA Member',
        bookerId: '',
        bookerPhone: flowState.bookerPhone,
        bookingFor: flowState.bookingFor,
        bedCount: flowState.effectiveBeds,
        guestDetails: flowState.guestDetails,
        paymentStatus: 'pending',
        advanceAmount: flowState.advanceAmount,
        bookingStatus: 'upcoming',
      );
    }
    return null;
  }

  /// Initiates Razorpay Order Creation and opens the payment gateway modal.
  Future<void> _handlePayment(BookingResponse booking) async {
    if (_isProcessing || !_termsAccepted) return;

    HapticFeedback.lightImpact();

    setState(() {
      _isProcessing = true;
      _errorMessage = null;
    });

    final repo = ref.read(tagaTowersRepositoryProvider);
    final authSession = ref.read(authProvider).value;
    final flowState = ref.read(bookingFlowNotifierProvider);

    // Backend-validated safety: Amount sent to Razorpay order creation
    // MUST strictly equal booking.advanceAmount * 100 (in paise).
    final int amountInPaise = booking.advanceAmount * 100;

    final String bookerName = booking.bookerName.isNotEmpty && booking.bookerName != 'Member'
        ? booking.bookerName
        : (authSession?.displayName ?? 'TAGA Member');
    final String bookerEmail = authSession?.memberEmail ?? '';
    final String bookerContact = (booking.bookerPhone ?? flowState.bookerPhone)
        .replaceAll(' ', '')
        .replaceAll('-', '');

    try {
      // Build order notes for admin notification & backend audit
      final notes = <String, dynamic>{
        'booking_id': booking.id,
        'room_name': booking.roomName,
        'bed_count': booking.bedCount.toString(),
        'check_in_date': booking.checkInDate,
        'check_out_date': booking.checkOutDate,
        'booker_name': bookerName,
        'booker_email': bookerEmail,
        'booker_phone': bookerContact,
        'booking_type': booking.bookingFor,
        'advance_amount_paise': amountInPaise.toString(),
        'advance_amount_inr': booking.advanceAmount.toString(),
      };

      // 1. Create Razorpay order on backend
      final orderResponse = await repo.createOrder(
        amountInPaise: amountInPaise,
        notes: notes,
      );

      if (!mounted) return;

      // 2. Prepare Razorpay SDK options
      final options = <String, dynamic>{
        'key': orderResponse.key,
        'amount': orderResponse.amount, // in paise
        'name': 'TAGA Towers',
        'description': 'Room Booking Advance',
        'order_id': orderResponse.orderId,
        'prefill': {
          'name': bookerName,
          'email': bookerEmail,
          'contact': bookerContact,
        },
        'theme': {
          'color': '#2D6A4F',
        },
      };

      // 3. Open Razorpay gateway
      _checkoutManager.openCheckout(
        options: options,
        onSuccess: (paymentId, orderId, signature) async {
          await _onPaymentSuccess(
            booking: booking,
            paymentId: paymentId,
            orderId: orderId.isNotEmpty ? orderId : orderResponse.orderId,
            signature: signature,
          );
        },
        onError: (errorMsg) {
          _onPaymentError(errorMsg);
        },
        onDismiss: () async {
          await _onPaymentDismissed(booking);
        },
      );
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isProcessing = false;
        _errorMessage = e.toString().replaceAll('Exception: ', '');
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Payment setup failed: $_errorMessage'),
          backgroundColor: AppColors.destructive,
          duration: const Duration(seconds: 4),
        ),
      );
    }
  }

  /// Handles successful gateway payment: verifies signature on backend.
  Future<void> _onPaymentSuccess({
    required BookingResponse booking,
    required String paymentId,
    required String orderId,
    required String signature,
  }) async {
    final repo = ref.read(tagaTowersRepositoryProvider);

    final verifyReq = VerifyTowersPaymentRequest(
      bookingId: booking.id,
      razorpayOrderId: orderId,
      razorpayPaymentId: paymentId,
      razorpaySignature: signature,
    );

    try {
      // Verify payment with backend
      await repo.verifyPayment(verifyReq);

      if (!mounted) return;

      // Clear wizard flow state
      ref.read(bookingFlowNotifierProvider.notifier).reset();

      final confirmedBooking = booking.copyWith(
        paymentStatus: 'confirmed',
        bookingStatus: 'upcoming',
        razorpayOrderId: orderId,
        razorpayPaymentId: paymentId,
      );

      // Navigate to confirmation screen
      context.go('/taga-towers/confirmation', extra: confirmedBooking);
    } catch (e) {
      debugPrint('[PaymentReviewScreen] Backend payment verification error: $e');

      // Backend verification failed or timed out:
      // Save payment credentials in flow state for recovery
      ref.read(bookingFlowNotifierProvider.notifier).setPaymentInterrupted(
            bookingId: booking.id,
            orderId: orderId,
            paymentId: paymentId,
            signature: signature,
          );

      if (!mounted) return;

      // Navigate immediately to the dedicated payment recovery screen
      context.go(
        '/taga-towers/payment-recovery',
        extra: {
          'booking': booking,
          'paymentId': paymentId,
          'orderId': orderId,
          'signature': signature,
          'error': e.toString(),
        },
      );
    }
  }

  /// Handles user closing or cancelling the Razorpay modal.
  /// Releases the 10-minute bed hold immediately via DELETE /api/towers/bookings/:id.
  Future<void> _onPaymentDismissed(BookingResponse booking) async {
    final repo = ref.read(tagaTowersRepositoryProvider);

    try {
      await repo.cancelBooking(booking.id);
    } catch (e) {
      debugPrint('[PaymentReviewScreen] Error releasing pending booking on dismiss: $e');
    }

    if (!mounted) return;

    setState(() {
      _isProcessing = false;
      _isBookingCancelled = true;
    });

    ref.read(bookingFlowNotifierProvider.notifier).setPendingBooking(null);

    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Payment cancelled. Booking removed.'),
        backgroundColor: AppColors.textSecondary,
        duration: Duration(seconds: 3),
      ),
    );
  }

  /// Handles payment failures returned by the gateway.
  void _onPaymentError(String errorMsg) {
    if (!mounted) return;

    setState(() {
      _isProcessing = false;
      _errorMessage = errorMsg;
    });

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('Payment Failed: $errorMsg'),
        backgroundColor: AppColors.destructive,
        duration: const Duration(seconds: 4),
        action: SnackBarAction(
          label: 'Dismiss',
          textColor: Colors.white,
          onPressed: () {
            ScaffoldMessenger.of(context).hideCurrentSnackBar();
          },
        ),
      ),
    );
  }

  /// Prompts confirmation before leaving to release bed hold.
  Future<bool> _onWillPop(BookingResponse booking) async {
    if (_isBookingCancelled) {
      return true;
    }

    final shouldCancel = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Cancel Pending Booking?'),
        content: const Text(
          'Leaving this screen will release your 10-minute reserved bed hold. '
          'Are you sure you want to cancel this booking?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('Stay & Pay'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: AppColors.destructive,
            ),
            onPressed: () => Navigator.of(ctx).pop(true),
            child: const Text('Cancel Booking'),
          ),
        ],
      ),
    );

    if (shouldCancel == true) {
      try {
        await ref.read(tagaTowersRepositoryProvider).cancelBooking(booking.id);
      } catch (e) {
        debugPrint('[PaymentReviewScreen] Cancel booking on back error: $e');
      }
      ref.read(bookingFlowNotifierProvider.notifier).setPendingBooking(null);
      return true;
    }

    return false;
  }

  @override
  Widget build(BuildContext context) {
    final flowState = ref.watch(bookingFlowNotifierProvider);
    final booking = _resolveBooking(flowState);

    // If no booking found to review
    if (booking == null) {
      return Scaffold(
        backgroundColor: TagaTowersDesignTokens.surfaceLight,
        appBar: AppBar(
          title: const Text('Payment Review'),
          backgroundColor: Colors.white,
          elevation: 0,
        ),
        body: Center(
          child: Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(
                  LucideIcons.calendarX,
                  size: 52,
                  color: AppColors.textSecondary,
                ),
                const SizedBox(height: 16),
                const Text(
                  'No Pending Booking Found',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
                const SizedBox(height: 8),
                const Text(
                  'Please select a room and review your stay details before proceeding to payment.',
                  textAlign: TextAlign.center,
                  style: TextStyle(
                    fontSize: 13.5,
                    color: AppColors.textSecondary,
                  ),
                ),
                const SizedBox(height: 24),
                ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: TagaTowersDesignTokens.primaryGreen,
                    foregroundColor: Colors.white,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  onPressed: () => context.go('/taga-towers'),
                  child: const Text('Return to TAGA Towers'),
                ),
              ],
            ),
          ),
        ),
      );
    }

    // If the booking was cancelled via dismissal or back
    if (_isBookingCancelled) {
      return Scaffold(
        backgroundColor: TagaTowersDesignTokens.surfaceLight,
        appBar: AppBar(
          title: const Text('Payment Cancelled'),
          backgroundColor: Colors.white,
          elevation: 0,
          leading: IconButton(
            icon: const Icon(LucideIcons.arrowLeft),
            onPressed: () => context.go('/taga-towers'),
          ),
        ),
        body: Center(
          child: Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: AppColors.destructive.withValues(alpha: 0.1),
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(
                    LucideIcons.xCircle,
                    size: 48,
                    color: AppColors.destructive,
                  ),
                ),
                const SizedBox(height: 16),
                const Text(
                  'Booking Hold Released',
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
                const SizedBox(height: 8),
                const Text(
                  'Your reservation has been cancelled and the room beds have been released back to availability.',
                  textAlign: TextAlign.center,
                  style: TextStyle(
                    fontSize: 13.5,
                    color: AppColors.textSecondary,
                    height: 1.4,
                  ),
                ),
                const SizedBox(height: 24),
                ElevatedButton.icon(
                  icon: const Icon(LucideIcons.rotateCcw, size: 18),
                  label: const Text('Start New Booking'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: TagaTowersDesignTokens.primaryGreen,
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  onPressed: () {
                    ref.read(bookingFlowNotifierProvider.notifier).reset();
                    context.go('/taga-towers');
                  },
                ),
              ],
            ),
          ),
        ),
      );
    }

    DateTime? checkInParsed;
    DateTime? checkOutParsed;
    try {
      checkInParsed = DateTime.parse(booking.checkInDate);
      checkOutParsed = DateTime.parse(booking.checkOutDate);
    } catch (_) {}

    final advanceAmount = booking.advanceAmount;
    final advancePerBed = booking.bedCount > 0 ? (advanceAmount ~/ booking.bedCount) : advanceAmount;

    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, result) async {
        if (didPop) return;
        final shouldLeave = await _onWillPop(booking);
        if (shouldLeave && context.mounted) {
          context.pop();
        }
      },
      child: Scaffold(
        backgroundColor: TagaTowersDesignTokens.surfaceLight,
        appBar: AppBar(
          title: const Text(
            'Payment Review',
            style: TextStyle(
              fontWeight: FontWeight.bold,
              fontSize: 18,
              color: AppColors.textPrimary,
            ),
          ),
          backgroundColor: Colors.white,
          elevation: 0,
          leading: IconButton(
            icon: const Icon(LucideIcons.arrowLeft, color: AppColors.textPrimary),
            onPressed: () async {
              final shouldLeave = await _onWillPop(booking);
              if (shouldLeave && context.mounted) {
                context.pop();
              }
            },
          ),
          actions: [
            Container(
              margin: const EdgeInsets.only(right: 14),
              child: Center(
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: TagaTowersDesignTokens.softMint,
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(
                      color: TagaTowersDesignTokens.mintHighlight,
                      width: 1,
                    ),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(
                        LucideIcons.lock,
                        size: 11,
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                      const SizedBox(width: 4),
                      Text(
                        'HOLD: #${booking.id.length > 8 ? booking.id.substring(0, 8) : booking.id}',
                        style: const TextStyle(
                          fontSize: 10.5,
                          fontWeight: FontWeight.bold,
                          color: TagaTowersDesignTokens.deepEmerald,
                          letterSpacing: 0.3,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ],
        ),
        body: SafeArea(
          child: Column(
            children: [
              // Scrollable Review Content
              Expanded(
                child: SingleChildScrollView(
                  physics: const BouncingScrollPhysics(),
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Wizard Step 4 Header
                      const BookingStepProgressBar(currentStep: 4),
                      const SizedBox(height: 16),

                      // Error message banner if order creation or payment failed
                      if (_errorMessage != null) ...[
                        _buildErrorBanner(_errorMessage!, () => _handlePayment(booking)),
                        const SizedBox(height: 16),
                      ],

                      // 10-Minute Hold Notice
                      _buildHoldCountdownNotice(),
                      const SizedBox(height: 16),

                      // Room & Reservation Details Card
                      _buildRoomDetailsCard(
                        booking: booking,
                        checkIn: checkInParsed,
                        checkOut: checkOutParsed,
                      ),
                      const SizedBox(height: 16),

                      // Prominent Advance Amount Box
                      _buildAdvanceAmountCard(
                        advanceAmount: advanceAmount,
                        bedCount: booking.bedCount,
                        advancePerBed: advancePerBed,
                      ),
                      const SizedBox(height: 16),

                      // Prominent Non-Refundable Cancellation Policy Alert
                      _buildNonRefundablePolicyNotice(),
                      const SizedBox(height: 16),

                      // Terms & Acknowledgement Checkbox
                      _buildTermsCheckbox(advanceAmount),
                      const SizedBox(height: 24),
                    ],
                  ),
                ),
              ),

              // Bottom Sticky Pay Action Area
              _buildBottomPayBar(booking, advanceAmount),
            ],
          ),
        ),
      ),
    );
  }

  /// Builds a prominent alert warning indicating the temporary 10-minute hold.
  Widget _buildHoldCountdownNotice() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFEFF8F3),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: TagaTowersDesignTokens.mintHighlight,
          width: 1.1,
        ),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(6),
            decoration: const BoxDecoration(
              color: TagaTowersDesignTokens.primaryGreen,
              shape: BoxShape.circle,
            ),
            child: const Icon(
              LucideIcons.timer,
              size: 14,
              color: Colors.white,
            ),
          ),
          const SizedBox(width: 10),
          const Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  '10-Minute Bed Hold Active',
                  style: TextStyle(
                    fontSize: 12.5,
                    fontWeight: FontWeight.bold,
                    color: TagaTowersDesignTokens.deepEmerald,
                  ),
                ),
                SizedBox(height: 2),
                Text(
                  'Your selected beds are locked. Complete payment to secure reservation.',
                  style: TextStyle(
                    fontSize: 11.5,
                    color: AppColors.textSecondary,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// Builds the room summary card showing room, stay dates, occupancy, and guest list.
  Widget _buildRoomDetailsCard({
    required BookingResponse booking,
    DateTime? checkIn,
    DateTime? checkOut,
  }) {
    final roomName = booking.roomName.isNotEmpty
        ? booking.roomName
        : TagaTowersDesignTokens.getRoomDisplayName(booking.roomId);

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 16,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Room title & badge
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      roomName,
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: AppColors.textPrimary,
                        letterSpacing: -0.3,
                      ),
                    ),
                    const SizedBox(height: 3),
                    const Row(
                      children: [
                        Icon(
                          LucideIcons.mapPin,
                          size: 13,
                          color: AppColors.textSecondary,
                        ),
                        SizedBox(width: 4),
                        Text(
                          'TAGA Towers, Egmore, Chennai',
                          style: TextStyle(
                            fontSize: 12,
                            color: AppColors.textSecondary,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                decoration: BoxDecoration(
                  color: TagaTowersDesignTokens.softMint,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  booking.bookingFor.toLowerCase() == 'guest' ? 'Guest Booking' : 'Self Booking',
                  style: const TextStyle(
                    fontSize: 11.5,
                    fontWeight: FontWeight.bold,
                    color: TagaTowersDesignTokens.deepEmerald,
                  ),
                ),
              ),
            ],
          ),
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 12),
            child: Divider(color: Color(0xFFE2EBE5), height: 1),
          ),

          // Stay Schedule (24-Hour Midnight Cycle)
          Row(
            children: [
              Expanded(
                child: _buildDateColumn(
                  label: 'CHECK-IN',
                  date: checkIn != null ? _dateFormat.format(checkIn) : booking.checkInDate,
                  time: TagaTowersDesignTokens.checkInTime,
                  icon: LucideIcons.calendarCheck,
                ),
              ),
              Container(
                height: 40,
                width: 1,
                color: const Color(0xFFE2EBE5),
                margin: const EdgeInsets.symmetric(horizontal: 12),
              ),
              Expanded(
                child: _buildDateColumn(
                  label: 'CHECK-OUT',
                  date: checkOut != null ? _dateFormat.format(checkOut) : booking.checkOutDate,
                  time: TagaTowersDesignTokens.checkOutTime,
                  icon: LucideIcons.calendarDays,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),

          // 24-hr Stay Cycle Disclaimer Badge
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
            decoration: BoxDecoration(
              color: const Color(0xFFF8FAF9),
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: const Color(0xFFE0EBE4)),
            ),
            child: const Row(
              children: [
                Icon(
                  LucideIcons.clock,
                  size: 13,
                  color: TagaTowersDesignTokens.primaryGreen,
                ),
                SizedBox(width: 6),
                Expanded(
                  child: Text(
                    '24-Hour Stay Schedule (12:00 AM Check-in to 12:00 AM Checkout)',
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                      color: TagaTowersDesignTokens.deepEmerald,
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),

          // Bed count & Guest summary row
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  const Icon(
                    LucideIcons.bedDouble,
                    size: 15,
                    color: AppColors.textSecondary,
                  ),
                  const SizedBox(width: 6),
                  Text(
                    '${booking.bedCount} Bed${booking.bedCount > 1 ? 's' : ''} Reserved',
                    style: const TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: AppColors.textPrimary,
                    ),
                  ),
                ],
              ),
              if (booking.gender != null && booking.gender!.isNotEmpty)
                Text(
                  'Gender: ${booking.gender![0].toUpperCase()}${booking.gender!.substring(1)}',
                  style: const TextStyle(
                    fontSize: 12,
                    color: AppColors.textSecondary,
                    fontWeight: FontWeight.w500,
                  ),
                ),
            ],
          ),

          // Optional Guest details list for guest booking
          if (booking.bookingFor.toLowerCase() == 'guest' &&
              booking.guestDetails.isNotEmpty) ...[
            const SizedBox(height: 10),
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: const Color(0xFFF9FBFA),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: const Color(0xFFE5EDE8)),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Guest Details:',
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textSecondary,
                    ),
                  ),
                  const SizedBox(height: 4),
                  ...booking.guestDetails.asMap().entries.map((entry) {
                    final idx = entry.key + 1;
                    final guest = entry.value;
                    return Padding(
                      padding: const EdgeInsets.only(top: 2),
                      child: Text(
                        '$idx. ${guest.name} (${guest.age} yrs, ${guest.gender.isNotEmpty ? guest.gender : "N/A"}) - ${guest.contact}',
                        style: const TextStyle(
                          fontSize: 11.5,
                          color: AppColors.textPrimary,
                        ),
                      ),
                    );
                  }),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildDateColumn({
    required String label,
    required String date,
    required String time,
    required IconData icon,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(icon, size: 12, color: TagaTowersDesignTokens.primaryGreen),
            const SizedBox(width: 4),
            Text(
              label,
              style: const TextStyle(
                fontSize: 10.5,
                fontWeight: FontWeight.w800,
                letterSpacing: 0.5,
                color: TagaTowersDesignTokens.primaryGreen,
              ),
            ),
          ],
        ),
        const SizedBox(height: 4),
        Text(
          date,
          style: const TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: 1),
        Text(
          time,
          style: const TextStyle(
            fontSize: 11,
            color: AppColors.textSecondary,
          ),
        ),
      ],
    );
  }

  /// Builds the prominent Advance Amount Box.
  Widget _buildAdvanceAmountCard({
    required int advanceAmount,
    required int bedCount,
    required int advancePerBed,
  }) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: const Color(0xFFF1F8E9),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: TagaTowersDesignTokens.mintHighlight,
          width: 1.5,
        ),
        boxShadow: TagaTowersDesignTokens.softShadow,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Advance Fee to Pay Now',
                    style: TextStyle(
                      fontSize: 13.5,
                      fontWeight: FontWeight.bold,
                      color: TagaTowersDesignTokens.deepEmerald,
                    ),
                  ),
                  SizedBox(height: 2),
                  Text(
                    'Non-refundable confirmation deposit',
                    style: TextStyle(
                      fontSize: 11.5,
                      color: AppColors.textSecondary,
                    ),
                  ),
                ],
              ),
              // Prominent Advance Amount Display
              Text(
                '₹$advanceAmount',
                style: const TextStyle(
                  fontSize: 28,
                  fontWeight: FontWeight.w900,
                  color: TagaTowersDesignTokens.primaryGreen,
                  letterSpacing: -0.5,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          const Divider(color: Color(0xFFCCE3D5), height: 1),
          const SizedBox(height: 10),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'Rate Calculation: ₹$advancePerBed × $bedCount Bed${bedCount > 1 ? 's' : ''}',
                style: const TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                  color: AppColors.textPrimary,
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(6),
                  border: Border.all(color: TagaTowersDesignTokens.mintHighlight),
                ),
                child: const Text(
                  'Payable via UPI / Cards',
                  style: TextStyle(
                    fontSize: 10.5,
                    fontWeight: FontWeight.w700,
                    color: TagaTowersDesignTokens.deepEmerald,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          const Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Padding(
                padding: EdgeInsets.only(top: 2),
                child: Icon(
                  LucideIcons.info,
                  size: 13,
                  color: AppColors.textSecondary,
                ),
              ),
              SizedBox(width: 6),
              Expanded(
                child: Text(
                  'The remaining room balance will be settled at the reception upon check-in or checkout.',
                  style: TextStyle(
                    fontSize: 11,
                    color: AppColors.textSecondary,
                    height: 1.35,
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  /// Builds the Non-Refundable Cancellation Policy Alert.
  Widget _buildNonRefundablePolicyNotice() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(14),
      decoration: TagaTowersDesignTokens.alertDecoration(
        borderRadius: 14,
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.all(6),
            decoration: BoxDecoration(
              color: TagaTowersDesignTokens.warningAmber.withValues(alpha: 0.2),
              shape: BoxShape.circle,
            ),
            child: const Icon(
              LucideIcons.alertTriangle,
              size: 18,
              color: TagaTowersDesignTokens.warningAmberDark,
            ),
          ),
          const SizedBox(width: 12),
          const Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Non-Refundable Policy Notice',
                  style: TextStyle(
                    fontSize: 13.5,
                    fontWeight: FontWeight.bold,
                    color: TagaTowersDesignTokens.warningAmberDark,
                  ),
                ),
                SizedBox(height: 4),
                Text(
                  'TAGA Towers operates on held accommodation inventory. Advance payments are strictly non-refundable upon cancellation.\n\n'
                  'Cancellations are only permitted strictly prior to your check-in date to release beds for fellow members.',
                  style: TextStyle(
                    fontSize: 11.5,
                    color: Color(0xFF78350F),
                    height: 1.4,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// Builds the Terms & Conditions acknowledgement checkbox.
  Widget _buildTermsCheckbox(int advanceAmount) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: _termsAccepted
              ? TagaTowersDesignTokens.primaryGreen
              : const Color(0xFFD6E2DA),
          width: 1.2,
        ),
      ),
      child: InkWell(
        onTap: _isProcessing
            ? null
            : () {
                HapticFeedback.selectionClick();
                setState(() {
                  _termsAccepted = !_termsAccepted;
                });
              },
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Checkbox(
              value: _termsAccepted,
              activeColor: TagaTowersDesignTokens.primaryGreen,
              materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
              onChanged: _isProcessing
                  ? null
                  : (val) {
                      HapticFeedback.selectionClick();
                      setState(() {
                        _termsAccepted = val ?? false;
                      });
                    },
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                'I acknowledge the 24-hour stay schedule (12:00 AM check-in / 12:00 AM check-out) and agree that this ₹$advanceAmount advance payment is strictly non-refundable upon cancellation.',
                style: const TextStyle(
                  fontSize: 12,
                  color: AppColors.textPrimary,
                  height: 1.4,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Builds an error banner with a quick retry button.
  Widget _buildErrorBanner(String message, VoidCallback onRetry) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: AppColors.destructive.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.destructive.withValues(alpha: 0.3)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          const Icon(LucideIcons.alertCircle, color: AppColors.destructive, size: 20),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              message,
              style: const TextStyle(fontSize: 12, color: AppColors.destructive),
            ),
          ),
          TextButton(
            onPressed: onRetry,
            child: const Text('Retry'),
          ),
        ],
      ),
    );
  }

  /// Builds the bottom sticky payment execution bar.
  Widget _buildBottomPayBar(BookingResponse booking, int advanceAmount) {
    final isPayEnabled = _termsAccepted && !_isProcessing;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.05),
            offset: const Offset(0, -3),
            blurRadius: 10,
          ),
        ],
        border: const Border(
          top: BorderSide(color: Color(0xFFEAEAEA), width: 1),
        ),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // Security disclaimer
          const Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                LucideIcons.shieldCheck,
                size: 13,
                color: TagaTowersDesignTokens.primaryGreen,
              ),
              SizedBox(width: 5),
              Text(
                '128-bit SSL Encrypted • Powered by Razorpay',
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.w500,
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),

          // Primary "Pay ₹XXX Now" Button
          BouncyButton(
            onTap: isPayEnabled ? () => _handlePayment(booking) : null,
            child: AnimatedContainer(
              duration: const Duration(milliseconds: 200),
              width: double.infinity,
              height: 52,
              decoration: BoxDecoration(
                color: isPayEnabled
                    ? TagaTowersDesignTokens.primaryGreen
                    : const Color(0xFFADC7BA),
                borderRadius: BorderRadius.circular(14),
                boxShadow: isPayEnabled ? TagaTowersDesignTokens.floatingGlow : null,
              ),
              child: Center(
                child: _isProcessing
                    ? const Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(
                              strokeWidth: 2.2,
                              valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                            ),
                          ),
                          SizedBox(width: 10),
                          Text(
                            'Connecting to Gateway...',
                            style: TextStyle(
                              color: Colors.white,
                              fontSize: 15,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ],
                      )
                    : Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Icon(
                            LucideIcons.creditCard,
                            color: Colors.white,
                            size: 18,
                          ),
                          const SizedBox(width: 8),
                          Text(
                            'Pay ₹$advanceAmount Now',
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                              letterSpacing: 0.2,
                            ),
                          ),
                        ],
                      ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
