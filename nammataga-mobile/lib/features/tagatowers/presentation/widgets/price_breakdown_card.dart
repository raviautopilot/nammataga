import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../theme/tagatowers_design_tokens.dart';

/// A comprehensive price breakdown card for TAGA Towers bookings.
///
/// Clearly displays:
/// - Room type / name
/// - Stay duration (X days / 24-hr cycles)
/// - Bed count
/// - Calculated advance payment payable now (e.g., ₹200 x 2 = ₹400)
/// - Remaining balance notice
/// - Non-refundable cancellation alert badge
/// - Optional primary action button to initiate payment
class PriceBreakdownCard extends StatelessWidget {
  /// Room identifier or title (e.g., 'apex-suite', 'gents-dorm', 'A/C Room')
  final String roomType;

  /// Optional room number or title qualifier (e.g., 'Room 1')
  final String? roomNumber;

  /// Duration of stay in days / 24-hour cycles (minimum 1)
  final int stayDurationDays;

  /// Number of beds booked
  final int bedCount;

  /// Booking recipient ('self' or 'guest')
  final String bookingFor;

  /// Optional override of the per-bed advance rate in INR.
  /// If null, auto-calculated using [TagaTowersDesignTokens.getAdvanceRatePerBed].
  final int? advanceRatePerBed;

  /// Optional override of the total advance in INR.
  /// If null, calculated as `advanceRatePerBed * bedCount`.
  final int? totalAdvanceAmount;

  /// Optional check-in date
  final DateTime? checkInDate;

  /// Optional check-out date
  final DateTime? checkOutDate;

  /// Custom remaining balance disclaimer
  final String remainingBalanceNotice;

  /// Custom cancellation policy message
  final String cancellationNotice;

  /// Optional callback to trigger payment
  final VoidCallback? onPayPressed;

  /// Whether payment is currently processing
  final bool isProcessingPayment;

  /// Custom CTA button text override
  final String? payButtonLabel;

  const PriceBreakdownCard({
    super.key,
    required this.roomType,
    this.roomNumber,
    this.stayDurationDays = 1,
    this.bedCount = 1,
    this.bookingFor = 'self',
    this.advanceRatePerBed,
    this.totalAdvanceAmount,
    this.checkInDate,
    this.checkOutDate,
    this.remainingBalanceNotice =
        'Remaining balance to be settled after stay upon checkout.',
    this.cancellationNotice =
        'Cancellation Policy: Please note that the advance booking payment is strictly non-refundable upon cancellation.',
    this.onPayPressed,
    this.isProcessingPayment = false,
    this.payButtonLabel,
  });

