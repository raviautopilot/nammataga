import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../auth/providers/auth_provider.dart';
import '../../profile/domain/member_profile.dart';
import '../../profile/providers/profile_provider.dart';
import '../domain/booking_model.dart';
import '../domain/room_model.dart';
import '../providers/booking_flow_provider.dart';
import 'theme/tagatowers_design_tokens.dart';
import 'widgets/booking_step_progress_bar.dart';

/// Step 1 of the TAGA Towers Booking Flow:
/// Booking Type & Bed Selection Screen.
///
/// Handles:
/// 1. Progress Header (Step 1 of 4)
/// 2. Selected Room & Stay Duration Summary
/// 3. Booking Type Selection:
///    - Self Booking: 1 bed for the authenticated member (multi-bed prohibited)
///    - Guest Booking: Multiple beds up to available capacity
/// 4. Bed Quantity Selector for guest booking (with live capacity constraint)
/// 5. Gender Selection for single-bed sharing (dorm-fixed or radio selector)
/// 6. Booker Phone Input with `+91 ` prefix and validation
/// 7. Navigation:
///    - Self: Autofills member profile into GuestDetail and routes to Step 3 (`/taga-towers/booking/review`)
///    - Guest: Routes to Step 2 (`/taga-towers/booking/guest-details`)
class BookingBedSelectionScreen extends ConsumerStatefulWidget {
  const BookingBedSelectionScreen({super.key});

  @override
  ConsumerState<BookingBedSelectionScreen> createState() =>
      _BookingBedSelectionScreenState();
}

