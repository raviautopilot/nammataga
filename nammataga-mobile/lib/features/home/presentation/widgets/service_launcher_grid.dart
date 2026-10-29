import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../../auth/providers/auth_provider.dart';
import 'guest_restriction_sheet.dart';

class ServiceLauncherGrid extends ConsumerWidget {
  const ServiceLauncherGrid({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sessionAsync = ref.watch(authProvider);
    final isGuest = sessionAsync.value?.isGuest ?? true;

    final services = [
      _ServiceItem(
        title: 'Office Bearers',
        subtitle: 'Leadership Directory',
        icon: LucideIcons.users,
        isRestricted: false,
        route: '/office-bearers',
      ),
      _ServiceItem(
        title: 'Events & Media',
        subtitle: 'Conferences & Gallery',
        icon: LucideIcons.calendarDays,
        isRestricted: false,
        route: '/events',
      ),
      _ServiceItem(
        title: 'Resources',
        subtitle: 'Documents & Bylaws',
        icon: LucideIcons.fileText,
        isRestricted: true,
        route: '/resources',
      ),
      _ServiceItem(
        title: 'Grievances',
        subtitle: 'Tickets & Petitions',
        icon: LucideIcons.messageSquareWarning,
        isRestricted: true,
        route: '/grievance',
      ),
      _ServiceItem(
        title: 'TAGA Towers',
        subtitle: 'Room & Dorm Booking',
        icon: LucideIcons.building2,
        isRestricted: true,
        route: '/taga-towers',
      ),
      _ServiceItem(
        title: 'Member Profile',
        subtitle: 'Subscription & ID',
        icon: LucideIcons.userCircle2,
        isRestricted: true,
        route: '/profile',
      ),
    ];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: const [
            Text(
              'Association Services',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
                letterSpacing: -0.2,
              ),
            ),
            Text(
              'Quick Access',
              style: TextStyle(
                fontSize: 11.5,
                fontWeight: FontWeight.w600,
                color: AppColors.primary,
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        GridView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: services.length,
          gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: 3,
            crossAxisSpacing: 10,
            mainAxisSpacing: 10,
            childAspectRatio: 0.92,
          ),
          itemBuilder: (context, index) {
            final item = services[index];

            return BouncyButton(
              onTap: () {
                if (isGuest && item.isRestricted) {
                  GuestRestrictionSheet.show(context, featureName: item.title);
                } else {
                  context.push(item.route);
                }
              },
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 8,
                  vertical: 10,
                ),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(
                    color: const Color(0xFFE2EBE5),
                    width: 1.1,
                  ),
                  boxShadow: const [
                    BoxShadow(
                      color: Color(0x061B4332),
                      blurRadius: 8,
                      offset: Offset(0, 2),
                    ),
                  ],
                ),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Stack(
                      clipBehavior: Clip.none,
                      children: [
                        Container(
                          width: 44,
                          height: 44,
                          decoration: BoxDecoration(
                            color: const Color(0xFFE8F5EE),
                            shape: BoxShape.circle,
                            border: Border.all(
                              color: const Color(0xFFCBE3D5),
                              width: 1.2,
                            ),
                          ),
                          child: Icon(
                            item.icon,
                            color: AppColors.primary,
                            size: 20,
                          ),
                        ),
                        if (isGuest && item.isRestricted)
                          Positioned(
                            right: -2,
                            top: -2,
                            child: Container(
                              padding: const EdgeInsets.all(3),
                              decoration: const BoxDecoration(
                                color: Color(0xFF1B4332),
                                shape: BoxShape.circle,
                              ),
                              child: const Icon(
                                LucideIcons.lock,
                                size: 9,
                                color: Colors.white,
                              ),
                            ),
                          ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Text(
                      item.title,
                      textAlign: TextAlign.center,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                        color: AppColors.textPrimary,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      item.subtitle,
                      textAlign: TextAlign.center,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        fontSize: 9.5,
                        color: AppColors.textMuted,
                      ),
                    ),
                  ],
                ),
              ),
            );
          },
        ),
      ],
    );
  }
}

class _ServiceItem {
  final String title;
  final String subtitle;
  final IconData icon;
  final bool isRestricted;
  final String route;

  const _ServiceItem({
    required this.title,
    required this.subtitle,
    required this.icon,
    required this.isRestricted,
    required this.route,
  });
}