  @override
  Widget build(BuildContext context) {
    final effectiveDays = stayDurationDays < 1 ? 1 : stayDurationDays;
    final effectiveBeds = bedCount < 1 ? 1 : bedCount;
    final ratePerBed = advanceRatePerBed ??
        TagaTowersDesignTokens.getAdvanceRatePerBed(roomType);
    final advancePayable = totalAdvanceAmount ?? (ratePerBed * effectiveBeds);
    final roomDisplayName = TagaTowersDesignTokens.getRoomDisplayName(roomType);

    final dateFormat = DateFormat('dd MMM yyyy');
    final hasDates = checkInDate != null && checkOutDate != null;

    return Container(
      width: double.infinity,
      decoration: TagaTowersDesignTokens.cardDecoration(
        color: Colors.white,
        borderRadius: 18,
      ),
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          // Header: Receipt icon and title
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
                      LucideIcons.receiptText,
                      size: 18,
                      color: TagaTowersDesignTokens.deepEmerald,
                    ),
                  ),
                  const SizedBox(width: 10),
                  const Text(
                    'Booking Summary',
                    style: TextStyle(
                      fontSize: 16.5,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                      letterSpacing: -0.3,
                    ),
                  ),
                ],
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 4),
                decoration: BoxDecoration(
                  color: const Color(0xFFF1F8E9),
                  borderRadius: BorderRadius.circular(999),
                  border: Border.all(
                    color: TagaTowersDesignTokens.mintHighlight,
                    width: 1,
                  ),
                ),
                child: const Text(
                  'Advance Fee',
                  style: TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w700,
                    color: TagaTowersDesignTokens.primaryGreen,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Line Items
          _buildLineItem(
            icon: LucideIcons.bedDouble,
            label: 'Room Type',
            value: roomNumber != null
                ? '$roomDisplayName ($roomNumber)'
                : roomDisplayName,
          ),
          const SizedBox(height: 10),

          _buildLineItem(
            icon: bookingFor.toLowerCase() == 'guest'
                ? LucideIcons.users
                : LucideIcons.user,
            label: 'Booking Type',
            value: bookingFor.toLowerCase() == 'guest'
                ? 'Guest Booking'
                : 'Self Booking',
          ),
          const SizedBox(height: 10),

          _buildLineItem(
            icon: LucideIcons.clock,
            label: 'Stay Duration',
            value: '$effectiveDays Day(s) • ${effectiveDays * 24} Hours',
            subValue: hasDates
                ? '${dateFormat.format(checkInDate!)} → ${dateFormat.format(checkOutDate!)}'
                : '12:00 AM Check-in to 12:00 AM Checkout',
          ),
          const SizedBox(height: 10),

          _buildLineItem(
            icon: LucideIcons.layers,
            label: 'Bed Count',
            value: '$effectiveBeds Bed${effectiveBeds > 1 ? 's' : ''}',
          ),
          const SizedBox(height: 10),

          _buildLineItem(
            icon: LucideIcons.indianRupee,
            label: 'Advance Rate / Bed',
            value: '₹$ratePerBed',
          ),

          const Padding(
            padding: EdgeInsets.symmetric(vertical: 14),
            child: Divider(
              color: Color(0xFFE2EBE5),
              height: 1,
              thickness: 1,
            ),
          ),

          // Total Advance Payable Box
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
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Advance Payment',
                      style: TextStyle(
                        fontSize: 14,
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
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
                Text(
                  '₹$advancePayable',
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

          // Remaining Balance Notice
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Padding(
                padding: EdgeInsets.only(top: 2),
                child: Icon(
                  LucideIcons.info,
                  size: 14,
                  color: AppColors.textSecondary,
                ),
              ),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  remainingBalanceNotice,
                  style: const TextStyle(
                    fontSize: 11.5,
                    color: AppColors.textSecondary,
                    height: 1.35,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),

          // Non-refundable Cancellation Alert Badge
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(12),
            decoration: TagaTowersDesignTokens.alertDecoration(
              borderRadius: 12,
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Padding(
                  padding: EdgeInsets.only(top: 2),
                  child: Icon(
                    LucideIcons.alertTriangle,
                    size: 16,
                    color: TagaTowersDesignTokens.warningAmberDark,
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    cancellationNotice,
                    style: const TextStyle(
                      fontSize: 11.5,
                      height: 1.45,
                      fontWeight: FontWeight.w600,
                      color: TagaTowersDesignTokens.warningAmberDark,
                    ),
                  ),
                ),
              ],
            ),
          ),

          // Optional Payment Action Button
          if (onPayPressed != null) ...[
            const SizedBox(height: 16),
            BouncyButton(
              onTap: isProcessingPayment ? null : onPayPressed,
              child: Container(
                width: double.infinity,
                height: 50,
                decoration: BoxDecoration(
                  color: isProcessingPayment
                      ? AppColors.primary.withValues(alpha: 0.6)
                      : AppColors.primary,
                  borderRadius: BorderRadius.circular(14),
                  boxShadow: [
                    BoxShadow(
                      color: AppColors.primary.withValues(alpha: 0.3),
                      blurRadius: 12,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: Center(
                  child: isProcessingPayment
                      ? const SizedBox(
                          width: 22,
                          height: 22,
                          child: CircularProgressIndicator(
                            strokeWidth: 2.2,
                            valueColor:
                                AlwaysStoppedAnimation<Color>(Colors.white),
                          ),
                        )
                      : Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            const Icon(
                              LucideIcons.shieldCheck,
                              size: 18,
                              color: Colors.white,
                            ),
                            const SizedBox(width: 8),
                            Text(
                              payButtonLabel ??
                                  'Proceed to Pay ₹$advancePayable',
                              style: const TextStyle(
                                color: Colors.white,
                                fontSize: 15,
                                fontWeight: FontWeight.bold,
                                letterSpacing: 0.2,
                              ),
                            ),
                          ],
                        ),
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildLineItem({
    required IconData icon,
    required String label,
    required String value,
    String? subValue,
  }) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(
          icon,
          size: 16,
          color: AppColors.textSecondary,
        ),
        const SizedBox(width: 10),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: const TextStyle(
                  fontSize: 12.5,
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
}
