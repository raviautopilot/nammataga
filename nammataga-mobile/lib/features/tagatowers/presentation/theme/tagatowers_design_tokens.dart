import 'package:flutter/material.dart';
import '../../../../core/constants/app_colors.dart';

/// Design tokens, typography rules, color schemes, and decoration helpers
/// strictly tailored for the TAGA Towers accommodation module.
class TagaTowersDesignTokens {
  TagaTowersDesignTokens._();

  // ═══════════════════════════════════════════════════════════════════════════
  // BRAND COLOR PALETTE (mapped to AppColors)
  // ═══════════════════════════════════════════════════════════════════════════

  /// Signature TAGA green (0xFF2D6A4F)
  static const Color primaryGreen = AppColors.primary;

  /// Deep forest emerald used for gradients, headers, and rich contrast (0xFF1B4332)
  static const Color deepEmerald = AppColors.primaryDark;

  /// Pale mint wash used for pill backgrounds, success states, and containers (0xFFD8F3DC)
  static const Color softMint = AppColors.secondary;

  /// Vibrant mint highlight used for badges, active borders, and accent rings (0xFF95D5B2)
  static const Color mintHighlight = AppColors.secondaryDark;

  /// Clean, warm off-white surface tone (0xFFFAFFFE)
  static const Color surfaceLight = AppColors.background;

  /// Subtle green-tinted border (10% opacity: 0x1A2D6A4F)
  static const Color subtleBorder = AppColors.borderLight;

  /// Defined green-tinted border (15% opacity: 0x262D6A4F)
  static const Color border = AppColors.border;

  /// Muted container background
  static const Color mutedBackground = AppColors.muted;

  /// Amber accents for notices, non-refundable warning, and caution banners
  static const Color warningAmber = Color(0xFFF59E0B);
  static const Color warningAmberDark = Color(0xFFB45309);
  static const Color warningAmberLight = Color(0xFFFEF3C7);
  static const Color warningAmberBorder = Color(0xFFFCD34D);

  // ═══════════════════════════════════════════════════════════════════════════
  // BOOKING RATES & BUSINESS LOGIC CONSTANTS
  // ═══════════════════════════════════════════════════════════════════════════

  /// Non-refundable advance fee per bed for dormitory accommodations (₹100)
  static const int dormitoryAdvance = 100;

  /// Non-refundable advance fee per bed for private/A.C. suites (₹200)
  static const int standardAdvance = 200;

  /// Maximum consecutive booking days permitted per session
  static const int maxBookingDays = 10;

  /// Standard check-in time (midnight cycle)
  static const String checkInTime = '12:00 AM';

  /// Standard check-out time (midnight cycle)
  static const String checkOutTime = '12:00 AM (Midnight)';

  /// Descriptive stay schedule policy notice
  static const String stayScheduleNotice =
      '24-Hour Stay Schedule (12:00 AM Check-in to 12:00 AM Checkout)';

  /// Facility physical location string
  static const String towerAddress =
      'Velanmai Illam, 45, Sait Colony, II Street, Egmore, Chennai - 600008';

  /// Reception & Helpdesk Contact
  static const String helpdeskPhone = '044 34919949';
  static const String helpdeskEmail = 'nammataga@gmail.com';

  /// Calculates the required advance rate in INR per bed based on room type or id
  static int getAdvanceRatePerBed(String roomTypeOrId) {
    final normalized = roomTypeOrId.toLowerCase().trim();
    if (normalized.contains('dorm') ||
        normalized == 'gents-dorm' ||
        normalized == 'ladies-dorm' ||
        normalized == 'gents dormitory' ||
        normalized == 'ladies dormitory') {
      return dormitoryAdvance;
    }
    return standardAdvance;
  }

  /// Maps internal room type identifiers to human-readable display names
  static String getRoomDisplayName(String roomTypeOrId) {
    final key = roomTypeOrId.toLowerCase().trim();
    switch (key) {
      case 'apex-suite':
      case 'apex-1':
        return 'Apex Suite A/C';
      case 'ac-room':
        return 'A/C Room';
      case 'gents-dorm':
        return 'Gents Dormitory';
      case 'ladies-dorm':
        return 'Ladies Dormitory';
      case 'kurinchi':
        return 'Kurinchi Suite';
      case 'pavalam':
        return 'Pavalam Suite';
      case 'malligai':
        return 'Malligai Suite';
      case 'kavery':
        return 'Kavery Suite';
      case 'vasantham':
        return 'Vasantham Suite';
      case 'pasumai':
        return 'Pasumai Suite';
      default:
        // Title-case fallback
        return roomTypeOrId
            .replaceAll('-', ' ')
            .replaceAll('_', ' ')
            .split(' ')
            .map((word) => word.isNotEmpty
                ? '${word[0].toUpperCase()}${word.substring(1)}'
                : '')
            .join(' ');
    }
  }

