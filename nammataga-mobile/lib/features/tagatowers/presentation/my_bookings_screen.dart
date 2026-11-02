import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../auth/providers/auth_provider.dart';
import '../../profile/providers/profile_provider.dart';
import '../domain/booking_model.dart';
import '../providers/tagatowers_provider.dart';
import 'theme/tagatowers_design_tokens.dart';

/// My Bookings Screen for TAGA Towers Accommodation module.
///
/// Provides two distinct tabs:
/// 1. "Active & Upcoming" - Displays current stays and upcoming reservations.
///    Supports booking details navigation and pre-check-in cancellation.
/// 2. "Past Bookings" - Displays completed and archived stays with a year filter.
class MyBookingsScreen extends ConsumerStatefulWidget {
  final int initialTabIndex;

  const MyBookingsScreen({
    super.key,
    this.initialTabIndex = 0,
  });

  @override
  ConsumerState<MyBookingsScreen> createState() => _MyBookingsScreenState();
}

class _MyBookingsScreenState extends ConsumerState<MyBookingsScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  String _selectedYear = 'All';

  // Dynamically generated years: Exactly the 3 most recent years (e.g. 2027, 2026, 2025)
  late final List<String> _availableYears;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(
      length: 2,
      vsync: this,
      initialIndex: widget.initialTabIndex.clamp(0, 1),
    );

    final currentYear = DateTime.now().year;
    // Exactly 3 years (e.g., current year and past 2 years: 2026, 2025, 2024 or 2027, 2026, 2025)
    _availableYears = [
      currentYear.toString(),
      (currentYear - 1).toString(),
      (currentYear - 2).toString(),
    ];
    _selectedYear = currentYear.toString();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  /// Extracts the active member's booker ID (TAGA ID or user identifier)
  String _getBookerId(WidgetRef ref) {
    final session = ref.watch(authProvider).value;
    final profile = ref.watch(memberProfileProvider).value;

    if (session?.tagaId != null && session!.tagaId.isNotEmpty) {
      return session.tagaId;
    }
    if (profile?.tagaId != null && profile!.tagaId.isNotEmpty) {
      return profile.tagaId;
    }
    final userMap = session?.user;
    final idVal = userMap?['taga_id'] ??
        userMap?['tagaId'] ??
        userMap?['sno']?.toString() ??
        userMap?['id']?.toString() ??
        '';
    return idVal.toString();
  }

  /// Computes the effective booking status for display and filtering
  static String getEffectiveBookingStatus(BookingResponse b) {
    final status = b.bookingStatus.toLowerCase().trim();
    final pStatus = b.paymentStatus.toLowerCase().trim();

    if (status == 'cancelled' || pStatus == 'cancelled' || pStatus == 'refunded') {
      return 'cancelled';
    }
    if (status == 'upcoming' || status == 'active' || status == 'completed') {
      return status;
    }

    try {
      final checkIn = DateTime.parse(b.checkInDate);
      final checkOut = DateTime.parse(b.checkOutDate);
      final now = DateTime.now();

      if (now.isBefore(checkIn)) return 'upcoming';
      if (now.isAfter(checkOut) || now.isAtSameMomentAs(checkOut)) return 'completed';
      return 'active';
    } catch (_) {
      return 'upcoming';
    }
  }

  Future<void> _refreshAll(String bookerId) async {
    HapticFeedback.lightImpact();
    ref.invalidate(userBookingsProvider(bookerId));
    ref.invalidate(pastBookingsProvider(_selectedYear == 'All' ? null : _selectedYear));
  }

  void _navigateToRoomBooking() {
    HapticFeedback.lightImpact();
    if (context.canPop()) {
      context.pop();
    } else {
      context.go('/taga-towers');
    }
  }

  @override
  Widget build(BuildContext context) {
    final bookerId = _getBookerId(ref);
    final userBookingsAsync = ref.watch(userBookingsProvider(bookerId));

    // Calculate active + upcoming count for tab pill badge
    final activeCount = userBookingsAsync.maybeWhen(
      data: (bookings) => bookings.where((b) {
        final st = getEffectiveBookingStatus(b);
        return st == 'upcoming' || st == 'active';
      }).length,
      orElse: () => 0,
    );

    return Scaffold(
      backgroundColor: TagaTowersDesignTokens.surfaceLight,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        surfaceTintColor: Colors.transparent,
        leading: IconButton(
          icon: const Icon(LucideIcons.arrowLeft, color: TagaTowersDesignTokens.deepEmerald),
          tooltip: 'Back',
          onPressed: () {
            if (context.canPop()) {
              context.pop();
            } else {
              context.go('/taga-towers');
            }
          },
        ),
        title: const Text(
          'My Bookings',
          style: TextStyle(
            fontSize: 20,
            fontWeight: FontWeight.bold,
            color: TagaTowersDesignTokens.deepEmerald,
            letterSpacing: -0.3,
          ),
        ),
        actions: [
          IconButton(
            icon: const Icon(LucideIcons.rotateCcw, color: TagaTowersDesignTokens.primaryGreen),
            tooltip: 'Refresh Bookings',
            onPressed: () => _refreshAll(bookerId),
          ),
          const SizedBox(width: 4),
        ],
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(54),
          child: Container(
            margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
            padding: const EdgeInsets.all(4),
            decoration: BoxDecoration(
              color: TagaTowersDesignTokens.mutedBackground,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: TagaTowersDesignTokens.subtleBorder),
            ),
            child: TabBar(
              controller: _tabController,
              indicatorSize: TabBarIndicatorSize.tab,
              dividerColor: Colors.transparent,
              indicator: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(10),
                boxShadow: TagaTowersDesignTokens.softShadow,
                border: Border.all(color: TagaTowersDesignTokens.subtleBorder),
              ),
              labelColor: TagaTowersDesignTokens.deepEmerald,
              unselectedLabelColor: AppColors.textSecondary,
              labelStyle: const TextStyle(
                fontSize: 13.5,
                fontWeight: FontWeight.w700,
              ),
              unselectedLabelStyle: const TextStyle(
                fontSize: 13.5,
                fontWeight: FontWeight.w500,
              ),
              tabs: [
                Tab(
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Text('Active & Upcoming'),
                      if (activeCount > 0) ...[
                        const SizedBox(width: 8),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 2),
                          decoration: BoxDecoration(
                            color: TagaTowersDesignTokens.softMint,
                            borderRadius: BorderRadius.circular(999),
                            border: Border.all(
                              color: TagaTowersDesignTokens.mintHighlight,
                              width: 0.8,
                            ),
                          ),
                          child: Text(
                            '$activeCount',
                            style: const TextStyle(
                              fontSize: 11,
                              fontWeight: FontWeight.bold,
                              color: TagaTowersDesignTokens.deepEmerald,
                            ),
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
                const Tab(text: 'Past Bookings'),
              ],
            ),
          ),
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          // Tab 1: Active & Upcoming Bookings
          _ActiveUpcomingTab(
            bookerId: bookerId,
            onRefresh: () => _refreshAll(bookerId),
            onBookRoom: _navigateToRoomBooking,
          ),

          // Tab 2: Past Bookings (with Year Filter)
          _PastBookingsTab(
            bookerId: bookerId,
            selectedYear: _selectedYear,
            availableYears: _availableYears,
            onYearSelected: (year) {
              setState(() {
                _selectedYear = year;
              });
              ref.invalidate(
                pastBookingsProvider(year == 'All' ? null : year),
              );
            },
            onRefresh: () => _refreshAll(bookerId),
            onBookRoom: _navigateToRoomBooking,
          ),
        ],
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// TAB 1: ACTIVE & UPCOMING BOOKINGS
// ═══════════════════════════════════════════════════════════════════════════

class _ActiveUpcomingTab extends ConsumerWidget {
  final String bookerId;
  final Future<void> Function() onRefresh;
  final VoidCallback onBookRoom;

  const _ActiveUpcomingTab({
    required this.bookerId,
    required this.onRefresh,
    required this.onBookRoom,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (bookerId.isEmpty) {
      return _EmptyBookingsView(
        title: 'Sign In Required',
        message: 'Please sign in with your active TAGA membership to view your accommodation bookings.',
        buttonLabel: 'Book a Room',
        onButtonPressed: onBookRoom,
        icon: LucideIcons.userCheck,
      );
    }

    final userBookingsAsync = ref.watch(userBookingsProvider(bookerId));

    return RefreshIndicator(
      color: TagaTowersDesignTokens.primaryGreen,
      backgroundColor: Colors.white,
      onRefresh: onRefresh,
      child: userBookingsAsync.when(
        loading: () => ListView(
          padding: const EdgeInsets.all(16),
          children: const [
            _BookingCardShimmer(),
            SizedBox(height: 16),
            _BookingCardShimmer(),
            SizedBox(height: 16),
            _BookingCardShimmer(),
          ],
        ),
        error: (error, stack) => _ErrorBookingsView(
          message: error.toString().replaceAll('Exception: ', ''),
          onRetry: () => ref.refresh(userBookingsProvider(bookerId)),
        ),
        data: (bookings) {
          // Filter strictly for active or upcoming bookings
          final activeAndUpcoming = bookings.where((b) {
            final st = _MyBookingsScreenState.getEffectiveBookingStatus(b);
            return st == 'upcoming' || st == 'active';
          }).toList()
            ..sort((a, b) => a.checkInDate.compareTo(b.checkInDate)); // chronological

          if (activeAndUpcoming.isEmpty) {
            return SingleChildScrollView(
              physics: const AlwaysScrollableScrollPhysics(),
              child: SizedBox(
                height: MediaQuery.of(context).size.height * 0.72,
                child: _EmptyBookingsView(
                  title: 'No Active or Upcoming Bookings',
                  message:
                      'You do not have any room stays scheduled at TAGA Towers right now. Reserve in advance for your upcoming trip to Egmore, Chennai.',
                  buttonLabel: 'Book a Room',
                  onButtonPressed: onBookRoom,
                  icon: LucideIcons.calendarX,
                ),
              ),
            );
          }

          return ListView.separated(
            padding: const EdgeInsets.all(16),
            itemCount: activeAndUpcoming.length,
            separatorBuilder: (context, index) => const SizedBox(height: 16),
            itemBuilder: (context, index) {
              final booking = activeAndUpcoming[index];
              return _BookingCard(
                booking: booking,
                bookerId: bookerId,
              );
            },
          );
        },
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// TAB 2: PAST BOOKINGS (WITH YEAR FILTER)
// ═══════════════════════════════════════════════════════════════════════════

class _PastBookingsTab extends ConsumerWidget {
  final String bookerId;
  final String selectedYear;
  final List<String> availableYears;
  final ValueChanged<String> onYearSelected;
  final Future<void> Function() onRefresh;
  final VoidCallback onBookRoom;

  const _PastBookingsTab({
    required this.bookerId,
    required this.selectedYear,
    required this.availableYears,
    required this.onYearSelected,
    required this.onRefresh,
    required this.onBookRoom,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (bookerId.isEmpty) {
      return _EmptyBookingsView(
        title: 'Sign In Required',
        message: 'Please sign in with your active TAGA membership to view your past accommodation history.',
        buttonLabel: 'Book a Room',
        onButtonPressed: onBookRoom,
        icon: LucideIcons.userCheck,
      );
    }

    final pastQueryYear = selectedYear == 'All' ? null : selectedYear;
    final pastBookingsAsync = ref.watch(pastBookingsProvider(pastQueryYear));
    final activeBookingsAsync = ref.watch(userBookingsProvider(bookerId));

    final isLoading = pastBookingsAsync.isLoading && !pastBookingsAsync.hasValue;

    return RefreshIndicator(
      color: TagaTowersDesignTokens.primaryGreen,
      backgroundColor: Colors.white,
      onRefresh: onRefresh,
      child: Column(
        children: [
          // Year Filter Chips Row
          Container(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
            color: Colors.white,
            child: Row(
              children: [
                const Icon(
                  LucideIcons.filter,
                  size: 16,
                  color: TagaTowersDesignTokens.primaryGreen,
                ),
                const SizedBox(width: 8),
                const Text(
                  'Filter by Year:',
                  style: TextStyle(
                    fontSize: 12.5,
                    fontWeight: FontWeight.w600,
                    color: AppColors.textSecondary,
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: SingleChildScrollView(
                    scrollDirection: Axis.horizontal,
                    child: Row(
                      children: availableYears.map((year) {
                        final isSelected = selectedYear == year;
                        return Padding(
                          padding: const EdgeInsets.only(right: 8),
                          child: ChoiceChip(
                            label: Text(year),
                            selected: isSelected,
                            onSelected: (_) {
                              HapticFeedback.selectionClick();
                              onYearSelected(year);
                            },
                            selectedColor: TagaTowersDesignTokens.primaryGreen,
                            backgroundColor: TagaTowersDesignTokens.mutedBackground,
                            labelStyle: TextStyle(
                              fontSize: 12,
                              fontWeight: isSelected ? FontWeight.bold : FontWeight.w500,
                              color: isSelected ? Colors.white : AppColors.textPrimary,
                            ),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(999),
                              side: BorderSide(
                                color: isSelected
                                    ? TagaTowersDesignTokens.primaryGreen
                                    : TagaTowersDesignTokens.subtleBorder,
                              ),
                            ),
                            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                            showCheckmark: false,
                          ),
                        );
                      }).toList(),
                    ),
                  ),
                ),
              ],
            ),
          ),
          const Divider(height: 1, thickness: 1, color: Color(0x1A2D6A4F)),

          // Body Content
          Expanded(
            child: isLoading
                ? ListView(
                    padding: const EdgeInsets.all(16),
                    children: const [
                      _BookingCardShimmer(),
                      SizedBox(height: 16),
                      _BookingCardShimmer(),
                    ],
                  )
                : Builder(
                    builder: (context) {
                      // 1. Get past archived stays
                      final archivedList = pastBookingsAsync.value ?? [];

                      // 2. Combine with completed or cancelled bookings from active list
                      final activeList = activeBookingsAsync.value ?? [];
                      final completedOrCancelledFromActive = activeList.where((b) {
                        final st = _MyBookingsScreenState.getEffectiveBookingStatus(b);
                        return st == 'completed' || st == 'cancelled';
                      }).where((b) {
                        if (selectedYear == 'All') return true;
                        return b.checkInDate.startsWith(selectedYear);
                      }).toList();

                      // 3. Deduplicate by Booking ID
                      final Map<String, BookingResponse> map = {};
                      for (final b in archivedList) {
                        if (selectedYear == 'All' || b.checkInDate.startsWith(selectedYear)) {
                          map[b.id] = b;
                        }
                      }
                      for (final b in completedOrCancelledFromActive) {
                        map[b.id] = b;
                      }

                      final combinedPast = map.values.toList()
                        ..sort((a, b) => b.checkInDate.compareTo(a.checkInDate)); // newest first

                      if (combinedPast.isEmpty) {
                        return SingleChildScrollView(
                          physics: const AlwaysScrollableScrollPhysics(),
                          child: SizedBox(
                            height: MediaQuery.of(context).size.height * 0.65,
                            child: _EmptyBookingsView(
                              title: 'No Past Bookings',
                              message: selectedYear == 'All'
                                  ? 'No completed or cancelled bookings found in your accommodation records.'
                                  : 'No booking records found for the year $selectedYear.',
                              buttonLabel: 'Book a Room',
                              onButtonPressed: onBookRoom,
                              icon: LucideIcons.calendarCheck2,
                            ),
                          ),
                        );
                      }

                      return ListView.separated(
                        padding: const EdgeInsets.all(16),
                        itemCount: combinedPast.length,
                        separatorBuilder: (context, index) => const SizedBox(height: 16),
                        itemBuilder: (context, index) {
                          return _BookingCard(
                            booking: combinedPast[index],
                            bookerId: bookerId,
                            isPastBooking: true,
                          );
                        },
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// BOOKING CARD COMPONENT
// ═══════════════════════════════════════════════════════════════════════════

class _BookingCard extends ConsumerWidget {
  final BookingResponse booking;
  final String bookerId;
  final bool isPastBooking;

  const _BookingCard({
    required this.booking,
    required this.bookerId,
    this.isPastBooking = false,
  });

  String _formatDisplayDate(String dateStr) {
    try {
      final date = DateTime.parse(dateStr);
      const months = [
        'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
        'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'
      ];
      final d = date.day.toString().padLeft(2, '0');
      final m = months[date.month - 1];
      final y = date.year;
      return '$d $m $y';
    } catch (_) {
      return dateStr;
    }
  }

  int _calculateStayNights(String checkIn, String checkOut) {
    try {
      final start = DateTime.parse(checkIn);
      final end = DateTime.parse(checkOut);
      final diff = end.difference(start).inDays;
      return diff > 0 ? diff : 1;
    } catch (_) {
      return 1;
    }
  }

  void _copyReference(BuildContext context) {
    Clipboard.setData(ClipboardData(text: booking.id));
    HapticFeedback.lightImpact();
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(LucideIcons.check, size: 18, color: Colors.white),
            const SizedBox(width: 8),
            Text('Booking Reference #${booking.id} copied to clipboard'),
          ],
        ),
        backgroundColor: TagaTowersDesignTokens.deepEmerald,
        behavior: SnackBarBehavior.floating,
        duration: const Duration(seconds: 2),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
      ),
    );
  }

  void _onViewDetails(BuildContext context) {
    HapticFeedback.lightImpact();
    // Navigates to the designated booking details route
    context.push('/taga-towers/booking/details/${booking.id}');
  }

  void _onCancelBooking(BuildContext context, WidgetRef ref) {
    HapticFeedback.mediumImpact();
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => _CancellationConfirmDialog(
        booking: booking,
        bookerId: bookerId,
      ),
    );
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final status = _MyBookingsScreenState.getEffectiveBookingStatus(booking);
    final roomDisplayName = booking.roomName.isNotEmpty && booking.roomName != 'Room'
        ? booking.roomName
        : TagaTowersDesignTokens.getRoomDisplayName(booking.roomId);
    final roomTypeLabel = TagaTowersDesignTokens.getRoomDisplayName(booking.roomId);
    final stayNights = _calculateStayNights(booking.checkInDate, booking.checkOutDate);

    final canCancel = booking.isCancellable;

    return Container(
      decoration: TagaTowersDesignTokens.cardDecoration(
        borderColor: status == 'active'
            ? const Color(0xFFBAE6FD)
            : TagaTowersDesignTokens.subtleBorder,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // ── Header: Reference & Status Badge ──────────────────────────────
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 14, 16, 10),
            child: Row(
              children: [
                // Reference Chip
                InkWell(
                  onTap: () => _copyReference(context),
                  borderRadius: BorderRadius.circular(8),
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                    decoration: BoxDecoration(
                      color: TagaTowersDesignTokens.mutedBackground,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: TagaTowersDesignTokens.subtleBorder),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(LucideIcons.hash, size: 12, color: AppColors.textSecondary),
                        const SizedBox(width: 2),
                        Text(
                          booking.id.length > 12 ? booking.id.substring(0, 12) : booking.id,
                          style: const TextStyle(
                            fontSize: 12,
                            fontWeight: FontWeight.bold,
                            fontFamily: 'monospace',
                            color: TagaTowersDesignTokens.deepEmerald,
                          ),
                        ),
                        const SizedBox(width: 4),
                        const Icon(LucideIcons.copy, size: 12, color: AppColors.textSecondary),
                      ],
                    ),
                  ),
                ),
                const Spacer(),
                // Status Badge
                _StatusBadge(status: status),
              ],
            ),
          ),

          const Divider(height: 1, thickness: 1, color: Color(0x122D6A4F)),

          // ── Body: Room & Stay Information ─────────────────────────────────
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Room Name & Category
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: TagaTowersDesignTokens.softMint.withValues(alpha: 0.5),
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: TagaTowersDesignTokens.subtleBorder),
                      ),
                      child: const Icon(
                        LucideIcons.hotel,
                        size: 22,
                        color: TagaTowersDesignTokens.primaryGreen,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            roomDisplayName,
                            style: const TextStyle(
                              fontSize: 16.5,
                              fontWeight: FontWeight.bold,
                              color: AppColors.textPrimary,
                              letterSpacing: -0.2,
                            ),
                          ),
                          const SizedBox(height: 3),
                          Text(
                            roomTypeLabel,
                            style: const TextStyle(
                              fontSize: 12.5,
                              fontWeight: FontWeight.w500,
                              color: AppColors.textSecondary,
                            ),
                          ),
                        ],
                      ),
                    ),
                    // Advance Amount Pill
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                      decoration: BoxDecoration(
                        color: const Color(0xFFE8F5E9),
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(color: const Color(0xFFA5D6A7)),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.end,
                        children: [
                          Text(
                            '₹${booking.advanceAmount}',
                            style: const TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.bold,
                              color: Color(0xFF2E7D32),
                            ),
                          ),
                          const Text(
                            'Advance Paid',
                            style: TextStyle(
                              fontSize: 9.5,
                              fontWeight: FontWeight.w600,
                              color: Color(0xFF388E3C),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),

                const SizedBox(height: 14),

                // Stay Dates Container
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF9FBF9),
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(color: const Color(0x1F2D6A4F)),
                  ),
                  child: Row(
                    children: [
                      // Check-In
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text(
                              'CHECK-IN (12:00 AM)',
                              style: TextStyle(
                                fontSize: 10,
                                fontWeight: FontWeight.bold,
                                color: AppColors.textSecondary,
                                letterSpacing: 0.3,
                              ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              _formatDisplayDate(booking.checkInDate),
                              style: const TextStyle(
                                fontSize: 13,
                                fontWeight: FontWeight.w700,
                                color: TagaTowersDesignTokens.deepEmerald,
                              ),
                            ),
                          ],
                        ),
                      ),
                      // Arrow & Duration
                      Column(
                        children: [
                          const Icon(
                            LucideIcons.arrowRight,
                            size: 16,
                            color: TagaTowersDesignTokens.primaryGreen,
                          ),
                          const SizedBox(height: 2),
                          Text(
                            '$stayNights night${stayNights > 1 ? 's' : ''}',
                            style: const TextStyle(
                              fontSize: 10.5,
                              fontWeight: FontWeight.w600,
                              color: AppColors.textSecondary,
                            ),
                          ),
                        ],
                      ),
                      // Check-Out
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.end,
                          children: [
                            const Text(
                              'CHECK-OUT (12:00 AM)',
                              style: TextStyle(
                                fontSize: 10,
                                fontWeight: FontWeight.bold,
                                color: AppColors.textSecondary,
                                letterSpacing: 0.3,
                              ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              _formatDisplayDate(booking.checkOutDate),
                              style: const TextStyle(
                                fontSize: 13,
                                fontWeight: FontWeight.w700,
                                color: TagaTowersDesignTokens.deepEmerald,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),

                const SizedBox(height: 12),

                // Details Chips (Beds Count & Booking For)
                Row(
                  children: [
                    _FeatureChip(
                      icon: LucideIcons.bedDouble,
                      label: '${booking.bedCount} Bed${booking.bedCount > 1 ? 's' : ''}',
                    ),
                    const SizedBox(width: 8),
                    _FeatureChip(
                      icon: booking.bookingFor.toLowerCase() == 'self'
                          ? LucideIcons.user
                          : LucideIcons.users,
                      label: booking.bookingFor.toLowerCase() == 'self'
                          ? 'Booking for Self'
                          : 'Guest Stay (${booking.guestDetails.isNotEmpty ? booking.guestDetails.length : booking.bedCount})',
                    ),
                  ],
                ),
              ],
            ),
          ),

          // ── Non-cancellable Note (if stay already underway or completed) ───
          if (!canCancel && !isPastBooking && status != 'cancelled')
            Container(
              margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: const Color(0xFFFFFBEB),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: const Color(0xFFFDE68A)),
              ),
              child: const Row(
                children: [
                  Icon(LucideIcons.info, size: 14, color: Color(0xFFB45309)),
                  SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'Cannot cancel after check-in • Stay is active or scheduled for today',
                      style: TextStyle(
                        fontSize: 11.5,
                        fontWeight: FontWeight.w600,
                        color: Color(0xFF92400E),
                      ),
                    ),
                  ),
                ],
              ),
            ),

          const SizedBox(height: 4),
          const Divider(height: 1, thickness: 1, color: Color(0x122D6A4F)),

          // ── Actions Row ───────────────────────────────────────────────────
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            child: Row(
              children: [
                // Cancel Booking button (Visible ONLY before check-in)
                if (canCancel) ...[
                  Expanded(
                    child: OutlinedButton.icon(
                      icon: const Icon(LucideIcons.xCircle, size: 16, color: Color(0xFFDC2626)),
                      label: const Text(
                        'Cancel Booking',
                        style: TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFFDC2626),
                        ),
                      ),
                      style: OutlinedButton.styleFrom(
                        side: const BorderSide(color: Color(0xFFFCA5A5)),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                        padding: const EdgeInsets.symmetric(vertical: 10),
                        backgroundColor: const Color(0xFFFEF2F2),
                      ),
                      onPressed: () => _onCancelBooking(context, ref),
                    ),
                  ),
                  const SizedBox(width: 10),
                ],

                // View Details Button
                Expanded(
                  child: BouncyButton(
                    onTap: () => _onViewDetails(context),
                    child: Container(
                      padding: const EdgeInsets.symmetric(vertical: 10),
                      decoration: BoxDecoration(
                        color: TagaTowersDesignTokens.primaryGreen,
                        borderRadius: BorderRadius.circular(10),
                        boxShadow: TagaTowersDesignTokens.softShadow,
                      ),
                      child: const Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(LucideIcons.fileText, size: 16, color: Colors.white),
                          SizedBox(width: 6),
                          Text(
                            'View Details',
                            style: TextStyle(
                              fontSize: 13,
                              fontWeight: FontWeight.bold,
                              color: Colors.white,
                            ),
                          ),
                        ],
                      ),
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
}

// ═══════════════════════════════════════════════════════════════════════════
// STATUS BADGE COMPONENT
// ═══════════════════════════════════════════════════════════════════════════

class _StatusBadge extends StatelessWidget {
  final String status;

  const _StatusBadge({required this.status});

  @override
  Widget build(BuildContext context) {
    Color bg;
    Color border;
    Color text;
    IconData icon;
    String label;

    switch (status.toLowerCase()) {
      case 'active':
        bg = const Color(0xFFE0F2FE);
        border = const Color(0xFFBAE6FD);
        text = const Color(0xFF0284C7);
        icon = LucideIcons.home;
        label = 'Active Stay';
        break;
      case 'completed':
        bg = TagaTowersDesignTokens.softMint;
        border = TagaTowersDesignTokens.mintHighlight;
        text = TagaTowersDesignTokens.deepEmerald;
        icon = LucideIcons.circleCheck;
        label = 'Completed';
        break;
      case 'cancelled':
        bg = const Color(0xFFFEE2E2);
        border = const Color(0xFFFECACA);
        text = const Color(0xFFDC2626);
        icon = LucideIcons.circleX;
        label = 'Cancelled';
        break;
      case 'upcoming':
      default:
        bg = const Color(0xFFFEF3C7);
        border = const Color(0xFFFCD34D);
        text = const Color(0xFFB45309);
        icon = LucideIcons.calendarClock;
        label = 'Upcoming';
        break;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 4),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: border, width: 1.0),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: text),
          const SizedBox(width: 4),
          Text(
            label,
            style: TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.bold,
              color: text,
              letterSpacing: 0.1,
            ),
          ),
        ],
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// FEATURE CHIP
// ═══════════════════════════════════════════════════════════════════════════

