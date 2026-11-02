import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../auth/providers/auth_provider.dart';
import '../domain/booking_model.dart';
import '../providers/tagatowers_provider.dart';
import 'theme/tagatowers_design_tokens.dart';

/// Cancellation Screen (Agent 14)
///
/// Features:
/// 1. Retrieval: Receives [booking], [bookingId], or [initialBooking].
/// 2. Pre-cancellation Verification: Checks backend rules (today < checkInDate and not cancelled).
///    If today >= checkInDate, displays warning "Cannot cancel booking on or after check-in date." and disables action.
/// 3. Cancellation Terms & No-Refund Notice (CRITICAL):
///    - Prominent Alert card: "No Refund Will Be Provided"
///    - "Please note that according to TAGA Towers booking policies, the advance payment of ₹XXX is strictly non-refundable upon cancellation. No refund will be credited."
///    - Terms checkbox: "I understand that cancelling this reservation is irreversible and that no refund will be provided."
/// 4. Booking Summary Review:
///    - Booking Reference ID, Room Name, Stay Dates, Number of Beds, Advance Paid.
/// 5. Actions:
///    - "Yes, Cancel Booking" button (disabled until checkbox checked).
///    - On success: refreshes providers (`ref.invalidate(userBookingsProvider)`), shows success card with return button.
///    - On failure: displays clear error message with a prominent "Retry Cancellation" button.
///    - "No, Keep Booking" button: safely navigates back.
class CancellationScreen extends ConsumerStatefulWidget {
  final BookingResponse? booking;
  final String? bookingId;
  final BookingResponse? initialBooking;

  const CancellationScreen({
    super.key,
    this.booking,
    this.bookingId,
    this.initialBooking,
  });

  @override
  ConsumerState<CancellationScreen> createState() => _CancellationScreenState();
}

class _CancellationScreenState extends ConsumerState<CancellationScreen> {
  BookingResponse? _booking;
  bool _isLoadingBooking = false;
  String? _fetchError;

  bool _acknowledgedNoRefund = false;
  bool _isCancelling = false;
  bool _isCancelledSuccess = false;
  String? _errorMessage;

  final DateFormat _dateFormat = DateFormat('EEE, dd MMM yyyy');

  @override
  void initState() {
    super.initState();
    final candidate = widget.booking ?? widget.initialBooking;
    if (candidate != null) {
      _booking = candidate;
    } else if (widget.bookingId != null && widget.bookingId!.isNotEmpty) {
      _loadBookingById();
    }
  }

