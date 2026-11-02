import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../../core/widgets/bouncy_button.dart';
import '../../auth/providers/auth_provider.dart';
import '../../profile/providers/profile_provider.dart';
import '../domain/booking_model.dart';
import '../providers/tagatowers_provider.dart';
import 'theme/tagatowers_design_tokens.dart';

/// Screen displaying the comprehensive inspection details for a TAGA Towers booking.
///
/// Supports loading via direct [initialBooking] payload or lazy fetching using [bookingId].
class BookingDetailsScreen extends ConsumerStatefulWidget {
  /// The unique reference ID of the room booking
  final String bookingId;

  /// Optional preloaded booking response passed via route extra
  final BookingResponse? initialBooking;

  const BookingDetailsScreen({
    super.key,
    required this.bookingId,
    this.initialBooking,
  });

  @override
  ConsumerState<BookingDetailsScreen> createState() => _BookingDetailsScreenState();
}

class _BookingDetailsScreenState extends ConsumerState<BookingDetailsScreen> {
  BookingResponse? _localBooking;

  @override
  void initState() {
    super.initState();
    _localBooking = widget.initialBooking;
  }

  Future<void> _refreshBooking() async {
    HapticFeedback.lightImpact();
    ref.invalidate(bookingByIdProvider(widget.bookingId));
    ref.invalidate(userBookingsProvider);
    ref.invalidate(pastBookingsProvider);
  }

