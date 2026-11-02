import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../domain/booking_model.dart';
import '../providers/booking_flow_provider.dart';
import 'theme/tagatowers_design_tokens.dart';

/// Caretaker contact model
class _CaretakerInfo {
  final String name;
  final String phone;

  const _CaretakerInfo({required this.name, required this.phone});
}

/// Official Caretaker contacts for TAGA Towers, Egmore, Chennai
const List<_CaretakerInfo> _kCaretakers = [
  _CaretakerInfo(name: 'Mr. Muthu', phone: '96007 63744'),
  _CaretakerInfo(name: 'Mr. Mariyappan', phone: '96770 10300'),
];

/// Post-booking confirmation screen for TAGA Towers accommodations.
///
/// Strict Rule:
/// - Display this screen ONLY after the backend confirms the booking
///   (or when passed a confirmed [BookingResponse]).
///
/// Features:
/// - Celebratory success header ("🎉 Booking Confirmed!").
/// - Official Booking Reference ID with one-tap copy button.
/// - Room Name & Type badge.
/// - Check-in Date & Check-out Date (12:00 AM Midnight times).
/// - Stay duration (N days / 24-hr cycles).
/// - Number of beds booked & Booking Type (Self / Guest).
/// - Guest list summary (if Guest booking).
/// - Payment Details card:
///   - Advance Amount Paid: ₹XXX
///   - Payment Gateway Status: Confirmed
///   - Razorpay Payment ID / Order ID (if available)
///   - Remaining Balance payable at check-in / check-out.
/// - Property check-in instructions & Caretaker contacts
///   (`Mr. Muthu: 96007 63744`, `Mr. Mariyappan: 96770 10300`, `Egmore, Chennai`).
/// - Non-Refundable Cancellation notice.
/// - Actions:
///   - "View in My Bookings" button -> navigates to `/taga-towers/my-bookings`.
///   - "Back to Home" button -> resets booking flow and navigates to `/taga-towers`.
class BookingConfirmationScreen extends ConsumerStatefulWidget {
  final BookingResponse booking;
  final String? razorpayPaymentId;
  final String? razorpayOrderId;

  const BookingConfirmationScreen({
    super.key,
    required this.booking,
    this.razorpayPaymentId,
    this.razorpayOrderId,
  });

  @override
  ConsumerState<BookingConfirmationScreen> createState() =>
      _BookingConfirmationScreenState();
}

