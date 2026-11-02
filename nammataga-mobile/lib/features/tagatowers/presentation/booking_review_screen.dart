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
import '../domain/room_model.dart';
import '../providers/tagatowers_provider.dart';
import 'theme/tagatowers_design_tokens.dart';
import 'widgets/booking_step_progress_bar.dart';
import 'widgets/price_breakdown_card.dart';

/// Step 3: Booking Review Screen (Agent 8)
///
/// Multi-step booking review screen displaying:
/// 1. Progress Header (Step 3 of 4: Review).
/// 2. Booking Summary Card (Room, dates, 12:00 AM timestamps, 24-hr cycles, bed count, booker phone).
/// 3. Guest Details Roster (if Guest Booking, expandable/clean cards with name, age, gender, contact).
/// 4. Financial Breakdown via [PriceBreakdownCard] (rate per bed, advance payable, remaining balance note, non-refundable policy).
/// 5. Seamless navigation & re-validation before backend creation holding beds for 10 min.
class BookingReviewScreen extends ConsumerStatefulWidget {
  const BookingReviewScreen({super.key});

  @override
  ConsumerState<BookingReviewScreen> createState() => _BookingReviewScreenState();
}

class _BookingReviewScreenState extends ConsumerState<BookingReviewScreen> {
  bool _isCreatingBooking = false;
  bool _isGuestRosterExpanded = true;
  String? _errorMessage;

  void _handleBack() {
    HapticFeedback.lightImpact();
    final flow = ref.read(bookingFlowNotifierProvider);

    // Step 1 for Self booking, Step 2 for Guest booking (Zero data loss)
    if (flow.bookingFor == 'guest') {
      if (context.canPop()) {
        context.pop();
      } else {
        context.pushReplacement('/taga-towers/booking/guest-details');
      }
    } else {
      if (context.canPop()) {
        context.pop();
      } else {
        context.pushReplacement('/taga-towers/booking/bed-selection');
      }
    }
  }

  void _onStepTapped(int stepIndex) {
    final flow = ref.read(bookingFlowNotifierProvider);
    if (stepIndex == 1) {
      if (context.canPop()) {
        context.pop();
      } else {
        context.pushReplacement('/taga-towers/booking/bed-selection');
      }
    } else if (stepIndex == 2 && flow.bookingFor == 'guest') {
      if (context.canPop()) {
        context.pop();
      } else {
        context.pushReplacement('/taga-towers/booking/guest-details');
      }
    }
  }

  Future<void> _onProceedToPayment() async {
    if (_isCreatingBooking) return;

    final flow = ref.read(bookingFlowNotifierProvider);
    final room = flow.selectedRoom;
    final checkIn = flow.checkInDate;
    final checkOut = flow.checkOutDate;

    if (room == null || checkIn == null || checkOut == null) {
      _showErrorSnackBar('Booking details are incomplete. Please select a room and stay dates.');
      return;
    }

    setState(() {
      _isCreatingBooking = true;
      _errorMessage = null;
    });

    try {
      final repo = ref.read(tagaTowersRepositoryProvider);

      // 1. Re-validate live availability over selected date range
      final params = DateRangeParams(checkIn: checkIn, checkOut: checkOut);
      final liveAvailMap = await repo.checkAvailabilityRange(
        params.checkInFormatted,
        params.checkOutFormatted,
      );

      final liveAvail = liveAvailMap[room.id];
      final requestedBeds = flow.effectiveBeds;

      if (liveAvail == null || !liveAvail.available || liveAvail.availableBeds < requestedBeds) {
        if (!mounted) return;
        final availableCount = liveAvail?.availableBeds ?? 0;
        _showAvailabilityConflictDialog(room.name, availableCount, requestedBeds);
        setState(() => _isCreatingBooking = false);
        return;
      }

      // 2. Resolve booker ID and contact phone
      final session = ref.read(authProvider).value;
      final userMap = session?.user;
      String bookerId = session?.tagaId ?? '';
      if (bookerId.isEmpty) {
        bookerId = (userMap?['taga_id'] ??
                userMap?['tagaId'] ??
                userMap?['sno']?.toString() ??
                userMap?['id']?.toString() ??
                '')
            .toString();
      }

      String phone = flow.bookerPhone.trim().replaceAll(' ', '');
      if (phone.isEmpty || phone == '+91') {
        phone = (userMap?['phone'] ??
                userMap?['contact'] ??
                userMap?['mobile'] ??
                '')
            .toString();
      }
      if (phone.isEmpty) {
        phone = '+919999999999';
      }

      // 3. Create pending booking on backend
      final req = CreateBookingRequest(
        roomId: room.id,
        checkInDate: params.checkInFormatted,
        checkOutDate: params.checkOutFormatted,
        bookerPhone: phone,
        bookingFor: flow.bookingFor,
        bedCount: flow.effectiveBeds,
        gender: flow.bookingFor == 'self' ? flow.bookingGender : null,
        guestDetails: flow.bookingFor == 'guest' ? flow.guestDetails : null,
      );

      final bookingRes = await repo.createBooking(req, bookerId);

      // 4. Update pendingBookingId in state
      ref.read(bookingFlowNotifierProvider.notifier).setPendingBookingId(bookingRes.id);

      if (!mounted) return;
      HapticFeedback.mediumImpact();

      // 5. Navigate to Payment Review Screen
      context.push('/taga-towers/booking/payment-review');
    } catch (e) {
      if (!mounted) return;
      final msg = e.toString().replaceAll('Exception:', '').trim();
      setState(() => _errorMessage = msg);
      _showErrorSnackBar(msg);
    } finally {
      if (mounted) {
        setState(() => _isCreatingBooking = false);
      }
    }
  }

