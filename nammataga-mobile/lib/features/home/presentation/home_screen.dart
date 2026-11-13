import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../core/constants/app_assets.dart';
import '../../../core/constants/app_colors.dart';
import '../../auth/providers/auth_provider.dart';
import '../providers/home_provider.dart';
import 'widgets/contact_glass_hub.dart';
import 'widgets/home_hero_card.dart';
import 'widgets/objectives_timeline.dart';
import 'widgets/service_launcher_grid.dart';

/// Professional, compact Home Dashboard for logged-in TAGA members.
/// Strictly ordered into 5 essential sections:
/// 1. Top Header ("TAGA Member Portal")
/// 2. Green Member Profile Card (TAGA, dynamic name, TAGA ID, association name, 4 G.O./Reg details)
/// 3. Our Objectives (4 concise bold statements)
/// 4. Association Services (6 simple cards without subtitles, "TAGA Tower")
/// 5. Contact Us (Contact information, Call, Email, Portal, Map)
class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
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
              children: const [
                Text(
                  'TAGA',
                  style: TextStyle(
                    fontFamily: 'Roboto',
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                    letterSpacing: -0.2,
                  ),
                ),
                Text(
                  'Member Portal',
                  style: TextStyle(
                    fontFamily: 'Roboto',
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
              // 1. Green Member Profile Card
              HomeHeroCard(),
              SizedBox(height: 16),

              // 2. Our Objectives (Directly below green profile card)
              ObjectivesTimeline(),
              SizedBox(height: 18),

              // 3. Association Services (Six simple cards)
              ServiceLauncherGrid(),
              SizedBox(height: 18),

              // 4. Contact Us (Final section of the dashboard)
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
              foregroundColor: Colors.white,
            ),
            onPressed: () {
              Navigator.of(ctx).pop();
              ref.read(authProvider.notifier).logout();
              context.go('/login');
            },
            child: const Text('Sign Out'),
          ),
        ],
      ),
    );
  }
}
