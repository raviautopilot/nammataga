import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/constants/app_assets.dart';
import '../../../../core/constants/app_colors.dart';
import '../../domain/association_info.dart';
import '../../providers/home_provider.dart';

class ObjectivesTimeline extends ConsumerWidget {
  const ObjectivesTimeline({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final objectivesAsync = ref.watch(objectivesProvider);

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
        boxShadow: const [
          BoxShadow(
            color: Color(0x061B4332),
            blurRadius: 10,
            offset: Offset(0, 3),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Section Title & Description from web
          const Text(
            'Our Objectives',
            style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: AppColors.primary,
              letterSpacing: -0.2,
            ),
          ),
          const SizedBox(height: 3),
          const Text(
            "Key goals that drive our association's activities and initiatives",
            style: TextStyle(
              fontSize: 12,
              color: AppColors.textSecondary,
              height: 1.35,
            ),
          ),
          const SizedBox(height: 14),

          // TAGA Towers architectural visual matching web's objectivesImage
          Container(
            height: 140,
            width: double.infinity,
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(14),
              gradient: const LinearGradient(
                colors: [Color(0xFF092014), Color(0xFF1B4332)],
              ),
            ),
            child: ClipRRect(
              borderRadius: BorderRadius.circular(14),
              child: Image.asset(
                AppAssets.tagatower,
                fit: BoxFit.cover,
                alignment: const Alignment(0.0, -0.3),
                errorBuilder: (context, error, stackTrace) =>
                    const SizedBox.shrink(),
              ),
            ),
          ),
          const SizedBox(height: 16),

          // Dynamic Objectives List from backend
          objectivesAsync.when(
            data: (objectives) {
              if (objectives.isEmpty) {
                return const SizedBox.shrink();
              }

              return Column(
                children: List.generate(objectives.length, (index) {
                  final obj = objectives[index];
                  final isLast = index == objectives.length - 1;

                  return _buildObjectiveItem(obj, index + 1, isLast);
                }),
              );
            },
            loading: () => const Center(
              child: Padding(
                padding: EdgeInsets.all(20),
                child: CircularProgressIndicator(color: AppColors.primary),
              ),
            ),
            error: (error, stack) => const SizedBox.shrink(),
          ),
        ],
      ),
    );
  }

  Widget _buildObjectiveItem(ObjectiveData obj, int index, bool isLast) {
    return Padding(
      padding: EdgeInsets.only(bottom: isLast ? 0 : 14),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Circular Badge with dot
          Container(
            width: 26,
            height: 26,
            decoration: BoxDecoration(
              color: const Color(0xFFE8F5EE),
              shape: BoxShape.circle,
              border: Border.all(color: const Color(0xFFCBE3D5), width: 1.2),
            ),
            child: Center(
              child: Container(
                width: 9,
                height: 9,
                decoration: const BoxDecoration(
                  color: AppColors.primary,
                  shape: BoxShape.circle,
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),

          // Title & Description strictly from backend
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (obj.title.isNotEmpty) ...[
                  Text(
                    obj.title,
                    style: const TextStyle(
                      fontSize: 13.5,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  const SizedBox(height: 2),
                ],
                Text(
                  obj.description,
                  style: const TextStyle(
                    fontSize: 12.5,
                    color: AppColors.textSecondary,
                    height: 1.45,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
