import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../domain/association_info.dart';
import '../../providers/home_provider.dart';

class AnimatedStatsTicker extends ConsumerWidget {
  const AnimatedStatsTicker({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final statsAsync = ref.watch(statsProvider);

    return statsAsync.when(
      data: (stats) {
        if (stats.isEmpty) return const SizedBox.shrink();

        return SizedBox(
          height: 94,
          child: ListView.separated(
            scrollDirection: Axis.horizontal,
            physics: const BouncingScrollPhysics(),
            itemCount: stats.length,
            separatorBuilder: (context, index) => const SizedBox(width: 10),
            itemBuilder: (context, index) {
              final stat = stats[index];
              final icon = _getStatIcon(stat.label);

              return _buildStatCard(stat, icon);
            },
          ),
        );
      },
      loading: () => SizedBox(
        height: 94,
        child: ListView.separated(
          scrollDirection: Axis.horizontal,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: 3,
          separatorBuilder: (context, index) => const SizedBox(width: 10),
          itemBuilder: (context, index) => Container(
            width: 132,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: const Color(0xFFE2EBE5)),
            ),
          ),
        ),
      ),
      error: (error, stack) => const SizedBox.shrink(),
    );
  }

  Widget _buildStatCard(StatData stat, IconData icon) {
    return Container(
      width: 136,
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFE0EBE4), width: 1.1),
        boxShadow: const [
          BoxShadow(
            color: Color(0x081B4332),
            blurRadius: 10,
            offset: Offset(0, 3),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Container(
                padding: const EdgeInsets.all(5),
                decoration: BoxDecoration(
                  color: const Color(0xFFE8F5E9),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Icon(icon, size: 14, color: AppColors.primary),
              ),
              const Icon(
                LucideIcons.trendingUp,
                size: 12,
                color: Color(0xFF52B788),
              ),
            ],
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              TweenAnimationBuilder<int>(
                tween: IntTween(begin: 0, end: stat.value),
                duration: const Duration(milliseconds: 1400),
                curve: Curves.easeOutCubic,
                builder: (context, val, child) {
                  final formatted = val.toString().replaceAllMapped(
                    RegExp(r'(\d{1,3})(?=(\d{3})+(?!\d))'),
                    (Match m) => '${m[1]},',
                  );
                  return Text(
                    '$formatted+',
                    style: const TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                      letterSpacing: -0.3,
                    ),
                  );
                },
              ),
              Text(
                stat.label,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.w500,
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  IconData _getStatIcon(String label) {
    final lower = label.toLowerCase();
    if (lower.contains('member')) {
      return LucideIcons.users;
    }
    if (lower.contains('year') || lower.contains('service')) {
      return LucideIcons.award;
    }
    if (lower.contains('district')) {
      return LucideIcons.mapPin;
    }
    if (lower.contains('event')) {
      return LucideIcons.calendar;
    }
    return LucideIcons.activity;
  }
}