  void _showErrorSnackBar(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(LucideIcons.alertCircle, color: Colors.white, size: 18),
            const SizedBox(width: 10),
            Expanded(
              child: Text(
                message,
                style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w500),
              ),
            ),
          ],
        ),
        backgroundColor: const Color(0xFFDC2626),
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        duration: const Duration(seconds: 4),
      ),
    );
  }

  void _showAvailabilityConflictDialog(String roomName, int availableBeds, int requestedBeds) {
    showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(18)),
        title: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: TagaTowersDesignTokens.warningAmberLight,
                borderRadius: BorderRadius.circular(10),
              ),
              child: const Icon(
                LucideIcons.alertTriangle,
                color: TagaTowersDesignTokens.warningAmberDark,
                size: 22,
              ),
            ),
            const SizedBox(width: 12),
            const Expanded(
              child: Text(
                'Live Bed Count Changed',
                style: TextStyle(fontSize: 16.5, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        content: Text(
          'Unfortunately, $roomName now has only $availableBeds bed(s) available for your selected dates (you requested $requestedBeds).\n\nPlease choose another room or change your dates.',
          style: const TextStyle(fontSize: 13.5, height: 1.45, color: AppColors.textPrimary),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('Stay Here'),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: TagaTowersDesignTokens.primaryGreen,
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
            ),
            onPressed: () {
              Navigator.of(ctx).pop();
              context.pushReplacement('/taga-towers/availability');
            },
            child: const Text('Change Room'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final flow = ref.watch(bookingFlowNotifierProvider);
    final room = flow.selectedRoom;
    final checkIn = flow.checkInDate;
    final checkOut = flow.checkOutDate;
    final session = ref.watch(authProvider).value;

    if (room == null || checkIn == null || checkOut == null) {
      return Scaffold(
        backgroundColor: TagaTowersDesignTokens.surfaceLight,
        appBar: AppBar(
          title: const Text('Review Booking'),
          backgroundColor: Colors.white,
          leading: IconButton(
            icon: const Icon(LucideIcons.arrowLeft),
            onPressed: () => context.pop(),
          ),
        ),
        body: Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(LucideIcons.bedDouble, size: 54, color: AppColors.textSecondary),
                const SizedBox(height: 16),
                const Text(
                  'No Room Selected',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                ),
                const SizedBox(height: 8),
                const Text(
                  'Please select a room and stay dates before reviewing your booking.',
                  textAlign: TextAlign.center,
                  style: TextStyle(color: AppColors.textSecondary, fontSize: 13.5),
                ),
                const SizedBox(height: 20),
                BouncyButton(
                  onTap: () => context.pushReplacement('/taga-towers/availability'),
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
                    decoration: BoxDecoration(
                      color: TagaTowersDesignTokens.primaryGreen,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Text(
                      'Browse Available Rooms',
                      style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      );
    }

    final isGuestBooking = flow.bookingFor == 'guest';
    final bookerName = session?.displayName.isNotEmpty == true ? session!.displayName : 'TAGA Member';
    final bookerPhone = flow.bookerPhone.isNotEmpty ? flow.bookerPhone : '+91 98765 43210';
    final dateFormat = DateFormat('EEE, dd MMM yyyy');

    return Scaffold(
      backgroundColor: TagaTowersDesignTokens.surfaceLight,
      appBar: AppBar(
        title: const Text(
          'Step 3: Review Booking',
          style: TextStyle(
            fontWeight: FontWeight.bold,
            fontSize: 18,
            color: TagaTowersDesignTokens.deepEmerald,
            letterSpacing: -0.2,
          ),
        ),
        backgroundColor: Colors.white,
        elevation: 0,
        surfaceTintColor: Colors.transparent,
        leading: IconButton(
          icon: const Icon(LucideIcons.arrowLeft, color: TagaTowersDesignTokens.deepEmerald),
          onPressed: _handleBack,
        ),
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(1),
          child: Container(
            color: TagaTowersDesignTokens.subtleBorder,
            height: 1,
          ),
        ),
      ),
      body: SingleChildScrollView(
        physics: const BouncingScrollPhysics(),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // 1. Progress Bar
            BookingStepProgressBar(
              currentStep: 3,
              onStepTapped: _onStepTapped,
            ),
            const SizedBox(height: 14),

            // Error Notice Banner if any
            if (_errorMessage != null) ...[
              Container(
                margin: const EdgeInsets.only(bottom: 12),
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFFFEF2F2),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: const Color(0xFFFCA5A5)),
                ),
                child: Row(
                  children: [
                    const Icon(LucideIcons.alertCircle, color: Color(0xFFDC2626), size: 18),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        _errorMessage!,
                        style: const TextStyle(
                          color: Color(0xFFB91C1C),
                          fontSize: 12.5,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ],

            // 2. Booking Summary Card
            _buildBookingSummaryCard(
              context: context,
              room: room,
              checkIn: checkIn,
              checkOut: checkOut,
              totalNights: flow.totalNights,
              effectiveBeds: flow.effectiveBeds,
              bookingFor: flow.bookingFor,
              bookerName: bookerName,
              bookerPhone: bookerPhone,
              dateFormat: dateFormat,
            ),
            const SizedBox(height: 14),

            // 3. Guest Details Roster (if Guest Booking)
            if (isGuestBooking) ...[
              _buildGuestDetailsRoster(
                context: context,
                guests: flow.guestDetails,
              ),
              const SizedBox(height: 14),
            ],

            // 4. Financial Breakdown Card
            PriceBreakdownCard(
              roomType: room.type,
              roomNumber: room.name,
              stayDurationDays: flow.totalNights,
              bedCount: flow.effectiveBeds,
              bookingFor: flow.bookingFor,
              advanceRatePerBed: flow.advanceRatePerBed,
              totalAdvanceAmount: flow.advanceAmount,
              checkInDate: checkIn,
              checkOutDate: checkOut,
              remainingBalanceNotice:
                  'Remaining balance is payable at the property upon check-in/checkout.',
              cancellationNotice:
                  'Cancellation Policy: Advance payment is strictly non-refundable upon cancellation.',
              isProcessingPayment: _isCreatingBooking,
              payButtonLabel: 'Proceed to Payment (₹${flow.advanceAmount})',
              onPayPressed: _onProceedToPayment,
            ),
            const SizedBox(height: 16),

            // Navigation Helpers & Data Safety Footnote
            _buildBottomNavigationRow(
              context: context,
              isGuestBooking: isGuestBooking,
            ),
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  /// Booking Summary Card displaying room details, stay timeline, and contact information.
  Widget _buildBookingSummaryCard({
    required BuildContext context,
    required Room room,
    required DateTime checkIn,
    required DateTime checkOut,
    required int totalNights,
    required int effectiveBeds,
    required String bookingFor,
    required String bookerName,
    required String bookerPhone,
    required DateFormat dateFormat,
  }) {
    final roomDisplayName = TagaTowersDesignTokens.getRoomDisplayName(room.type);
    final isSelf = bookingFor == 'self';

    return Container(
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 18,
      ),
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Room Name & Badges
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      room.name,
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: TagaTowersDesignTokens.deepEmerald,
                        letterSpacing: -0.3,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      roomDisplayName,
                      style: const TextStyle(
                        fontSize: 12.5,
                        color: AppColors.textSecondary,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4.5),
                decoration: BoxDecoration(
                  color: isSelf
                      ? TagaTowersDesignTokens.softMint
                      : const Color(0xFFE0F2FE),
                  borderRadius: BorderRadius.circular(999),
                  border: Border.all(
                    color: isSelf
                        ? TagaTowersDesignTokens.mintHighlight
                        : const Color(0xFFBAE6FD),
                    width: 1,
                  ),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      isSelf ? LucideIcons.user : LucideIcons.users,
                      size: 12,
                      color: isSelf
                          ? TagaTowersDesignTokens.deepEmerald
                          : const Color(0xFF0369A1),
                    ),
                    const SizedBox(width: 5),
                    Text(
                      isSelf ? 'Self Booking' : 'Guest Booking',
                      style: TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.w800,
                        color: isSelf
                            ? TagaTowersDesignTokens.deepEmerald
                            : const Color(0xFF0369A1),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Stay Dates & Midnight Schedule Container
          Container(
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: const Color(0xFFF7FAF8),
              borderRadius: BorderRadius.circular(14),
              border: Border.all(
                color: const Color(0xFFE3EDE6),
                width: 1,
              ),
            ),
            child: Column(
              children: [
                // Check-in Row
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Container(
                      padding: const EdgeInsets.all(7),
                      decoration: BoxDecoration(
                        color: TagaTowersDesignTokens.softMint,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: const Icon(
                        LucideIcons.logIn,
                        size: 15,
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text(
                            'Check-in Date & Time',
                            style: TextStyle(
                              fontSize: 11.5,
                              color: AppColors.textSecondary,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            '${dateFormat.format(checkIn)}, 12:00 AM (Midnight)',
                            style: const TextStyle(
                              fontSize: 13,
                              fontWeight: FontWeight.bold,
                              color: AppColors.textPrimary,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: 10),
                  child: Divider(height: 1, color: Color(0xFFE3EDE6)),
                ),

                // Check-out Row
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Container(
                      padding: const EdgeInsets.all(7),
                      decoration: BoxDecoration(
                        color: const Color(0xFFFEF3C7),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: const Icon(
                        LucideIcons.logOut,
                        size: 15,
                        color: Color(0xFFB45309),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text(
                            'Check-out Date & Time',
                            style: TextStyle(
                              fontSize: 11.5,
                              color: AppColors.textSecondary,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            '${dateFormat.format(checkOut)}, 12:00 AM (Midnight)',
                            style: const TextStyle(
                              fontSize: 13,
                              fontWeight: FontWeight.bold,
                              color: AppColors.textPrimary,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(height: 14),

          // Total Nights & 24-hr Stay Cycle Pill
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: TagaTowersDesignTokens.mutedBackground,
              borderRadius: BorderRadius.circular(10),
            ),
            child: Row(
              children: [
                const Icon(
                  LucideIcons.clock,
                  size: 14,
                  color: TagaTowersDesignTokens.primaryGreen,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    '$totalNights Night${totalNights > 1 ? 's' : ''} • ${totalNights * 24}-Hour Stay Schedule',
                    style: const TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                      color: TagaTowersDesignTokens.deepEmerald,
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),

          // Bed Count & Booker Info Rows
          _buildInfoRow(
            icon: LucideIcons.bedDouble,
            label: 'Total Beds Booked',
            value: '$effectiveBeds Bed${effectiveBeds > 1 ? 's' : ''}',
            badgeText: 'Capacity: ${room.capacity}',
          ),
          const SizedBox(height: 10),
          _buildInfoRow(
            icon: LucideIcons.userCheck,
            label: 'Booked By',
            value: bookerName,
          ),
          const SizedBox(height: 10),
          _buildInfoRow(
            icon: LucideIcons.phone,
            label: 'Booker Contact',
            value: bookerPhone,
          ),
        ],
      ),
    );
  }

  /// Guest Details Roster card for Guest Bookings.
  Widget _buildGuestDetailsRoster({
    required BuildContext context,
    required List<GuestDetail> guests,
  }) {
    return Container(
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 18,
      ),
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header with expand/collapse toggle
          InkWell(
            onTap: () {
              HapticFeedback.selectionClick();
              setState(() {
                _isGuestRosterExpanded = !_isGuestRosterExpanded;
              });
            },
            borderRadius: BorderRadius.circular(8),
            child: Row(
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
                        LucideIcons.users,
                        size: 17,
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                    const SizedBox(width: 10),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Guest Details Roster',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            color: AppColors.textPrimary,
                            letterSpacing: -0.2,
                          ),
                        ),
                        Text(
                          '${guests.length} Guest${guests.length > 1 ? 's' : ''} Registered',
                          style: const TextStyle(
                            fontSize: 11.5,
                            color: AppColors.textSecondary,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
                Icon(
                  _isGuestRosterExpanded
                      ? LucideIcons.chevronUp
                      : LucideIcons.chevronDown,
                  size: 20,
                  color: AppColors.textSecondary,
                ),
              ],
            ),
          ),

          // Expanded Guest List
          if (_isGuestRosterExpanded) ...[
            const SizedBox(height: 14),
            const Divider(height: 1, color: Color(0xFFE2EBE5)),
            const SizedBox(height: 12),
            ListView.separated(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: guests.length,
              separatorBuilder: (_, _) => const Padding(
                padding: EdgeInsets.symmetric(vertical: 8),
                child: Divider(height: 1, color: Color(0xFFF1F5F3)),
              ),
              itemBuilder: (context, index) {
                final guest = guests[index];
                final displayName = guest.name.trim().isNotEmpty
                    ? guest.name.trim()
                    : 'Guest ${index + 1}';
                final displayGender = guest.gender.trim().isNotEmpty
                    ? (guest.gender.trim()[0].toUpperCase() +
                        guest.gender.trim().substring(1).toLowerCase())
                    : 'Gender unspecified';
                final displayAge = guest.age > 0 ? '${guest.age} yrs' : 'Age not provided';
                final displayPhone = guest.contact.trim().isNotEmpty
                    ? guest.contact.trim()
                    : 'No contact provided';

                return Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Guest Badge Icon
                    Container(
                      width: 32,
                      height: 32,
                      decoration: BoxDecoration(
                        color: TagaTowersDesignTokens.softMint,
                        shape: BoxShape.circle,
                        border: Border.all(
                          color: TagaTowersDesignTokens.mintHighlight,
                          width: 1,
                        ),
                      ),
                      child: Center(
                        child: Text(
                          'G${index + 1}',
                          style: const TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w800,
                            color: TagaTowersDesignTokens.deepEmerald,
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),

                    // Guest Info Details
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            displayName,
                            style: const TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.bold,
                              color: AppColors.textPrimary,
                            ),
                          ),
                          const SizedBox(height: 4),
                          Wrap(
                            spacing: 6,
                            runSpacing: 4,
                            children: [
                              _buildGuestChip(displayAge),
                              _buildGuestChip(displayGender),
                              _buildGuestChip(displayPhone, icon: LucideIcons.phone),
                            ],
                          ),
                        ],
                      ),
                    ),
                  ],
                );
              },
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildGuestChip(String text, {IconData? icon}) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2.5),
      decoration: BoxDecoration(
        color: const Color(0xFFF4F7F5),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(
          color: const Color(0xFFDDE6E1),
          width: 0.8,
        ),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (icon != null) ...[
            Icon(icon, size: 10.5, color: AppColors.textSecondary),
            const SizedBox(width: 3.5),
          ],
          Text(
            text,
            style: const TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.w600,
              color: AppColors.textSecondary,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildInfoRow({
    required IconData icon,
    required String label,
    required String value,
    String? badgeText,
  }) {
    return Row(
      children: [
        Icon(icon, size: 16, color: AppColors.textSecondary),
        const SizedBox(width: 10),
        Text(
          label,
          style: const TextStyle(
            fontSize: 12.5,
            color: AppColors.textSecondary,
            fontWeight: FontWeight.w500,
          ),
        ),
        const Spacer(),
        if (badgeText != null) ...[
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
            decoration: BoxDecoration(
              color: const Color(0xFFF1F5F3),
              borderRadius: BorderRadius.circular(4),
            ),
            child: Text(
              badgeText,
              style: const TextStyle(
                fontSize: 10.5,
                fontWeight: FontWeight.w600,
                color: Color(0xFF5A7366),
              ),
            ),
          ),
          const SizedBox(width: 6),
        ],
        Text(
          value,
          style: const TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
        ),
      ],
    );
  }

  /// Bottom navigation row providing an explicit Back button with zero data loss.
  Widget _buildBottomNavigationRow({
    required BuildContext context,
    required bool isGuestBooking,
  }) {
    final backLabel = isGuestBooking ? 'Edit Guest Info' : 'Edit Room & Beds';

    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        TextButton.icon(
          onPressed: _handleBack,
          icon: const Icon(LucideIcons.arrowLeft, size: 16, color: TagaTowersDesignTokens.primaryGreen),
          label: Text(
            backLabel,
            style: const TextStyle(
              fontSize: 13,
              fontWeight: FontWeight.bold,
              color: TagaTowersDesignTokens.primaryGreen,
            ),
          ),
        ),
      ],
    );
  }
}
