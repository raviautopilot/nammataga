import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../domain/room_model.dart';
import '../providers/tagatowers_provider.dart';
import 'theme/tagatowers_design_tokens.dart';
import 'widgets/date_range_picker_card.dart';

/// Room Availability Screen for TAGA Towers.
///
/// Implements:
/// 1. All Visible Room Types Together (`apex-1`, `pavalam`, `malligai`, `kavery`, `vasantham`, `gents-dorm`, `ladies-dorm`).
/// 2. Prominent Available-Bed counts (`Available beds: 8 / 12`), Capacity (`Capacity: 12 persons`), and Gender restrictions.
/// 3. Advance rates display (₹100/bed for dorms, ₹200/bed for suites/rooms).
/// 4. Unavailable rooms remain visible with clear "Full / Unavailable" badge and disabled CTA.
/// 5. Enabled "Book Room" button updating `bookingFlowNotifierProvider` and routing to `/taga-towers/booking/bed-selection`.
/// 6. Interactive Stay Dates header bar displaying duration (`2 Days (48 Hours)`), quick modifier chips, and date picker.
/// 7. Shimmer loading skeleton, pull-to-refresh, and retryable error state.
class RoomAvailabilityScreen extends ConsumerStatefulWidget {
  const RoomAvailabilityScreen({super.key});

  @override
  ConsumerState<RoomAvailabilityScreen> createState() =>
      _RoomAvailabilityScreenState();
}