class _BookingBedSelectionScreenState
    extends ConsumerState<BookingBedSelectionScreen> {
  late final TextEditingController _phoneController;
  final FocusNode _phoneFocusNode = FocusNode();

  String _bookingFor = 'self'; // 'self' | 'guest'
  int _bedCount = 1;
  String _selectedGender = 'male'; // 'male' | 'female'
  String? _phoneError;
  bool _initializedFromState = false;

  @override
  void initState() {
    super.initState();
    _phoneController = TextEditingController(text: '+91 ');
    _phoneController.addListener(_onPhoneChanged);
  }

  @override
  void dispose() {
    _phoneController.removeListener(_onPhoneChanged);
    _phoneController.dispose();
    _phoneFocusNode.dispose();
    super.dispose();
  }

  void _onPhoneChanged() {
    final text = _phoneController.text;
    // Keep prefix if user deletes everything
    if (!text.startsWith('+91')) {
      _phoneController.text = '+91 ';
      _phoneController.selection = TextSelection.fromPosition(
        const TextPosition(offset: 4),
      );
      return;
    }

    if (_phoneError != null) {
      setState(() {
        _phoneError = _validatePhone(text);
      });
    }
  }

  String? _validatePhone(String val) {
    final clean = val.replaceAll(RegExp(r'\s+'), '');
    if (clean.isEmpty || clean == '+91') {
      return 'Booker contact number is required';
    }
    // Web regex: ^(\+\d{1,4})?[0-9]{7,14}$
    final phoneRegex = RegExp(r'^(\+\d{1,4})?[0-9]{7,14}$');
    if (!phoneRegex.hasMatch(clean)) {
      return 'Invalid mobile number format (7–14 digits)';
    }
    return null;
  }

  bool get _isPhoneValid => _validatePhone(_phoneController.text) == null;

  void _syncInitialData(BookingFlowState flowState) {
    if (_initializedFromState) return;
    _initializedFromState = true;

    _bookingFor = flowState.bookingFor;
    _bedCount = flowState.bedCount > 0 ? flowState.bedCount : 1;
    _selectedGender = flowState.bookingGender.isNotEmpty
        ? flowState.bookingGender
        : 'male';

    // Booker phone pre-population priority:
    // 1. Existing flowState.bookerPhone (if not default '+91 ')
    // 2. Logged-in session user data
    // 3. Member profile data
    if (flowState.bookerPhone.trim() != '+91' &&
        flowState.bookerPhone.trim().isNotEmpty) {
      _formatAndSetPhone(flowState.bookerPhone);
    } else {
      final session = ref.read(authProvider).value;
      final rawMobile = session?.user?['mobile_number'] ??
          session?.user?['mobileNumber'] ??
          session?.user?['phone'];
      if (rawMobile != null && rawMobile.toString().isNotEmpty) {
        _formatAndSetPhone(rawMobile.toString());
      }
    }

    // Pre-populate gender from session or room restriction
    final avail = flowState.roomAvailability;
    final room = flowState.selectedRoom;
    if (avail?.genderRestriction != null &&
        avail!.genderRestriction!.isNotEmpty) {
      _selectedGender = avail.genderRestriction!.toLowerCase();
    } else if (room?.type == 'gents-dorm') {
      _selectedGender = 'male';
    } else if (room?.type == 'ladies-dorm') {
      _selectedGender = 'female';
    }
  }

  void _formatAndSetPhone(String phone) {
    String clean = phone.trim();
    if (clean.startsWith('+91')) {
      final digits = clean.substring(3).trim();
      _phoneController.text = '+91 $digits';
    } else if (clean.startsWith('91') && clean.length > 10) {
      final digits = clean.substring(2).trim();
      _phoneController.text = '+91 $digits';
    } else {
      _phoneController.text = '+91 $clean';
    }
  }

  int _calculateMemberAge(String dob) {
    if (dob.isEmpty) return 0;
    try {
      DateTime? parsed;
      if (dob.contains('-')) {
        final parts = dob.split('-');
        if (parts.length == 3) {
          if (parts[0].length == 4) {
            parsed = DateTime(
              int.parse(parts[0]),
              int.parse(parts[1]),
              int.parse(parts[2]),
            );
          } else {
            parsed = DateTime(
              int.parse(parts[2]),
              int.parse(parts[1]),
              int.parse(parts[0]),
            );
          }
        }
      } else if (dob.contains('/')) {
        final parts = dob.split('/');
        if (parts.length == 3) {
          if (parts[2].length == 4) {
            parsed = DateTime(
              int.parse(parts[2]),
              int.parse(parts[1]),
              int.parse(parts[0]),
            );
          } else {
            parsed = DateTime(
              int.parse(parts[0]),
              int.parse(parts[1]),
              int.parse(parts[2]),
            );
          }
        }
      }
      if (parsed != null) {
        final now = DateTime.now();
        var age = now.year - parsed.year;
        if (now.month < parsed.month ||
            (now.month == parsed.month && now.day < parsed.day)) {
          age--;
        }
        return (age > 0 && age < 120) ? age : 0;
      }
    } catch (_) {}
    return 0;
  }

  void _onContinuePressed({
    required Room room,
    required RoomAvailability? avail,
    required MemberProfile? profile,
  }) {
    // 1. Phone validation
    final phoneError = _validatePhone(_phoneController.text);
    if (phoneError != null) {
      setState(() {
        _phoneError = phoneError;
      });
      _phoneFocusNode.requestFocus();
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(phoneError),
          backgroundColor: AppColors.destructive,
          behavior: SnackBarBehavior.floating,
        ),
      );
      return;
    }

    final cleanPhone = _phoneController.text.trim();
    final notifier = ref.read(bookingFlowNotifierProvider.notifier);

    // 2. Compute effective gender
    String effectiveGender = _selectedGender;
    if (room.type == 'gents-dorm') {
      effectiveGender = 'male';
    } else if (room.type == 'ladies-dorm') {
      effectiveGender = 'female';
    } else if (avail?.genderRestriction != null &&
        avail!.genderRestriction!.isNotEmpty) {
      effectiveGender = avail.genderRestriction!.toLowerCase();
    }

    // 3. Save selections to Riverpod notifier
    notifier.setBookingFor(_bookingFor);
    notifier.setBedCount(_bookingFor == 'self' ? 1 : _bedCount);
    notifier.setBookerPhone(cleanPhone);
    notifier.setBookingGender(effectiveGender);

    HapticFeedback.mediumImpact();

    // 4. Branch based on Booking Type
    if (_bookingFor == 'self') {
      // Autofill member's details into GuestDetail
      final session = ref.read(authProvider).value;
      final memberName = (profile?.name.isNotEmpty == true)
          ? profile!.fullNameWithInitial
          : session?.displayName ?? 'Member';

      final dob = profile?.dateOfBirth ??
          session?.user?['date_of_birth']?.toString() ??
          session?.user?['dateOfBirth']?.toString() ??
          '';
      final age = _calculateMemberAge(dob);

      final guestGender = (profile?.gender.isNotEmpty == true)
          ? profile!.gender.toLowerCase()
          : effectiveGender;

      final selfGuest = GuestDetail(
        name: memberName,
        age: age,
        contact: cleanPhone,
        gender: (guestGender == 'male' || guestGender == 'female')
            ? guestGender
            : effectiveGender,
      );

      notifier.setGuestDetails([selfGuest]);

      // Direct route to Step 3 (Review Screen)
      context.push('/taga-towers/booking/review');
    } else {
      // Guest Booking: Route to Step 2 (Guest Details Screen)
      context.push('/taga-towers/booking/guest-details');
    }
  }

  @override
  Widget build(BuildContext context) {
    final flowState = ref.watch(bookingFlowNotifierProvider);
    _syncInitialData(flowState);

    final room = flowState.selectedRoom;
    final avail = flowState.roomAvailability;
    final checkIn = flowState.checkInDate ?? DateTime.now();
    final checkOut =
        flowState.checkOutDate ?? DateTime.now().add(const Duration(days: 1));
    final stayNights = flowState.totalNights;

    // Asynchronously watch MemberProfile for fallback autofill
    final profileAsync = ref.watch(memberProfileProvider);
    final profile = profileAsync.value;

    // Edge-case guard: If screen is accessed without selecting a room
    if (room == null) {
      return Scaffold(
        backgroundColor: TagaTowersDesignTokens.surfaceLight,
        appBar: AppBar(
          title: const Text(
            'Booking Setup',
            style: TextStyle(
              fontWeight: FontWeight.bold,
              color: TagaTowersDesignTokens.deepEmerald,
            ),
          ),
          backgroundColor: Colors.white,
          elevation: 0,
        ),
        body: Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(
                  LucideIcons.bedDouble,
                  size: 48,
                  color: AppColors.textSecondary,
                ),
                const SizedBox(height: 16),
                const Text(
                  'No Room Selected',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: TagaTowersDesignTokens.deepEmerald,
                  ),
                ),
                const SizedBox(height: 8),
                const Text(
                  'Please select a bookable room and stay dates before configuring beds.',
                  textAlign: TextAlign.center,
                  style: TextStyle(color: AppColors.textSecondary),
                ),
                const SizedBox(height: 20),
                BouncyButton(
                  onTap: () => context.go('/taga-towers'),
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 24,
                      vertical: 12,
                    ),
                    decoration: BoxDecoration(
                      color: TagaTowersDesignTokens.primaryGreen,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Text(
                      'Browse Available Rooms',
                      style: TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.bold,
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

    final advanceRatePerBed =
        TagaTowersDesignTokens.getAdvanceRatePerBed(room.type);
    final effectiveBeds = _bookingFor == 'self'
        ? 1
        : (room.allowSingleBed ? _bedCount : room.capacity);
    final totalAdvance = advanceRatePerBed * effectiveBeds;

    // Available beds limit for guest picker
    final maxAvailableBeds = (avail?.availableBeds ?? room.capacity).clamp(
      1,
      room.capacity,
    );

    return Scaffold(
      backgroundColor: TagaTowersDesignTokens.surfaceLight,
      appBar: AppBar(
        title: const Text(
          'Room & Bed Selection',
          style: TextStyle(
            fontWeight: FontWeight.bold,
            color: TagaTowersDesignTokens.deepEmerald,
            fontSize: 18,
          ),
        ),
        backgroundColor: Colors.white,
        elevation: 0,
        scrolledUnderElevation: 1,
        leading: IconButton(
          icon: const Icon(
            LucideIcons.arrowLeft,
            color: TagaTowersDesignTokens.deepEmerald,
          ),
          onPressed: () => context.pop(),
        ),
      ),
      body: SafeArea(
        child: Column(
          children: [
            Expanded(
              child: SingleChildScrollView(
                physics: const BouncingScrollPhysics(),
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // 1. Wizard Progress Bar Header (Step 1 of 4)
                    const BookingStepProgressBar(currentStep: 1),
                    const SizedBox(height: 16),

                    // 2. Selected Room & Dates Summary Card
                    _buildSelectedRoomSummaryCard(
                      room: room,
                      avail: avail,
                      checkIn: checkIn,
                      checkOut: checkOut,
                      stayNights: stayNights,
                    ),
                    const SizedBox(height: 20),

                    // 3. Booker Mobile Number Input Card
                    _buildBookerPhoneCard(),
                    const SizedBox(height: 20),

                    // 4. Booking Type Selection (Self vs Guest)
                    _buildBookingTypeSelectionCard(advanceRate: advanceRatePerBed),
                    const SizedBox(height: 20),

                    // 5. Bed Quantity Selector (Guest Booking only)
                    if (_bookingFor == 'guest' && room.allowSingleBed) ...[
                      _buildBedQuantitySelectorCard(
                        maxAvailableBeds: maxAvailableBeds,
                        totalRoomCapacity: room.capacity,
                        advanceRate: advanceRatePerBed,
                      ),
                      const SizedBox(height: 20),
                    ],

                    // 6. Gender Selection for Single-Bed Sharing
                    if (room.allowSingleBed) ...[
                      _buildGenderSelectionSection(
                        room: room,
                        avail: avail,
                        bookingFor: _bookingFor,
                      ),
                      const SizedBox(height: 20),
                    ],

                    // 7. Advance Payment & Cancellation Summary Card
                    _buildAdvanceBreakdownCard(
                      room: room,
                      effectiveBeds: effectiveBeds,
                      ratePerBed: advanceRatePerBed,
                      totalAdvance: totalAdvance,
                      stayNights: stayNights,
                    ),
                    const SizedBox(height: 16),
                  ],
                ),
              ),
            ),

            // 8. Sticky Bottom Action Bar
            _buildBottomActionBar(
              room: room,
              avail: avail,
              profile: profile,
              effectiveBeds: effectiveBeds,
              totalAdvance: totalAdvance,
            ),
          ],
        ),
      ),
    );
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // SECTION BUILDERS
  // ═══════════════════════════════════════════════════════════════════════════

  /// Section 2: Selected Room & Dates Summary
  Widget _buildSelectedRoomSummaryCard({
    required Room room,
    required RoomAvailability? avail,
    required DateTime checkIn,
    required DateTime checkOut,
    required int stayNights,
  }) {
    final dateFormat = DateFormat('EEE, dd MMM yyyy');
    final availableBeds = avail?.availableBeds ?? room.capacity;

    return Container(
      width: double.infinity,
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 16,
      ),
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: TagaTowersDesignTokens.softMint,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Icon(
                  LucideIcons.building,
                  size: 20,
                  color: TagaTowersDesignTokens.deepEmerald,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      room.name,
                      style: const TextStyle(
                        fontSize: 16.5,
                        fontWeight: FontWeight.bold,
                        color: TagaTowersDesignTokens.deepEmerald,
                        letterSpacing: -0.2,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      TagaTowersDesignTokens.getRoomDisplayName(room.type),
                      style: const TextStyle(
                        fontSize: 12.5,
                        color: AppColors.textSecondary,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 9,
                  vertical: 4.5,
                ),
                decoration: BoxDecoration(
                  color: availableBeds > 0
                      ? TagaTowersDesignTokens.softMint
                      : Colors.red.shade50,
                  borderRadius: BorderRadius.circular(999),
                  border: Border.all(
                    color: availableBeds > 0
                        ? TagaTowersDesignTokens.mintHighlight
                        : Colors.red.shade200,
                  ),
                ),
                child: Text(
                  '$availableBeds/${room.capacity} Beds Avail',
                  style: TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                    color: availableBeds > 0
                        ? TagaTowersDesignTokens.deepEmerald
                        : Colors.red.shade800,
                  ),
                ),
              ),
            ],
          ),
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 12),
            child: Divider(height: 1, color: Color(0xFFE8EFEA)),
          ),

          // Date & Stay Duration Grid
          Row(
            children: [
              Expanded(
                child: _buildStayDetailTile(
                  icon: LucideIcons.calendarCheck,
                  title: 'Check-in',
                  value: dateFormat.format(checkIn),
                  time: TagaTowersDesignTokens.checkInTime,
                ),
              ),
              Container(width: 1, height: 42, color: const Color(0xFFE8EFEA)),
              Expanded(
                child: Padding(
                  padding: const EdgeInsets.only(left: 12),
                  child: _buildStayDetailTile(
                    icon: LucideIcons.calendarX,
                    title: 'Check-out',
                    value: dateFormat.format(checkOut),
                    time: TagaTowersDesignTokens.checkOutTime,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),

          // Stay Duration Pill
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
            decoration: BoxDecoration(
              color: TagaTowersDesignTokens.mutedBackground,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Row(
              children: [
                const Icon(
                  LucideIcons.clock,
                  size: 13,
                  color: TagaTowersDesignTokens.primaryGreen,
                ),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    'Stay Duration: $stayNights Day(s) • ${stayNights * 24} Hours (Midnight-to-Midnight Cycle)',
                    style: const TextStyle(
                      fontSize: 11.5,
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

  Widget _buildStayDetailTile({
    required IconData icon,
    required String title,
    required String value,
    required String time,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(icon, size: 13, color: AppColors.textSecondary),
            const SizedBox(width: 4),
            Text(
              title,
              style: const TextStyle(
                fontSize: 11.5,
                fontWeight: FontWeight.w600,
                color: AppColors.textSecondary,
              ),
            ),
          ],
        ),
        const SizedBox(height: 3),
        Text(
          value,
          style: const TextStyle(
            fontSize: 12.5,
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
        ),
        Text(
          time,
          style: const TextStyle(fontSize: 10.5, color: Color(0xFF718E81)),
        ),
      ],
    );
  }

  /// Section 3: Booker Phone Input Card
  Widget _buildBookerPhoneCard() {
    return Container(
      width: double.infinity,
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
              const Text(
                'Booker Mobile Number',
                style: TextStyle(
                  fontSize: 14.5,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
              if (_isPhoneValid)
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 2.5,
                  ),
                  decoration: BoxDecoration(
                    color: TagaTowersDesignTokens.softMint,
                    borderRadius: BorderRadius.circular(999),
                  ),
                  child: const Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        LucideIcons.check,
                        size: 11,
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                      SizedBox(width: 4),
                      Text(
                        'Valid number',
                        style: TextStyle(
                          fontSize: 10.5,
                          fontWeight: FontWeight.bold,
                          color: TagaTowersDesignTokens.deepEmerald,
                        ),
                      ),
                    ],
                  ),
                ),
            ],
          ),
          const SizedBox(height: 4),
          const Text(
            'Confirmation SMS, gate check-in pass, and stay assistance will be sent to this mobile number.',
            style: TextStyle(
              fontSize: 11.5,
              color: AppColors.textSecondary,
              height: 1.3,
            ),
          ),
          const SizedBox(height: 12),

          // Input field with country code prefix
          TextField(
            controller: _phoneController,
            focusNode: _phoneFocusNode,
            keyboardType: TextInputType.phone,
            inputFormatters: [
              FilteringTextInputFormatter.allow(RegExp(r'[0-9+\s]')),
              LengthLimitingTextInputFormatter(18),
            ],
            style: const TextStyle(
              fontSize: 15,
              fontWeight: FontWeight.w600,
              color: AppColors.textPrimary,
              letterSpacing: 0.5,
            ),
            decoration: InputDecoration(
              filled: true,
              fillColor: TagaTowersDesignTokens.mutedBackground,
              prefixIcon: const Padding(
                padding: EdgeInsets.symmetric(horizontal: 12),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      LucideIcons.phone,
                      size: 18,
                      color: TagaTowersDesignTokens.primaryGreen,
                    ),
                    SizedBox(width: 8),
                    Text(
                      '+91',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                  ],
                ),
              ),
              hintText: '+91 98765 43210',
              hintStyle: TextStyle(
                color: AppColors.textSecondary.withValues(alpha: 0.5),
                fontSize: 14,
              ),
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 14,
              ),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
                borderSide: const BorderSide(
                  color: TagaTowersDesignTokens.subtleBorder,
                ),
              ),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
                borderSide: BorderSide(
                  color: _phoneError != null
                      ? AppColors.destructive
                      : (_isPhoneValid
                          ? TagaTowersDesignTokens.mintHighlight
                          : TagaTowersDesignTokens.subtleBorder),
                  width: 1.2,
                ),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
                borderSide: const BorderSide(
                  color: TagaTowersDesignTokens.primaryGreen,
                  width: 1.8,
                ),
              ),
            ),
          ),

          if (_phoneError != null) ...[
            const SizedBox(height: 6),
            Row(
              children: [
                const Icon(
                  LucideIcons.alertCircle,
                  size: 13,
                  color: AppColors.destructive,
                ),
                const SizedBox(width: 4),
                Text(
                  _phoneError!,
                  style: const TextStyle(
                    fontSize: 11.5,
                    color: AppColors.destructive,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  /// Section 4: Booking Type Selection (Self vs Guest)
  Widget _buildBookingTypeSelectionCard({required int advanceRate}) {
    return Container(
      width: double.infinity,
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 16,
      ),
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'Select Booking Type',
            style: TextStyle(
              fontSize: 14.5,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 3),
          const Text(
            'Choose whether this accommodation is for yourself or for visiting guests.',
            style: TextStyle(
              fontSize: 11.5,
              color: AppColors.textSecondary,
              height: 1.3,
            ),
          ),
          const SizedBox(height: 14),

          // Options Row
          Row(
            children: [
              // Option A: Self Booking
              Expanded(
                child: _buildBookingTypeOptionTile(
                  id: 'self',
                  title: 'Self Booking',
                  subtitle: 'Single bed for member',
                  rateSubtitle: 'Advance: ₹$advanceRate (1 Bed)',
                  icon: LucideIcons.user,
                  isSelected: _bookingFor == 'self',
                  onTap: () {
                    HapticFeedback.selectionClick();
                    setState(() {
                      _bookingFor = 'self';
                      _bedCount = 1; // Strict rule: Self is strictly 1 bed
                    });
                  },
                ),
              ),
              const SizedBox(width: 12),

              // Option B: Guest Booking
              Expanded(
                child: _buildBookingTypeOptionTile(
                  id: 'guest',
                  title: 'Guest Booking',
                  subtitle: 'For friends & family',
                  rateSubtitle: 'Advance: ₹$advanceRate / bed',
                  icon: LucideIcons.users,
                  isSelected: _bookingFor == 'guest',
                  onTap: () {
                    HapticFeedback.selectionClick();
                    setState(() {
                      _bookingFor = 'guest';
                    });
                  },
                ),
              ),
            ],
          ),

          if (_bookingFor == 'self') ...[
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
              decoration: BoxDecoration(
                color: TagaTowersDesignTokens.softMint.withValues(alpha: 0.6),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(
                  color: TagaTowersDesignTokens.mintHighlight.withValues(
                    alpha: 0.5,
                  ),
                ),
              ),
              child: const Row(
                children: [
                  Icon(
                    LucideIcons.shieldCheck,
                    size: 14,
                    color: TagaTowersDesignTokens.deepEmerald,
                  ),
                  SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'Self Booking is strictly limited to 1 bed. Your member profile details will be automatically attached.',
                      style: TextStyle(
                        fontSize: 11,
                        color: TagaTowersDesignTokens.deepEmerald,
                        fontWeight: FontWeight.w500,
                        height: 1.3,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildBookingTypeOptionTile({
    required String id,
    required String title,
    required String subtitle,
    required String rateSubtitle,
    required IconData icon,
    required bool isSelected,
    required VoidCallback onTap,
  }) {
    return BouncyButton(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        curve: Curves.easeInOut,
        padding: const EdgeInsets.all(13),
        decoration: BoxDecoration(
          color: isSelected
              ? TagaTowersDesignTokens.softMint.withValues(alpha: 0.7)
              : Colors.white,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(
            color: isSelected
                ? TagaTowersDesignTokens.primaryGreen
                : const Color(0xFFD6E3DC),
            width: isSelected ? 2.0 : 1.2,
          ),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: TagaTowersDesignTokens.primaryGreen.withValues(
                      alpha: 0.12,
                    ),
                    blurRadius: 10,
                    offset: const Offset(0, 3),
                  ),
                ]
              : null,
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Icon(
                  icon,
                  size: 20,
                  color: isSelected
                      ? TagaTowersDesignTokens.primaryGreen
                      : AppColors.textSecondary,
                ),
                Container(
                  width: 18,
                  height: 18,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    border: Border.all(
                      color: isSelected
                          ? TagaTowersDesignTokens.primaryGreen
                          : const Color(0xFF9EAF9B),
                      width: 2,
                    ),
                    color: isSelected
                        ? TagaTowersDesignTokens.primaryGreen
                        : Colors.transparent,
                  ),
                  child: isSelected
                      ? const Center(
                          child: Icon(
                            Icons.circle,
                            size: 6,
                            color: Colors.white,
                          ),
                        )
                      : null,
                ),
              ],
            ),
            const SizedBox(height: 10),
            Text(
              title,
              style: TextStyle(
                fontSize: 13.5,
                fontWeight: FontWeight.bold,
                color: isSelected
                    ? TagaTowersDesignTokens.deepEmerald
                    : AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: 2),
            Text(
              subtitle,
              style: const TextStyle(
                fontSize: 11,
                color: AppColors.textSecondary,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              rateSubtitle,
              style: TextStyle(
                fontSize: 10.5,
                fontWeight: FontWeight.w700,
                color: isSelected
                    ? TagaTowersDesignTokens.primaryGreen
                    : const Color(0xFF5D7A6B),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Section 5: Bed Quantity Selector (Guest Booking only)
  Widget _buildBedQuantitySelectorCard({
    required int maxAvailableBeds,
    required int totalRoomCapacity,
    required int advanceRate,
  }) {
    return Container(
      width: double.infinity,
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
              const Text(
                'Number of Beds',
                style: TextStyle(
                  fontSize: 14.5,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 8,
                  vertical: 3,
                ),
                decoration: BoxDecoration(
                  color: TagaTowersDesignTokens.mutedBackground,
                  borderRadius: BorderRadius.circular(999),
                ),
                child: Text(
                  'Max: $maxAvailableBeds Available',
                  style: const TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w600,
                    color: TagaTowersDesignTokens.deepEmerald,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 4),
          Text(
            'Select the exact bed count required. Up to $maxAvailableBeds beds can be reserved for this stay.',
            style: const TextStyle(
              fontSize: 11.5,
              color: AppColors.textSecondary,
              height: 1.3,
            ),
          ),
          const SizedBox(height: 16),

          // Stepper Counter Row
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: TagaTowersDesignTokens.mutedBackground,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: TagaTowersDesignTokens.subtleBorder),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      '$_bedCount Bed${_bedCount > 1 ? 's' : ''} Selected',
                      style: const TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.bold,
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      'Advance: ₹${advanceRate * _bedCount} (₹$advanceRate × $_bedCount)',
                      style: const TextStyle(
                        fontSize: 11.5,
                        color: AppColors.textSecondary,
                      ),
                    ),
                  ],
                ),

                // Decrement & Increment Stepper Controls
                Row(
                  children: [
                    // Decrement Button
                    BouncyButton(
                      onTap: _bedCount > 1
                          ? () {
                              HapticFeedback.selectionClick();
                              setState(() {
                                _bedCount--;
                              });
                            }
                          : null,
                      child: Container(
                        width: 36,
                        height: 36,
                        decoration: BoxDecoration(
                          color: _bedCount > 1
                              ? Colors.white
                              : Colors.white.withValues(alpha: 0.5),
                          shape: BoxShape.circle,
                          border: Border.all(
                            color: _bedCount > 1
                                ? TagaTowersDesignTokens.mintHighlight
                                : Colors.grey.shade300,
                          ),
                          boxShadow: _bedCount > 1
                              ? [
                                  BoxShadow(
                                    color: Colors.black.withValues(alpha: 0.05),
                                    blurRadius: 4,
                                    offset: const Offset(0, 1),
                                  ),
                                ]
                              : null,
                        ),
                        child: Icon(
                          LucideIcons.minus,
                          size: 16,
                          color: _bedCount > 1
                              ? TagaTowersDesignTokens.primaryGreen
                              : Colors.grey.shade400,
                        ),
                      ),
                    ),
                    const SizedBox(width: 14),

                    // Current count display
                    Text(
                      '$_bedCount',
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.w800,
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                    const SizedBox(width: 14),

                    // Increment Button
                    BouncyButton(
                      onTap: _bedCount < maxAvailableBeds
                          ? () {
                              HapticFeedback.selectionClick();
                              setState(() {
                                _bedCount++;
                              });
                            }
                          : null,
                      child: Container(
                        width: 36,
                        height: 36,
                        decoration: BoxDecoration(
                          color: _bedCount < maxAvailableBeds
                              ? TagaTowersDesignTokens.primaryGreen
                              : Colors.grey.shade300,
                          shape: BoxShape.circle,
                          boxShadow: _bedCount < maxAvailableBeds
                              ? [
                                  BoxShadow(
                                    color: TagaTowersDesignTokens.primaryGreen
                                        .withValues(alpha: 0.3),
                                    blurRadius: 6,
                                    offset: const Offset(0, 2),
                                  ),
                                ]
                              : null,
                        ),
                        child: Icon(
                          LucideIcons.plus,
                          size: 16,
                          color: _bedCount < maxAvailableBeds
                              ? Colors.white
                              : Colors.grey.shade500,
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),

          // Quick selection chips if max beds is small (e.g., up to 6)
          if (maxAvailableBeds > 1 && maxAvailableBeds <= 6) ...[
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: List.generate(maxAvailableBeds, (index) {
                final count = index + 1;
                final isCurrent = _bedCount == count;
                return BouncyButton(
                  onTap: () {
                    HapticFeedback.selectionClick();
                    setState(() {
                      _bedCount = count;
                    });
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 14,
                      vertical: 6,
                    ),
                    decoration: BoxDecoration(
                      color: isCurrent
                          ? TagaTowersDesignTokens.primaryGreen
                          : Colors.white,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(
                        color: isCurrent
                            ? TagaTowersDesignTokens.primaryGreen
                            : const Color(0xFFD6E3DC),
                      ),
                    ),
                    child: Text(
                      '$count Bed${count > 1 ? 's' : ''}',
                      style: TextStyle(
                        fontSize: 11.5,
                        fontWeight: FontWeight.bold,
                        color: isCurrent ? Colors.white : AppColors.textPrimary,
                      ),
                    ),
                  ),
                );
              }),
            ),
          ],
        ],
      ),
    );
  }

  /// Section 6: Gender Selection for Single-Bed Sharing
  Widget _buildGenderSelectionSection({
    required Room room,
    required RoomAvailability? avail,
    required String bookingFor,
  }) {
    final restriction = avail?.genderRestriction;

    // Case A: Dormitory Room with fixed gender
    if (room.type == 'gents-dorm') {
      return Container(
        width: double.infinity,
        decoration: TagaTowersDesignTokens.cardDecoration(
          color: Colors.blue.shade50.withValues(alpha: 0.7),
          borderColor: Colors.blue.shade200,
          borderRadius: 14,
        ),
        padding: const EdgeInsets.all(14),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: Colors.blue.shade100,
                borderRadius: BorderRadius.circular(10),
              ),
              child: Icon(
                LucideIcons.shieldCheck,
                size: 20,
                color: Colors.blue.shade800,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Gents Dormitory — Male Only',
                    style: TextStyle(
                      fontSize: 13.5,
                      fontWeight: FontWeight.bold,
                      color: Colors.blue.shade900,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    'Accommodation is strictly reserved for male guests. Single bed shared arrangement.',
                    style: TextStyle(
                      fontSize: 11,
                      color: Colors.blue.shade800,
                      height: 1.3,
                    ),
                  ),
                ],
              ),
            ),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
              decoration: BoxDecoration(
                color: Colors.blue.shade700,
                borderRadius: BorderRadius.circular(999),
              ),
              child: const Text(
                'MALE ONLY',
                style: TextStyle(
                  fontSize: 9.5,
                  fontWeight: FontWeight.w800,
                  color: Colors.white,
                ),
              ),
            ),
          ],
        ),
      );
    }

    if (room.type == 'ladies-dorm') {
      return Container(
        width: double.infinity,
        decoration: TagaTowersDesignTokens.cardDecoration(
          color: Colors.pink.shade50.withValues(alpha: 0.7),
          borderColor: Colors.pink.shade200,
          borderRadius: 14,
        ),
        padding: const EdgeInsets.all(14),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: Colors.pink.shade100,
                borderRadius: BorderRadius.circular(10),
              ),
              child: Icon(
                LucideIcons.shieldCheck,
                size: 20,
                color: Colors.pink.shade800,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Ladies Dormitory — Female Only',
                    style: TextStyle(
                      fontSize: 13.5,
                      fontWeight: FontWeight.bold,
                      color: Colors.pink.shade900,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    'Accommodation is strictly reserved for female guests. Single bed shared arrangement.',
                    style: TextStyle(
                      fontSize: 11,
                      color: Colors.pink.shade800,
                      height: 1.3,
                    ),
                  ),
                ],
              ),
            ),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
              decoration: BoxDecoration(
                color: Colors.pink.shade700,
                borderRadius: BorderRadius.circular(999),
              ),
              child: const Text(
                'FEMALE ONLY',
                style: TextStyle(
                  fontSize: 9.5,
                  fontWeight: FontWeight.w800,
                  color: Colors.white,
                ),
              ),
            ),
          ],
        ),
      );
    }

    // Case B: Partially occupied room with active gender restriction
    if (restriction != null && restriction.isNotEmpty) {
      final isMale = restriction.toLowerCase() == 'male';
      return Container(
        width: double.infinity,
        decoration: TagaTowersDesignTokens.alertDecoration(borderRadius: 14),
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  LucideIcons.alertTriangle,
                  size: 16,
                  color: TagaTowersDesignTokens.warningAmberDark,
                ),
                const SizedBox(width: 8),
                Text(
                  'Room Gender Restriction: ${restriction.toUpperCase()} ONLY',
                  style: const TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.bold,
                    color: TagaTowersDesignTokens.warningAmberDark,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 6),
            Text(
              'This room is currently partially occupied by ${isMale ? 'male' : 'female'} guests. '
              'To comply with building sharing policy, only ${isMale ? 'male' : 'female'} guests can book the remaining beds.',
              style: const TextStyle(
                fontSize: 11.5,
                color: TagaTowersDesignTokens.warningAmberDark,
                height: 1.35,
              ),
            ),
          ],
        ),
      );
    }

    // Case C: AC Room / Suite sharing for Self Booking
    if (bookingFor == 'self') {
      return Container(
        width: double.infinity,
        decoration: TagaTowersDesignTokens.cardDecoration(
          color: Colors.white,
          borderRadius: 16,
        ),
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Gender for Shared Room',
              style: TextStyle(
                fontSize: 14.5,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: 3),
            const Text(
              'TAGA Towers room sharing policy requires occupants of the same shared room to be of the same gender.',
              style: TextStyle(
                fontSize: 11.5,
                color: AppColors.textSecondary,
                height: 1.3,
              ),
            ),
            const SizedBox(height: 14),

            // Male / Female Selector Radios
            Row(
              children: [
                Expanded(
                  child: _buildGenderRadioOption(
                    title: 'Male Occupant',
                    genderValue: 'male',
                    icon: LucideIcons.userCheck,
                    isSelected: _selectedGender.toLowerCase() == 'male',
                    onTap: () {
                      HapticFeedback.selectionClick();
                      setState(() {
                        _selectedGender = 'male';
                      });
                    },
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: _buildGenderRadioOption(
                    title: 'Female Occupant',
                    genderValue: 'female',
                    icon: LucideIcons.userCheck,
                    isSelected: _selectedGender.toLowerCase() == 'female',
                    onTap: () {
                      HapticFeedback.selectionClick();
                      setState(() {
                        _selectedGender = 'female';
                      });
                    },
                  ),
                ),
              ],
            ),
          ],
        ),
      );
    }

    // Case D: Guest Booking Notice
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: TagaTowersDesignTokens.mutedBackground,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: TagaTowersDesignTokens.subtleBorder),
      ),
      child: const Row(
        children: [
          Icon(
            LucideIcons.info,
            size: 16,
            color: TagaTowersDesignTokens.primaryGreen,
          ),
          SizedBox(width: 8),
          Expanded(
            child: Text(
              'Guest genders will be captured individually for each guest on Step 2 (Guest Details).',
              style: TextStyle(
                fontSize: 11.5,
                color: TagaTowersDesignTokens.deepEmerald,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildGenderRadioOption({
    required String title,
    required String genderValue,
    required IconData icon,
    required bool isSelected,
    required VoidCallback onTap,
  }) {
    return BouncyButton(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
        decoration: BoxDecoration(
          color: isSelected ? TagaTowersDesignTokens.softMint : Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: isSelected
                ? TagaTowersDesignTokens.primaryGreen
                : const Color(0xFFD6E3DC),
            width: isSelected ? 1.8 : 1.2,
          ),
        ),
        child: Row(
          children: [
            Container(
              width: 18,
              height: 18,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                border: Border.all(
                  color: isSelected
                      ? TagaTowersDesignTokens.primaryGreen
                      : const Color(0xFF9EAF9B),
                  width: 2,
                ),
                color: isSelected
                    ? TagaTowersDesignTokens.primaryGreen
                    : Colors.transparent,
              ),
              child: isSelected
                  ? const Center(
                      child: Icon(
                        Icons.circle,
                        size: 6,
                        color: Colors.white,
                      ),
                    )
                  : null,
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                title,
                style: TextStyle(
                  fontSize: 12.5,
                  fontWeight: FontWeight.bold,
                  color: isSelected
                      ? TagaTowersDesignTokens.deepEmerald
                      : AppColors.textPrimary,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Section 7: Advance Payment Breakdown Card
  Widget _buildAdvanceBreakdownCard({
    required Room room,
    required int effectiveBeds,
    required int ratePerBed,
    required int totalAdvance,
    required int stayNights,
  }) {
    return Container(
      width: double.infinity,
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
              const Row(
                children: [
                  Icon(
                    LucideIcons.receipt,
                    size: 16,
                    color: TagaTowersDesignTokens.primaryGreen,
                  ),
                  SizedBox(width: 8),
                  Text(
                    'Advance Estimation',
                    style: TextStyle(
                      fontSize: 14.5,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
                  ),
                ],
              ),
              Text(
                '₹$ratePerBed / bed',
                style: const TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),

          // Computation Box
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
            decoration: BoxDecoration(
              color: TagaTowersDesignTokens.mutedBackground,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: TagaTowersDesignTokens.mintHighlight.withValues(
                  alpha: 0.8,
                ),
              ),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Advance Amount Payable Now',
                      style: TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.bold,
                        color: TagaTowersDesignTokens.deepEmerald,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      '₹$ratePerBed × $effectiveBeds Bed${effectiveBeds > 1 ? 's' : ''}',
                      style: const TextStyle(
                        fontSize: 11.5,
                        color: AppColors.textSecondary,
                      ),
                    ),
                  ],
                ),
                Text(
                  '₹$totalAdvance',
                  style: const TextStyle(
                    fontSize: 22,
                    fontWeight: FontWeight.w900,
                    color: TagaTowersDesignTokens.primaryGreen,
                    letterSpacing: -0.5,
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 10),

          // Remaining balance disclaimer
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
                  'Remaining balance is payable after stay upon checkout at TAGA Towers reception.',
                  style: TextStyle(
                    fontSize: 11.5,
                    color: AppColors.textSecondary,
                    height: 1.35,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),

          // Non-refundable cancellation policy notice
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(10),
            decoration: TagaTowersDesignTokens.alertDecoration(borderRadius: 10),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Padding(
                  padding: EdgeInsets.only(top: 1),
                  child: Icon(
                    LucideIcons.alertTriangle,
                    size: 14,
                    color: TagaTowersDesignTokens.warningAmberDark,
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    'Cancellation Policy: Advance booking payments are non-refundable upon cancellation.',
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                      color: TagaTowersDesignTokens.warningAmberDark,
                      height: 1.35,
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

  /// Section 8: Sticky Bottom Action Bar
  Widget _buildBottomActionBar({
    required Room room,
    required RoomAvailability? avail,
    required MemberProfile? profile,
    required int effectiveBeds,
    required int totalAdvance,
  }) {
    final isSelf = _bookingFor == 'self';

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        border: const Border(
          top: BorderSide(color: TagaTowersDesignTokens.subtleBorder),
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.04),
            blurRadius: 10,
            offset: const Offset(0, -3),
          ),
        ],
      ),
      child: Row(
        children: [
          // Total Advance Price Indicator
          Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Advance ($effectiveBeds Bed${effectiveBeds > 1 ? 's' : ''})',
                style: const TextStyle(
                  fontSize: 11.5,
                  color: AppColors.textSecondary,
                  fontWeight: FontWeight.w500,
                ),
              ),
              Text(
                '₹$totalAdvance',
                style: const TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.w900,
                  color: TagaTowersDesignTokens.primaryGreen,
                  letterSpacing: -0.4,
                ),
              ),
            ],
          ),
          const SizedBox(width: 16),

          // Continue CTA Button
          Expanded(
            child: BouncyButton(
              onTap: () => _onContinuePressed(
                room: room,
                avail: avail,
                profile: profile,
              ),
              child: Container(
                height: 48,
                decoration: BoxDecoration(
                  color: TagaTowersDesignTokens.primaryGreen,
                  borderRadius: BorderRadius.circular(12),
                  boxShadow: [
                    BoxShadow(
                      color: TagaTowersDesignTokens.primaryGreen.withValues(
                        alpha: 0.3,
                      ),
                      blurRadius: 8,
                      offset: const Offset(0, 3),
                    ),
                  ],
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      isSelf ? 'Review Booking' : 'Guest Details',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 14.5,
                        fontWeight: FontWeight.bold,
                        letterSpacing: 0.2,
                      ),
                    ),
                    const SizedBox(width: 8),
                    const Icon(
                      LucideIcons.arrowRight,
                      size: 16,
                      color: Colors.white,
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