class _FeatureChip extends StatelessWidget {
  final IconData icon;
  final String label;

  const _FeatureChip({required this.icon, required this.label});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 4),
      decoration: BoxDecoration(
        color: TagaTowersDesignTokens.mutedBackground,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: TagaTowersDesignTokens.subtleBorder),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 13, color: TagaTowersDesignTokens.primaryGreen),
          const SizedBox(width: 5),
          Text(
            label,
            style: const TextStyle(
              fontSize: 11.5,
              fontWeight: FontWeight.w600,
              color: AppColors.textPrimary,
            ),
          ),
        ],
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// CANCELLATION CONFIRMATION BOTTOM SHEET
// ═══════════════════════════════════════════════════════════════════════════

class _CancellationConfirmDialog extends ConsumerStatefulWidget {
  final BookingResponse booking;
  final String bookerId;

  const _CancellationConfirmDialog({
    required this.booking,
    required this.bookerId,
  });

  @override
  ConsumerState<_CancellationConfirmDialog> createState() =>
      _CancellationConfirmDialogState();
}

class _CancellationConfirmDialogState
    extends ConsumerState<_CancellationConfirmDialog> {
  bool _isCancelling = false;
  String? _errorMessage;

  Future<void> _performCancellation() async {
    setState(() {
      _isCancelling = true;
      _errorMessage = null;
    });

    try {
      final repo = ref.read(tagaTowersRepositoryProvider);
      await repo.cancelBooking(widget.booking.id);

      // Invalidate providers to force a fresh fetch from server
      ref.invalidate(userBookingsProvider(widget.bookerId));
      ref.invalidate(pastBookingsProvider(null));

      if (mounted) {
        Navigator.of(context).pop();
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Row(
              children: [
                const Icon(LucideIcons.checkCircle2, color: Colors.white, size: 20),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    'Booking #${widget.booking.id} has been successfully cancelled.',
                    style: const TextStyle(fontWeight: FontWeight.w600),
                  ),
                ),
              ],
            ),
            backgroundColor: TagaTowersDesignTokens.primaryGreen,
            behavior: SnackBarBehavior.floating,
            duration: const Duration(seconds: 3),
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _isCancelling = false;
          _errorMessage = e.toString().replaceAll('Exception: ', '');
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.fromLTRB(
        20,
        16,
        20,
        MediaQuery.of(context).viewInsets.bottom + 24,
      ),
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // Drag handle
          Center(
            child: Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: Colors.grey.shade300,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
          ),
          const SizedBox(height: 18),

          // Icon & Header
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: const Color(0xFFFEE2E2),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: const Color(0xFFFECACA)),
                ),
                child: const Icon(
                  LucideIcons.alertTriangle,
                  size: 24,
                  color: Color(0xFFDC2626),
                ),
              ),
              const SizedBox(width: 14),
              const Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Cancel Booking?',
                      style: TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: AppColors.textPrimary,
                      ),
                    ),
                    SizedBox(height: 2),
                    Text(
                      'This action cannot be undone.',
                      style: TextStyle(
                        fontSize: 12.5,
                        color: AppColors.textSecondary,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),

          const SizedBox(height: 16),

          // Booking Summary Box
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: TagaTowersDesignTokens.mutedBackground,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: TagaTowersDesignTokens.subtleBorder),
            ),
            child: Column(
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text(
                      'Reference:',
                      style: TextStyle(fontSize: 12.5, color: AppColors.textSecondary),
                    ),
                    Text(
                      '#${widget.booking.id}',
                      style: const TextStyle(
                        fontSize: 12.5,
                        fontWeight: FontWeight.bold,
                        fontFamily: 'monospace',
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 6),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text(
                      'Room / Beds:',
                      style: TextStyle(fontSize: 12.5, color: AppColors.textSecondary),
                    ),
                    Text(
                      '${widget.booking.roomName} (${widget.booking.bedCount} Bed${widget.booking.bedCount > 1 ? 's' : ''})',
                      style: const TextStyle(
                        fontSize: 12.5,
                        fontWeight: FontWeight.bold,
                        color: AppColors.textPrimary,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 6),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text(
                      'Advance Paid:',
                      style: TextStyle(fontSize: 12.5, color: AppColors.textSecondary),
                    ),
                    Text(
                      '₹${widget.booking.advanceAmount}',
                      style: const TextStyle(
                        fontSize: 12.5,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF2E7D32),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),

          const SizedBox(height: 14),

          // Association Policy Disclaimer Banner
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: const Color(0xFFFFFBEB),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: const Color(0xFFFDE68A)),
            ),
            child: const Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Icon(LucideIcons.info, size: 16, color: Color(0xFFB45309)),
                SizedBox(width: 10),
                Expanded(
                  child: Text(
                    'Association Advance Policy: Please note that the advance deposit is non-refundable upon cancellation. Your bed allocation will immediately be made available for other members.',
                    style: TextStyle(
                      fontSize: 11.5,
                      height: 1.4,
                      color: Color(0xFF92400E),
                    ),
                  ),
                ),
              ],
            ),
          ),

          if (_errorMessage != null) ...[
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: const Color(0xFFFEF2F2),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: const Color(0xFFFECACA)),
              ),
              child: Text(
                _errorMessage!,
                style: const TextStyle(
                  fontSize: 12,
                  color: Color(0xFFDC2626),
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
          ],

          const SizedBox(height: 20),

          // Buttons
          Row(
            children: [
              Expanded(
                child: OutlinedButton(
                  onPressed: _isCancelling ? null : () => Navigator.of(context).pop(),
                  style: OutlinedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    side: const BorderSide(color: Color(0xFFCBD5E1)),
                  ),
                  child: const Text(
                    'Keep Booking',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: ElevatedButton(
                  onPressed: _isCancelling ? null : _performCancellation,
                  style: ElevatedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    backgroundColor: const Color(0xFFDC2626),
                    foregroundColor: Colors.white,
                    elevation: 0,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                  child: _isCancelling
                      ? const SizedBox(
                          height: 18,
                          width: 18,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: Colors.white,
                          ),
                        )
                      : const Text(
                          'Yes, Cancel',
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// EMPTY STATE COMPONENT
// ═══════════════════════════════════════════════════════════════════════════

class _EmptyBookingsView extends StatelessWidget {
  final String title;
  final String message;
  final String buttonLabel;
  final VoidCallback onButtonPressed;
  final IconData icon;

  const _EmptyBookingsView({
    required this.title,
    required this.message,
    required this.buttonLabel,
    required this.onButtonPressed,
    required this.icon,
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 80,
              height: 80,
              decoration: BoxDecoration(
                color: TagaTowersDesignTokens.softMint.withValues(alpha: 0.6),
                shape: BoxShape.circle,
                border: Border.all(
                  color: TagaTowersDesignTokens.mintHighlight.withValues(alpha: 0.8),
                  width: 1.5,
                ),
                boxShadow: TagaTowersDesignTokens.softShadow,
              ),
              child: Icon(
                icon,
                size: 38,
                color: TagaTowersDesignTokens.primaryGreen,
              ),
            ),
            const SizedBox(height: 20),
            Text(
              title,
              textAlign: TextAlign.center,
              style: const TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: TagaTowersDesignTokens.deepEmerald,
                letterSpacing: -0.2,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              message,
              textAlign: TextAlign.center,
              style: const TextStyle(
                fontSize: 13,
                height: 1.45,
                color: AppColors.textSecondary,
              ),
            ),
            const SizedBox(height: 24),
            BouncyButton(
              onTap: onButtonPressed,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                decoration: BoxDecoration(
                  color: TagaTowersDesignTokens.primaryGreen,
                  borderRadius: BorderRadius.circular(12),
                  boxShadow: TagaTowersDesignTokens.floatingGlow,
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(LucideIcons.bedSingle, size: 18, color: Colors.white),
                    const SizedBox(width: 8),
                    Text(
                      buttonLabel,
                      style: const TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
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
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// ERROR STATE COMPONENT
// ═══════════════════════════════════════════════════════════════════════════

class _ErrorBookingsView extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;

  const _ErrorBookingsView({
    required this.message,
    required this.onRetry,
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: const Color(0xFFFEF2F2),
                shape: BoxShape.circle,
                border: Border.all(color: const Color(0xFFFECACA)),
              ),
              child: const Icon(
                LucideIcons.alertCircle,
                size: 36,
                color: Color(0xFFDC2626),
              ),
            ),
            const SizedBox(height: 16),
            const Text(
              'Unable to Load Bookings',
              style: TextStyle(
                fontSize: 17,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              message.isNotEmpty ? message : 'Network request failed. Please check your connection.',
              textAlign: TextAlign.center,
              style: const TextStyle(
                fontSize: 13,
                color: AppColors.textSecondary,
              ),
            ),
            const SizedBox(height: 20),
            ElevatedButton.icon(
              onPressed: onRetry,
              icon: const Icon(LucideIcons.rotateCcw, size: 16),
              label: const Text('Try Again'),
              style: ElevatedButton.styleFrom(
                backgroundColor: TagaTowersDesignTokens.primaryGreen,
                foregroundColor: Colors.white,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// SHIMMER LOADING SKELETON
// ═══════════════════════════════════════════════════════════════════════════

class _BookingCardShimmer extends StatefulWidget {
  const _BookingCardShimmer();

  @override
  State<_BookingCardShimmer> createState() => _BookingCardShimmerState();
}

class _BookingCardShimmerState extends State<_BookingCardShimmer>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1400),
    )..repeat();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Widget _buildPulseBox({
    required double width,
    required double height,
    double borderRadius = 8,
  }) {
    return AnimatedBuilder(
      animation: _controller,
      builder: (context, child) {
        return Container(
          width: width,
          height: height,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(borderRadius),
            gradient: LinearGradient(
              begin: Alignment(-1.0 + 2.0 * _controller.value, -0.3),
              end: Alignment(1.0 + 2.0 * _controller.value, 0.3),
              colors: const [
                Color(0xFFEAEFEA),
                Color(0xFFF7FAF7),
                Color(0xFFEAEFEA),
              ],
            ),
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: TagaTowersDesignTokens.cardDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              _buildPulseBox(width: 110, height: 22, borderRadius: 6),
              _buildPulseBox(width: 80, height: 22, borderRadius: 999),
            ],
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              _buildPulseBox(width: 44, height: 44, borderRadius: 12),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _buildPulseBox(width: 140, height: 16),
                    const SizedBox(height: 6),
                    _buildPulseBox(width: 90, height: 12),
                  ],
                ),
              ),
              _buildPulseBox(width: 65, height: 32, borderRadius: 10),
            ],
          ),
          const SizedBox(height: 16),
          _buildPulseBox(width: double.infinity, height: 48, borderRadius: 12),
          const SizedBox(height: 14),
          Row(
            children: [
              _buildPulseBox(width: 80, height: 24, borderRadius: 8),
              const SizedBox(width: 8),
              _buildPulseBox(width: 110, height: 24, borderRadius: 8),
            ],
          ),
          const SizedBox(height: 16),
          _buildPulseBox(width: double.infinity, height: 40, borderRadius: 10),
        ],
      ),
    );
  }
}