  /// Alias for [getRoomDisplayName] for compatibility
  static String getRoomTypeLabel(String roomTypeOrId) =>
      getRoomDisplayName(roomTypeOrId);

  // ═══════════════════════════════════════════════════════════════════════════
  // SHADOW TOKENS (Tinted Ambient Shadows)
  // ═══════════════════════════════════════════════════════════════════════════

  static const List<BoxShadow> softShadow = [
    BoxShadow(
      color: Color(0x0C1B4332), // 5% deep emerald
      blurRadius: 16,
      offset: Offset(0, 4),
    ),
  ];

  static const List<BoxShadow> elevatedShadow = [
    BoxShadow(
      color: Color(0x141B4332), // 8% deep emerald
      blurRadius: 24,
      offset: Offset(0, 8),
    ),
  ];

  static const List<BoxShadow> floatingGlow = [
    BoxShadow(
      color: Color(0x262D6A4F), // 15% primary green
      blurRadius: 20,
      offset: Offset(0, 6),
    ),
  ];

  // ═══════════════════════════════════════════════════════════════════════════
  // CARD DECORATION HELPERS
  // ═══════════════════════════════════════════════════════════════════════════

  /// Standard card decoration with subtle emerald border and ambient shadow
  static BoxDecoration cardDecoration({
    Color color = Colors.white,
    Color borderColor = subtleBorder,
    double borderRadius = 16.0,
    double borderWidth = 1.2,
    List<BoxShadow>? boxShadow,
  }) {
    return BoxDecoration(
      color: color,
      borderRadius: BorderRadius.circular(borderRadius),
      border: Border.all(
        color: borderColor,
        width: borderWidth,
      ),
      boxShadow: boxShadow ?? softShadow,
    );
  }

  /// Frosted glass decoration with soft translucency and white-tint border
  static BoxDecoration frostedDecoration({
    double borderRadius = 16.0,
    Color backgroundColor = const Color(0xD9FFFFFF),
    Color borderColor = const Color(0x4DFFFFFF),
  }) {
    return BoxDecoration(
      color: backgroundColor,
      borderRadius: BorderRadius.circular(borderRadius),
      border: Border.all(
        color: borderColor,
        width: 1.2,
      ),
    );
  }

  /// Dark emerald hero container gradient
  static BoxDecoration heroGradientDecoration({
    double borderRadius = 20.0,
  }) {
    return BoxDecoration(
      gradient: const LinearGradient(
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
        colors: [
          Color(0xFF0D281E),
          Color(0xFF1B4332),
          Color(0xFF2D6A4F),
        ],
      ),
      borderRadius: BorderRadius.circular(borderRadius),
      boxShadow: elevatedShadow,
    );
  }

  // ═══════════════════════════════════════════════════════════════════════════
  // BADGE & PILL DECORATIONS
  // ═══════════════════════════════════════════════════════════════════════════

  /// Mint badge decoration for positive badges, checkmarks, and active indicators
  static BoxDecoration badgeDecoration({
    Color backgroundColor = softMint,
    Color borderColor = mintHighlight,
    double borderRadius = 999.0,
  }) {
    return BoxDecoration(
      color: backgroundColor,
      borderRadius: BorderRadius.circular(borderRadius),
      border: Border.all(
        color: borderColor,
        width: 1.0,
      ),
    );
  }

  /// Dark translucent pill decoration for hero overlays
  static BoxDecoration darkPillDecoration({
    double borderRadius = 999.0,
  }) {
    return BoxDecoration(
      color: deepEmerald.withValues(alpha: 0.75),
      borderRadius: BorderRadius.circular(borderRadius),
      border: Border.all(
        color: mintHighlight.withValues(alpha: 0.4),
        width: 1.0,
      ),
    );
  }

  /// Warning alert container decoration for cancellation policies
  static BoxDecoration alertDecoration({
    double borderRadius = 12.0,
  }) {
    return BoxDecoration(
      color: warningAmberLight,
      borderRadius: BorderRadius.circular(borderRadius),
      border: Border.all(
        color: warningAmberBorder,
        width: 1.2,
      ),
    );
  }
}