class _RoomAvailabilityScreenState
    extends ConsumerState<RoomAvailabilityScreen> {
  int _selectedFilterIndex = 0; // 0: All, 1: Suites & AC, 2: Dormitories

  /// Fallback static list guaranteeing all 7 visible room types are present
  static const List<Room> _defaultVisibleRooms = [
    Room(
      id: 'apex-1',
      name: 'Apex Suite A/C',
      type: 'apex-suite',
      capacity: 3,
      allowSingleBed: true,
    ),
    Room(
      id: 'pavalam',
      name: 'Pavalam',
      type: 'ac-room',
      capacity: 2,
      allowSingleBed: true,
    ),
    Room(
      id: 'malligai',
      name: 'Malligai',
      type: 'ac-room',
      capacity: 2,
      allowSingleBed: true,
    ),
    Room(
      id: 'kavery',
      name: 'Kavery',
      type: 'ac-room',
      capacity: 2,
      allowSingleBed: true,
    ),
    Room(
      id: 'vasantham',
      name: 'Vasantham',
      type: 'ac-room',
      capacity: 2,
      allowSingleBed: true,
    ),
    Room(
      id: 'gents-dorm',
      name: 'Gents Dormitory',
      type: 'gents-dorm',
      capacity: 12,
      allowSingleBed: true,
    ),
    Room(
      id: 'ladies-dorm',
      name: 'Ladies Dormitory',
      type: 'ladies-dorm',
      capacity: 8,
      allowSingleBed: true,
    ),
  ];

  void _onBookRoom(Room room, RoomAvailability? avail) {
    HapticFeedback.mediumImpact();
    ref.read(bookingFlowNotifierProvider.notifier).selectRoom(room, avail);
    context.push('/taga-towers/booking/bed-selection');
  }

  void _showRoomUnavailableToast(Room room) {
    HapticFeedback.lightImpact();
    ScaffoldMessenger.of(context).hideCurrentSnackBar();
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(LucideIcons.alertCircle, color: Colors.white, size: 18),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                '${room.name} is fully booked for these dates. Please try another room or change dates.',
                style: const TextStyle(fontSize: 13),
              ),
            ),
          ],
        ),
        backgroundColor: AppColors.destructive,
        duration: const Duration(seconds: 3),
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
      ),
    );
  }

  void _openDatePickerSheet(
    BuildContext context,
    DateTime currentIn,
    DateTime currentOut,
  ) {
    HapticFeedback.lightImpact();
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (modalContext) {
        return Padding(
          padding: EdgeInsets.only(
            bottom: MediaQuery.of(modalContext).viewInsets.bottom,
          ),
          child: Container(
            decoration: const BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
            ),
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  width: 44,
                  height: 4,
                  decoration: BoxDecoration(
                    color: Colors.grey.shade300,
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
                const SizedBox(height: 16),
                DateRangePickerCard(
                  checkInDate: currentIn,
                  checkOutDate: currentOut,
                  onCheckInChanged: (newIn) {
                    ref
                        .read(bookingFlowNotifierProvider.notifier)
                        .setDates(newIn, currentOut);
                  },
                  onCheckOutChanged: (newOut) {
                    ref
                        .read(bookingFlowNotifierProvider.notifier)
                        .setDates(currentIn, newOut);
                  },
                ),
                const SizedBox(height: 16),
                BouncyButton(
                  onTap: () => Navigator.of(modalContext).pop(),
                  child: Container(
                    width: double.infinity,
                    height: 48,
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        colors: AppColors.brandGradient,
                      ),
                      borderRadius: BorderRadius.circular(14),
                      boxShadow: TagaTowersDesignTokens.softShadow,
                    ),
                    alignment: Alignment.center,
                    child: const Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(LucideIcons.check, color: Colors.white, size: 18),
                        SizedBox(width: 8),
                        Text(
                          'Apply Dates & Refresh',
                          style: TextStyle(
                            color: Colors.white,
                            fontSize: 15,
                            fontWeight: FontWeight.bold,
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
      },
    );
  }

  void _setQuickStayDuration(int days) {
    HapticFeedback.selectionClick();
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final bookingFlow = ref.read(bookingFlowNotifierProvider);
    final currentIn = bookingFlow.checkInDate ?? today;
    final checkIn = currentIn.isBefore(today) ? today : currentIn;
    final checkOut = checkIn.add(Duration(days: days));
    ref.read(bookingFlowNotifierProvider.notifier).setDates(checkIn, checkOut);
  }

  void _showPolicyDialog() {
    HapticFeedback.lightImpact();
    showDialog<void>(
      context: context,
      builder: (context) {
        return AlertDialog(
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
          title: const Row(
            children: [
              Icon(LucideIcons.info, color: AppColors.primary, size: 22),
              SizedBox(width: 8),
              Text(
                'Stay Policies',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
            ],
          ),
          content: const SingleChildScrollView(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  '24-Hour Stay Schedule',
                  style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
                ),
                SizedBox(height: 4),
                Text(
                  'Check-in is at 12:00 AM Midnight. Checkout is at 12:00 AM Midnight the following day. Each night booked equals a 24-hour cycle.',
                  style: TextStyle(fontSize: 12.5, color: AppColors.textSecondary),
                ),
                SizedBox(height: 12),
                Text(
                  'Non-Refundable Advance Fee',
                  style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
                ),
                SizedBox(height: 4),
                Text(
                  '• Dormitories: ₹100 / bed advance\n• Suites & A/C Rooms: ₹200 / bed advance\nAdvances are non-refundable upon cancellation.',
                  style: TextStyle(fontSize: 12.5, color: AppColors.textSecondary),
                ),
                SizedBox(height: 12),
                Text(
                  'Gender Restrictions',
                  style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
                ),
                SizedBox(height: 4),
                Text(
                  '• Gents Dormitory is strictly for male members/guests.\n• Ladies Dormitory is strictly for female members/guests.\n• When private rooms have active bed bookings, remaining beds are restricted to guests of the same gender.',
                  style: TextStyle(fontSize: 12.5, color: AppColors.textSecondary),
                ),
                SizedBox(height: 12),
                Text(
                  'Helpdesk Support',
                  style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
                ),
                SizedBox(height: 4),
                Text(
                  'Reception: ${TagaTowersDesignTokens.helpdeskPhone}\nEmail: ${TagaTowersDesignTokens.helpdeskEmail}',
                  style: TextStyle(fontSize: 12.5, color: AppColors.textSecondary),
                ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text(
                'Got It',
                style: TextStyle(
                  color: AppColors.primary,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ],
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final bookingState = ref.watch(bookingFlowNotifierProvider);
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final checkIn = bookingState.checkInDate ?? today;
    final checkOut =
        bookingState.checkOutDate ?? checkIn.add(const Duration(days: 1));
    final params = DateRangeParams(checkIn: checkIn, checkOut: checkOut);

    final roomsAsync = ref.watch(towersRoomsProvider);
    final availabilityAsync =
        ref.watch(towersAvailabilityRangeProvider(params));

    final diffDays = checkOut.difference(checkIn).inDays;
    final nights = diffDays > 0 ? diffDays : 1;
    final totalHours = nights * 24;

    final dateFormat = DateFormat('MMM d, yyyy');
    final formattedDateRange =
        '${dateFormat.format(checkIn)} to ${dateFormat.format(checkOut)} • $nights Day${nights > 1 ? 's' : ''} ($totalHours Hours)';

    return Scaffold(
      backgroundColor: TagaTowersDesignTokens.surfaceLight,
      appBar: AppBar(
        title: const Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Room Availability',
              style: TextStyle(
                fontWeight: FontWeight.bold,
                fontSize: 18,
                color: TagaTowersDesignTokens.deepEmerald,
              ),
            ),
            Text(
              'All Rooms & Available Beds',
              style: TextStyle(
                fontSize: 11,
                color: AppColors.textSecondary,
                fontWeight: FontWeight.normal,
              ),
            ),
          ],
        ),
        backgroundColor: Colors.white,
        elevation: 0,
        surfaceTintColor: Colors.transparent,
        leading: IconButton(
          icon: const Icon(
            LucideIcons.arrowLeft,
            color: TagaTowersDesignTokens.deepEmerald,
          ),
          onPressed: () {
            if (context.canPop()) {
              context.pop();
            } else {
              context.go('/home');
            }
          },
        ),
        actions: [
          IconButton(
            icon: const Icon(
              LucideIcons.info,
              color: TagaTowersDesignTokens.deepEmerald,
            ),
            tooltip: 'Stay Policies & Rates',
            onPressed: _showPolicyDialog,
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
      body: RefreshIndicator(
        color: TagaTowersDesignTokens.primaryGreen,
        backgroundColor: Colors.white,
        onRefresh: () async {
          HapticFeedback.lightImpact();
          ref.invalidate(towersRoomsProvider);
          ref.invalidate(towersAvailabilityRangeProvider(params));
        },
        child: CustomScrollView(
          physics: const AlwaysScrollableScrollPhysics(
            parent: BouncingScrollPhysics(),
          ),
          slivers: [
            // ── Section 1: Interactive Stay Dates Header Bar ──────────────────
            SliverToBoxAdapter(
              child: Container(
                margin: const EdgeInsets.fromLTRB(16, 14, 16, 10),
                padding: const EdgeInsets.all(16),
                decoration: TagaTowersDesignTokens.cardDecoration(
                  color: Colors.white,
                  borderRadius: 18,
                  borderColor: TagaTowersDesignTokens.subtleBorder,
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Header title & Change Dates button
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Row(
                          children: [
                            Container(
                              padding: const EdgeInsets.all(7),
                              decoration: BoxDecoration(
                                color: TagaTowersDesignTokens.softMint,
                                borderRadius: BorderRadius.circular(8),
                              ),
                              child: const Icon(
                                LucideIcons.calendarDays,
                                size: 16,
                                color: TagaTowersDesignTokens.deepEmerald,
                              ),
                            ),
                            const SizedBox(width: 8),
                            const Text(
                              'STAY SCHEDULE',
                              style: TextStyle(
                                fontSize: 11.5,
                                fontWeight: FontWeight.w800,
                                letterSpacing: 0.6,
                                color: AppColors.textPrimary,
                              ),
                            ),
                          ],
                        ),
                        BouncyButton(
                          onTap: () =>
                              _openDatePickerSheet(context, checkIn, checkOut),
                          child: Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 10,
                              vertical: 5,
                            ),
                            decoration: BoxDecoration(
                              color: TagaTowersDesignTokens.softMint,
                              borderRadius: BorderRadius.circular(999),
                              border: Border.all(
                                color: TagaTowersDesignTokens.mintHighlight,
                                width: 1,
                              ),
                            ),
                            child: const Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Icon(
                                  LucideIcons.calendar,
                                  size: 13,
                                  color: TagaTowersDesignTokens.deepEmerald,
                                ),
                                SizedBox(width: 4),
                                Text(
                                  'Change Dates',
                                  style: TextStyle(
                                    fontSize: 11.5,
                                    fontWeight: FontWeight.bold,
                                    color: TagaTowersDesignTokens.deepEmerald,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),

                    // Date range highlight text
                    Text(
                      formattedDateRange,
                      style: const TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                        color: TagaTowersDesignTokens.deepEmerald,
                        letterSpacing: -0.2,
                      ),
                    ),
                    const SizedBox(height: 4),
                    const Text(
                      '24-Hour Cycle • 12:00 AM Check-in to 12:00 AM Checkout',
                      style: TextStyle(
                        fontSize: 11.5,
                        color: AppColors.textSecondary,
                        fontWeight: FontWeight.w500,
                      ),
                    ),

                    const SizedBox(height: 14),

                    // Quick duration changer buttons row
                    Row(
                      children: [
                        _buildQuickDurationChip(
                          label: '1 Day (24h)',
                          isSelected: nights == 1,
                          onTap: () => _setQuickStayDuration(1),
                        ),
                        const SizedBox(width: 8),
                        _buildQuickDurationChip(
                          label: '2 Days (48h)',
                          isSelected: nights == 2,
                          onTap: () => _setQuickStayDuration(2),
                        ),
                        const SizedBox(width: 8),
                        _buildQuickDurationChip(
                          label: '3 Days (72h)',
                          isSelected: nights == 3,
                          onTap: () => _setQuickStayDuration(3),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),

            // ── Section 2: Filter Category Chips Row ─────────────────────────
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
                child: Row(
                  children: [
                    _buildFilterChip('All Rooms (7)', 0),
                    const SizedBox(width: 8),
                    _buildFilterChip('Private & Suites (5)', 1),
                    const SizedBox(width: 8),
                    _buildFilterChip('Dormitories (2)', 2),
                  ],
                ),
              ),
            ),

            const SliverToBoxAdapter(child: SizedBox(height: 8)),

            // ── Section 3: Room List with Live Availability ──────────────────
            _buildRoomListSliver(
              context: context,
              roomsAsync: roomsAsync,
              availabilityAsync: availabilityAsync,
              params: params,
            ),

            // Bottom Spacing
            const SliverToBoxAdapter(child: SizedBox(height: 32)),
          ],
        ),
      ),
    );
  }

  Widget _buildQuickDurationChip({
    required String label,
    required bool isSelected,
    required VoidCallback onTap,
  }) {
    return Expanded(
      child: BouncyButton(
        onTap: onTap,
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 7),
          decoration: BoxDecoration(
            color: isSelected
                ? TagaTowersDesignTokens.deepEmerald
                : const Color(0xFFF1F8E9),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: isSelected
                  ? TagaTowersDesignTokens.deepEmerald
                  : TagaTowersDesignTokens.subtleBorder,
              width: 1,
            ),
          ),
          alignment: Alignment.center,
          child: Text(
            label,
            style: TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.bold,
              color: isSelected ? Colors.white : AppColors.textPrimary,
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildFilterChip(String label, int index) {
    final isSelected = _selectedFilterIndex == index;
    return Expanded(
      child: BouncyButton(
        onTap: () {
          HapticFeedback.selectionClick();
          setState(() => _selectedFilterIndex = index);
        },
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 8),
          decoration: BoxDecoration(
            color: isSelected
                ? TagaTowersDesignTokens.primaryGreen
                : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: isSelected
                  ? TagaTowersDesignTokens.primaryGreen
                  : TagaTowersDesignTokens.subtleBorder,
              width: 1.1,
            ),
            boxShadow: isSelected ? TagaTowersDesignTokens.softShadow : null,
          ),
          alignment: Alignment.center,
          child: Text(
            label,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: TextStyle(
              fontSize: 11.5,
              fontWeight: isSelected ? FontWeight.bold : FontWeight.w600,
              color: isSelected ? Colors.white : AppColors.textSecondary,
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildRoomListSliver({
    required BuildContext context,
    required AsyncValue<List<Room>> roomsAsync,
    required AsyncValue<Map<String, RoomAvailability>> availabilityAsync,
    required DateRangeParams params,
  }) {
    // 1. Resolve raw room list
    final List<Room> rooms =
        roomsAsync.value != null && roomsAsync.value!.isNotEmpty
            ? roomsAsync.value!
            : _defaultVisibleRooms;

    // 2. Apply category filtering
    final filteredRooms = rooms.where((room) {
      final isDorm = room.type == 'gents-dorm' ||
          room.type == 'ladies-dorm' ||
          room.id.contains('dorm');
      if (_selectedFilterIndex == 1) return !isDorm;
      if (_selectedFilterIndex == 2) return isDorm;
      return true; // All
    }).toList();

    // 3. Shimmer loading state when availability data is being fetched
    if (availabilityAsync.isLoading) {
      return SliverPadding(
        padding: const EdgeInsets.symmetric(horizontal: 16),
        sliver: SliverList(
          delegate: SliverChildBuilderDelegate(
            (context, index) => const Padding(
              padding: EdgeInsets.only(bottom: 14),
              child: _RoomCardShimmer(),
            ),
            childCount: filteredRooms.length.clamp(3, 7),
          ),
        ),
      );
    }

    // 4. Error state with retry action
    if (availabilityAsync.hasError) {
      return SliverToBoxAdapter(
        child: Container(
          margin: const EdgeInsets.all(16),
          padding: const EdgeInsets.all(20),
          decoration: TagaTowersDesignTokens.cardDecoration(
            color: Colors.white,
            borderRadius: 16,
            borderColor: AppColors.destructive.withValues(alpha: 0.3),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(
                LucideIcons.alertTriangle,
                size: 36,
                color: AppColors.destructive,
              ),
              const SizedBox(height: 12),
              const Text(
                'Unable to Load Room Availability',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
              const SizedBox(height: 6),
              Text(
                '${availabilityAsync.error}',
                textAlign: TextAlign.center,
                style: const TextStyle(
                  fontSize: 12,
                  color: AppColors.textSecondary,
                ),
              ),
              const SizedBox(height: 16),
              BouncyButton(
                onTap: () {
                  HapticFeedback.lightImpact();
                  ref.invalidate(towersAvailabilityRangeProvider(params));
                },
                child: Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 20,
                    vertical: 10,
                  ),
                  decoration: BoxDecoration(
                    color: TagaTowersDesignTokens.primaryGreen,
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: const Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(LucideIcons.refreshCw, color: Colors.white, size: 14),
                      SizedBox(width: 8),
                      Text(
                        'Retry Availability Check',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 13,
                          fontWeight: FontWeight.bold,
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

    final availabilityMap = availabilityAsync.value ?? {};

    // 5. Render room cards
    return SliverPadding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      sliver: SliverList(
        delegate: SliverChildBuilderDelegate(
          (context, index) {
            final room = filteredRooms[index];
            final avail = availabilityMap[room.id];

            // If avail is not null, evaluate availability & available beds
            final bool isAvailable = avail != null
                ? (avail.available && avail.availableBeds > 0)
                : true;
            final int availableBeds = avail?.availableBeds ?? room.capacity;
            final int totalBeds = room.capacity;
            final int advanceRate =
                TagaTowersDesignTokens.getAdvanceRatePerBed(room.type);
            final String? genderRestriction = avail?.genderRestriction;

            return Padding(
              padding: const EdgeInsets.only(bottom: 14),
              child: _buildRoomCard(
                context: context,
                room: room,
                avail: avail,
                isAvailable: isAvailable,
                availableBeds: availableBeds,
                totalBeds: totalBeds,
                advanceRate: advanceRate,
                genderRestriction: genderRestriction,
              ),
            );
          },
          childCount: filteredRooms.length,
        ),
      ),
    );
  }

  Widget _buildRoomCard({
    required BuildContext context,
    required Room room,
    required RoomAvailability? avail,
    required bool isAvailable,
    required int availableBeds,
    required int totalBeds,
    required int advanceRate,
    required String? genderRestriction,
  }) {
    final isGentsDorm = room.id == 'gents-dorm' || room.type == 'gents-dorm';
    final isLadiesDorm = room.id == 'ladies-dorm' || room.type == 'ladies-dorm';

    // Bed availability progress ratio (0.0 to 1.0)
    final double bedRatio = totalBeds > 0
        ? (availableBeds / totalBeds).clamp(0.0, 1.0)
        : 0.0;

    // Color coordination for bed availability meter
    final Color progressColor = isAvailable
        ? (bedRatio > 0.4
            ? TagaTowersDesignTokens.primaryGreen
            : const Color(0xFFE76F51))
        : const Color(0xFFD4183D);

    return Container(
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 18,
        borderColor: isAvailable
            ? TagaTowersDesignTokens.subtleBorder
            : const Color(0x33D4183D),
        boxShadow: TagaTowersDesignTokens.softShadow,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // ── Header: Room Type Tag & Status Badge ────────────────────────
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Room Title
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        room.name,
                        style: const TextStyle(
                          fontSize: 22,
                          fontWeight: FontWeight.w900,
                          color: TagaTowersDesignTokens.deepEmerald,
                          letterSpacing: -0.4,
                        ),
                      ),
                      const SizedBox(height: 3),
                      Row(
                        children: [
                          const Icon(
                            LucideIcons.users,
                            size: 13,
                            color: AppColors.textSecondary,
                          ),
                          const SizedBox(width: 4),
                          Text(
                            'Capacity: $totalBeds ${totalBeds == 1 ? 'person' : 'persons'}',
                            style: const TextStyle(
                              fontSize: 12,
                              color: AppColors.textSecondary,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),

                // Live Availability Status Pill
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 10,
                    vertical: 5,
                  ),
                  decoration: BoxDecoration(
                    color: isAvailable
                        ? (availableBeds < totalBeds
                            ? const Color(0xFFFFF3CD)
                            : TagaTowersDesignTokens.softMint)
                        : const Color(0xFFFFEBEE),
                    borderRadius: BorderRadius.circular(999),
                    border: Border.all(
                      color: isAvailable
                          ? (availableBeds < totalBeds
                              ? const Color(0xFFFFEEBA)
                              : TagaTowersDesignTokens.mintHighlight)
                          : const Color(0xFFFFCDD2),
                      width: 1,
                    ),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Container(
                        width: 7,
                        height: 7,
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          color: isAvailable
                              ? (availableBeds < totalBeds
                                  ? const Color(0xFFB45309)
                                  : TagaTowersDesignTokens.primaryGreen)
                              : const Color(0xFFD4183D),
                        ),
                      ),
                      const SizedBox(width: 5),
                      Text(
                        isAvailable
                            ? (availableBeds < totalBeds
                                ? '$availableBeds bed(s) left'
                                : 'Available')
                            : 'Unavailable / Full',
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.bold,
                          color: isAvailable
                              ? (availableBeds < totalBeds
                                  ? const Color(0xFFB45309)
                                  : TagaTowersDesignTokens.deepEmerald)
                              : const Color(0xFFD4183D),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),

          const Divider(
            height: 1,
            color: TagaTowersDesignTokens.subtleBorder,
          ),

          // ── Middle Section: Bed Availability & Gender Restriction ────────
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Prominent Available-Bed Count (CRITICAL REQUIREMENT)
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Row(
                      children: [
                        Icon(
                          LucideIcons.bedDouble,
                          size: 16,
                          color: isAvailable
                              ? TagaTowersDesignTokens.primaryGreen
                              : Colors.grey.shade500,
                        ),
                        const SizedBox(width: 6),
                        Text(
                          'Available beds: $availableBeds / $totalBeds',
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.bold,
                            color: isAvailable
                                ? TagaTowersDesignTokens.deepEmerald
                                : Colors.grey.shade700,
                          ),
                        ),
                      ],
                    ),
                    Text(
                      isAvailable
                          ? '$availableBeds bed(s) available'
                          : '0 beds left',
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                        color: isAvailable
                            ? TagaTowersDesignTokens.primaryGreen
                            : const Color(0xFFD4183D),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),

                // Visual Bed Capacity Indicator Bar (Meter)
                ClipRRect(
                  borderRadius: BorderRadius.circular(999),
                  child: Container(
                    height: 6,
                    width: double.infinity,
                    color: const Color(0xFFE8EFEA),
                    child: FractionallySizedBox(
                      alignment: Alignment.centerLeft,
                      widthFactor: bedRatio,
                      child: Container(
                        decoration: BoxDecoration(
                          color: progressColor,
                          borderRadius: BorderRadius.circular(999),
                        ),
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 12),

                // Pricing & Advance Rate Row
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    // Advance Rate per bed
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 5,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF1F8E9),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(
                          color: TagaTowersDesignTokens.subtleBorder,
                          width: 1,
                        ),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Icon(
                            LucideIcons.indianRupee,
                            size: 13,
                            color: TagaTowersDesignTokens.deepEmerald,
                          ),
                          const SizedBox(width: 3),
                          Text(
                            'Advance: ₹$advanceRate / bed',
                            style: const TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.bold,
                              color: TagaTowersDesignTokens.deepEmerald,
                            ),
                          ),
                        ],
                      ),
                    ),

                    // Single bed vs full room note
                    Text(
                      room.allowSingleBed
                          ? 'Per-bed booking allowed'
                          : 'Full room booking',
                      style: const TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.w500,
                        color: AppColors.textSecondary,
                      ),
                    ),
                  ],
                ),

                // Gender Restriction Banners / Badges
                if (isGentsDorm) ...[
                  const SizedBox(height: 10),
                  _buildGenderRestrictionPill(
                    icon: LucideIcons.user,
                    label: 'Gents Only — Strictly Male Guests',
                    backgroundColor: const Color(0xFFE3F2FD),
                    borderColor: const Color(0xFF90CAF9),
                    textColor: const Color(0xFF1565C0),
                  ),
                ] else if (isLadiesDorm) ...[
                  const SizedBox(height: 10),
                  _buildGenderRestrictionPill(
                    icon: LucideIcons.userCheck,
                    label: 'Ladies Only — Strictly Female Guests',
                    backgroundColor: const Color(0xFFFCE4EC),
                    borderColor: const Color(0xFFF48FB1),
                    textColor: const Color(0xFFAD1457),
                  ),
                ] else if (genderRestriction != null &&
                    genderRestriction.trim().isNotEmpty &&
                    genderRestriction.toLowerCase() != 'none') ...[
                  const SizedBox(height: 10),
                  _buildGenderRestrictionPill(
                    icon: LucideIcons.alertCircle,
                    label:
                        'Partially booked — $genderRestriction guests only',
                    backgroundColor: const Color(0xFFFFF8E1),
                    borderColor: const Color(0xFFFFE082),
                    textColor: const Color(0xFFE65100),
                  ),
                ],
              ],
            ),
          ),

          // ── Bottom Action Button (Book Room vs Room Full) ────────────────
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
            child: isAvailable
                ? BouncyButton(
                    onTap: () => _onBookRoom(room, avail),
                    child: Container(
                      width: double.infinity,
                      height: 44,
                      decoration: BoxDecoration(
                        gradient: const LinearGradient(
                          colors: AppColors.brandGradient,
                        ),
                        borderRadius: BorderRadius.circular(12),
                        boxShadow: TagaTowersDesignTokens.softShadow,
                      ),
                      child: const Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            LucideIcons.calendarCheck,
                            color: Colors.white,
                            size: 16,
                          ),
                          SizedBox(width: 8),
                          Text(
                            'Book Room',
                            style: TextStyle(
                              color: Colors.white,
                              fontSize: 14,
                              fontWeight: FontWeight.bold,
                              letterSpacing: 0.2,
                            ),
                          ),
                          SizedBox(width: 6),
                          Icon(
                            LucideIcons.arrowRight,
                            color: Colors.white,
                            size: 15,
                          ),
                        ],
                      ),
                    ),
                  )
                : GestureDetector(
                    onTap: () => _showRoomUnavailableToast(room),
                    behavior: HitTestBehavior.opaque,
                    child: Container(
                      width: double.infinity,
                      height: 44,
                      decoration: BoxDecoration(
                        color: Colors.grey.shade200,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: Colors.grey.shade300),
                      ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            LucideIcons.ban,
                            color: Colors.grey.shade600,
                            size: 16,
                          ),
                          const SizedBox(width: 8),
                          Text(
                            'Room Full / Unavailable',
                            style: TextStyle(
                              color: Colors.grey.shade600,
                              fontSize: 13.5,
                              fontWeight: FontWeight.bold,
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

  Widget _buildGenderRestrictionPill({
    required IconData icon,
    required String label,
    required Color backgroundColor,
    required Color borderColor,
    required Color textColor,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: borderColor, width: 1),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: textColor),
          const SizedBox(width: 6),
          Expanded(
            child: Text(
              label,
              style: TextStyle(
                fontSize: 11.5,
                fontWeight: FontWeight.w700,
                color: textColor,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// Shimmer card placeholder displayed while availability is loading
class _RoomCardShimmer extends StatefulWidget {
  const _RoomCardShimmer();

  @override
  State<_RoomCardShimmer> createState() => _RoomCardShimmerState();
}

class _RoomCardShimmerState extends State<_RoomCardShimmer>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1100),
    )..repeat(reverse: true);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _controller,
      builder: (context, child) {
        final opacity = 0.35 + (_controller.value * 0.45);
        return Opacity(opacity: opacity, child: child);
      },
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: TagaTowersDesignTokens.cardDecoration(
          borderColor: const Color(0xFFE2EBE5),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Container(
                  width: 140,
                  height: 18,
                  decoration: BoxDecoration(
                    color: Colors.grey.shade300,
                    borderRadius: BorderRadius.circular(6),
                  ),
                ),
                Container(
                  width: 80,
                  height: 22,
                  decoration: BoxDecoration(
                    color: Colors.grey.shade300,
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Container(
              width: 100,
              height: 14,
              decoration: BoxDecoration(
                color: Colors.grey.shade200,
                borderRadius: BorderRadius.circular(4),
              ),
            ),
            const SizedBox(height: 12),
            Container(
              width: double.infinity,
              height: 8,
              decoration: BoxDecoration(
                color: Colors.grey.shade200,
                borderRadius: BorderRadius.circular(4),
              ),
            ),
            const SizedBox(height: 14),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Container(
                  width: 110,
                  height: 16,
                  decoration: BoxDecoration(
                    color: Colors.grey.shade200,
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
                Container(
                  width: 80,
                  height: 16,
                  decoration: BoxDecoration(
                    color: Colors.grey.shade200,
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Container(
              width: double.infinity,
              height: 44,
              decoration: BoxDecoration(
                color: Colors.grey.shade300,
                borderRadius: BorderRadius.circular(12),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