  Future<void> _copyToClipboard(String text, String label) async {
    await Clipboard.setData(ClipboardData(text: text));
    HapticFeedback.lightImpact();
    if (mounted) {
      ScaffoldMessenger.of(context).clearSnackBars();
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Row(
            children: [
              const Icon(LucideIcons.checkCheck, color: Colors.white, size: 18),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  '$label copied to clipboard',
                  style: const TextStyle(fontWeight: FontWeight.w500),
                ),
              ),
            ],
          ),
          backgroundColor: TagaTowersDesignTokens.deepEmerald,
          behavior: SnackBarBehavior.floating,
          duration: const Duration(seconds: 2),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        ),
      );
    }
  }

  Future<void> _makePhoneCall(String phoneNumber) async {
    HapticFeedback.lightImpact();
    final cleanNumber = phoneNumber.replaceAll(RegExp(r'\s+'), '');
    final uri = Uri(scheme: 'tel', path: cleanNumber);
    try {
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri);
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Could not open phone dialer for $phoneNumber'),
              backgroundColor: Colors.redAccent,
              behavior: SnackBarBehavior.floating,
            ),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to initiate call: $e'),
            backgroundColor: Colors.redAccent,
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    }
  }

  void _navigateToCancellation(BookingResponse booking) {
    HapticFeedback.mediumImpact();
    context.push(
      '/taga-towers/booking/cancel/${booking.id}',
      extra: booking,
    );
  }

  DateTime? _parseDate(String dateStr) {
    if (dateStr.isEmpty) return null;
    try {
      return DateTime.parse(dateStr);
    } catch (_) {
      return null;
    }
  }

  int _calculateStayDays(DateTime? checkIn, DateTime? checkOut) {
    if (checkIn == null || checkOut == null) return 1;
    final diff = checkOut.difference(checkIn).inDays;
    return diff > 0 ? diff : 1;
  }

  bool _isPastOrOnCheckIn(BookingResponse booking) {
    final checkIn = _parseDate(booking.checkInDate);
    if (checkIn == null) return false;
    final now = DateTime.now();
    final todayOnly = DateTime(now.year, now.month, now.day);
    final checkInOnly = DateTime(checkIn.year, checkIn.month, checkIn.day);
    return !todayOnly.isBefore(checkInOnly); // today >= checkIn
  }

  @override
  Widget build(BuildContext context) {
    if (_localBooking != null) {
      return _buildScaffold(_localBooking!);
    }

    final bookingAsync = ref.watch(bookingByIdProvider(widget.bookingId));

    return bookingAsync.when(
      loading: () => _buildLoadingScaffold(),
      error: (error, _) => _buildErrorScaffold(error.toString()),
      data: (booking) {
        if (booking == null) {
          return _buildNotFoundScaffold();
        }
        return _buildScaffold(booking);
      },
    );
  }

  Widget _buildLoadingScaffold() {
    return Scaffold(
      backgroundColor: TagaTowersDesignTokens.surfaceLight,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        surfaceTintColor: Colors.transparent,
        title: const Text(
          'Booking Details',
          style: TextStyle(
            color: TagaTowersDesignTokens.deepEmerald,
            fontWeight: FontWeight.w700,
            fontSize: 18,
          ),
        ),
        leading: IconButton(
          icon: const Icon(LucideIcons.arrowLeft, color: TagaTowersDesignTokens.deepEmerald),
          onPressed: () => Navigator.maybePop(context),
        ),
      ),
      body: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const CircularProgressIndicator(
              valueColor: AlwaysStoppedAnimation<Color>(TagaTowersDesignTokens.primaryGreen),
              strokeWidth: 3,
            ),
            const SizedBox(height: 16),
            Text(
              'Retrieving booking details...',
              style: TextStyle(
                color: Colors.grey.shade700,
                fontSize: 14,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildNotFoundScaffold() {
    return Scaffold(
      backgroundColor: TagaTowersDesignTokens.surfaceLight,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        title: const Text(
          'Booking Not Found',
          style: TextStyle(color: TagaTowersDesignTokens.deepEmerald, fontWeight: FontWeight.w700),
        ),
        leading: IconButton(
          icon: const Icon(LucideIcons.arrowLeft, color: TagaTowersDesignTokens.deepEmerald),
          onPressed: () => Navigator.maybePop(context),
        ),
      ),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.grey.shade100,
                  shape: BoxShape.circle,
                ),
                child: Icon(LucideIcons.searchX, size: 48, color: Colors.grey.shade500),
              ),
              const SizedBox(height: 20),
              const Text(
                'Booking Record Not Found',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              Text(
                'We could not locate booking reference "${widget.bookingId}". It may have expired or been removed.',
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 14, color: Colors.grey.shade600),
              ),
              const SizedBox(height: 24),
              BouncyButton(
                onTap: _refreshBooking,
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                  decoration: BoxDecoration(
                    color: TagaTowersDesignTokens.primaryGreen,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Text(
                    'Retry Search',
                    style: TextStyle(color: Colors.white, fontWeight: FontWeight.w600),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildErrorScaffold(String errorMessage) {
    return Scaffold(
      backgroundColor: TagaTowersDesignTokens.surfaceLight,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        title: const Text(
          'Booking Details',
          style: TextStyle(color: TagaTowersDesignTokens.deepEmerald, fontWeight: FontWeight.w700),
        ),
        leading: IconButton(
          icon: const Icon(LucideIcons.arrowLeft, color: TagaTowersDesignTokens.deepEmerald),
          onPressed: () => Navigator.maybePop(context),
        ),
      ),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.red.shade50,
                  shape: BoxShape.circle,
                ),
                child: Icon(LucideIcons.alertCircle, size: 48, color: Colors.red.shade600),
              ),
              const SizedBox(height: 20),
              const Text(
                'Unable to Load Booking',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              Text(
                errorMessage,
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 14, color: Colors.grey.shade600),
              ),
              const SizedBox(height: 24),
              BouncyButton(
                onTap: _refreshBooking,
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                  decoration: BoxDecoration(
                    color: TagaTowersDesignTokens.primaryGreen,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Text(
                    'Try Again',
                    style: TextStyle(color: Colors.white, fontWeight: FontWeight.w600),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildScaffold(BookingResponse booking) {
    final checkIn = _parseDate(booking.checkInDate);
    final checkOut = _parseDate(booking.checkOutDate);
    final totalStayDays = _calculateStayDays(checkIn, checkOut);
    final isGuestBooking = booking.bookingFor.toLowerCase() == 'guest';
    final isCancelled = booking.isCancelled;
    final canCancel = booking.isCancellable;
    final pastCheckIn = _isPastOrOnCheckIn(booking);

    return Scaffold(
      backgroundColor: TagaTowersDesignTokens.surfaceLight,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0.5,
        surfaceTintColor: Colors.transparent,
        title: const Text(
          'Booking Details',
          style: TextStyle(
            color: TagaTowersDesignTokens.deepEmerald,
            fontWeight: FontWeight.w700,
            fontSize: 18,
          ),
        ),
        leading: IconButton(
          icon: const Icon(LucideIcons.arrowLeft, color: TagaTowersDesignTokens.deepEmerald),
          onPressed: () => Navigator.maybePop(context),
        ),
        actions: [
          IconButton(
            icon: const Icon(LucideIcons.rotateCw, color: TagaTowersDesignTokens.deepEmerald, size: 20),
            tooltip: 'Refresh Booking',
            onPressed: _refreshBooking,
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _refreshBooking,
        color: TagaTowersDesignTokens.primaryGreen,
        backgroundColor: Colors.white,
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(parent: BouncingScrollPhysics()),
          padding: const EdgeInsets.fromLTRB(16, 16, 16, 32),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // 1. Header with Room Name, Room Type badge & Status badge
              _buildHeaderCard(booking),
              const SizedBox(height: 14),

              // 2. Booking Reference ID card with copy button
              _buildBookingReferenceCard(booking),
              const SizedBox(height: 14),

              // 3. Stay Schedule (Check-in, Check-out, Duration, 24-hr Midnight notice)
              _buildStayScheduleCard(booking, checkIn, checkOut, totalStayDays),
              const SizedBox(height: 14),

              // 4. Occupancy & Category Card
              _buildOccupancyCategoryCard(booking),
              const SizedBox(height: 14),

              // 5. Guest Details Roster (if Guest Booking)
              if (isGuestBooking) ...[
                _buildGuestDetailsRosterCard(booking),
                const SizedBox(height: 14),
              ],

              // 6. Payment & Financial Card
              _buildPaymentFinancialCard(booking),
              const SizedBox(height: 14),

              // 7. Property Location & Caretakers
              _buildLocationCaretakersCard(),
              const SizedBox(height: 14),

              // 8. Cancellation Policy Card
              _buildCancellationPolicyCard(),
              const SizedBox(height: 20),

              // 9. Cancellation Action Section
              _buildCancellationActionSection(
                booking: booking,
                isCancelled: isCancelled,
                canCancel: canCancel,
                pastCheckIn: pastCheckIn,
              ),
            ],
          ),
        ),
      ),
    );
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // SECTION BUILDERS
  // ═══════════════════════════════════════════════════════════════════════════

  /// 1. Header Card with Room Name, Room Type badge, and Status badge
  Widget _buildHeaderCard(BookingResponse booking) {
    final roomDisplayName = booking.roomName.isNotEmpty
        ? booking.roomName
        : TagaTowersDesignTokens.getRoomDisplayName(booking.roomId);

    final roomTypeBadgeText = TagaTowersDesignTokens.getRoomDisplayName(booking.roomId);

    return Container(
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 18,
      ),
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: TagaTowersDesignTokens.softMint.withValues(alpha: 0.6),
                  borderRadius: BorderRadius.circular(14),
                  border: Border.all(
                    color: TagaTowersDesignTokens.mintHighlight.withValues(alpha: 0.4),
                  ),
                ),
                child: const Icon(
                  LucideIcons.building2,
                  color: TagaTowersDesignTokens.deepEmerald,
                  size: 26,
                ),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      roomDisplayName,
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.w700,
                        color: Color(0xFF0F241D),
                        height: 1.25,
                      ),
                    ),
                    const SizedBox(height: 6),
                    // Room Type badge
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF1F8E9),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: const Color(0xFFC8E6C9), width: 1),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Icon(LucideIcons.doorOpen, size: 12, color: TagaTowersDesignTokens.deepEmerald),
                          const SizedBox(width: 5),
                          Text(
                            roomTypeBadgeText,
                            style: const TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.w600,
                              color: TagaTowersDesignTokens.deepEmerald,
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
          const SizedBox(height: 14),
          const Divider(height: 1, thickness: 1, color: Color(0xFFEEF3F0)),
          const SizedBox(height: 12),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'Current Status',
                style: TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.w500,
                  color: Colors.grey.shade600,
                ),
              ),
              _buildStatusBadge(booking),
            ],
          ),
        ],
      ),
    );
  }

  /// Status badge displaying Upcoming, Active, Completed, or Cancelled
  Widget _buildStatusBadge(BookingResponse booking) {
    final status = booking.bookingStatus.toLowerCase();
    final isCancelled = booking.isCancelled;

    Color bg;
    Color border;
    Color text;
    IconData icon;
    String label;

    if (isCancelled) {
      bg = const Color(0xFFFFE4E6);
      border = const Color(0xFFFECDD3);
      text = const Color(0xFFBE123C);
      icon = LucideIcons.xCircle;
      label = 'Cancelled';
    } else if (status == 'active') {
      bg = const Color(0xFFD8F3DC);
      border = const Color(0xFF95D5B2);
      text = const Color(0xFF1B4332);
      icon = LucideIcons.checkCircle2;
      label = 'Active Stay';
    } else if (status == 'completed') {
      bg = const Color(0xFFF1F5F9);
      border = const Color(0xFFCBD5E1);
      text = const Color(0xFF475569);
      icon = LucideIcons.checkCheck;
      label = 'Completed';
    } else {
      bg = const Color(0xFFE0F2FE);
      border = const Color(0xFFBAE6FD);
      text = const Color(0xFF0369A1);
      icon = LucideIcons.calendar;
      label = 'Upcoming';
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: border, width: 1.2),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 13, color: text),
          const SizedBox(width: 5),
          Text(
            label,
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w700,
              color: text,
              letterSpacing: 0.2,
            ),
          ),
        ],
      ),
    );
  }

  /// 2. Booking Reference ID card with copy button
  Widget _buildBookingReferenceCard(BookingResponse booking) {
    return Container(
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 16,
      ),
      padding: const EdgeInsets.all(16),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: const Color(0xFFF0FDF4),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: const Color(0xFFBBF7D0)),
            ),
            child: const Icon(LucideIcons.hash, size: 20, color: TagaTowersDesignTokens.deepEmerald),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'BOOKING REFERENCE ID',
                  style: TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w600,
                    letterSpacing: 0.6,
                    color: Colors.grey.shade600,
                  ),
                ),
                const SizedBox(height: 3),
                Text(
                  booking.id,
                  style: const TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w700,
                    color: Color(0xFF1E293B),
                    fontFamily: 'monospace',
                  ),
                ),
              ],
            ),
          ),
          BouncyButton(
            onTap: () => _copyToClipboard(booking.id, 'Booking ID'),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: TagaTowersDesignTokens.softMint.withValues(alpha: 0.5),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(
                  color: TagaTowersDesignTokens.mintHighlight.withValues(alpha: 0.5),
                ),
              ),
              child: const Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(LucideIcons.copy, size: 14, color: TagaTowersDesignTokens.deepEmerald),
                  SizedBox(width: 6),
                  Text(
                    'Copy',
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                      color: TagaTowersDesignTokens.deepEmerald,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// 3. Stay Schedule Card (Check-in, Check-out, Duration, 24-hr midnight cycle)
  Widget _buildStayScheduleCard(
    BookingResponse booking,
    DateTime? checkIn,
    DateTime? checkOut,
    int totalStayDays,
  ) {
    final dateFormat = DateFormat('EEE, dd MMM yyyy');
    final checkInFormatted = checkIn != null ? dateFormat.format(checkIn) : booking.checkInDate;
    final checkOutFormatted = checkOut != null ? dateFormat.format(checkOut) : booking.checkOutDate;

    return Container(
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 18,
      ),
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(LucideIcons.calendarClock, size: 20, color: TagaTowersDesignTokens.deepEmerald),
              const SizedBox(width: 8),
              const Text(
                'Stay Schedule',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF0F241D),
                ),
              ),
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                decoration: BoxDecoration(
                  color: TagaTowersDesignTokens.softMint.withValues(alpha: 0.5),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(
                  '$totalStayDays ${totalStayDays == 1 ? 'Cycle' : 'Cycles'}',
                  style: const TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w700,
                    color: TagaTowersDesignTokens.deepEmerald,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Check-in & Check-out row
          Container(
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: const Color(0xFFF9FBFA),
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: TagaTowersDesignTokens.subtleBorder),
            ),
            child: Row(
              children: [
                // Check-in column
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Container(
                            width: 8,
                            height: 8,
                            decoration: const BoxDecoration(
                              color: TagaTowersDesignTokens.primaryGreen,
                              shape: BoxShape.circle,
                            ),
                          ),
                          const SizedBox(width: 6),
                          Text(
                            'CHECK-IN',
                            style: TextStyle(
                              fontSize: 11,
                              fontWeight: FontWeight.w700,
                              color: Colors.grey.shade600,
                              letterSpacing: 0.5,
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 6),
                      Text(
                        checkInFormatted,
                        style: const TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w700,
                          color: Color(0xFF1E293B),
                        ),
                      ),
                      const SizedBox(height: 2),
                      const Text(
                        '12:00 AM',
                        style: TextStyle(
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                          color: TagaTowersDesignTokens.primaryGreen,
                        ),
                      ),
                    ],
                  ),
                ),

                // Separator arrow
                Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    shape: BoxShape.circle,
                    border: Border.all(color: TagaTowersDesignTokens.subtleBorder),
                  ),
                  child: const Icon(
                    LucideIcons.arrowRight,
                    size: 14,
                    color: TagaTowersDesignTokens.deepEmerald,
                  ),
                ),

                // Check-out column
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.only(left: 12.0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Container(
                              width: 8,
                              height: 8,
                              decoration: const BoxDecoration(
                                color: Color(0xFFE11D48),
                                shape: BoxShape.circle,
                              ),
                            ),
                            const SizedBox(width: 6),
                            Text(
                              'CHECK-OUT',
                              style: TextStyle(
                                fontSize: 11,
                                fontWeight: FontWeight.w700,
                                color: Colors.grey.shade600,
                                letterSpacing: 0.5,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 6),
                        Text(
                          checkOutFormatted,
                          style: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w700,
                            color: Color(0xFF1E293B),
                          ),
                        ),
                        const SizedBox(height: 2),
                        const Text(
                          '12:00 AM Midnight',
                          style: TextStyle(
                            fontSize: 12,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFFE11D48),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),

          // Total Stay Duration pill & 24-hr midnight cycle badge
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
            decoration: BoxDecoration(
              color: const Color(0xFFF1F8E9),
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: const Color(0xFFC8E6C9)),
            ),
            child: Row(
              children: [
                const Icon(LucideIcons.clock4, size: 15, color: TagaTowersDesignTokens.deepEmerald),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Total Stay Duration: $totalStayDays ${totalStayDays == 1 ? 'Day' : 'Days'} ($totalStayDays × 24-hr cycle)',
                    style: const TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                      color: TagaTowersDesignTokens.deepEmerald,
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

  /// 4. Occupancy & Category Card
  Widget _buildOccupancyCategoryCard(BookingResponse booking) {
    final isGuest = booking.bookingFor.toLowerCase() == 'guest';

    // Resolve booker phone with comprehensive fallback:
    // 1. Direct booking.bookerPhone
    // 2. Member profile mobileNumber (if authenticated member)
    // 3. Session user mobileNumber
    final profile = ref.watch(memberProfileProvider).value;
    final session = ref.watch(authProvider).value;
    final fallbackPhone = profile?.mobileNumber.isNotEmpty == true
        ? profile!.mobileNumber
        : (session?.user?['mobileNumber'] ??
                session?.user?['mobile_number'] ??
                session?.user?['contact'] ??
                session?.user?['phone'])
            ?.toString();

    String bookerPhoneText = 'Not Provided';
    if (booking.bookerPhone != null &&
        booking.bookerPhone!.trim().isNotEmpty &&
        booking.bookerPhone!.trim() != '+91') {
      bookerPhoneText = booking.bookerPhone!;
    } else if (fallbackPhone != null &&
        fallbackPhone.trim().isNotEmpty &&
        fallbackPhone.trim() != '+91') {
      bookerPhoneText = fallbackPhone.startsWith('+') ? fallbackPhone : '+91 $fallbackPhone';
    }

    return Container(
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 18,
      ),
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Row(
            children: [
              Icon(LucideIcons.users, size: 20, color: TagaTowersDesignTokens.deepEmerald),
              SizedBox(width: 8),
              Text(
                'Occupancy & Category',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF0F241D),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Booking For & Number of Beds row
          Row(
            children: [
              Expanded(
                child: _buildInfoTile(
                  icon: isGuest ? LucideIcons.userPlus : LucideIcons.userCheck,
                  label: 'BOOKING FOR',
                  value: isGuest ? 'Guest Booking' : 'Self (Member)',
                  valueColor: isGuest ? const Color(0xFF0369A1) : TagaTowersDesignTokens.deepEmerald,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _buildInfoTile(
                  icon: LucideIcons.bedDouble,
                  label: 'NUMBER OF BEDS',
                  value: '${booking.bedCount} ${booking.bedCount == 1 ? 'Bed' : 'Beds'}',
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          const Divider(height: 1, thickness: 1, color: Color(0xFFEEF3F0)),
          const SizedBox(height: 12),

          // Booker Information
          Text(
            'BOOKER DETAILS',
            style: TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.w700,
              letterSpacing: 0.6,
              color: Colors.grey.shade600,
            ),
          ),
          const SizedBox(height: 8),
          _buildDetailRow(label: 'Booker Name', value: booking.bookerName),
          const SizedBox(height: 6),
          _buildDetailRow(label: 'TAGA ID', value: booking.bookerId, isMonospace: true),
          const SizedBox(height: 6),
          _buildDetailRow(label: 'Phone Number', value: bookerPhoneText),
        ],
      ),
    );
  }

  /// 5. Guest Details Roster (if Guest Booking)
  Widget _buildGuestDetailsRosterCard(BookingResponse booking) {
    final guests = booking.guestDetails;

    return Container(
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 18,
      ),
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(LucideIcons.userCheck, size: 20, color: TagaTowersDesignTokens.deepEmerald),
              const SizedBox(width: 8),
              const Text(
                'Guest Details Roster',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF0F241D),
                ),
              ),
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                decoration: BoxDecoration(
                  color: const Color(0xFFE0F2FE),
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(
                  '${guests.length} ${guests.length == 1 ? 'Guest' : 'Guests'}',
                  style: const TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w700,
                    color: Color(0xFF0369A1),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            'Registered accompanying guests for this accommodation reservation:',
            style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
          ),
          const SizedBox(height: 14),

          if (guests.isEmpty)
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.grey.shade50,
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: Colors.grey.shade200),
              ),
              child: Row(
                children: [
                  Icon(LucideIcons.info, size: 16, color: Colors.grey.shade600),
                  const SizedBox(width: 8),
                  Text(
                    'No individual guest profiles recorded.',
                    style: TextStyle(fontSize: 13, color: Colors.grey.shade700),
                  ),
                ],
              ),
            )
          else
            ListView.separated(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: guests.length,
              separatorBuilder: (context, _) => const SizedBox(height: 10),
              itemBuilder: (context, index) {
                final guest = guests[index];
                return _buildGuestRosterItem(index + 1, guest);
              },
            ),
        ],
      ),
    );
  }

  /// Guest Roster Item Tile
  Widget _buildGuestRosterItem(int index, GuestDetail guest) {
    final genderDisplay = guest.gender.isNotEmpty
        ? guest.gender[0].toUpperCase() + guest.gender.substring(1)
        : 'Not specified';

    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFFAFCFA),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFFE2E8F0)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 28,
            height: 28,
            decoration: BoxDecoration(
              color: TagaTowersDesignTokens.softMint,
              borderRadius: BorderRadius.circular(8),
            ),
            alignment: Alignment.center,
            child: Text(
              '$index',
              style: const TextStyle(
                fontWeight: FontWeight.w700,
                fontSize: 13,
                color: TagaTowersDesignTokens.deepEmerald,
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  guest.name.isNotEmpty ? guest.name : 'Guest $index',
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w700,
                    color: Color(0xFF1E293B),
                  ),
                ),
                const SizedBox(height: 4),
                Row(
                  children: [
                    if (guest.age > 0) ...[
                      Text(
                        'Age: ${guest.age}',
                        style: TextStyle(fontSize: 12, color: Colors.grey.shade700),
                      ),
                      const SizedBox(width: 8),
                      Container(width: 3, height: 3, decoration: const BoxDecoration(color: Colors.grey, shape: BoxShape.circle)),
                      const SizedBox(width: 8),
                    ],
                    Text(
                      'Gender: $genderDisplay',
                      style: TextStyle(fontSize: 12, color: Colors.grey.shade700),
                    ),
                  ],
                ),
                if (guest.contact.isNotEmpty) ...[
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      const Icon(LucideIcons.phone, size: 12, color: TagaTowersDesignTokens.primaryGreen),
                      const SizedBox(width: 5),
                      Text(
                        guest.contact,
                        style: const TextStyle(
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF334155),
                        ),
                      ),
                    ],
                  ),
                ],
              ],
            ),
          ),
          if (guest.contact.isNotEmpty)
            BouncyButton(
              onTap: () => _makePhoneCall(guest.contact),
              child: Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: TagaTowersDesignTokens.softMint.withValues(alpha: 0.6),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: TagaTowersDesignTokens.mintHighlight.withValues(alpha: 0.5)),
                ),
                child: const Icon(LucideIcons.phoneCall, size: 16, color: TagaTowersDesignTokens.deepEmerald),
              ),
            ),
        ],
      ),
    );
  }

  /// 6. Payment & Financial Card
  Widget _buildPaymentFinancialCard(BookingResponse booking) {
    final paymentStatusUpper = booking.paymentStatus.toUpperCase();
    final isConfirmed = booking.isConfirmed;
    final isPending = booking.isPending;
    final isCancelled = booking.isCancelled;

    Color badgeBg;
    Color badgeBorder;
    Color badgeText;
    IconData badgeIcon;

    if (isConfirmed) {
      badgeBg = const Color(0xFFD8F3DC);
      badgeBorder = const Color(0xFF95D5B2);
      badgeText = const Color(0xFF1B4332);
      badgeIcon = LucideIcons.shieldCheck;
    } else if (isPending) {
      badgeBg = const Color(0xFFFEF3C7);
      badgeBorder = const Color(0xFFFCD34D);
      badgeText = const Color(0xFFB45309);
      badgeIcon = LucideIcons.clock;
    } else if (isCancelled) {
      badgeBg = const Color(0xFFFFE4E6);
      badgeBorder = const Color(0xFFFECDD3);
      badgeText = const Color(0xFFBE123C);
      badgeIcon = LucideIcons.xCircle;
    } else {
      badgeBg = const Color(0xFFF1F5F9);
      badgeBorder = const Color(0xFFCBD5E1);
      badgeText = const Color(0xFF475569);
      badgeIcon = LucideIcons.info;
    }

    return Container(
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 18,
      ),
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Row(
            children: [
              Icon(LucideIcons.creditCard, size: 20, color: TagaTowersDesignTokens.deepEmerald),
              SizedBox(width: 8),
              Text(
                'Payment & Financial Summary',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF0F241D),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Advance Amount Highlight
          Container(
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: [Color(0xFFF0FDF4), Color(0xFFDCFCE7)],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: const Color(0xFF86EFAC)),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'ADVANCE AMOUNT PAID',
                      style: TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        letterSpacing: 0.6,
                        color: Colors.green.shade800,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      '₹${booking.advanceAmount}',
                      style: const TextStyle(
                        fontSize: 24,
                        fontWeight: FontWeight.w800,
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                  ],
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                  decoration: BoxDecoration(
                    color: badgeBg,
                    borderRadius: BorderRadius.circular(999),
                    border: Border.all(color: badgeBorder, width: 1.2),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(badgeIcon, size: 14, color: badgeText),
                      const SizedBox(width: 5),
                      Text(
                        paymentStatusUpper,
                        style: TextStyle(
                          fontSize: 12,
                          fontWeight: FontWeight.w700,
                          color: badgeText,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 14),

          // Razorpay Transaction Details (if available)
          if (booking.razorpayOrderId != null && booking.razorpayOrderId!.isNotEmpty) ...[
            _buildTransactionRow(
              label: 'Razorpay Order ID',
              value: booking.razorpayOrderId!,
              onCopy: () => _copyToClipboard(booking.razorpayOrderId!, 'Order ID'),
            ),
            const SizedBox(height: 8),
          ],

          if (booking.razorpayPaymentId != null && booking.razorpayPaymentId!.isNotEmpty) ...[
            _buildTransactionRow(
              label: 'Razorpay Payment ID',
              value: booking.razorpayPaymentId!,
              onCopy: () => _copyToClipboard(booking.razorpayPaymentId!, 'Payment ID'),
            ),
            const SizedBox(height: 8),
          ],

          // Remaining Balance Info Box
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: const Color(0xFFF8FAFC),
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: const Color(0xFFE2E8F0)),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Icon(LucideIcons.badgePercent, size: 16, color: TagaTowersDesignTokens.deepEmerald),
                const SizedBox(width: 8),
                Expanded(
                  child: RichText(
                    text: TextSpan(
                      style: TextStyle(fontSize: 12, color: Colors.grey.shade700, height: 1.35),
                      children: const [
                        TextSpan(
                          text: 'Remaining Balance: ',
                          style: TextStyle(fontWeight: FontWeight.w700, color: Color(0xFF1E293B)),
                        ),
                        TextSpan(
                          text: 'Room tariff balance is payable directly at reception during check-in or checkout.',
                        ),
                      ],
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

  Widget _buildTransactionRow({
    required String label,
    required String value,
    required VoidCallback onCopy,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: const Color(0xFFFAFAFA),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: TextStyle(fontSize: 11, color: Colors.grey.shade600, fontWeight: FontWeight.w500),
                ),
                Text(
                  value,
                  style: const TextStyle(
                    fontSize: 13,
                    fontFamily: 'monospace',
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF1E293B),
                  ),
                ),
              ],
            ),
          ),
          IconButton(
            icon: const Icon(LucideIcons.copy, size: 14, color: TagaTowersDesignTokens.deepEmerald),
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints(),
            tooltip: 'Copy $label',
            onPressed: onCopy,
          ),
        ],
      ),
    );
  }

  /// 7. Property Location & Caretaker Contacts Card
  Widget _buildLocationCaretakersCard() {
    return Container(
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 18,
      ),
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Row(
            children: [
              Icon(LucideIcons.mapPin, size: 20, color: TagaTowersDesignTokens.deepEmerald),
              SizedBox(width: 8),
              Text(
                'Property Location & Caretakers',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF0F241D),
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),

          // Address Container
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: const Color(0xFFF8FAFC),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: const Color(0xFFE2E8F0)),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Icon(LucideIcons.navigation, size: 16, color: TagaTowersDesignTokens.deepEmerald),
                const SizedBox(width: 10),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'TAGA Towers / Velanmai Illam',
                        style: TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.w700,
                          color: Color(0xFF1E293B),
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        '45, Sait Colony, II Street, Egmore, Chennai - 600008, Tamil Nadu, India',
                        style: TextStyle(fontSize: 12, color: Colors.grey.shade700, height: 1.3),
                      ),
                    ],
                  ),
                ),
                IconButton(
                  icon: const Icon(LucideIcons.copy, size: 15, color: TagaTowersDesignTokens.deepEmerald),
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(),
                  tooltip: 'Copy Address',
                  onPressed: () => _copyToClipboard(
                    'Velanmai Illam, 45, Sait Colony, II Street, Egmore, Chennai - 600008',
                    'Address',
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 14),

          // Caretaker contacts header
          Text(
            'ON-SITE CARETAKERS',
            style: TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.w700,
              letterSpacing: 0.6,
              color: Colors.grey.shade600,
            ),
          ),
          const SizedBox(height: 8),

          // Mr. Muthu
          _buildCaretakerTile(
            name: 'Mr. Muthu',
            phone: '96007 63744',
            role: 'Caretaker',
          ),
          const SizedBox(height: 8),

          // Mr. Mariyappan
          _buildCaretakerTile(
            name: 'Mr. Mariyappan',
            phone: '96770 10300',
            role: 'Caretaker',
          ),
          const SizedBox(height: 10),

          // Association Reception
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: const Color(0xFFF1F5F9),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Row(
              children: [
                const Icon(LucideIcons.headset, size: 14, color: Color(0xFF475569)),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Reception Helpdesk: 044 34919949 • nammataga@gmail.com',
                    style: TextStyle(fontSize: 11, color: Colors.grey.shade700, fontWeight: FontWeight.w500),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCaretakerTile({
    required String name,
    required String phone,
    required String role,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFF9FBFA),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: TagaTowersDesignTokens.subtleBorder),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: TagaTowersDesignTokens.softMint,
              borderRadius: BorderRadius.circular(10),
            ),
            child: const Icon(LucideIcons.user, size: 16, color: TagaTowersDesignTokens.deepEmerald),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  name,
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w700,
                    color: Color(0xFF1E293B),
                  ),
                ),
                Text(
                  phone,
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w500,
                    color: Colors.grey.shade700,
                  ),
                ),
              ],
            ),
          ),
          BouncyButton(
            onTap: () => _makePhoneCall(phone),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 7),
              decoration: BoxDecoration(
                color: TagaTowersDesignTokens.primaryGreen,
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(LucideIcons.phoneCall, size: 13, color: Colors.white),
                  SizedBox(width: 6),
                  Text(
                    'Call',
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                      color: Colors.white,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// 8. Cancellation Policy Card
  Widget _buildCancellationPolicyCard() {
    return Container(
      decoration: TagaTowersDesignTokens.alertDecoration(borderRadius: 16),
      padding: const EdgeInsets.all(16),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(
            LucideIcons.alertTriangle,
            size: 20,
            color: TagaTowersDesignTokens.warningAmberDark,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Cancellation Policy',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w700,
                    color: TagaTowersDesignTokens.warningAmberDark,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  'Advance payment is non-refundable upon cancellation. '
                  'Cancellations are strictly allowed only BEFORE the scheduled check-in date. '
                  'Cancellations on or after the check-in date (12:00 AM) cannot be processed.',
                  style: TextStyle(
                    fontSize: 12,
                    height: 1.4,
                    color: Colors.brown.shade800,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// 9. Cancellation Action Section
  Widget _buildCancellationActionSection({
    required BookingResponse booking,
    required bool isCancelled,
    required bool canCancel,
    required bool pastCheckIn,
  }) {
    if (isCancelled) {
      return Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: const Color(0xFFFFF1F2),
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: const Color(0xFFFECDD3)),
        ),
        child: Row(
          children: [
            const Icon(LucideIcons.xCircle, size: 20, color: Color(0xFFBE123C)),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Booking Cancelled',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w700,
                      color: Color(0xFFBE123C),
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    'This booking reservation was cancelled. Advance payment was non-refundable.',
                    style: TextStyle(
                      fontSize: 12,
                      color: Colors.red.shade800,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      );
    }

    if (canCancel) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          BouncyButton(
            onTap: () => _navigateToCancellation(booking),
            child: Container(
              height: 52,
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(14),
                border: Border.all(
                  color: const Color(0xFFE11D48),
                  width: 1.5,
                ),
                boxShadow: const [
                  BoxShadow(
                    color: Color(0x0DE11D48),
                    blurRadius: 10,
                    offset: Offset(0, 3),
                  ),
                ],
              ),
              alignment: Alignment.center,
              child: const Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(LucideIcons.xCircle, size: 18, color: Color(0xFFE11D48)),
                  SizedBox(width: 8),
                  Text(
                    'Cancel Booking',
                    style: TextStyle(
                      fontSize: 15,
                      fontWeight: FontWeight.w700,
                      color: Color(0xFFE11D48),
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 8),
          Center(
            child: Text(
              'Eligible for cancellation before check-in date',
              style: TextStyle(
                fontSize: 12,
                color: Colors.grey.shade600,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
        ],
      );
    }

    // Past or on check-in date
    if (pastCheckIn) {
      return Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: const Color(0xFFF1F5F9),
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: const Color(0xFFCBD5E1)),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Icon(LucideIcons.info, size: 18, color: Color(0xFF475569)),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Cancellation Not Permitted',
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w700,
                      color: Color(0xFF1E293B),
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    'Cancellation is not permitted on or after check-in date as per TAGA accommodation rules.',
                    style: TextStyle(
                      fontSize: 12,
                      color: Colors.grey.shade700,
                      height: 1.35,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      );
    }

    return const SizedBox.shrink();
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // REUSABLE HELPER TILES
  // ═══════════════════════════════════════════════════════════════════════════

  Widget _buildInfoTile({
    required IconData icon,
    required String label,
    required String value,
    Color? valueColor,
  }) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFF9FBFA),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: TagaTowersDesignTokens.subtleBorder),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, size: 14, color: TagaTowersDesignTokens.deepEmerald),
              const SizedBox(width: 5),
              Expanded(
                child: Text(
                  label,
                  style: TextStyle(
                    fontSize: 10,
                    fontWeight: FontWeight.w700,
                    letterSpacing: 0.5,
                    color: Colors.grey.shade600,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Text(
            value,
            style: TextStyle(
              fontSize: 13,
              fontWeight: FontWeight.w700,
              color: valueColor ?? const Color(0xFF1E293B),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDetailRow({
    required String label,
    required String value,
    bool isMonospace = false,
  }) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          label,
          style: TextStyle(
            fontSize: 13,
            color: Colors.grey.shade600,
            fontWeight: FontWeight.w500,
          ),
        ),
        Text(
          value,
          style: TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w600,
            color: const Color(0xFF1E293B),
            fontFamily: isMonospace ? 'monospace' : null,
          ),
        ),
      ],
    );
  }
}