  /// Attempts to fetch the booking by its ID from active or past bookings
  Future<void> _loadBookingById() async {
    final bId = widget.bookingId?.trim() ?? '';
    if (bId.isEmpty) return;

    setState(() {
      _isLoadingBooking = true;
      _fetchError = null;
    });

    try {
      final session = ref.read(authProvider).value;
      final userMap = session?.user;
      final bookerId = (userMap?['taga_id'] ??
              userMap?['tagaId'] ??
              userMap?['sno']?.toString() ??
              userMap?['id']?.toString() ??
              '')
          .toString();

      final repo = ref.read(tagaTowersRepositoryProvider);
      final activeBookings = await repo.getUserBookings(bookerId);
      final match =
          activeBookings.where((b) => b.id.trim() == bId).firstOrNull;

      if (!mounted) return;

      if (match != null) {
        setState(() {
          _booking = match;
          _isLoadingBooking = false;
        });
        return;
      }

      // Check archived past bookings as fallback
      final pastBookings = await repo.getPastBookings(bookerId);
      final pastMatch =
          pastBookings.where((b) => b.id.trim() == bId).firstOrNull;

      if (!mounted) return;

      if (pastMatch != null) {
        setState(() {
          _booking = pastMatch;
          _isLoadingBooking = false;
        });
      } else {
        setState(() {
          _isLoadingBooking = false;
          _fetchError =
              'Reservation #$bId could not be found. It may have already been purged or access is restricted.';
        });
      }
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isLoadingBooking = false;
        _fetchError =
            'Failed to retrieve booking: ${e.toString().replaceAll('Exception: ', '')}';
      });
    }
  }

  /// Whether the booking is already cancelled
  bool get _isAlreadyCancelled {
    if (_booking == null) return false;
    return _booking!.paymentStatus.toLowerCase() == 'cancelled' ||
        _booking!.bookingStatus.toLowerCase() == 'cancelled';
  }

  /// Whether today is on or after the scheduled check-in date
  bool get _isPastOrTodayCheckIn {
    if (_booking == null) return false;
    DateTime? checkIn = DateTime.tryParse(_booking!.checkInDate);
    if (checkIn == null) return false;

    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final checkInDay = DateTime(checkIn.year, checkIn.month, checkIn.day);

    // Rule: cannot cancel on or after check-in date
    return today.isAtSameMomentAs(checkInDay) || today.isAfter(checkInDay);
  }

  /// Pre-cancellation verification:
  /// Cancellation is ONLY permitted if today < checkInDate and status is not already cancelled.
  bool get _canCancel =>
      _booking != null &&
      !_isAlreadyCancelled &&
      !_isPastOrTodayCheckIn;

  /// Executes booking cancellation via repository
  Future<void> _handleCancelBooking() async {
    if (!_canCancel || !_acknowledgedNoRefund || _isCancelling) return;

    final booking = _booking!;

    // Show native confirmation modal
    final shouldCancel = await showDialog<bool>(
      context: context,
      builder: (dialogCtx) => AlertDialog(
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(18),
        ),
        title: const Row(
          children: [
            Icon(
              LucideIcons.alertTriangle,
              color: AppColors.destructive,
              size: 24,
            ),
            SizedBox(width: 10),
            Expanded(
              child: Text(
                'Confirm Cancellation',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
            ),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Are you sure you want to cancel booking #${booking.id} for ${TagaTowersDesignTokens.getRoomDisplayName(booking.roomName)}?',
              style: const TextStyle(
                fontSize: 14,
                color: AppColors.textPrimary,
                height: 1.4,
              ),
            ),
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: const Color(0xFFFEF2F2),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(
                  color: const Color(0xFFFCA5A5),
                  width: 1,
                ),
              ),
              child: Text(
                'Reminder: The advance payment of ₹${booking.advanceAmount} is strictly non-refundable and will not be refunded.',
                style: const TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF991B1B),
                ),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(dialogCtx).pop(false),
            child: const Text(
              'Keep Booking',
              style: TextStyle(
                color: AppColors.textSecondary,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          ElevatedButton(
            onPressed: () => Navigator.of(dialogCtx).pop(true),
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.destructive,
              foregroundColor: Colors.white,
              elevation: 0,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            child: const Text(
              'Yes, Cancel Booking',
              style: TextStyle(fontWeight: FontWeight.bold),
            ),
          ),
        ],
      ),
    );

    if (shouldCancel != true || !mounted) return;

    setState(() {
      _isCancelling = true;
      _errorMessage = null;
    });

    try {
      final repo = ref.read(tagaTowersRepositoryProvider);
      await repo.cancelBooking(booking.id);

      // Invalidate active bookings provider to refresh caches
      ref.invalidate(userBookingsProvider);
      ref.invalidate(pastBookingsProvider);

      if (!mounted) return;

      HapticFeedback.heavyImpact();

      setState(() {
        _isCancelling = false;
        _isCancelledSuccess = true;
      });
    } catch (e) {
      if (!mounted) return;

      HapticFeedback.lightImpact();

      setState(() {
        _isCancelling = false;
        _errorMessage = e
            .toString()
            .replaceAll('Exception: ', '')
            .replaceAll('DioException', 'Network Error');
      });
    }
  }

  void _navigateSafelyBack() {
    HapticFeedback.lightImpact();
    if (context.canPop()) {
      context.pop();
    } else {
      context.go('/taga-towers');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF9FAF7),
      appBar: PreferredSize(
        preferredSize: const Size.fromHeight(68),
        child: _buildHeader(context),
      ),
      body: SafeArea(
        top: false,
        child: _buildBody(context),
      ),
    );
  }

  /// Dark emerald header bar matching design tokens
  Widget _buildHeader(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            Color(0xFF061A0F),
            Color(0xFF0D2C1B),
            Color(0xFF1B4332),
          ],
        ),
        borderRadius: BorderRadius.vertical(bottom: Radius.circular(20)),
        boxShadow: [
          BoxShadow(
            color: Color(0x33061A0F),
            blurRadius: 10,
            offset: Offset(0, 3),
          ),
        ],
      ),
      child: SafeArea(
        bottom: false,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
          child: Row(
            children: [
              BouncyButton(
                onTap: _navigateSafelyBack,
                child: Container(
                  width: 38,
                  height: 38,
                  decoration: BoxDecoration(
                    color: Colors.white.withValues(alpha: 0.12),
                    shape: BoxShape.circle,
                    border: Border.all(
                      color: Colors.white.withValues(alpha: 0.22),
                      width: 1,
                    ),
                  ),
                  child: const Icon(
                    LucideIcons.arrowLeft,
                    color: Colors.white,
                    size: 20,
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Cancel Reservation',
                      style: TextStyle(
                        color: Colors.white,
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        letterSpacing: -0.2,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      'TAGA Towers Egmore',
                      style: TextStyle(
                        color: TagaTowersDesignTokens.mintHighlight
                            .withValues(alpha: 0.9),
                        fontSize: 12,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                decoration: BoxDecoration(
                  color: const Color(0xFFFEF2F2).withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(999),
                  border: Border.all(
                    color: const Color(0xFFF87171).withValues(alpha: 0.4),
                    width: 1,
                  ),
                ),
                child: const Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      LucideIcons.alertTriangle,
                      size: 13,
                      color: Color(0xFFFCA5A5),
                    ),
                    SizedBox(width: 4),
                    Text(
                      'Strict Policy',
                      style: TextStyle(
                        color: Color(0xFFFCA5A5),
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildBody(BuildContext context) {
    if (_isLoadingBooking) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              CircularProgressIndicator(
                valueColor: AlwaysStoppedAnimation<Color>(
                  TagaTowersDesignTokens.primaryGreen,
                ),
              ),
              SizedBox(height: 16),
              Text(
                'Loading reservation details...',
                style: TextStyle(
                  color: AppColors.textSecondary,
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ],
          ),
        ),
      );
    }

    if (_fetchError != null && _booking == null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Container(
            padding: const EdgeInsets.all(20),
            decoration: TagaTowersDesignTokens.cardDecoration(
              color: Colors.white,
              borderRadius: 16,
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(
                  LucideIcons.calendarX,
                  size: 48,
                  color: AppColors.destructive,
                ),
                const SizedBox(height: 14),
                const Text(
                  'Reservation Not Found',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  _fetchError!,
                  textAlign: TextAlign.center,
                  style: const TextStyle(
                    fontSize: 13,
                    color: AppColors.textSecondary,
                    height: 1.4,
                  ),
                ),
                const SizedBox(height: 20),
                Row(
                  children: [
                    if (widget.bookingId != null) ...[
                      Expanded(
                        child: OutlinedButton.icon(
                          onPressed: _loadBookingById,
                          icon: const Icon(LucideIcons.refreshCw, size: 16),
                          label: const Text('Retry'),
                          style: OutlinedButton.styleFrom(
                            foregroundColor: TagaTowersDesignTokens.primaryGreen,
                            side: const BorderSide(
                              color: TagaTowersDesignTokens.mintHighlight,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                    ],
                    Expanded(
                      child: ElevatedButton(
                        onPressed: _navigateSafelyBack,
                        style: ElevatedButton.styleFrom(
                          backgroundColor: TagaTowersDesignTokens.primaryGreen,
                          foregroundColor: Colors.white,
                        ),
                        child: const Text('Back to Bookings'),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      );
    }

    if (_booking == null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(
                LucideIcons.info,
                size: 42,
                color: AppColors.textSecondary,
              ),
              const SizedBox(height: 12),
              const Text(
                'No booking selected for cancellation',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: _navigateSafelyBack,
                style: ElevatedButton.styleFrom(
                  backgroundColor: TagaTowersDesignTokens.primaryGreen,
                  foregroundColor: Colors.white,
                ),
                child: const Text('Return to Bookings'),
              ),
            ],
          ),
        ),
      );
    }

    // Success State View
    if (_isCancelledSuccess) {
      return _buildSuccessState(context, _booking!);
    }

    final booking = _booking!;

    return SingleChildScrollView(
      physics: const BouncingScrollPhysics(),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // Pre-cancellation Verification Warnings
          if (_isAlreadyCancelled) ...[
            _buildAlreadyCancelledCard(),
            const SizedBox(height: 16),
          ] else if (_isPastOrTodayCheckIn) ...[
            _buildCheckInDateWarningCard(booking),
            const SizedBox(height: 16),
          ],

          // Critical No-Refund Alert Card
          _buildNoRefundAlertCard(booking),
          const SizedBox(height: 16),

          // Booking Summary Review
          _buildBookingSummaryCard(booking),
          const SizedBox(height: 16),

          // Terms Acknowledgement Checkbox
          if (_canCancel) ...[
            _buildTermsAcknowledgementCheckbox(),
            const SizedBox(height: 16),
          ],

          // Error Card on Failure (with prominent retry button)
          if (_errorMessage != null) ...[
            _buildCancellationErrorCard(),
            const SizedBox(height: 16),
          ],

          // Action Buttons
          _buildActionButtons(booking),
          const SizedBox(height: 24),
        ],
      ),
    );
  }

  /// Card displayed when the booking has already been cancelled
  Widget _buildAlreadyCancelledCard() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFFF3F4F6),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: const Color(0xFFD1D5DB),
          width: 1.2,
        ),
      ),
      child: const Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(
            LucideIcons.ban,
            size: 22,
            color: Color(0xFF6B7280),
          ),
          SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Booking Already Cancelled',
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.bold,
                    color: Color(0xFF1F2937),
                  ),
                ),
                SizedBox(height: 4),
                Text(
                  'This reservation has already been cancelled. Room slots were released back to availability and no further action is permitted.',
                  style: TextStyle(
                    fontSize: 12.5,
                    color: Color(0xFF4B5563),
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

  /// Card displayed when check-in date is today or in the past
  Widget _buildCheckInDateWarningCard(BookingResponse booking) {
    DateTime? checkIn = DateTime.tryParse(booking.checkInDate);
    final formattedCheckIn =
        checkIn != null ? _dateFormat.format(checkIn) : booking.checkInDate;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFFFEF2F2),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: const Color(0xFFFCA5A5),
          width: 1.4,
        ),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(
            LucideIcons.alertCircle,
            size: 24,
            color: AppColors.destructive,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Cannot cancel booking on or after check-in date.',
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.bold,
                    color: Color(0xFF991B1B),
                  ),
                ),
                const SizedBox(height: 6),
                Text(
                  'Scheduled check-in: $formattedCheckIn at 12:00 AM. According to TAGA Towers regulations, cancellation requests must be made prior to the arrival date.',
                  style: const TextStyle(
                    fontSize: 12.5,
                    color: Color(0xFF7F1D1D),
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

  /// CRITICAL Alert Card: Prominent, unmistakable "No Refund Will Be Provided" notice
  Widget _buildNoRefundAlertCard(BookingResponse booking) {
    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: const Color(0xFFFFFBEB),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: const Color(0xFFF59E0B),
          width: 1.4,
        ),
        boxShadow: const [
          BoxShadow(
            color: Color(0x14F59E0B),
            blurRadius: 14,
            offset: Offset(0, 4),
          ),
        ],
      ),
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: const Color(0xFFFEF3C7),
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(
                    color: const Color(0xFFFCD34D),
                    width: 1,
                  ),
                ),
                child: const Icon(
                  LucideIcons.alertTriangle,
                  size: 20,
                  color: Color(0xFFB45309),
                ),
              ),
              const SizedBox(width: 12),
              const Expanded(
                child: Text(
                  'No Refund Will Be Provided',
                  style: TextStyle(
                    fontSize: 16.5,
                    fontWeight: FontWeight.w800,
                    color: Color(0xFF92400E),
                    letterSpacing: -0.2,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            'Please note that according to TAGA Towers booking policies, the advance payment of ₹${booking.advanceAmount} is strictly non-refundable upon cancellation. No refund will be credited.',
            style: const TextStyle(
              fontSize: 13.5,
              fontWeight: FontWeight.w600,
              color: Color(0xFF78350F),
              height: 1.45,
            ),
          ),
          const SizedBox(height: 8),
          const Text(
            'The reserved room and bed slots will be made available immediately to other association members.',
            style: TextStyle(
              fontSize: 12,
              color: Color(0xFF92400E),
              height: 1.35,
            ),
          ),
        ],
      ),
    );
  }

  /// Booking Summary Review Card displaying reference, room, dates, beds, advance
  Widget _buildBookingSummaryCard(BookingResponse booking) {
    DateTime? checkIn = DateTime.tryParse(booking.checkInDate);
    DateTime? checkOut = DateTime.tryParse(booking.checkOutDate);

    final checkInStr =
        checkIn != null ? _dateFormat.format(checkIn) : booking.checkInDate;
    final checkOutStr =
        checkOut != null ? _dateFormat.format(checkOut) : booking.checkOutDate;

    int stayNights = 1;
    if (checkIn != null && checkOut != null) {
      final diff = checkOut.difference(checkIn).inDays;
      if (diff > 0) stayNights = diff;
    }

    final roomDisplayName =
        TagaTowersDesignTokens.getRoomDisplayName(booking.roomName);

    return Container(
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 16,
      ),
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: TagaTowersDesignTokens.softMint,
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: const Icon(
                      LucideIcons.clipboardCheck,
                      size: 18,
                      color: TagaTowersDesignTokens.deepEmerald,
                    ),
                  ),
                  const SizedBox(width: 10),
                  const Text(
                    'Booking Summary',
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
                  ),
                ],
              ),
              _buildStatusPill(booking),
            ],
          ),
          const SizedBox(height: 14),
          const Divider(color: Color(0xFFE2EBE5), height: 1),
          const SizedBox(height: 14),

          // Booking Reference ID with Copy Button
          Row(
            crossAxisAlignment: CrossAxisAlignment.center,
            children: [
              const Icon(
                LucideIcons.receiptText,
                size: 16,
                color: AppColors.textSecondary,
              ),
              const SizedBox(width: 10),
              const Text(
                'Reference ID',
                style: TextStyle(
                  fontSize: 13,
                  color: AppColors.textSecondary,
                  fontWeight: FontWeight.w500,
                ),
              ),
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                decoration: BoxDecoration(
                  color: const Color(0xFFF1F8E9),
                  borderRadius: BorderRadius.circular(6),
                  border: Border.all(
                    color: TagaTowersDesignTokens.mintHighlight,
                    width: 0.8,
                  ),
                ),
                child: Text(
                  booking.id,
                  style: const TextStyle(
                    fontFamily: 'monospace',
                    fontSize: 12.5,
                    fontWeight: FontWeight.w700,
                    color: TagaTowersDesignTokens.deepEmerald,
                  ),
                ),
              ),
              const SizedBox(width: 6),
              BouncyButton(
                onTap: () {
                  Clipboard.setData(ClipboardData(text: booking.id));
                  HapticFeedback.selectionClick();
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text('Booking ID copied to clipboard'),
                      duration: Duration(seconds: 2),
                    ),
                  );
                },
                child: const Padding(
                  padding: EdgeInsets.all(4),
                  child: Icon(
                    LucideIcons.copy,
                    size: 16,
                    color: TagaTowersDesignTokens.primaryGreen,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),

          // Room Name
          _buildSummaryRow(
            icon: LucideIcons.bedDouble,
            label: 'Room Name',
            value: roomDisplayName,
          ),
          const SizedBox(height: 12),

          // Stay Dates & Timings
          _buildSummaryRow(
            icon: LucideIcons.calendar,
            label: 'Stay Dates',
            value: '$checkInStr → $checkOutStr',
            subValue: '$stayNights Night(s) • 12:00 AM Cycle',
          ),
          const SizedBox(height: 12),

          // Number of Beds
          _buildSummaryRow(
            icon: LucideIcons.layers,
            label: 'Number of Beds',
            value: '${booking.bedCount} Bed${booking.bedCount > 1 ? 's' : ''}',
          ),
          const SizedBox(height: 12),

          // Booker Name & Target
          _buildSummaryRow(
            icon: booking.bookingFor.toLowerCase() == 'guest'
                ? LucideIcons.users
                : LucideIcons.user,
            label: 'Booked For',
            value: '${booking.bookerName} (${booking.bookingFor.toUpperCase()})',
          ),
          const SizedBox(height: 14),

          // Advance Paid Highlight Box
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
            decoration: BoxDecoration(
              color: const Color(0xFFF1F8E9),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: TagaTowersDesignTokens.mintHighlight,
                width: 1.2,
              ),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Advance Paid',
                      style: TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.bold,
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                    SizedBox(height: 2),
                    Text(
                      'Strictly non-refundable fee',
                      style: TextStyle(
                        fontSize: 11,
                        color: AppColors.textSecondary,
                        fontStyle: FontStyle.italic,
                      ),
                    ),
                  ],
                ),
                Text(
                  '₹${booking.advanceAmount}',
                  style: const TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.w900,
                    color: TagaTowersDesignTokens.primaryGreen,
                    letterSpacing: -0.3,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSummaryRow({
    required IconData icon,
    required String label,
    required String value,
    String? subValue,
  }) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: 16, color: AppColors.textSecondary),
        const SizedBox(width: 10),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: const TextStyle(
                  fontSize: 13,
                  color: AppColors.textSecondary,
                  fontWeight: FontWeight.w500,
                ),
              ),
              if (subValue != null) ...[
                const SizedBox(height: 2),
                Text(
                  subValue,
                  style: const TextStyle(
                    fontSize: 11,
                    color: Color(0xFF718E81),
                    fontStyle: FontStyle.italic,
                  ),
                ),
              ],
            ],
          ),
        ),
        const SizedBox(width: 8),
        Text(
          value,
          textAlign: TextAlign.right,
          style: const TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w700,
            color: AppColors.textPrimary,
          ),
        ),
      ],
    );
  }

  Widget _buildStatusPill(BookingResponse booking) {
    Color bg;
    Color border;
    Color text;
    String label;

    if (_isAlreadyCancelled) {
      bg = const Color(0xFFFEE2E2);
      border = const Color(0xFFFCA5A5);
      text = const Color(0xFF991B1B);
      label = 'CANCELLED';
    } else if (booking.bookingStatus.toLowerCase() == 'active') {
      bg = const Color(0xFFDBEAFE);
      border = const Color(0xFF93C5FD);
      text = const Color(0xFF1E40AF);
      label = 'ACTIVE';
    } else {
      bg = TagaTowersDesignTokens.softMint;
      border = TagaTowersDesignTokens.mintHighlight;
      text = TagaTowersDesignTokens.deepEmerald;
      label = booking.bookingStatus.toUpperCase();
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: border, width: 1),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 10.5,
          fontWeight: FontWeight.w800,
          letterSpacing: 0.5,
          color: text,
        ),
      ),
    );
  }

  /// Checkbox to acknowledge that cancellation is irreversible and no refund is provided
  Widget _buildTermsAcknowledgementCheckbox() {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: _acknowledgedNoRefund
              ? AppColors.destructive.withValues(alpha: 0.5)
              : TagaTowersDesignTokens.border,
          width: 1.2,
        ),
        boxShadow: TagaTowersDesignTokens.softShadow,
      ),
      child: Material(
        color: Colors.transparent,
        borderRadius: BorderRadius.circular(14),
        child: InkWell(
          borderRadius: BorderRadius.circular(14),
          onTap: _isCancelling
              ? null
              : () {
                  HapticFeedback.selectionClick();
                  setState(() {
                    _acknowledgedNoRefund = !_acknowledgedNoRefund;
                  });
                },
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                SizedBox(
                  width: 24,
                  height: 24,
                  child: Checkbox(
                    value: _acknowledgedNoRefund,
                    activeColor: AppColors.destructive,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(5),
                    ),
                    onChanged: _isCancelling
                        ? null
                        : (val) {
                            HapticFeedback.selectionClick();
                            setState(() {
                              _acknowledgedNoRefund = val ?? false;
                            });
                          },
                  ),
                ),
                const SizedBox(width: 12),
                const Expanded(
                  child: Text(
                    'I understand that cancelling this reservation is irreversible and that no refund will be provided.',
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: AppColors.textPrimary,
                      height: 1.4,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  /// Cancellation failure error card with a prominent retry button
  Widget _buildCancellationErrorCard() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFFFEF2F2),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: const Color(0xFFFCA5A5),
          width: 1.3,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(
                LucideIcons.circleAlert,
                size: 20,
                color: AppColors.destructive,
              ),
              const SizedBox(width: 8),
              const Text(
                'Cancellation Failed',
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF991B1B),
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            _errorMessage ?? 'Unable to complete cancellation request.',
            style: const TextStyle(
              fontSize: 12.5,
              color: Color(0xFF7F1D1D),
              height: 1.35,
            ),
          ),
          const SizedBox(height: 14),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              onPressed: _isCancelling ? null : _handleCancelBooking,
              icon: const Icon(LucideIcons.refreshCw, size: 16),
              label: const Text(
                'Retry Cancellation',
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.destructive,
                foregroundColor: Colors.white,
                elevation: 0,
                padding: const EdgeInsets.symmetric(vertical: 12),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(10),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// Action buttons: "Yes, Cancel Booking" and "No, Keep Booking"
  Widget _buildActionButtons(BookingResponse booking) {
    final isCancelButtonActive =
        _canCancel && _acknowledgedNoRefund && !_isCancelling;

    return Column(
      children: [
        // Destructive cancellation button
        BouncyButton(
          onTap: isCancelButtonActive ? _handleCancelBooking : null,
          child: Container(
            width: double.infinity,
            height: 52,
            decoration: BoxDecoration(
              color: isCancelButtonActive
                  ? AppColors.destructive
                  : const Color(0xFFCBD5E1),
              borderRadius: BorderRadius.circular(14),
              boxShadow: isCancelButtonActive
                  ? [
                      BoxShadow(
                        color: AppColors.destructive.withValues(alpha: 0.35),
                        blurRadius: 12,
                        offset: const Offset(0, 4),
                      ),
                    ]
                  : null,
            ),
            child: Center(
              child: _isCancelling
                  ? const Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(
                            strokeWidth: 2.2,
                            valueColor:
                                AlwaysStoppedAnimation<Color>(Colors.white),
                          ),
                        ),
                        SizedBox(width: 10),
                        Text(
                          'Cancelling Reservation...',
                          style: TextStyle(
                            color: Colors.white,
                            fontSize: 15,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    )
                  : const Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          LucideIcons.trash2,
                          size: 18,
                          color: Colors.white,
                        ),
                        SizedBox(width: 8),
                        Text(
                          'Yes, Cancel Booking',
                          style: TextStyle(
                            color: Colors.white,
                            fontSize: 15.5,
                            fontWeight: FontWeight.bold,
                            letterSpacing: 0.2,
                          ),
                        ),
                      ],
                    ),
            ),
          ),
        ),
        const SizedBox(height: 12),

        // Safe "No, Keep Booking" return button
        BouncyButton(
          onTap: _navigateSafelyBack,
          child: Container(
            width: double.infinity,
            height: 50,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(
                color: TagaTowersDesignTokens.border,
                width: 1.2,
              ),
              boxShadow: TagaTowersDesignTokens.softShadow,
            ),
            child: const Center(
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    LucideIcons.shieldCheck,
                    size: 18,
                    color: TagaTowersDesignTokens.primaryGreen,
                  ),
                  SizedBox(width: 8),
                  Text(
                    'No, Keep Booking',
                    style: TextStyle(
                      color: TagaTowersDesignTokens.deepEmerald,
                      fontSize: 15,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }

  /// Full screen success view displayed upon successful cancellation
  Widget _buildSuccessState(BuildContext context, BookingResponse booking) {
    return Center(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Container(
          width: double.infinity,
          padding: const EdgeInsets.all(24),
          decoration: TagaTowersDesignTokens.cardDecoration(
            color: Colors.white,
            borderRadius: 20,
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 72,
                height: 72,
                decoration: BoxDecoration(
                  color: TagaTowersDesignTokens.softMint,
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: TagaTowersDesignTokens.mintHighlight,
                    width: 2,
                  ),
                ),
                child: const Center(
                  child: Icon(
                    LucideIcons.check,
                    size: 38,
                    color: TagaTowersDesignTokens.deepEmerald,
                  ),
                ),
              ),
              const SizedBox(height: 18),
              const Text(
                'Booking Cancelled',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 22,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                  letterSpacing: -0.3,
                ),
              ),
              const SizedBox(height: 8),
              const Text(
                'Your reservation has been cancelled.',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: TagaTowersDesignTokens.deepEmerald,
                ),
              ),
              const SizedBox(height: 10),
              Text(
                'The room inventory for #${booking.id} (${TagaTowersDesignTokens.getRoomDisplayName(booking.roomName)}) has been released. The advance fee of ₹${booking.advanceAmount} is retained as non-refundable advance.',
                textAlign: TextAlign.center,
                style: const TextStyle(
                  fontSize: 13,
                  color: AppColors.textSecondary,
                  height: 1.45,
                ),
              ),
              const SizedBox(height: 24),
              BouncyButton(
                onTap: () {
                  HapticFeedback.lightImpact();
                  if (context.canPop()) {
                    context.pop();
                  } else {
                    context.go('/taga-towers');
                  }
                },
                child: Container(
                  width: double.infinity,
                  height: 50,
                  decoration: BoxDecoration(
                    color: TagaTowersDesignTokens.primaryGreen,
                    borderRadius: BorderRadius.circular(14),
                    boxShadow: [
                      BoxShadow(
                        color: TagaTowersDesignTokens.primaryGreen
                            .withValues(alpha: 0.35),
                        blurRadius: 10,
                        offset: const Offset(0, 4),
                      ),
                    ],
                  ),
                  child: const Center(
                    child: Text(
                      'Return to My Bookings',
                      style: TextStyle(
                        color: Colors.white,
                        fontSize: 15.5,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
