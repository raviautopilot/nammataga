import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';

/// Association Services section with 6 simple cards:
/// 1. Office Bearers
/// 2. Events & Media
/// 3. Resources
/// 4. Grievances
/// 5. TAGA Tower
/// 6. Member Profile
class ServiceLauncherGrid extends ConsumerWidget {
  const ServiceLauncherGrid({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final services = [
      _ServiceItem(
        title: 'Office Bearers',
        icon: LucideIcons.users,
        route: '/office-bearers',
      ),
      _ServiceItem(
        title: 'Events & Media',
        icon: LucideIcons.calendarDays,
        route: '/events',
      ),
      _ServiceItem(
        title: 'Resources',
        icon: LucideIcons.fileText,
        route: '/resources',
      ),
      _ServiceItem(
        title: 'Grievances',
        icon: LucideIcons.messageSquareWarning,
        route: '/grievance',
      ),
      _ServiceItem(
        title: 'TAGA Tower',
        icon: LucideIcons.building2,
        route: '/taga-towers',
      ),
      _ServiceItem(
        title: 'Member Profile',
        icon: LucideIcons.userCircle2,
        route: '/profile',
      ),
    ];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Container(
              width: 3.5,
              height: 16,
              decoration: BoxDecoration(
                color: AppColors.primary,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            const SizedBox(width: 8),
            const Text(
              'Association Services',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
                letterSpacing: -0.2,
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
            childAspectRatio: 0.98,
          ),
          itemBuilder: (context, index) {
            final item = services[index];

            return BouncyButton(
              onTap: () {
                context.push(item.route);
              },
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 6,
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
                    const SizedBox(height: 8),
                    Text(
                      item.title,
                      textAlign: TextAlign.center,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                        color: AppColors.textPrimary,
                        height: 1.2,
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
  final IconData icon;
  final String route;

  const _ServiceItem({
    required this.title,
    required this.icon,
    required this.route,
  });
}
