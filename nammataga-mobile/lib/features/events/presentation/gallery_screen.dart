import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../providers/events_provider.dart';
import 'widgets/event_shimmer.dart';
import 'widgets/gallery_tile.dart';

/// Tab page presenting photo memories and event archives from NammaTAGA.
/// Displays 1 image per row in an editorial feed without any cropping or black bars.
class GalleryScreen extends ConsumerWidget {
  const GalleryScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final yearsAsync = ref.watch(galleryYearsProvider);
    final selectedYear = ref.watch(selectedGalleryYearProvider);
    final imagesAsync = ref.watch(galleryImagesProvider(selectedYear));

    return RefreshIndicator(
      color: AppColors.primary,
      backgroundColor: Colors.white,
      onRefresh: () async {
        ref.invalidate(galleryYearsProvider);
        ref.invalidate(galleryImagesProvider(selectedYear));
      },
      child: CustomScrollView(
        physics: const AlwaysScrollableScrollPhysics(
          parent: BouncingScrollPhysics(),
        ),
        slivers: [
          // Year Filter Bar
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  yearsAsync.when(
                    loading: () => const SizedBox(
                      height: 38,
                      child: Center(
                        child: SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: AppColors.primary,
                          ),
                        ),
                      ),
                    ),
                    error: (e, _) => const SizedBox.shrink(),
                    data: (years) {
                      return SizedBox(
                        height: 38,
                        child: ListView(
                          scrollDirection: Axis.horizontal,
                          physics: const BouncingScrollPhysics(),
                          children: [
                            // "All" filter chip
                            _FilterChip(
                              label: 'All Memories',
                              isSelected: selectedYear == null,
                              onTap: () {
                                HapticFeedback.lightImpact();
                                ref
                                    .read(selectedGalleryYearProvider.notifier)
                                    .setYear(null);
                              },
                            ),
                            const SizedBox(width: 8),
                            // Year chips
                            ...years.map((y) {
                              return Padding(
                                padding: const EdgeInsets.only(right: 8),
                                child: _FilterChip(
                                  label: y.toString(),
                                  isSelected: selectedYear == y,
                                  onTap: () {
                                    HapticFeedback.lightImpact();
                                    ref
                                        .read(selectedGalleryYearProvider.notifier)
                                        .setYear(y);
                                  },
                                ),
                              );
                            }),
                          ],
                        ),
                      );
                    },
                  ),

                  const SizedBox(height: 14),

                  // Image count & archive status
                  imagesAsync.maybeWhen(
                    data: (images) => Row(
                      children: [
                        const Icon(
                          LucideIcons.camera,
                          size: 14,
                          color: Color(0xFF2D6A4F),
                        ),
                        const SizedBox(width: 6),
                        Text(
                          images.isEmpty
                              ? 'No Photos Found'
                              : 'Showing ${images.length} ${images.length == 1 ? 'Photograph' : 'Photographs'}${selectedYear != null ? ' from $selectedYear' : ' (All Archives)'}',
                          style: const TextStyle(
                            fontSize: 12.5,
                            fontWeight: FontWeight.w700,
                            color: Color(0xFF1B4332),
                          ),
                        ),
                      ],
                    ),
                    orElse: () => const SizedBox.shrink(),
                  ),
                ],
              ),
            ),
          ),

          // Images Feed or Status States
          imagesAsync.when(
            loading: () => const SliverFillRemaining(
              child: GalleryShimmerFeed(),
            ),
            error: (err, _) => SliverFillRemaining(
              hasScrollBody: false,
              child: Center(
                child: Padding(
                  padding: const EdgeInsets.all(24),
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
                          LucideIcons.imageOff,
                          size: 28,
                          color: Color(0xFFD4183D),
                        ),
                      ),
                      const SizedBox(height: 14),
                      const Text(
                        'Failed to load gallery photos',
                        style: TextStyle(
                          fontSize: 16.5,
                          fontWeight: FontWeight.w800,
                          color: AppColors.textPrimary,
                        ),
                      ),
                      const SizedBox(height: 6),
                      const Text(
                        'Could not retrieve photo memories from the server. Please try again.',
                        textAlign: TextAlign.center,
                        style: TextStyle(
                          fontSize: 13,
                          color: Color(0xFF6B7280),
                        ),
                      ),
                      const SizedBox(height: 18),
                      ElevatedButton.icon(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: AppColors.primary,
                          foregroundColor: Colors.white,
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12),
                          ),
                        ),
                        onPressed: () {
                          ref.invalidate(galleryImagesProvider(selectedYear));
                        },
                        icon: const Icon(LucideIcons.refreshCw, size: 16),
                        label: const Text('Retry'),
                      ),
                    ],
                  ),
                ),
              ),
            ),
            data: (images) {
              if (images.isEmpty) {
                return SliverFillRemaining(
                  hasScrollBody: false,
                  child: Padding(
                    padding: const EdgeInsets.all(28),
                    child: Center(
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 24,
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
                            ),
                          ],
                        ),
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Container(
                              width: 64,
                              height: 64,
                              decoration: BoxDecoration(
                                color: const Color(0xFFF1F8E9),
                                shape: BoxShape.circle,
                                border: Border.all(
                                  color: const Color(0xFFD8F3DC),
                                  width: 2,
                                ),
                              ),
                              child: const Icon(
                                LucideIcons.camera,
                                size: 30,
                                color: AppColors.primary,
                              ),
                            ),
                            const SizedBox(height: 16),
                            const Text(
                              'No Photos Available For This Period',
                              textAlign: TextAlign.center,
                              style: TextStyle(
                                fontSize: 17,
                                fontWeight: FontWeight.w800,
                                color: AppColors.textPrimary,
                              ),
                            ),
                            const SizedBox(height: 6),
                            const Text(
                              'Gallery photographs for this year have not yet been published or archived.',
                              textAlign: TextAlign.center,
                              style: TextStyle(
                                fontSize: 13,
                                color: Color(0xFF6B7280),
                                height: 1.45,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                );
              }

              // 1 image per row in a single continuous feed
              return SliverPadding(
                padding: const EdgeInsets.fromLTRB(16, 4, 16, 28),
                sliver: SliverList(
                  delegate: SliverChildBuilderDelegate(
                    (context, index) {
                      final img = images[index];
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 18),
                        child: GalleryTile(
                          image: img,
                          allImages: images,
                          index: index,
                        )
                            .animate()
                            .fadeIn(
                              duration: 350.ms,
                              delay: (index * 40).ms,
                            )
                            .slideY(
                              begin: 0.05,
                              end: 0,
                            ),
                      );
                    },
                    childCount: images.length,
                  ),
                ),
              );
            },
          ),
        ],
      ),
    );
  }
}

class _FilterChip extends StatelessWidget {
  final String label;
  final bool isSelected;
  final VoidCallback onTap;

  const _FilterChip({
    required this.label,
    required this.isSelected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        decoration: BoxDecoration(
          color: isSelected ? AppColors.primary : Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: isSelected ? AppColors.primary : const Color(0xFFD8F3DC),
            width: 1.1,
          ),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: AppColors.primary.withValues(alpha: 0.28),
                    blurRadius: 8,
                    offset: const Offset(0, 3),
                  ),
                ]
              : null,
        ),
        child: Center(
          child: Text(
            label,
            style: TextStyle(
              fontSize: 12.5,
              fontWeight: isSelected ? FontWeight.w700 : FontWeight.w600,
              color: isSelected ? Colors.white : AppColors.textPrimary,
            ),
          ),
        ),
      ),
    );
  }
}
