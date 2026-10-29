import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../core/constants/app_assets.dart';
import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../auth/providers/auth_provider.dart';
import '../providers/home_provider.dart';
import 'widgets/about_us_section.dart';
import 'widgets/animated_stats_ticker.dart';
import 'widgets/contact_glass_hub.dart';
import 'widgets/home_hero_card.dart';
import 'widgets/mission_vision_section.dart';
import 'widgets/objectives_timeline.dart';
import 'widgets/service_launcher_grid.dart';

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sessionAsync = ref.watch(authProvider);
    final isGuest = sessionAsync.value?.isGuest ?? true;

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        surfaceTintColor: Colors.transparent,
        title: Row(
          children: [
            Container(
              width: 36,
              height: 36,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: Colors.white,
                border: Border.all(color: const Color(0xFFD8F3DC), width: 1.5),
                boxShadow: [
                  BoxShadow(
                    color: AppColors.primary.withValues(alpha: 0.15),
                    blurRadius: 6,
                  ),
                ],
              ),
              padding: const EdgeInsets.all(1.5),
              child: ClipOval(
                child: Image.asset(AppAssets.logo, fit: BoxFit.contain),
              ),
            ),
            const SizedBox(width: 10),
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Namma TAGA',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                    letterSpacing: -0.2,
                  ),
                ),
                Text(
                  isGuest ? 'Guest Portal' : 'Member Portal',
                  style: const TextStyle(
                    fontSize: 11,
                    color: AppColors.primary,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
          ],
        ),
        actions: [
          if (isGuest)
            Padding(
              padding: const EdgeInsets.only(right: 12),
              child: BouncyButton(
                onTap: () => context.go('/login'),
                child: Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 6,
                  ),
                  decoration: BoxDecoration(
                    color: AppColors.primary,
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: const Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(LucideIcons.logIn, color: Colors.white, size: 13),
                      SizedBox(width: 5),
                      Text(
                        'Sign In',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 12,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            )
          else
            Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                IconButton(
                  icon: const Icon(
                    LucideIcons.user,
                    color: AppColors.primary,
                    size: 22,
                  ),
                  tooltip: 'My Profile',
                  onPressed: () => context.push('/profile'),
                ),
                IconButton(
                  icon: const Icon(
                    LucideIcons.logOut,
                    color: AppColors.destructive,
                    size: 20,
                  ),
                  tooltip: 'Sign Out',
                  onPressed: () => _confirmSignOut(context, ref),
                ),
              ],
            ),
        ],
      ),
      body: RefreshIndicator(
        color: AppColors.primary,
        onRefresh: () async {
          HapticFeedback.lightImpact();
          ref.invalidate(aboutInfoProvider);
          ref.invalidate(statsProvider);
          ref.invalidate(objectivesProvider);
          ref.invalidate(contactInfoProvider);
        },
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(
            parent: BouncingScrollPhysics(),
          ),
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 32),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: const [
              // 1. Dynamic Hero Header Card (strictly from /api/public/about)
              HomeHeroCard(),
              SizedBox(height: 16),

              // 2. Real-time Animated Statistics Counters (strictly from /api/public/about/stats)
              AnimatedStatsTicker(),
              SizedBox(height: 18),

              // 3. Quick Association Services Launcher (Member feature routing)
              ServiceLauncherGrid(),
              SizedBox(height: 18),

              // 4. Official About Us Section (strictly from /api/public/about description)
              AboutUsSection(),
              SizedBox(height: 18),

              // 5. Mission & Vision Dual Cards (strictly from /api/public/about mission & vision)
              MissionVisionSection(),
              SizedBox(height: 18),

              // 6. Strategic Objectives & TAGA Towers Showcase (strictly from /api/public/about/objectives)
              ObjectivesTimeline(),
              SizedBox(height: 18),

              // 7. Contact Us Hub (strictly from /api/public/about/contact)
              ContactGlassHub(),
            ],
          ),
        ),
      ),
    );
  }

  void _confirmSignOut(BuildContext context, WidgetRef ref) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Sign Out'),
        content: const Text(
          'Are you sure you want to sign out of your TAGA member account?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.destructive,
            ),
            onPressed: () async {
              Navigator.of(ctx).pop();
              await ref.read(authProvider.notifier).logout();
              if (context.mounted) {
                context.go('/login');
              }
            },
            child: const Text(
              'Sign Out',
              style: TextStyle(color: Colors.white),
            ),
          ),
        ],
      ),
    );
  }
}
