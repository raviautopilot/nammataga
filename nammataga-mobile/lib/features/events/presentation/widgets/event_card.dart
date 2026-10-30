import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../domain/event_model.dart';

/// Premium, high-polish card representation of an upcoming association event.
class EventCard extends StatelessWidget {
  final EventModel event;
  final VoidCallback onTap;

  const EventCard({
    super.key,
    required this.event,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return BouncyButton(
      onTap: () {
        HapticFeedback.lightImpact();
        onTap();
      },
      child: Container(
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(22),
          border: Border.all(
            color: const Color(0xFFE2EBE5),
            width: 1.2,
          ),
          boxShadow: [
            BoxShadow(
              color: const Color(0xFF1B4332).withValues(alpha: 0.06),
              blurRadius: 18,
              offset: const Offset(0, 5),
            ),
          ],
        ),
        clipBehavior: Clip.antiAlias,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Card Body Content
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Distinctive Emerald Calendar Block
                  Container(
                    width: 64,
                    height: 74,
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                        colors: [
                          Color(0xFF061A0F),
                          Color(0xFF0D2C1B),
                          Color(0xFF1B4332),
                        ],
                      ),
                      borderRadius: BorderRadius.circular(16),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0xFF1B4332).withValues(alpha: 0.28),
                          blurRadius: 10,
                          offset: const Offset(0, 3),
                        ),
                      ],
                    ),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(
                          event.formattedMonth,
                          style: const TextStyle(
                            color: Color(0xFFB7E4C7),
                            fontSize: 11,
                            fontWeight: FontWeight.w800,
                            letterSpacing: 1.2,
                          ),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          event.formattedDay,
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 25,
                            fontWeight: FontWeight.w900,
                            height: 1.0,
                          ),
                        ),
                      ],
                    ),
                  ),

                  const SizedBox(width: 14),

                  // Event Info Details
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Days Remaining Pill (if upcoming)
                        if (event.daysRemaining != null) ...[
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 8,
                              vertical: 3,
                            ),
                            decoration: BoxDecoration(
                              color: event.daysRemaining! <= 7
                                  ? const Color(0xFFFFECEB)
                                  : const Color(0xFFE8F5E9),
                              borderRadius: BorderRadius.circular(6),
                            ),
                            child: Text(
                              event.daysRemaining == 0
                                  ? 'Today'
                                  : 'In ${event.daysRemaining} days',
                              style: TextStyle(
                                fontSize: 10.5,
                                fontWeight: FontWeight.bold,
                                color: event.daysRemaining! <= 7
                                    ? const Color(0xFFD4183D)
                                    : AppColors.primaryDark,
                              ),
                            ),
                          ),
                          const SizedBox(height: 6),
                        ],

                        Text(
                          event.title,
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w800,
                            color: AppColors.textPrimary,
                            height: 1.25,
                            letterSpacing: -0.2,
                          ),
                        ),
                        const SizedBox(height: 8),

                        // Venue with pin icon
                        Row(
                          children: [
                            const Icon(
                              LucideIcons.mapPin,
                              size: 13,
                              color: AppColors.primary,
                            ),
                            const SizedBox(width: 5),
                            Expanded(
                              child: Text(
                                event.location,
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: const TextStyle(
                                  fontSize: 12.5,
                                  fontWeight: FontWeight.w500,
                                  color: Color(0xFF4B6B5D),
                                ),
                              ),
                            ),
                          ],
                        ),

                        const SizedBox(height: 4),

                        // Time with clock icon
                        Row(
                          children: [
                            const Icon(
                              LucideIcons.clock,
                              size: 13,
                              color: Color(0xFF6B7280),
                            ),
                            const SizedBox(width: 5),
                            Text(
                              event.formattedTime,
                              style: const TextStyle(
                                fontSize: 12,
                                fontWeight: FontWeight.w500,
                                color: Color(0xFF6B7280),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),

            // Card Footer
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 11),
              decoration: const BoxDecoration(
                color: Color(0xFFFAFFFE),
                border: Border(
                  top: BorderSide(
                    color: Color(0xFFF1F5F2),
                    width: 1.0,
                  ),
                ),
              ),
              child: Row(
                children: [
                  const Icon(
                    LucideIcons.calendarCheck,
                    size: 14,
                    color: Color(0xFF2D6A4F),
                  ),
                  const SizedBox(width: 6),
                  Text(
                    event.formattedDate,
                    style: const TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                      color: Color(0xFF2D6A4F),
                    ),
                  ),
                  const Spacer(),

                  // Interactive "View Details" callout
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: const [
                      Text(
                        'Details',
                        style: TextStyle(
                          fontSize: 12.5,
                          fontWeight: FontWeight.w700,
                          color: AppColors.primary,
                        ),
                      ),
                      SizedBox(width: 4),
                      Icon(
                        LucideIcons.arrowRight,
                        size: 14,
                        color: AppColors.primary,
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
