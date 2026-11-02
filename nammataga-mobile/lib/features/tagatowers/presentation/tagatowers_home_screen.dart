import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../auth/providers/auth_provider.dart';
import '../../profile/providers/profile_provider.dart';
import '../providers/booking_flow_provider.dart';
import 'tagatowers_access_restricted_screen.dart';
import 'theme/tagatowers_design_tokens.dart';
import 'widgets/date_range_picker_card.dart';
import 'widgets/tagatowers_hero_card.dart';

/// TAGA Towers Accommodation Home / Landing Screen (Agent 4)
///
/// Strictly enforces the date-first booking workflow:
/// 1. Subscription Guard: Restricts access to paid members only. Unpaid members
///    are shown [TagatowersAccessRestrictedScreen].
/// 2. Date-First Requirement: Before searching, members see ONLY the Hero Card,
///    Essential Stay Schedule, Contact & Caretakers information, Date Range selector,
///    and the Search CTA. Absolutely NO room cards are displayed on this screen!
/// 3. Booking Horizon & Stay Rules: Enforces max 10-day advance booking horizon
///    and 24-hour midnight-to-midnight cycle (12:00 AM check-in / check-out).
/// 4. Search Availability Action: Updates [bookingFlowNotifierProvider] with the
///    chosen dates and navigates directly to `/taga-towers/availability`.
class TagaTowersHomeScreen extends ConsumerStatefulWidget {
  const TagaTowersHomeScreen({super.key});

  @override
  ConsumerState<TagaTowersHomeScreen> createState() => _TagaTowersHomeScreenState();
}

class _TagaTowersHomeScreenState extends ConsumerState<TagaTowersHomeScreen> {
  late DateTime _checkInDate;
  late DateTime _checkOutDate;