class _BookingConfirmationScreenState
    extends ConsumerState<BookingConfirmationScreen> {
  final Set<String> _copiedKeys = {};

  bool get _isConfirmed =>
      widget.booking.paymentStatus.trim().toLowerCase() == 'confirmed';

  String get _effectivePaymentId {
    final fromBooking = widget.booking.razorpayPaymentId?.trim();
    if (fromBooking != null && fromBooking.isNotEmpty) return fromBooking;
    final fromParam = widget.razorpayPaymentId?.trim();
    if (fromParam != null && fromParam.isNotEmpty) return fromParam;
    return '';
  }

  String get _effectiveOrderId {
    final fromBooking = widget.booking.razorpayOrderId?.trim();
    if (fromBooking != null && fromBooking.isNotEmpty) return fromBooking;
    final fromParam = widget.razorpayOrderId?.trim();
    if (fromParam != null && fromParam.isNotEmpty) return fromParam;
    return '';
  }

  DateTime? _parseDate(String dateStr) {
    if (dateStr.isEmpty) return null;
    return DateTime.tryParse(dateStr);
  }

  String _formatDisplayDate(String dateStr) {
    final parsed = _parseDate(dateStr);
    if (parsed == null) return dateStr;
    return DateFormat('EEE, dd MMM yyyy').format(parsed);
  }

  int _calculateStayDays() {
    final checkIn = _parseDate(widget.booking.checkInDate);
    final checkOut = _parseDate(widget.booking.checkOutDate);
    if (checkIn != null && checkOut != null) {
      final days = checkOut.difference(checkIn).inDays;
      return days > 0 ? days : 1;
    }
    return 1;
  }

  Future<void> _copyText(String key, String text, String label) async {
    setState(() {
      _copiedKeys.add(key);
    });

    await Clipboard.setData(ClipboardData(text: text));
    HapticFeedback.lightImpact();

    if (mounted) {
      ScaffoldMessenger.of(context).hideCurrentSnackBar();
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Row(
            children: [
              const Icon(LucideIcons.check, color: Colors.white, size: 16),
              const SizedBox(width: 8),
              Text(
                '$label copied to clipboard',
                style: const TextStyle(fontWeight: FontWeight.w600),
              ),
            ],
          ),
          backgroundColor: TagaTowersDesignTokens.deepEmerald,
          behavior: SnackBarBehavior.floating,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
          duration: const Duration(seconds: 2),
        ),
      );
    }

    Future.delayed(const Duration(seconds: 2), () {
      if (mounted) {
        setState(() {
          _copiedKeys.remove(key);
        });
      }
    });
  }

  Future<void> _makePhoneCall(String phoneNumber) async {
    final cleanNumber = phoneNumber.replaceAll(RegExp(r'\s+'), '');
    final uri = Uri.parse('tel:$cleanNumber');
    try {
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri);
      } else {
        await _copyText('caretaker_$cleanNumber', cleanNumber, 'Phone number');
      }
    } catch (_) {
      await _copyText('caretaker_$cleanNumber', cleanNumber, 'Phone number');
    }
  }

  void _onViewInMyBookings() {
    context.go('/taga-towers/my-bookings');
  }

  void _onBackToHome() {
    ref.read(bookingFlowNotifierProvider.notifier).reset();
    context.go('/taga-towers');
  }

  @override
  Widget build(BuildContext context) {
    // Intercept back navigation to return home rather than revisiting payment screens
    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, result) {
        if (didPop) return;
        _onBackToHome();
      },
      child: Scaffold(
        backgroundColor: TagaTowersDesignTokens.surfaceLight,
        appBar: AppBar(
          backgroundColor: Colors.white,
          elevation: 0,
          scrolledUnderElevation: 0,
          centerTitle: true,
          automaticallyImplyLeading: false,
          leading: IconButton(
            icon: const Icon(LucideIcons.x, color: AppColors.textPrimary),
            tooltip: 'Close and Return Home',
            onPressed: _onBackToHome,
          ),
          title: const Text(
            'Booking Confirmation',
            style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.w700,
              color: AppColors.textPrimary,
              letterSpacing: -0.3,
            ),
          ),
          actions: [
            IconButton(
              icon: const Icon(LucideIcons.home, color: AppColors.textPrimary),
              tooltip: 'Home',
              onPressed: _onBackToHome,
            ),
          ],
          bottom: PreferredSize(
            preferredSize: const Size.fromHeight(1.0),
            child: Container(
              color: TagaTowersDesignTokens.subtleBorder,
              height: 1.0,
            ),
          ),
        ),
        body: SafeArea(
          child: !_isConfirmed
              ? _buildUnconfirmedState()
              : SingleChildScrollView(
                  physics: const BouncingScrollPhysics(),
                  padding: const EdgeInsets.symmetric(
                    horizontal: 20.0,
                    vertical: 20.0,
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      // 1. Celebratory Success Header
                      _buildCelebratoryHeader(),
                      const SizedBox(height: 20),

                      // 2. Official Booking Reference ID Card
                      _buildBookingReferenceCard(),
                      const SizedBox(height: 16),

                      // 3. Room & Stay Details Card (with 24-hr midnight cycle)
                      _buildStayDetailsCard(),
                      const SizedBox(height: 16),

                      // 4. Guest List Summary Card (if guest booking)
                      if (widget.booking.bookingFor.trim().toLowerCase() ==
                              'guest' &&
                          widget.booking.guestDetails.isNotEmpty) ...[
                        _buildGuestListCard(),
                        const SizedBox(height: 16),
                      ],

                      // 5. Payment Details Card
                      _buildPaymentDetailsCard(),
                      const SizedBox(height: 16),

                      // 6. Property Check-in Instructions & Caretaker Contacts
                      _buildCheckInAndCaretakersCard(),
                      const SizedBox(height: 16),

                      // 7. Non-Refundable Cancellation Notice
                      _buildCancellationNoticeCard(),
                      const SizedBox(height: 28),

                      // 8. Navigation Action Buttons
                      _buildActionButtons(),
                      const SizedBox(height: 20),
                    ],
                  ),
                ),
        ),
      ),
    );
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // 1. CELEBRATORY SUCCESS HEADER
  // ═══════════════════════════════════════════════════════════════════════════

  Widget _buildCelebratoryHeader() {
    return Column(
      children: [
        // Pulsing glowing success circle
        Container(
          width: 84,
          height: 84,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: TagaTowersDesignTokens.softMint,
            border: Border.all(
              color: TagaTowersDesignTokens.mintHighlight,
              width: 2.5,
            ),
            boxShadow: [
              BoxShadow(
                color:
                    TagaTowersDesignTokens.primaryGreen.withValues(alpha: 0.22),
                blurRadius: 24,
                spreadRadius: 2,
                offset: const Offset(0, 6),
              ),
            ],
          ),
          child: const Center(
            child: Icon(
              LucideIcons.check,
              size: 42,
              color: TagaTowersDesignTokens.deepEmerald,
            ),
          ),
        )
            .animate()
            .scale(
              duration: 500.ms,
              curve: Curves.elasticOut,
            )
            .fadeIn(duration: 350.ms),
        const SizedBox(height: 18),

        // Celebratory headline
        const Text(
          '🎉 Booking Confirmed!',
          textAlign: TextAlign.center,
          style: TextStyle(
            fontSize: 25,
            fontWeight: FontWeight.w800,
            color: TagaTowersDesignTokens.deepEmerald,
            letterSpacing: -0.4,
          ),
        )
            .animate()
            .fadeIn(duration: 400.ms, delay: 150.ms)
            .slideY(begin: 0.15, end: 0, curve: Curves.easeOutCubic),
        const SizedBox(height: 8),

        // Warm subtitle
        Text(
          widget.booking.bookerName.trim().isNotEmpty
              ? 'Thank you, ${widget.booking.bookerName.trim()}! Your accommodation at TAGA Towers Chennai is confirmed and secured.'
              : 'Your accommodation at TAGA Towers Chennai has been confirmed and secured.',
          textAlign: TextAlign.center,
          style: const TextStyle(
            fontSize: 14.5,
            color: AppColors.textSecondary,
            height: 1.45,
            fontWeight: FontWeight.w400,
          ),
        )
            .animate()
            .fadeIn(duration: 400.ms, delay: 250.ms)
            .slideY(begin: 0.1, end: 0, curve: Curves.easeOutCubic),
      ],
    );
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // 2. OFFICIAL BOOKING REFERENCE ID CARD
  // ═══════════════════════════════════════════════════════════════════════════

  Widget _buildBookingReferenceCard() {
    final isCopied = _copiedKeys.contains('booking_id');
    final bookingId = widget.booking.id;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 16,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: const [
                  Icon(
                    LucideIcons.ticket,
                    size: 16,
                    color: TagaTowersDesignTokens.primaryGreen,
                  ),
                  SizedBox(width: 8),
                  Text(
                    'OFFICIAL BOOKING REFERENCE ID',
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w800,
                      letterSpacing: 0.7,
                      color: AppColors.textSecondary,
                    ),
                  ),
                ],
              ),
              BouncyButton(
                onTap: () => _copyText('booking_id', bookingId, 'Booking Reference ID'),
                child: Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                  decoration: BoxDecoration(
                    color: isCopied
                        ? TagaTowersDesignTokens.primaryGreen
                        : TagaTowersDesignTokens.softMint,
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(
                      color: isCopied
                          ? TagaTowersDesignTokens.primaryGreen
                          : TagaTowersDesignTokens.mintHighlight,
                      width: 1.0,
                    ),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        isCopied ? LucideIcons.check : LucideIcons.copy,
                        size: 13,
                        color: isCopied
                            ? Colors.white
                            : TagaTowersDesignTokens.deepEmerald,
                      ),
                      const SizedBox(width: 5),
                      Text(
                        isCopied ? 'Copied' : 'Copy',
                        style: TextStyle(
                          fontSize: 11.5,
                          fontWeight: FontWeight.w700,
                          color: isCopied
                              ? Colors.white
                              : TagaTowersDesignTokens.deepEmerald,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
            decoration: BoxDecoration(
              color: TagaTowersDesignTokens.mutedBackground,
              borderRadius: BorderRadius.circular(10),
              border: Border.all(
                color: TagaTowersDesignTokens.border,
                width: 1.0,
              ),
            ),
            child: Row(
              children: [
                Expanded(
                  child: SelectableText(
                    bookingId,
                    style: const TextStyle(
                      fontSize: 16.5,
                      fontWeight: FontWeight.w800,
                      fontFamily: 'monospace',
                      letterSpacing: 1.0,
                      color: TagaTowersDesignTokens.deepEmerald,
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 8),
          const Text(
            'Keep this Booking Reference ID ready upon arrival at TAGA Towers check-in desk.',
            style: TextStyle(
              fontSize: 11.5,
              color: AppColors.textSecondary,
              height: 1.35,
            ),
          ),
        ],
      ),
    );
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // 3. STAY DETAILS & 24-HR MIDNIGHT CYCLE CARD
  // ═══════════════════════════════════════════════════════════════════════════

  Widget _buildStayDetailsCard() {
    final stayDays = _calculateStayDays();
    final totalHours = stayDays * 24;
    final isGuest = widget.booking.bookingFor.trim().toLowerCase() == 'guest';
    final roomDisplayName =
        TagaTowersDesignTokens.getRoomDisplayName(widget.booking.roomId);
    final roomName = widget.booking.roomName.trim().isNotEmpty
        ? widget.booking.roomName.trim()
        : roomDisplayName;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 16,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Section Title
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: const [
                  Icon(
                    LucideIcons.calendarDays,
                    size: 16,
                    color: TagaTowersDesignTokens.primaryGreen,
                  ),
                  SizedBox(width: 8),
                  Text(
                    'RESERVATION DETAILS',
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w800,
                      letterSpacing: 0.7,
                      color: AppColors.textSecondary,
                    ),
                  ),
                ],
              ),
              // Booking Type Badge (Self / Guest)
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 9, vertical: 4),
                decoration: BoxDecoration(
                  color: isGuest
                      ? const Color(0xFFEFF6FF)
                      : TagaTowersDesignTokens.softMint,
                  borderRadius: BorderRadius.circular(999),
                  border: Border.all(
                    color: isGuest
                        ? const Color(0xFFBFDBFE)
                        : TagaTowersDesignTokens.mintHighlight,
                    width: 1.0,
                  ),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      isGuest ? LucideIcons.users : LucideIcons.user,
                      size: 12,
                      color: isGuest
                          ? const Color(0xFF1D4ED8)
                          : TagaTowersDesignTokens.deepEmerald,
                    ),
                    const SizedBox(width: 4),
                    Text(
                      isGuest ? 'Guest Booking' : 'Self Booking',
                      style: TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        color: isGuest
                            ? const Color(0xFF1D4ED8)
                            : TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),

          // Room Title & Badges
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Container(
                width: 44,
                height: 44,
                decoration: BoxDecoration(
                  color: TagaTowersDesignTokens.softMint,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Center(
                  child: Icon(
                    LucideIcons.bedDouble,
                    size: 22,
                    color: TagaTowersDesignTokens.primaryGreen,
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      roomName,
                      style: const TextStyle(
                        fontSize: 17,
                        fontWeight: FontWeight.w800,
                        color: AppColors.textPrimary,
                        letterSpacing: -0.2,
                      ),
                    ),
                    const SizedBox(height: 5),
                    Wrap(
                      spacing: 6,
                      runSpacing: 4,
                      children: [
                        _buildMiniBadge(
                          label: roomDisplayName,
                          color: TagaTowersDesignTokens.deepEmerald,
                          background: TagaTowersDesignTokens.softMint,
                        ),
                        _buildMiniBadge(
                          label:
                              '${widget.booking.bedCount} Bed${widget.booking.bedCount > 1 ? 's' : ''} Booked',
                          color: AppColors.textPrimary,
                          background: const Color(0xFFF1F5F9),
                        ),
                        if (widget.booking.gender != null &&
                            widget.booking.gender!.isNotEmpty)
                          _buildMiniBadge(
                            label: widget.booking.gender!.toUpperCase(),
                            color: AppColors.textPrimary,
                            background: const Color(0xFFF1F5F9),
                          ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          const Divider(height: 1, color: TagaTowersDesignTokens.subtleBorder),
          const SizedBox(height: 14),

          // Check-in & Check-out visual timeline
          Row(
            children: [
              // Check-in box
              Expanded(
                child: _buildDateBox(
                  label: 'CHECK-IN',
                  dateStr: _formatDisplayDate(widget.booking.checkInDate),
                  timeStr: '12:00 AM (Midnight)',
                  icon: LucideIcons.logIn,
                ),
              ),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 8.0),
                child: Icon(
                  LucideIcons.arrowRight,
                  size: 18,
                  color: TagaTowersDesignTokens.mintHighlight,
                ),
              ),
              // Check-out box
              Expanded(
                child: _buildDateBox(
                  label: 'CHECK-OUT',
                  dateStr: _formatDisplayDate(widget.booking.checkOutDate),
                  timeStr: '12:00 AM (Midnight)',
                  icon: LucideIcons.logOut,
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),

          // Stay Duration Banner (24-hr Midnight cycle)
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
            decoration: BoxDecoration(
              color: TagaTowersDesignTokens.mutedBackground,
              borderRadius: BorderRadius.circular(10),
              border: Border.all(
                color: TagaTowersDesignTokens.subtleBorder,
                width: 1.0,
              ),
            ),
            child: Row(
              children: [
                const Icon(
                  LucideIcons.clock,
                  size: 16,
                  color: TagaTowersDesignTokens.primaryGreen,
                ),
                const SizedBox(width: 9),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Stay Duration: $stayDays Day${stayDays > 1 ? 's' : ''} ($stayDays × 24-Hr Cycle${stayDays > 1 ? 's' : ''} • $totalHours Hours)',
                        style: const TextStyle(
                          fontSize: 12.5,
                          fontWeight: FontWeight.w700,
                          color: TagaTowersDesignTokens.deepEmerald,
                        ),
                      ),
                      const SizedBox(height: 2),
                      const Text(
                        'TAGA Towers operates on a strict 12:00 AM Midnight to 12:00 AM Midnight cycle.',
                        style: TextStyle(
                          fontSize: 11,
                          color: AppColors.textSecondary,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDateBox({
    required String label,
    required String dateStr,
    required String timeStr,
    required IconData icon,
  }) {
    return Container(
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: const Color(0xFFF8FAFC),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: const Color(0xFFE2E8F0), width: 1.0),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, size: 12, color: TagaTowersDesignTokens.primaryGreen),
              const SizedBox(width: 5),
              Text(
                label,
                style: const TextStyle(
                  fontSize: 10,
                  fontWeight: FontWeight.w800,
                  color: AppColors.textSecondary,
                  letterSpacing: 0.5,
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            dateStr,
            style: const TextStyle(
              fontSize: 13,
              fontWeight: FontWeight.w700,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 3),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
            decoration: BoxDecoration(
              color: TagaTowersDesignTokens.softMint,
              borderRadius: BorderRadius.circular(4),
            ),
            child: Text(
              timeStr,
              style: const TextStyle(
                fontSize: 10,
                fontWeight: FontWeight.w700,
                color: TagaTowersDesignTokens.deepEmerald,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMiniBadge({
    required String label,
    required Color color,
    required Color background,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: color,
        ),
      ),
    );
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // 4. GUEST LIST SUMMARY CARD (if Guest booking)
  // ═══════════════════════════════════════════════════════════════════════════

  Widget _buildGuestListCard() {
    final guests = widget.booking.guestDetails;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 16,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(
                LucideIcons.users,
                size: 16,
                color: TagaTowersDesignTokens.primaryGreen,
              ),
              const SizedBox(width: 8),
              Text(
                'GUEST INFORMATION (${guests.length})',
                style: const TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 0.7,
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: guests.length,
            separatorBuilder: (context, index) => const SizedBox(height: 8),
            itemBuilder: (context, index) {
              final guest = guests[index];
              return Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFFF8FAFC),
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(
                    color: const Color(0xFFE2E8F0),
                    width: 1.0,
                  ),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Container(
                      width: 28,
                      height: 28,
                      decoration: const BoxDecoration(
                        color: TagaTowersDesignTokens.softMint,
                        shape: BoxShape.circle,
                      ),
                      child: Center(
                        child: Text(
                          '${index + 1}',
                          style: const TextStyle(
                            fontSize: 12,
                            fontWeight: FontWeight.w800,
                            color: TagaTowersDesignTokens.deepEmerald,
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            guest.name.trim().isNotEmpty
                                ? guest.name.trim()
                                : 'Guest ${index + 1}',
                            style: const TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.w700,
                              color: AppColors.textPrimary,
                            ),
                          ),
                          const SizedBox(height: 4),
                          Wrap(
                            spacing: 10,
                            children: [
                              if (guest.age > 0)
                                Text(
                                  'Age: ${guest.age}',
                                  style: const TextStyle(
                                    fontSize: 12,
                                    color: AppColors.textSecondary,
                                  ),
                                ),
                              if (guest.gender.isNotEmpty)
                                Text(
                                  'Gender: ${guest.gender.toUpperCase()}',
                                  style: const TextStyle(
                                    fontSize: 12,
                                    color: AppColors.textSecondary,
                                  ),
                                ),
                              if (guest.contact.isNotEmpty)
                                Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    const Icon(
                                      LucideIcons.phone,
                                      size: 11,
                                      color: AppColors.textSecondary,
                                    ),
                                    const SizedBox(width: 3),
                                    Text(
                                      guest.contact,
                                      style: const TextStyle(
                                        fontSize: 12,
                                        color: AppColors.textSecondary,
                                      ),
                                    ),
                                  ],
                                ),
                            ],
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              );
            },
          ),
        ],
      ),
    );
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // 5. PAYMENT DETAILS CARD
  // ═══════════════════════════════════════════════════════════════════════════

  Widget _buildPaymentDetailsCard() {
    final paymentId = _effectivePaymentId;
    final orderId = _effectiveOrderId;
    final isPaymentIdCopied = _copiedKeys.contains('payment_id');
    final isOrderIdCopied = _copiedKeys.contains('order_id');

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 16,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: const [
              Icon(
                LucideIcons.creditCard,
                size: 16,
                color: TagaTowersDesignTokens.primaryGreen,
              ),
              SizedBox(width: 8),
              Text(
                'PAYMENT DETAILS',
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 0.7,
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),

          // Advance Amount Paid Row
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text(
                'Advance Amount Paid',
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: AppColors.textPrimary,
                ),
              ),
              Text(
                '₹${widget.booking.advanceAmount}',
                style: const TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.w800,
                  color: TagaTowersDesignTokens.deepEmerald,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),

          // Payment Gateway Status Row
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text(
                'Payment Gateway Status',
                style: TextStyle(
                  fontSize: 13.5,
                  color: AppColors.textSecondary,
                ),
              ),
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: TagaTowersDesignTokens.badgeDecoration(),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: const [
                    Icon(
                      LucideIcons.checkCircle2,
                      size: 13,
                      color: TagaTowersDesignTokens.deepEmerald,
                    ),
                    SizedBox(width: 5),
                    Text(
                      'Confirmed',
                      style: TextStyle(
                        fontSize: 11.5,
                        fontWeight: FontWeight.w700,
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),

          // Razorpay Payment ID (if available)
          if (paymentId.isNotEmpty) ...[
            const SizedBox(height: 10),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  'Razorpay Payment ID',
                  style: TextStyle(
                    fontSize: 13,
                    color: AppColors.textSecondary,
                  ),
                ),
                BouncyButton(
                  onTap: () =>
                      _copyText('payment_id', paymentId, 'Payment ID'),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        paymentId,
                        style: const TextStyle(
                          fontSize: 12.5,
                          fontFamily: 'monospace',
                          fontWeight: FontWeight.w700,
                          color: AppColors.textPrimary,
                        ),
                      ),
                      const SizedBox(width: 6),
                      Icon(
                        isPaymentIdCopied
                            ? LucideIcons.check
                            : LucideIcons.copy,
                        size: 13,
                        color: isPaymentIdCopied
                            ? TagaTowersDesignTokens.primaryGreen
                            : AppColors.textSecondary,
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ],

          // Razorpay Order ID (if available)
          if (orderId.isNotEmpty) ...[
            const SizedBox(height: 8),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  'Order ID',
                  style: TextStyle(
                    fontSize: 13,
                    color: AppColors.textSecondary,
                  ),
                ),
                BouncyButton(
                  onTap: () => _copyText('order_id', orderId, 'Order ID'),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        orderId,
                        style: const TextStyle(
                          fontSize: 12.5,
                          fontFamily: 'monospace',
                          fontWeight: FontWeight.w600,
                          color: AppColors.textSecondary,
                        ),
                      ),
                      const SizedBox(width: 6),
                      Icon(
                        isOrderIdCopied
                            ? LucideIcons.check
                            : LucideIcons.copy,
                        size: 13,
                        color: isOrderIdCopied
                            ? TagaTowersDesignTokens.primaryGreen
                            : AppColors.textSecondary,
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ],

          const SizedBox(height: 14),
          const Divider(height: 1, color: TagaTowersDesignTokens.subtleBorder),
          const SizedBox(height: 12),

          // Remaining balance notice
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
            decoration: BoxDecoration(
              color: TagaTowersDesignTokens.mutedBackground,
              borderRadius: BorderRadius.circular(10),
              border: Border.all(
                color: TagaTowersDesignTokens.subtleBorder,
                width: 1.0,
              ),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: const [
                Icon(
                  LucideIcons.info,
                  size: 15,
                  color: TagaTowersDesignTokens.primaryGreen,
                ),
                SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Remaining Balance: The remainder of the room tariff is payable at check-in or check-out directly at TAGA Towers reception.',
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w500,
                      color: AppColors.textPrimary,
                      height: 1.4,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // 6. PROPERTY CHECK-IN INSTRUCTIONS & CARETAKER CONTACTS
  // ═══════════════════════════════════════════════════════════════════════════

  Widget _buildCheckInAndCaretakersCard() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 16,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: const [
              Icon(
                LucideIcons.mapPin,
                size: 16,
                color: TagaTowersDesignTokens.primaryGreen,
              ),
              SizedBox(width: 8),
              Text(
                'CHECK-IN INSTRUCTIONS & CONTACTS',
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 0.7,
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),

          // Property Address
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: const Color(0xFFF8FAFC),
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: const Color(0xFFE2E8F0), width: 1.0),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: const [
                    Icon(
                      LucideIcons.building,
                      size: 14,
                      color: TagaTowersDesignTokens.primaryGreen,
                    ),
                    SizedBox(width: 6),
                    Text(
                      'TAGA Towers Location',
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w700,
                        color: AppColors.textPrimary,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 4),
                Text(
                  TagaTowersDesignTokens.towerAddress,
                  style: const TextStyle(
                    fontSize: 12.5,
                    color: AppColors.textSecondary,
                    height: 1.35,
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 14),

          // Caretakers List
          const Text(
            'Facility Caretakers (Egmore, Chennai):',
            style: TextStyle(
              fontSize: 13,
              fontWeight: FontWeight.w700,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 8),
          ..._kCaretakers.map((caretaker) {
            return Padding(
              padding: const EdgeInsets.only(bottom: 8.0),
              child: Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(
                    color: TagaTowersDesignTokens.subtleBorder,
                    width: 1.0,
                  ),
                ),
                child: Row(
                  children: [
                    Container(
                      width: 32,
                      height: 32,
                      decoration: const BoxDecoration(
                        color: TagaTowersDesignTokens.softMint,
                        shape: BoxShape.circle,
                      ),
                      child: const Center(
                        child: Icon(
                          LucideIcons.userCheck,
                          size: 16,
                          color: TagaTowersDesignTokens.deepEmerald,
                        ),
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            caretaker.name,
                            style: const TextStyle(
                              fontSize: 13.5,
                              fontWeight: FontWeight.w700,
                              color: AppColors.textPrimary,
                            ),
                          ),
                          Text(
                            caretaker.phone,
                            style: const TextStyle(
                              fontSize: 12,
                              color: AppColors.textSecondary,
                            ),
                          ),
                        ],
                      ),
                    ),
                    BouncyButton(
                      onTap: () => _makePhoneCall(caretaker.phone),
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 10,
                          vertical: 6,
                        ),
                        decoration: BoxDecoration(
                          color: TagaTowersDesignTokens.softMint,
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(
                            color: TagaTowersDesignTokens.mintHighlight,
                            width: 1.0,
                          ),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: const [
                            Icon(
                              LucideIcons.phoneCall,
                              size: 12,
                              color: TagaTowersDesignTokens.deepEmerald,
                            ),
                            SizedBox(width: 4),
                            Text(
                              'Call',
                              style: TextStyle(
                                fontSize: 11.5,
                                fontWeight: FontWeight.w700,
                                color: TagaTowersDesignTokens.deepEmerald,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            );
          }),
          const SizedBox(height: 6),

          // Instructions bullet list
          _buildInstructionBullet(
            'Check-in starts at 12:00 AM Midnight on your arrival date.',
          ),
          const SizedBox(height: 4),
          _buildInstructionBullet(
            'Present your Booking Reference ID and a valid government photo ID upon check-in.',
          ),
          const SizedBox(height: 4),
          _buildInstructionBullet(
            'Please contact the caretakers upon arrival at Egmore for key allocation.',
          ),
        ],
      ),
    );
  }

  Widget _buildInstructionBullet(String text) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Padding(
          padding: EdgeInsets.only(top: 4.0),
          child: Icon(
            LucideIcons.check,
            size: 12,
            color: TagaTowersDesignTokens.primaryGreen,
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            text,
            style: const TextStyle(
              fontSize: 12,
              color: AppColors.textSecondary,
              height: 1.35,
            ),
          ),
        ),
      ],
    );
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // 7. NON-REFUNDABLE CANCELLATION NOTICE
  // ═══════════════════════════════════════════════════════════════════════════

  Widget _buildCancellationNoticeCard() {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: TagaTowersDesignTokens.alertDecoration(borderRadius: 14),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(
            LucideIcons.alertTriangle,
            size: 18,
            color: TagaTowersDesignTokens.warningAmberDark,
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Cancellation Policy (Non-Refundable Advance)',
                  style: TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w800,
                    color: TagaTowersDesignTokens.warningAmberDark,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  'The advance payment of ₹${widget.booking.advanceAmount} is strictly non-refundable. Cancellations can only be initiated strictly prior to your scheduled check-in date. No refund will be issued upon cancellation or no-show.',
                  style: TextStyle(
                    fontSize: 12,
                    color: TagaTowersDesignTokens.warningAmberDark
                        .withValues(alpha: 0.95),
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

  // ═══════════════════════════════════════════════════════════════════════════
  // 8. ACTIONS: "VIEW IN MY BOOKINGS" & "BACK TO HOME"
  // ═══════════════════════════════════════════════════════════════════════════

  Widget _buildActionButtons() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        // Primary: View in My Bookings
        BouncyButton(
          onTap: _onViewInMyBookings,
          child: Container(
            height: 52,
            decoration: BoxDecoration(
              color: TagaTowersDesignTokens.primaryGreen,
              borderRadius: BorderRadius.circular(14),
              boxShadow: [
                BoxShadow(
                  color: TagaTowersDesignTokens.primaryGreen
                      .withValues(alpha: 0.3),
                  blurRadius: 14,
                  offset: const Offset(0, 4),
                ),
              ],
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: const [
                Icon(
                  LucideIcons.calendarCheck2,
                  size: 18,
                  color: Colors.white,
                ),
                SizedBox(width: 8),
                Text(
                  'View in My Bookings',
                  style: TextStyle(
                    fontSize: 15.5,
                    fontWeight: FontWeight.w700,
                    color: Colors.white,
                    letterSpacing: -0.2,
                  ),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 12),

        // Secondary: Back to Home
        BouncyButton(
          onTap: _onBackToHome,
          child: Container(
            height: 48,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(
                color: TagaTowersDesignTokens.border,
                width: 1.2,
              ),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: const [
                Icon(
                  LucideIcons.home,
                  size: 17,
                  color: TagaTowersDesignTokens.deepEmerald,
                ),
                SizedBox(width: 8),
                Text(
                  'Back to Home',
                  style: TextStyle(
                    fontSize: 14.5,
                    fontWeight: FontWeight.w700,
                    color: TagaTowersDesignTokens.deepEmerald,
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // FALLBACK: UNCONFIRMED BOOKING SAFETY STATE
  // ═══════════════════════════════════════════════════════════════════════════

  Widget _buildUnconfirmedState() {
    return Padding(
      padding: const EdgeInsets.all(24.0),
      child: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 72,
              height: 72,
              decoration: BoxDecoration(
                color: TagaTowersDesignTokens.warningAmberLight,
                shape: BoxShape.circle,
                border: Border.all(
                  color: TagaTowersDesignTokens.warningAmberBorder,
                  width: 2,
                ),
              ),
              child: const Center(
                child: Icon(
                  LucideIcons.alertCircle,
                  size: 36,
                  color: TagaTowersDesignTokens.warningAmberDark,
                ),
              ),
            ),
            const SizedBox(height: 20),
            const Text(
              'Awaiting Confirmation',
              style: TextStyle(
                fontSize: 22,
                fontWeight: FontWeight.w800,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: 10),
            Text(
              'This confirmation screen is only displayed once the booking has been confirmed by the backend. Current payment status: "${widget.booking.paymentStatus}".',
              textAlign: TextAlign.center,
              style: const TextStyle(
                fontSize: 14,
                color: AppColors.textSecondary,
                height: 1.45,
              ),
            ),
            const SizedBox(height: 28),
            BouncyButton(
              onTap: _onViewInMyBookings,
              child: Container(
                height: 48,
                padding: const EdgeInsets.symmetric(horizontal: 24),
                decoration: BoxDecoration(
                  color: TagaTowersDesignTokens.primaryGreen,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Center(
                  child: Text(
                    'Check Status in My Bookings',
                    style: TextStyle(
                      fontSize: 14.5,
                      fontWeight: FontWeight.w700,
                      color: Colors.white,
                    ),
                  ),
                ),
              ),
            ),
            const SizedBox(height: 12),
            BouncyButton(
              onTap: _onBackToHome,
              child: Container(
                height: 44,
                padding: const EdgeInsets.symmetric(horizontal: 24),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: TagaTowersDesignTokens.border,
                    width: 1.2,
                  ),
                ),
                child: const Center(
                  child: Text(
                    'Back to Home',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w700,
                      color: TagaTowersDesignTokens.deepEmerald,
                    ),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
