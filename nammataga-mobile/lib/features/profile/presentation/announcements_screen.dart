import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../domain/member_notification_model.dart';
import '../providers/profile_provider.dart';
import 'widgets/profile_subpage_scaffold.dart';

class AnnouncementsScreen extends ConsumerWidget {
  const AnnouncementsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final announcementsAsync = ref.watch(memberAnnouncementsProvider);

    return ProfileSubpageScaffold(
      title: 'Announcements',
      watermarkIcon: LucideIcons.megaphone,
      showEditNotice: false,
      body: announcementsAsync.when(
        data: (announcements) {
          if (announcements.isEmpty) {
            return Container(
              width: double.infinity,
              padding: const EdgeInsets.all(32),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
              ),
              child: const Column(
                children: [
                  Icon(
                    LucideIcons.megaphone,
                    size: 40,
                    color: Color(0xFF9CA3AF),
                  ),
                  SizedBox(height: 12),
                  Text(
                    'No Announcements',
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  SizedBox(height: 4),
                  Text(
                    'You are all caught up on official association announcements.',
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      fontSize: 12.5,
                      color: AppColors.textSecondary,
                    ),
                  ),
                ],
              ),
            );
          }

          return Column(
            children: announcements.map((a) {
              return _buildAnnouncementCard(context, ref, a);
            }).toList(),
          );
        },
        loading: () => const Center(
          child: Padding(
            padding: EdgeInsets.all(40),
            child: CircularProgressIndicator(color: AppColors.primary),
          ),
        ),
        error: (error, stack) => Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Text(
              'Unable to load announcements.\n$error',
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: AppColors.destructive,
                fontSize: 13,
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildAnnouncementCard(
    BuildContext context,
    WidgetRef ref,
    MemberNotification announcement,
  ) {
    final isUrgent = announcement.priority.toLowerCase() == 'urgent';
    final isHigh = announcement.priority.toLowerCase() == 'high';

    final badgeColor = isUrgent
        ? const Color(0xFFFEE2E2)
        : (isHigh ? const Color(0xFFFEF3C7) : const Color(0xFFE8F5EE));

    final badgeTextColor = isUrgent
        ? const Color(0xFFB91C1C)
        : (isHigh ? const Color(0xFFB45309) : AppColors.primary);

    return InkWell(
      borderRadius: BorderRadius.circular(16),
      onTap: () async {
        HapticFeedback.lightImpact();
        if (!announcement.isRead) {
          await ref
              .read(profileRepositoryProvider)
              .markNotificationRead(announcement.id);
          ref.invalidate(memberAnnouncementsProvider);
          ref.invalidate(unreadAnnouncementsCountProvider);
        }
      },
      child: Container(
        margin: const EdgeInsets.only(bottom: 12),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: announcement.isRead
                ? const Color(0xFFE5ECE7)
                : const Color(0xFF52B788),
            width: announcement.isRead ? 1 : 1.3,
          ),
          boxShadow: const [
            BoxShadow(
              color: Color(0x061B4332),
              blurRadius: 10,
              offset: Offset(0, 3),
            ),
          ],
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: announcement.isRead
                    ? const Color(0xFFF3F4F6)
                    : const Color(0xFFE8F5EE),
                shape: BoxShape.circle,
              ),
              child: Icon(
                LucideIcons.megaphone,
                size: 18,
                color: announcement.isRead
                    ? const Color(0xFF9CA3AF)
                    : AppColors.primary,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Expanded(
                        child: Text(
                          announcement.title,
                          style: TextStyle(
                            fontSize: 14.5,
                            fontWeight: announcement.isRead
                                ? FontWeight.w600
                                : FontWeight.bold,
                            color: AppColors.textPrimary,
                          ),
                        ),
                      ),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 7,
                          vertical: 2,
                        ),
                        decoration: BoxDecoration(
                          color: badgeColor,
                          borderRadius: BorderRadius.circular(6),
                        ),
                        child: Text(
                          announcement.priority.toUpperCase(),
                          style: TextStyle(
                            fontSize: 9.5,
                            fontWeight: FontWeight.bold,
                            color: badgeTextColor,
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 5),
                  Text(
                    announcement.message,
                    style: const TextStyle(
                      fontSize: 12.5,
                      color: AppColors.textSecondary,
                      height: 1.4,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        announcement.formattedDate,
                        style: const TextStyle(
                          fontSize: 11,
                          color: Color(0xFF9CA3AF),
                        ),
                      ),
                      if (!announcement.isRead)
                        const Row(
                          children: [
                            Icon(
                              LucideIcons.check,
                              size: 12,
                              color: AppColors.primary,
                            ),
                            SizedBox(width: 4),
                            Text(
                              'Tap to mark read',
                              style: TextStyle(
                                fontSize: 11,
                                color: AppColors.primary,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ],
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