  @override
  void initState() {
    super.initState();
    final today = _stripTime(DateTime.now());
    _checkInDate = today;
    _checkOutDate = today.add(const Duration(days: 1));

    // Synchronize dates from existing booking flow state if present and valid
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final existingState = ref.read(bookingFlowNotifierProvider);
      if (existingState.checkInDate != null && existingState.checkOutDate != null) {
        final existingCheckIn = _stripTime(existingState.checkInDate!);
        final existingCheckOut = _stripTime(existingState.checkOutDate!);
        final maxAllowed = today.add(
          const Duration(days: TagaTowersDesignTokens.maxBookingDays),
        );

        if (!existingCheckIn.isBefore(today) &&
            !existingCheckIn.isAfter(today.add(const Duration(days: 9))) &&
            existingCheckOut.isAfter(existingCheckIn) &&
            !existingCheckOut.isAfter(maxAllowed)) {
          setState(() {
            _checkInDate = existingCheckIn;
            _checkOutDate = existingCheckOut;
          });
        }
      }
    });
  }

  DateTime _stripTime(DateTime dt) => DateTime(dt.year, dt.month, dt.day);

  int get _stayNights {
    final diff = _checkOutDate.difference(_checkInDate).inDays;
    return diff > 0 ? diff : 1;
  }

  void _onSearchAvailability() {
    HapticFeedback.mediumImpact();
    // 1. Update Riverpod booking flow state
    ref.read(bookingFlowNotifierProvider.notifier).setDates(_checkInDate, _checkOutDate);

    // 2. Navigate to Room Availability Screen (Agent 5)
    context.push('/taga-towers/availability');
  }

  Future<void> _makePhoneCall(String phoneNumber) async {
    HapticFeedback.lightImpact();
    final cleanNumber = phoneNumber.replaceAll(RegExp(r'\s+'), '');
    final uri = Uri(scheme: 'tel', path: cleanNumber);
    try {
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri);
      } else {
        _copyToClipboard(cleanNumber, 'Phone number');
      }
    } catch (_) {
      _copyToClipboard(cleanNumber, 'Phone number');
    }
  }

  Future<void> _sendEmail(String email) async {
    HapticFeedback.lightImpact();
    final uri = Uri(scheme: 'mailto', path: email);
    try {
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri);
      } else {
        _copyToClipboard(email, 'Email address');
      }
    } catch (_) {
      _copyToClipboard(email, 'Email address');
    }
  }

  void _copyToClipboard(String text, String label) {
    Clipboard.setData(ClipboardData(text: text));
    HapticFeedback.lightImpact();
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(LucideIcons.checkCheck, color: Colors.white, size: 16),
            const SizedBox(width: 8),
            Text('$label copied to clipboard'),
          ],
        ),
        backgroundColor: TagaTowersDesignTokens.deepEmerald,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        duration: const Duration(seconds: 2),
      ),
    );
  }

  void _applyQuickPreset(int nights) {
    HapticFeedback.lightImpact();
    final today = _stripTime(DateTime.now());
    final absoluteMax = today.add(
      const Duration(days: TagaTowersDesignTokens.maxBookingDays),
    );

    var newCheckIn = _checkInDate;
    if (newCheckIn.isBefore(today)) {
      newCheckIn = today;
    }

    var newCheckOut = newCheckIn.add(Duration(days: nights));
    if (newCheckOut.isAfter(absoluteMax)) {
      newCheckOut = absoluteMax;
      newCheckIn = newCheckOut.subtract(Duration(days: nights));
      if (newCheckIn.isBefore(today)) {
        newCheckIn = today;
      }
    }

    setState(() {
      _checkInDate = newCheckIn;
      _checkOutDate = newCheckOut;
    });
  }

  @override
  Widget build(BuildContext context) {
    // ═════════════════════════════════════════════════════════════════════════
    // 1. SUBSCRIPTION VERIFICATION GUARD
    // ═════════════════════════════════════════════════════════════════════════
    final profileAsync = ref.watch(memberProfileProvider);
    final authSession = ref.watch(authProvider).value;

    final isPaid = authSession?.isPaid == true ||
        (profileAsync.hasValue &&
            (profileAsync.value!.isPaid ||
                profileAsync.value!.subscriptionActive));

    // If profile is still resolving and session has no confirmed paid state
    if (profileAsync.isLoading && authSession?.isPaid != true) {
      return Scaffold(
        backgroundColor: TagaTowersDesignTokens.surfaceLight,
        body: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              SizedBox(
                width: 32,
                height: 32,
                child: CircularProgressIndicator(
                  strokeWidth: 2.5,
                  valueColor: AlwaysStoppedAnimation<Color>(
                    TagaTowersDesignTokens.primaryGreen,
                  ),
                ),
              ),
              const SizedBox(height: 14),
              const Text(
                'Checking membership status...',
                style: TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.w500,
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ),
        ),
      );
    }

    // Access restricted for non-subscribers
    if (!isPaid) {
      return const TagatowersAccessRestrictedScreen();
    }

    // ═════════════════════════════════════════════════════════════════════════
    // 2. MAIN HOME / LANDING VIEW
    // ═════════════════════════════════════════════════════════════════════════
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAF7),
      appBar: PreferredSize(
        preferredSize: const Size.fromHeight(64),
        child: Container(
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [
                Color(0xFF061A0F),
                Color(0xFF0D281E),
                Color(0xFF1B4332),
              ],
            ),
            borderRadius: const BorderRadius.vertical(
              bottom: Radius.circular(20),
            ),
            boxShadow: [
              BoxShadow(
                color: const Color(0xFF061A0F).withValues(alpha: 0.28),
                blurRadius: 12,
                offset: const Offset(0, 4),
              ),
            ],
          ),
          child: SafeArea(
            bottom: false,
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
              child: Row(
                children: [
                  // Back to Dashboard
                  BouncyButton(
                    onTap: () {
                      HapticFeedback.lightImpact();
                      if (context.canPop()) {
                        context.pop();
                      } else {
                        context.go('/home');
                      }
                    },
                    child: Container(
                      width: 38,
                      height: 38,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: Colors.white.withValues(alpha: 0.12),
                        border: Border.all(
                          color: Colors.white.withValues(alpha: 0.2),
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

                  // Title & Subtitle
                  const Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(
                          'TAGA Towers',
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.w800,
                            color: Colors.white,
                            letterSpacing: -0.2,
                          ),
                        ),
                        Text(
                          'Accommodation & Guest House',
                          style: TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w400,
                            color: Color(0xFFD8F3DC),
                          ),
                        ),
                      ],
                    ),
                  ),

                  // "My Bookings" Direct Access Action
                  BouncyButton(
                    onTap: () {
                      HapticFeedback.lightImpact();
                      context.push('/taga-towers/my-bookings');
                    },
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 6),
                      decoration: BoxDecoration(
                        color: Colors.white.withValues(alpha: 0.14),
                        borderRadius: BorderRadius.circular(999),
                        border: Border.all(
                          color: TagaTowersDesignTokens.mintHighlight.withValues(alpha: 0.5),
                          width: 1,
                        ),
                      ),
                      child: const Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            LucideIcons.calendarCheck,
                            size: 14,
                            color: Color(0xFFD8F3DC),
                          ),
                          SizedBox(width: 5),
                          Text(
                            'My Bookings',
                            style: TextStyle(
                              fontSize: 11.5,
                              fontWeight: FontWeight.w700,
                              color: Colors.white,
                              letterSpacing: 0.1,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
      body: RefreshIndicator(
        color: TagaTowersDesignTokens.primaryGreen,
        backgroundColor: Colors.white,
        onRefresh: () async {
          HapticFeedback.lightImpact();
          ref.invalidate(memberProfileProvider);
          ref.invalidate(authProvider);
        },
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(
            parent: BouncingScrollPhysics(),
          ),
          child: Center(
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 580),
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    // 1. Building Image Hero Card (Full View of TAGA Tower)
                    const TagaTowersHeroCard(
                      showLocationPill: true,
                    ),
                    const SizedBox(height: 14),

                    // 2. Essential Stay Schedule & Horizon Notice Card
                    _buildStayScheduleNoticeCard(),
                    const SizedBox(height: 14),

                    // 3. Premium Date-Range Selection Section
                    DateRangePickerCard(
                      checkInDate: _checkInDate,
                      checkOutDate: _checkOutDate,
                      onCheckInChanged: (newCheckIn) {
                        setState(() {
                          _checkInDate = newCheckIn;
                        });
                      },
                      onCheckOutChanged: (newCheckOut) {
                        setState(() {
                          _checkOutDate = newCheckOut;
                        });
                      },
                    ),
                    const SizedBox(height: 10),

                    // Quick duration chips (1 Night, 2 Nights, 3 Nights)
                    _buildQuickDurationChips(),
                    const SizedBox(height: 16),

                    // 4. Prominent "Search Availability" CTA Button
                    BouncyButton(
                      onTap: _onSearchAvailability,
                      child: Container(
                        width: double.infinity,
                        height: 52,
                        decoration: BoxDecoration(
                          gradient: const LinearGradient(
                            begin: Alignment.topLeft,
                            end: Alignment.bottomRight,
                            colors: [
                              Color(0xFF2D6A4F),
                              Color(0xFF1B4332),
                            ],
                          ),
                          borderRadius: BorderRadius.circular(16),
                          boxShadow: TagaTowersDesignTokens.floatingGlow,
                        ),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Container(
                              padding: const EdgeInsets.all(6),
                              decoration: BoxDecoration(
                                color: Colors.white.withValues(alpha: 0.18),
                                shape: BoxShape.circle,
                              ),
                              child: const Icon(
                                LucideIcons.search,
                                size: 16,
                                color: Colors.white,
                              ),
                            ),
                            const SizedBox(width: 10),
                            const Text(
                              'Search Availability',
                              style: TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.w800,
                                color: Colors.white,
                                letterSpacing: 0.2,
                              ),
                            ),
                            const SizedBox(width: 8),
                            const Icon(
                              LucideIcons.arrowRight,
                              size: 16,
                              color: Color(0xFFD8F3DC),
                            ),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(height: 18),

                    // 5. Direct Access to "My Bookings" Banner Card
                    _buildMyBookingsBannerCard(context),
                    const SizedBox(height: 14),

                    // 6. Contact & Caretakers Information Card
                    _buildContactAndCaretakersCard(),
                    const SizedBox(height: 24),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // SUB-COMPONENTS
  // ═══════════════════════════════════════════════════════════════════════════

  /// Essential Booking Schedule & Rules Info Card
  Widget _buildStayScheduleNoticeCard() {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFFF1F8E9),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: const Color(0xFFB7E4C7).withValues(alpha: 0.9),
          width: 1.2,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(5),
                decoration: const BoxDecoration(
                  color: Color(0xFFD8F3DC),
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  LucideIcons.clock,
                  color: AppColors.primaryDark,
                  size: 15,
                ),
              ),
              const SizedBox(width: 8),
              const Text(
                'BOOKING SCHEDULE & POLICIES',
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 0.6,
                  color: AppColors.primaryDark,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          _buildScheduleRow(
            icon: LucideIcons.moon,
            title: '12:00 AM to 12:00 AM Stay Cycle',
            description:
                'For your comfort and convenience, TAGA Towers operates on a complete 24-hour cycle. Check-in commences at 12:00 AM Midnight, and check-out is at 12:00 AM Midnight the following day, offering you a full, uninterrupted stay experience.',
          ),
          const SizedBox(height: 8),
          _buildScheduleRow(
            icon: LucideIcons.calendarClock,
            title: '10-Day Booking Horizon',
            description: 'Reservations can be booked up to 10 days in advance of arrival.',
          ),
          const SizedBox(height: 8),
          _buildScheduleRow(
            icon: LucideIcons.shieldAlert,
            title: 'Non-Refundable Advance Fee',
            description:
                '₹100/bed for dorms, ₹200/bed for A/C suites. Advance is non-refundable upon cancellation.',
          ),
        ],
      ),
    );
  }

  Widget _buildScheduleRow({
    required IconData icon,
    required String title,
    required String description,
  }) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(top: 2),
          child: Icon(icon, size: 13, color: AppColors.primary),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Text.rich(
            TextSpan(
              style: const TextStyle(
                fontSize: 11.5,
                color: AppColors.textPrimary,
                height: 1.35,
              ),
              children: [
                TextSpan(
                  text: '$title: ',
                  style: const TextStyle(fontWeight: FontWeight.bold),
                ),
                TextSpan(
                  text: description,
                  style: const TextStyle(color: AppColors.textSecondary),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }

  /// Quick duration chips (1 Night, 2 Nights, 3 Nights)
  Widget _buildQuickDurationChips() {
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      physics: const BouncingScrollPhysics(),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Text(
            'Quick Stay:',
            style: TextStyle(
              fontSize: 11.5,
              fontWeight: FontWeight.w600,
              color: AppColors.textSecondary,
            ),
          ),
          const SizedBox(width: 8),
          _buildChip(label: '1 Night (24h)', nights: 1),
          const SizedBox(width: 6),
          _buildChip(label: '2 Nights (48h)', nights: 2),
          const SizedBox(width: 6),
          _buildChip(label: '3 Nights (72h)', nights: 3),
        ],
      ),
    );
  }

  Widget _buildChip({required String label, required int nights}) {
    final isSelected = _stayNights == nights;
    return BouncyButton(
      onTap: () => _applyQuickPreset(nights),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 180),
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
        decoration: BoxDecoration(
          color: isSelected
              ? TagaTowersDesignTokens.primaryGreen
              : Colors.white,
          borderRadius: BorderRadius.circular(999),
          border: Border.all(
            color: isSelected
                ? TagaTowersDesignTokens.primaryGreen
                : const Color(0xFFD0D7DE),
            width: 1,
          ),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: TagaTowersDesignTokens.primaryGreen.withValues(alpha: 0.25),
                    blurRadius: 6,
                    offset: const Offset(0, 2),
                  ),
                ]
              : null,
        ),
        child: Text(
          label,
          style: TextStyle(
            fontSize: 11,
            fontWeight: isSelected ? FontWeight.bold : FontWeight.w500,
            color: isSelected ? Colors.white : AppColors.textPrimary,
          ),
        ),
      ),
    );
  }

  /// Direct Access Banner to "My Bookings"
  Widget _buildMyBookingsBannerCard(BuildContext context) {
    return BouncyButton(
      onTap: () {
        HapticFeedback.lightImpact();
        context.push('/taga-towers/my-bookings');
      },
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        decoration: TagaTowersDesignTokens.cardDecoration(
          color: Colors.white,
          borderRadius: 16,
        ),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: const Color(0xFFE8F5E9),
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Icon(
                LucideIcons.calendarCheck,
                color: AppColors.primary,
                size: 20,
              ),
            ),
            const SizedBox(width: 12),
            const Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'View My Bookings',
                    style: TextStyle(
                      fontSize: 14.5,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  SizedBox(height: 2),
                  Text(
                    'Check reservation status, bed details, or past stays',
                    style: TextStyle(
                      fontSize: 11.5,
                      color: AppColors.textSecondary,
                    ),
                  ),
                ],
              ),
            ),
            const Icon(
              LucideIcons.chevronRight,
              size: 18,
              color: AppColors.textSecondary,
            ),
          ],
        ),
      ),
    );
  }

  /// Contact & Caretakers Information Card
  Widget _buildContactAndCaretakersCard() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 20,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: const Color(0xFFE8F5E9),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  LucideIcons.phoneCall,
                  color: AppColors.primary,
                  size: 16,
                ),
              ),
              const SizedBox(width: 8),
              const Text(
                'PROPERTY & CARETAKERS CONTACT',
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 0.8,
                  color: AppColors.textPrimary,
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          const Text(
            'For check-in assistance, room key handover, or facility queries:',
            style: TextStyle(
              fontSize: 11.5,
              color: AppColors.textSecondary,
              height: 1.35,
            ),
          ),
          const SizedBox(height: 14),

          // Caretaker 1: Mr. Muthu
          _buildContactTile(
            icon: LucideIcons.userCheck,
            title: 'Mr. Muthu (Caretaker)',
            subtitle: '96007 63744',
            onCall: () => _makePhoneCall('9600763744'),
            onCopy: () => _copyToClipboard('9600763744', 'Mr. Muthu Phone'),
          ),
          const SizedBox(height: 10),

          // Caretaker 2: Mr. Mariyappan
          _buildContactTile(
            icon: LucideIcons.userCheck,
            title: 'Mr. Mariyappan (Caretaker)',
            subtitle: '96770 10300',
            onCall: () => _makePhoneCall('9677010300'),
            onCopy: () => _copyToClipboard('9677010300', 'Mr. Mariyappan Phone'),
          ),
          const SizedBox(height: 10),

          // Landline / Helpdesk
          _buildContactTile(
            icon: LucideIcons.phone,
            title: 'Reception & Helpdesk',
            subtitle: '044 34919949',
            onCall: () => _makePhoneCall('04434919949'),
            onCopy: () => _copyToClipboard('04434919949', 'Reception Landline'),
          ),
          const SizedBox(height: 10),

          // Association Email
          _buildContactTile(
            icon: LucideIcons.mail,
            title: 'Official Email',
            subtitle: 'nammataga@gmail.com',
            onCall: () => _sendEmail('nammataga@gmail.com'),
            onCopy: () => _copyToClipboard('nammataga@gmail.com', 'Association Email'),
            actionIcon: LucideIcons.send,
            actionTooltip: 'Send Email',
          ),
          const SizedBox(height: 12),

          // Physical Address Card
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: const Color(0xFFF7FAF7),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: TagaTowersDesignTokens.subtleBorder,
              ),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Padding(
                  padding: EdgeInsets.only(top: 2),
                  child: Icon(
                    LucideIcons.mapPin,
                    size: 15,
                    color: AppColors.primary,
                  ),
                ),
                const SizedBox(width: 8),
                const Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Velanmai Illam, Egmore, Chennai',
                        style: TextStyle(
                          fontSize: 12.5,
                          fontWeight: FontWeight.bold,
                          color: AppColors.textPrimary,
                        ),
                      ),
                      SizedBox(height: 2),
                      Text(
                        'Velanmai Illam, 45, Sait Colony, II Street, Egmore, Chennai - 600008',
                        style: TextStyle(
                          fontSize: 11,
                          color: AppColors.textSecondary,
                          height: 1.3,
                        ),
                      ),
                    ],
                  ),
                ),
                IconButton(
                  tooltip: 'Copy Address',
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                  icon: const Icon(
                    LucideIcons.copy,
                    size: 15,
                    color: AppColors.primary,
                  ),
                  onPressed: () => _copyToClipboard(
                    'Velanmai Illam, 45, Sait Colony, II Street, Egmore, Chennai - 600008',
                    'Address',
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildContactTile({
    required IconData icon,
    required String title,
    required String subtitle,
    required VoidCallback onCall,
    required VoidCallback onCopy,
    IconData actionIcon = LucideIcons.phoneCall,
    String actionTooltip = 'Call',
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: const Color(0xFFFAFCFA),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: TagaTowersDesignTokens.subtleBorder),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(7),
            decoration: const BoxDecoration(
              color: Color(0xFFE8F5E9),
              shape: BoxShape.circle,
            ),
            child: Icon(icon, size: 14, color: AppColors.primary),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(
                    fontSize: 11.5,
                    fontWeight: FontWeight.w600,
                    color: AppColors.textSecondary,
                  ),
                ),
                Text(
                  subtitle,
                  style: const TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                    letterSpacing: 0.2,
                  ),
                ),
              ],
            ),
          ),
          IconButton(
            tooltip: 'Copy',
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
            icon: const Icon(
              LucideIcons.copy,
              size: 14,
              color: AppColors.textSecondary,
            ),
            onPressed: onCopy,
          ),
          IconButton(
            tooltip: actionTooltip,
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
            icon: Icon(
              actionIcon,
              size: 15,
              color: AppColors.primary,
            ),
            onPressed: onCall,
          ),
        ],
      ),
    );
  }
}
