import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../providers/events_provider.dart';
import 'widgets/event_card.dart';
import 'widgets/event_detail_modal.dart';
import 'widgets/event_shimmer.dart';

/// Tab page presenting upcoming association events, workshops, and conventions.
/// Shows strictly real backend data in a focused, clutter-free single feed.
class UpcomingEventsScreen extends ConsumerWidget {
  const UpcomingEventsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final eventsAsync = ref.watch(upcomingEventsProvider);

    return RefreshIndicator(
      color: AppColors.primary,
      backgroundColor: Colors.white,
      onRefresh: () => ref.read(upcomingEventsProvider.notifier).refresh(),
      child: eventsAsync.when(
        loading: () => const UpcomingEventsShimmer(),
        error: (error, _) => ListView(
          physics: const AlwaysScrollableScrollPhysics(
            parent: BouncingScrollPhysics(),
          ),
          padding: const EdgeInsets.all(24),
          children: [
            const SizedBox(height: 48),
            Center(
              child: Container(
                padding: const EdgeInsets.all(28),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(24),
                  border: Border.all(
                    color: const Color(0xFFE2EBE5),
                    width: 1.1,
                  ),
                  boxShadow: [
                    BoxShadow(
                      color: const Color(0xFF1B4332).withValues(alpha: 0.05),
                      blurRadius: 16,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Container(
                      width: 56,
                      height: 56,
                      decoration: const BoxDecoration(
                        color: Color(0xFFFFECEB),
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(
                        LucideIcons.circleAlert,
                        size: 28,
                        color: Color(0xFFD4183D),
                      ),
                    ),
                    const SizedBox(height: 16),
                    const Text(
                      'Unable to Connect',
                      style: TextStyle(
                        fontSize: 16.5,
                        fontWeight: FontWeight.w800,
                        color: AppColors.textPrimary,
                      ),
                    ),
                    const SizedBox(height: 6),
                    const Text(
                      'Could not refresh upcoming events. Please check your connection and tap below to try again.',
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        fontSize: 13,
                        color: Color(0xFF6B7280),
                        height: 1.45,
                      ),
                    ),
                    const SizedBox(height: 20),
                    ElevatedButton.icon(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: AppColors.primary,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(
                          horizontal: 24,
                          vertical: 12,
                        ),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                      ),
                      onPressed: () =>
                          ref.read(upcomingEventsProvider.notifier).refresh(),
                      icon: const Icon(LucideIcons.refreshCw, size: 16),
                      label: const Text(
                        'Try Again',
                        style: TextStyle(fontWeight: FontWeight.bold),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
        data: (rawEvents) {
          final events = rawEvents.where((e) => e.isUpcoming).toList();
          if (events.isEmpty) {
            return ListView(
              physics: const AlwaysScrollableScrollPhysics(
                parent: BouncingScrollPhysics(),
              ),
              padding: const EdgeInsets.all(24),
              children: [
                const SizedBox(height: 48),
                Center(
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 26,
                      vertical: 36,
                    ),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(24),
                      border: Border.all(
                        color: const Color(0xFFE2EBE5),
                        width: 1.1,
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0xFF1B4332).withValues(alpha: 0.05),
                          blurRadius: 16,
                          offset: const Offset(0, 4),
                        ),
                      ],
                    ),
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          width: 68,
                          height: 68,
                          decoration: BoxDecoration(
                            color: const Color(0xFFE8F5E9),
                            shape: BoxShape.circle,
                            border: Border.all(
                              color: const Color(0xFFB7E4C7),
                              width: 1.5,
                            ),
                          ),
                          child: const Icon(
                            LucideIcons.calendarHeart,
                            size: 32,
                            color: AppColors.primary,
                          ),
                        ),
                        const SizedBox(height: 18),
                        const Text(
                          'No Upcoming Events Right Now',
                          textAlign: TextAlign.center,
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.w800,
                            color: AppColors.textPrimary,
                            letterSpacing: -0.3,
                          ),
                        ),
                        const SizedBox(height: 8),
                        const Text(
                          'We are currently organizing new conferences, seminars, and technical workshops for our members.\n\nNew event schedules and registration links will be published here as soon as dates are confirmed.',
                          textAlign: TextAlign.center,
                          style: TextStyle(
                            fontSize: 13.5,
                            color: Color(0xFF52796F),
                            height: 1.5,
                          ),
                        ),
                        const SizedBox(height: 22),
                        BouncyButton(
                          onTap: () => ref
                              .read(upcomingEventsProvider.notifier)
                              .refresh(),
                          child: Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 20,
                              vertical: 11,
                            ),
                            decoration: BoxDecoration(
                              color: const Color(0xFFF1F8E9),
                              borderRadius: BorderRadius.circular(12),
                              border: Border.all(
                                color: const Color(0xFFD8F3DC),
                                width: 1.1,
                              ),
                            ),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: const [
                                Icon(
                                  LucideIcons.refreshCw,
                                  size: 15,
                                  color: AppColors.primary,
                                ),
                                SizedBox(width: 8),
                                Text(
                                  'Check for Updates',
                                  style: TextStyle(
                                    fontSize: 13,
                                    fontWeight: FontWeight.w700,
                                    color: AppColors.primaryDark,
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
              ],
            );
          }

          return CustomScrollView(
            physics: const AlwaysScrollableScrollPhysics(
              parent: BouncingScrollPhysics(),
            ),
            slivers: [
              // Header Summary Count
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(20, 16, 20, 12),
                  child: Row(
                    children: [
                      Container(
                        width: 8,
                        height: 8,
                        decoration: const BoxDecoration(
                          shape: BoxShape.circle,
                          color: Color(0xFF2D6A4F),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(
                        '${events.length} Upcoming ${events.length == 1 ? 'Program' : 'Programs'} Scheduled',
                        style: const TextStyle(
                          fontSize: 13.5,
                          fontWeight: FontWeight.w700,
                          color: Color(0xFF1B4332),
                        ),
                      ),
                    ],
                  ),
                ),
              ),

              // Events List (1 card per row)
              SliverPadding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 28),
                sliver: SliverList(
                  delegate: SliverChildBuilderDelegate(
                    (context, index) {
                      final event = events[index];
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 16),
                        child: EventCard(
                          event: event,
                          onTap: () {
                            EventDetailModal.show(context, event);
                          },
                        )
                            .animate()
                            .fadeIn(
                              duration: 350.ms,
                              delay: (index * 60).ms,
                            )
                            .slideY(begin: 0.06, end: 0),
                      );
                    },
                    childCount: events.length,
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}
