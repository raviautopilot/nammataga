import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../domain/gallery_image_model.dart';
import 'gallery_lightbox.dart';

/// Editorial photo card showcasing a full, uncropped photograph without any black bars.
/// Displays 1 image per row with clean white card aesthetics, event tags, and tap-to-expand.
class GalleryTile extends StatelessWidget {
  final GalleryImageModel image;
  final List<GalleryImageModel> allImages;
  final int index;

  const GalleryTile({
    super.key,
    required this.image,
    required this.allImages,
    required this.index,
  });

  @override
  Widget build(BuildContext context) {
    return BouncyButton(
      onTap: () {
        HapticFeedback.lightImpact();
        GalleryLightbox.open(
          context,
          images: allImages,
          initialIndex: index,
        );
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
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Top Bar: Clean Year Badge and Fullscreen Expand Icon
            Padding(
              padding: const EdgeInsets.fromLTRB(14, 12, 14, 10),
              child: Row(
                children: [
                  // Year Badge
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 4.5,
                    ),
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        colors: [
                          Color(0xFF0D2C1B),
                          Color(0xFF1B4332),
                        ],
                      ),
                      borderRadius: BorderRadius.circular(8),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0xFF1B4332).withValues(alpha: 0.18),
                          blurRadius: 6,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(
                          LucideIcons.calendar,
                          size: 12,
                          color: Color(0xFFB7E4C7),
                        ),
                        const SizedBox(width: 5),
                        Text(
                          image.year.toString(),
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 11.5,
                            fontWeight: FontWeight.w800,
                            letterSpacing: 0.5,
                          ),
                        ),
                      ],
                    ),
                  ),

                  const Spacer(),

                  // Fullscreen Action Hint Circle
                  Container(
                    width: 32,
                    height: 32,
                    decoration: BoxDecoration(
                      color: const Color(0xFFF1F8E9),
                      shape: BoxShape.circle,
                      border: Border.all(
                        color: const Color(0xFFD8F3DC),
                        width: 1.0,
                      ),
                    ),
                    child: const Center(
                      child: Icon(
                        LucideIcons.expand,
                        size: 14,
                        color: AppColors.primary,
                      ),
                    ),
                  ),
                ],
              ),
            ),

            // Full, Uncropped Photograph (Natural aspect ratio, no cropping)
            Padding(
              padding: const EdgeInsets.fromLTRB(10, 0, 10, 10),
              child: ClipRRect(
                borderRadius: BorderRadius.circular(16),
                child: Hero(
                  tag: 'gallery_image_${image.id}',
                  child: image.fullImageUrl.isNotEmpty
                      ? CachedNetworkImage(
                          imageUrl: image.fullImageUrl,
                          width: double.infinity,
                          fit: BoxFit.fitWidth, // Natural proportion, 100% visible
                          placeholder: (context, url) => Container(
                            height: 220,
                            decoration: BoxDecoration(
                              color: const Color(0xFFF1F8E9),
                              borderRadius: BorderRadius.circular(16),
                            ),
                            child: const Center(
                              child: SizedBox(
                                width: 26,
                                height: 26,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                  color: AppColors.primary,
                                ),
                              ),
                            ),
                          ),
                          errorWidget: (context, url, error) => Container(
                            height: 200,
                            decoration: BoxDecoration(
                              color: const Color(0xFFF6FAF6),
                              borderRadius: BorderRadius.circular(16),
                              border: Border.all(
                                color: const Color(0xFFE2EBE5),
                              ),
                            ),
                            child: Center(
                              child: Column(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  Container(
                                    width: 44,
                                    height: 44,
                                    decoration: const BoxDecoration(
                                      color: Color(0xFFE8F5E9),
                                      shape: BoxShape.circle,
                                    ),
                                    child: const Icon(
                                      LucideIcons.image,
                                      size: 22,
                                      color: AppColors.primary,
                                    ),
                                  ),
                                  const SizedBox(height: 10),
                                  const Text(
                                    'TAGA Archival Photograph',
                                    style: TextStyle(
                                      color: AppColors.primaryDark,
                                      fontSize: 13,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                  const SizedBox(height: 3),
                                  const Text(
                                    'Tap to preview image',
                                    style: TextStyle(
                                      color: Color(0xFF6B7280),
                                      fontSize: 11.5,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        )
                      : Container(
                          height: 200,
                          decoration: BoxDecoration(
                            color: const Color(0xFFF6FAF6),
                            borderRadius: BorderRadius.circular(16),
                          ),
                          child: const Center(
                            child: Icon(
                              LucideIcons.image,
                              size: 38,
                              color: Color(0xFF52796F),
                            ),
                          ),
                        ),
                ),
              ),
            ),

            // Photo Description & Action Section
            Container(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 14),
              decoration: const BoxDecoration(
                color: Color(0xFFFAFFFE),
                border: Border(
                  top: BorderSide(
                    color: Color(0xFFF1F5F2),
                    width: 1.0,
                  ),
                ),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    image.displayDescription,
                    style: const TextStyle(
                      fontSize: 14.5,
                      fontWeight: FontWeight.w700,
                      color: AppColors.textPrimary,
                      height: 1.35,
                      letterSpacing: -0.2,
                    ),
                  ),
                  const SizedBox(height: 10),
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 3.5,
                        ),
                        decoration: BoxDecoration(
                          color: const Color(0xFFE8F5E9),
                          borderRadius: BorderRadius.circular(6),
                          border: Border.all(
                            color: const Color(0xFFB7E4C7),
                            width: 0.8,
                          ),
                        ),
                        child: Text(
                          '${image.year} Memory',
                          style: const TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w700,
                            color: Color(0xFF1B4332),
                          ),
                        ),
                      ),
                      const Spacer(),
                      Row(
                        mainAxisSize: MainAxisSize.min,
                        children: const [
                          Text(
                            'Full View',
                            style: TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.w700,
                              color: AppColors.primary,
                            ),
                          ),
                          SizedBox(width: 4),
                          Icon(
                            LucideIcons.arrowUpRight,
                            size: 14,
                            color: AppColors.primary,
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
