import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../theme/tagatowers_design_tokens.dart';

/// A card widget allowing members to select check-in and check-out dates
/// for TAGA Towers within the strictly enforced 10-day booking horizon.
///
/// Follows the 24-hour midnight-to-midnight stay cycle:
/// - Check-in: between today and (today + 9 days)
/// - Check-out: between (check-in + 1 day) and (today + 10 days)
class DateRangePickerCard extends StatelessWidget {
  final DateTime checkInDate;
  final DateTime checkOutDate;
  final ValueChanged<DateTime> onCheckInChanged;
  final ValueChanged<DateTime> onCheckOutChanged;
  final int maxHorizonDays;

  const DateRangePickerCard({
    super.key,
    required this.checkInDate,
    required this.checkOutDate,
    required this.onCheckInChanged,
    required this.onCheckOutChanged,
    this.maxHorizonDays = TagaTowersDesignTokens.maxBookingDays,
  });

  DateTime _stripTime(DateTime dt) => DateTime(dt.year, dt.month, dt.day);

  int get totalNights {
    final diff = checkOutDate.difference(checkInDate).inDays;
    return diff > 0 ? diff : 1;
  }

  ThemeData _buildPickerTheme(BuildContext context) {
    return Theme.of(context).copyWith(
      colorScheme: const ColorScheme.light(
        primary: TagaTowersDesignTokens.primaryGreen,
        onPrimary: Colors.white,
        onSurface: AppColors.textPrimary,
        surface: Colors.white,
        secondary: TagaTowersDesignTokens.softMint,
      ),
      dialogTheme: DialogThemeData(
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(24),
        ),
        elevation: 16,
      ),
      datePickerTheme: DatePickerThemeData(
        backgroundColor: Colors.white,
        surfaceTintColor: Colors.transparent,
        headerBackgroundColor: TagaTowersDesignTokens.deepEmerald,
        headerForegroundColor: Colors.white,
        headerHeadlineStyle: const TextStyle(
          fontSize: 22,
          fontWeight: FontWeight.w800,
          color: Colors.white,
          letterSpacing: -0.3,
        ),
        headerHelpStyle: const TextStyle(
          fontSize: 11.5,
          fontWeight: FontWeight.w700,
          color: Color(0xFFD8F3DC),
          letterSpacing: 0.8,
        ),
        weekdayStyle: const TextStyle(
          fontSize: 12.5,
          fontWeight: FontWeight.w800,
          color: TagaTowersDesignTokens.deepEmerald,
        ),
        dayStyle: const TextStyle(
          fontSize: 14,
          fontWeight: FontWeight.w700,
        ),
        dayForegroundColor: WidgetStateProperty.resolveWith<Color>((states) {
          if (states.contains(WidgetState.disabled)) {
            return const Color(0xFFB0BEC5); // Crisp visible soft-gray for non-horizon days
          }
          if (states.contains(WidgetState.selected)) {
            return Colors.white;
          }
          return TagaTowersDesignTokens.deepEmerald; // High-contrast, sharp dark emerald for available days
        }),
        dayBackgroundColor: WidgetStateProperty.resolveWith<Color?>((states) {
          if (states.contains(WidgetState.selected)) {
            return TagaTowersDesignTokens.primaryGreen;
          }
          return null;
        }),
        todayForegroundColor: const WidgetStatePropertyAll<Color>(
          TagaTowersDesignTokens.primaryGreen,
        ),
        todayBorder: const BorderSide(
          color: TagaTowersDesignTokens.primaryGreen,
          width: 1.8,
        ),
        dividerColor: const Color(0x262D6A4F),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(24),
        ),
        cancelButtonStyle: TextButton.styleFrom(
          foregroundColor: AppColors.textSecondary,
          textStyle: const TextStyle(fontWeight: FontWeight.w700, fontSize: 13),
        ),
        confirmButtonStyle: TextButton.styleFrom(
          foregroundColor: TagaTowersDesignTokens.deepEmerald,
          textStyle: const TextStyle(fontWeight: FontWeight.w800, fontSize: 13),
        ),
      ),
    );
  }

  Future<void> _pickCheckInDate(BuildContext context) async {
    HapticFeedback.lightImpact();
    final today = _stripTime(DateTime.now());
    // Since minimum stay is 1 night and max checkout is today + maxHorizonDays,
    // the latest valid check-in date is today + (maxHorizonDays - 1).
    final lastCheckIn = today.add(Duration(days: maxHorizonDays - 1));

    final picked = await showDatePicker(
      context: context,
      initialDate: checkInDate.isBefore(today)
          ? today
          : (checkInDate.isAfter(lastCheckIn) ? lastCheckIn : checkInDate),
      firstDate: today,
      lastDate: lastCheckIn,
      initialEntryMode: DatePickerEntryMode.calendarOnly,
      barrierColor: const Color(0xB3000000), // Rich 70% black backdrop preventing background text bleed-through
      helpText: 'SELECT CHECK-IN DATE (12:00 AM)',
      confirmText: 'SELECT',
      cancelText: 'CANCEL',
      builder: (context, child) => Theme(
        data: _buildPickerTheme(context),
        child: child!,
      ),
    );

    if (picked != null) {
      final normalizedPicked = _stripTime(picked);
      onCheckInChanged(normalizedPicked);

      // If existing check-out is now invalid (<= new check-in or exceeds max allowed date)
      final minCheckout = normalizedPicked.add(const Duration(days: 1));
      final absoluteMaxCheckout = today.add(Duration(days: maxHorizonDays));

      if (checkOutDate.isBefore(minCheckout) ||
          checkOutDate.isAtSameMomentAs(normalizedPicked)) {
        onCheckOutChanged(minCheckout);
      } else if (checkOutDate.isAfter(absoluteMaxCheckout)) {
        onCheckOutChanged(absoluteMaxCheckout);
      }
    }
  }

  Future<void> _pickCheckOutDate(BuildContext context) async {
    HapticFeedback.lightImpact();
    final today = _stripTime(DateTime.now());
    final minCheckout = _stripTime(checkInDate).add(const Duration(days: 1));
    final absoluteMaxCheckout = today.add(Duration(days: maxHorizonDays));

    if (minCheckout.isAfter(absoluteMaxCheckout)) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text(
            'Check-in is on the latest available day. Check-out must be the following day.',
          ),
          duration: Duration(seconds: 2),
        ),
      );
      return;
    }

    final initial = checkOutDate.isBefore(minCheckout)
        ? minCheckout
        : (checkOutDate.isAfter(absoluteMaxCheckout)
            ? absoluteMaxCheckout
            : checkOutDate);

    final picked = await showDatePicker(
      context: context,
      initialDate: initial,
      firstDate: minCheckout,
      lastDate: absoluteMaxCheckout,
      initialEntryMode: DatePickerEntryMode.calendarOnly,
      barrierColor: const Color(0xB3000000), // Rich 70% black backdrop preventing background text bleed-through
      helpText: 'SELECT CHECK-OUT DATE (12:00 AM)',
      confirmText: 'SELECT',
      cancelText: 'CANCEL',
      builder: (context, child) => Theme(
        data: _buildPickerTheme(context),
        child: child!,
      ),
    );

    if (picked != null) {
      onCheckOutChanged(_stripTime(picked));
    }
  }

  @override
  Widget build(BuildContext context) {
    final dateFormat = DateFormat('EEE, dd MMM yyyy');
    final dayFormat = DateFormat('dd');
    final monthYearFormat = DateFormat('MMM yyyy');
    final weekdayFormat = DateFormat('EEEE');

    return Container(
      padding: const EdgeInsets.all(18),
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 20,
        boxShadow: TagaTowersDesignTokens.softShadow,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header Row: Section Label + Nights Badge
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(6),
                    decoration: BoxDecoration(
                      color: TagaTowersDesignTokens.softMint,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Icon(
                      LucideIcons.calendarDays,
                      color: AppColors.primaryDark,
                      size: 16,
                    ),
                  ),
                  const SizedBox(width: 8),
                  const Text(
                    'SELECT STAY DATES',
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                      letterSpacing: 0.8,
                      color: AppColors.textPrimary,
                    ),
                  ),
                ],
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: TagaTowersDesignTokens.badgeDecoration(
                  backgroundColor: TagaTowersDesignTokens.softMint,
                  borderColor: TagaTowersDesignTokens.mintHighlight,
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(
                      LucideIcons.moon,
                      size: 12,
                      color: AppColors.primaryDark,
                    ),
                    const SizedBox(width: 4),
                    Text(
                      '$totalNights ${totalNights == 1 ? 'Night' : 'Nights'} (${totalNights * 24}h)',
                      style: const TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        color: AppColors.primaryDark,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Date Tiles Grid (Check-In & Check-Out)
          Row(
            children: [
              // Check-In Selector Tile
              Expanded(
                child: BouncyButton(
                  onTap: () => _pickCheckInDate(context),
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                        colors: [
                          Color(0xFFFAFCF8),
                          Color(0xFFF1F7EE),
                        ],
                      ),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color: TagaTowersDesignTokens.primaryGreen.withValues(alpha: 0.35),
                        width: 1.3,
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: TagaTowersDesignTokens.deepEmerald.withValues(alpha: 0.04),
                          blurRadius: 8,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Container(
                              padding: const EdgeInsets.all(4),
                              decoration: BoxDecoration(
                                color: TagaTowersDesignTokens.softMint,
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: const Icon(
                                LucideIcons.logIn,
                                size: 12,
                                color: TagaTowersDesignTokens.deepEmerald,
                              ),
                            ),
                            const SizedBox(width: 5),
                            const Expanded(
                              child: Text(
                                'CHECK-IN (12 AM)',
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: TextStyle(
                                  fontSize: 9.5,
                                  fontWeight: FontWeight.w800,
                                  letterSpacing: 0.3,
                                  color: AppColors.primaryDark,
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 10),
                        Row(
                          crossAxisAlignment: CrossAxisAlignment.baseline,
                          textBaseline: TextBaseline.alphabetic,
                          children: [
                            Text(
                              dayFormat.format(checkInDate),
                              style: const TextStyle(
                                fontSize: 32,
                                fontWeight: FontWeight.w900,
                                color: TagaTowersDesignTokens.deepEmerald,
                                height: 1.0,
                                letterSpacing: -0.5,
                              ),
                            ),
                            const SizedBox(width: 8),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    monthYearFormat.format(checkInDate),
                                    style: const TextStyle(
                                      fontSize: 13,
                                      fontWeight: FontWeight.w800,
                                      color: AppColors.textPrimary,
                                      height: 1.1,
                                    ),
                                  ),
                                  const SizedBox(height: 2),
                                  Text(
                                    weekdayFormat.format(checkInDate),
                                    style: const TextStyle(
                                      fontSize: 11,
                                      fontWeight: FontWeight.w500,
                                      color: AppColors.textSecondary,
                                      height: 1.1,
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
                ),
              ),

              const SizedBox(width: 10),

              // Arrow Divider
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    colors: [
                      Color(0xFFE8F5E9),
                      Color(0xFFD8F3DC),
                    ],
                  ),
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: TagaTowersDesignTokens.primaryGreen.withValues(alpha: 0.2),
                    width: 1,
                  ),
                  boxShadow: [
                    BoxShadow(
                      color: TagaTowersDesignTokens.deepEmerald.withValues(alpha: 0.08),
                      blurRadius: 4,
                      offset: const Offset(0, 2),
                    ),
                  ],
                ),
                child: const Icon(
                  LucideIcons.arrowRight,
                  size: 15,
                  color: TagaTowersDesignTokens.deepEmerald,
                ),
              ),

              const SizedBox(width: 10),

              // Check-Out Selector Tile
              Expanded(
                child: BouncyButton(
                  onTap: () => _pickCheckOutDate(context),
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                        colors: [
                          Color(0xFFFAFCF8),
                          Color(0xFFF1F7EE),
                        ],
                      ),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color: TagaTowersDesignTokens.primaryGreen.withValues(alpha: 0.35),
                        width: 1.3,
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: TagaTowersDesignTokens.deepEmerald.withValues(alpha: 0.04),
                          blurRadius: 8,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Container(
                              padding: const EdgeInsets.all(4),
                              decoration: BoxDecoration(
                                color: TagaTowersDesignTokens.softMint,
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: const Icon(
                                LucideIcons.logOut,
                                size: 12,
                                color: TagaTowersDesignTokens.deepEmerald,
                              ),
                            ),
                            const SizedBox(width: 5),
                            const Expanded(
                              child: Text(
                                'CHECK-OUT (12 AM)',
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: TextStyle(
                                  fontSize: 9.5,
                                  fontWeight: FontWeight.w800,
                                  letterSpacing: 0.3,
                                  color: AppColors.primaryDark,
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 10),
                        Row(
                          crossAxisAlignment: CrossAxisAlignment.baseline,
                          textBaseline: TextBaseline.alphabetic,
                          children: [
                            Text(
                              dayFormat.format(checkOutDate),
                              style: const TextStyle(
                                fontSize: 32,
                                fontWeight: FontWeight.w900,
                                color: TagaTowersDesignTokens.deepEmerald,
                                height: 1.0,
                                letterSpacing: -0.5,
                              ),
                            ),
                            const SizedBox(width: 8),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    monthYearFormat.format(checkOutDate),
                                    style: const TextStyle(
                                      fontSize: 13,
                                      fontWeight: FontWeight.w800,
                                      color: AppColors.textPrimary,
                                      height: 1.1,
                                    ),
                                  ),
                                  const SizedBox(height: 2),
                                  Text(
                                    weekdayFormat.format(checkOutDate),
                                    style: const TextStyle(
                                      fontSize: 11,
                                      fontWeight: FontWeight.w500,
                                      color: AppColors.textSecondary,
                                      height: 1.1,
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
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),

          // Informational footer pill
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
            decoration: BoxDecoration(
              color: const Color(0xFFF1F8E9),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Row(
              children: [
                const Icon(
                  LucideIcons.info,
                  size: 13,
                  color: AppColors.primary,
                ),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    '24-Hour stay window: ${dateFormat.format(checkInDate)} 12:00 AM to ${dateFormat.format(checkOutDate)} 12:00 AM.',
                    style: const TextStyle(
                      fontSize: 11,
                      color: AppColors.textPrimary,
                      fontWeight: FontWeight.w500,
                      height: 1.3,
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
