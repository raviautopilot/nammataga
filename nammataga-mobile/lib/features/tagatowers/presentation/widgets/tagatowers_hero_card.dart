import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_assets.dart';
import '../theme/tagatowers_design_tokens.dart';

/// An immersive, high-end hero card showcasing the TAGA Towers building banner,
/// an emerald gradient overlay, typography, and Chennai location pill.
class TagaTowersHeroCard extends StatelessWidget {
  /// Custom image URL or asset if overriding the default [AppAssets.tagatower]
  final String? bannerImageUrl;

  /// Optional card height. Defaults to natural aspect ratio if not specified.
  final double? height;

  /// Whether to display the 24-hour check-in/out badge
  final bool showTimingBadge;

  /// Whether to display the Chennai location pill
  final bool showLocationPill;

  /// Optional custom title
  final String title;

  /// Optional custom subtitle
  final String subtitle;

  /// Optional card tap callback
  final VoidCallback? onTap;

  /// Optional trailing widget in the top badge row
  final Widget? topTrailing;

  const TagaTowersHeroCard({
    super.key,
    this.bannerImageUrl,
    this.height,
    this.showTimingBadge = false,
    this.showLocationPill = true,
    this.title = 'TAGA Towers',
    this.subtitle =
        'Comfortable accommodation facility for TAGA members in Chennai',
    this.onTap,
    this.topTrailing,
  });

  @override
  Widget build(BuildContext context) {
    final cardContent = RepaintBoundary(
      child: Stack(
        fit: StackFit.expand,
        children: [
          // 1. Background Image - Direct local high-performance asset with memory-efficient cache dimensions
          _buildBannerImage(context, bannerImageUrl),

          // 2. Subtle Vignette Overlay: Leaves center building 100% clear & bright
          Container(
            decoration: BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: [
                  const Color(0xFF061A0F).withValues(alpha: 0.35),
                  Colors.transparent,
                  Colors.transparent,
                  const Color(0xFF061A0F).withValues(alpha: 0.75),
                ],
                stops: const [0.0, 0.20, 0.65, 1.0],
              ),
            ),
          ),

          // 3. Foreground Card Content
          Padding(
            padding: const EdgeInsets.all(16.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                // Top Pill Row: Location & Optional Trailing
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    if (showLocationPill)
                      _buildPill(
                        icon: LucideIcons.mapPin,
                        label: 'Egmore, Chennai',
                      )
                    else
                      const SizedBox.shrink(),
                    if (topTrailing != null)
                      topTrailing!
                    else
                      _buildPill(
                        icon: LucideIcons.building2,
                        label: 'Velanmai Illam',
                      ),
                  ],
                ),

                // Bottom Content: Title & Subtitle
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      title,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 26,
                        fontWeight: FontWeight.w800,
                        letterSpacing: -0.5,
                        shadows: [
                          Shadow(
                            color: Color(0xCC000000),
                            blurRadius: 8,
                            offset: Offset(0, 2),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      subtitle,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                        color: Colors.white.withValues(alpha: 0.95),
                        fontSize: 12.5,
                        height: 1.35,
                        fontWeight: FontWeight.w500,
                        shadows: const [
                          Shadow(
                            color: Color(0x99000000),
                            blurRadius: 6,
                            offset: Offset(0, 1),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );

    Widget cardWrapper;
    if (height != null) {
      cardWrapper = Container(
        height: height,
        width: double.infinity,
        decoration: BoxDecoration(
          color: TagaTowersDesignTokens.deepEmerald,
          borderRadius: BorderRadius.circular(20),
          boxShadow: TagaTowersDesignTokens.elevatedShadow,
        ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(20),
          child: cardContent,
        ),
      );
    } else {
      // Natural 975 / 1300 aspect ratio (0.75) to show the FULL VIEW of the TAGA Tower building
      cardWrapper = Container(
        width: double.infinity,
        decoration: BoxDecoration(
          color: TagaTowersDesignTokens.deepEmerald,
          borderRadius: BorderRadius.circular(20),
          boxShadow: TagaTowersDesignTokens.elevatedShadow,
        ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(20),
          child: AspectRatio(
            aspectRatio: 975 / 1300,
            child: cardContent,
          ),
        ),
      );
    }

    return Semantics(
      label: '$title. $subtitle',
      child: GestureDetector(
        onTap: onTap,
        child: cardWrapper,
      ),
    );
  }

  /// Builds local asset hero banner directly with screen-size aware cache bounds
  Widget _buildBannerImage(BuildContext context, String? overrideUrl) {
    if (overrideUrl != null && overrideUrl.startsWith('http')) {
      return CachedNetworkImage(
        imageUrl: overrideUrl,
        fit: BoxFit.cover,
        placeholder: (context, _) => Container(
          color: TagaTowersDesignTokens.deepEmerald,
          child: Center(
            child: Icon(
              LucideIcons.building2,
              size: 36,
              color: Colors.white.withValues(alpha: 0.4),
            ),
          ),
        ),
        errorWidget: (context, url, error) => Image.asset(
          AppAssets.tagatower,
          fit: BoxFit.cover,
        ),
      );
    }

    // Determine target pixel cache size based on device layout to eliminate memory bloat & raster lag
    final mediaQuery = MediaQuery.maybeOf(context);
    final screenWidth = mediaQuery?.size.width ?? 400.0;
    final pixelRatio = mediaQuery?.devicePixelRatio ?? 2.0;
    // Cap decoding to max 580dp chassis width * DPR to prevent excessive GPU decoding work
    final displayWidth = screenWidth > 580 ? 580.0 : screenWidth;
    final targetPixelWidth = (displayWidth * pixelRatio).round().clamp(600, 1200);

    return Image.asset(
      AppAssets.tagatower,
      fit: BoxFit.cover,
      cacheWidth: targetPixelWidth,
      filterQuality: FilterQuality.medium,
      errorBuilder: (context, error, stackTrace) => Container(
        color: TagaTowersDesignTokens.deepEmerald,
        child: const Center(
          child: Icon(
            LucideIcons.building2,
            size: 48,
            color: TagaTowersDesignTokens.softMint,
          ),
        ),
      ),
    );
  }

  /// High-performance pill badge avoiding GPU BackdropFilter shader cost
  Widget _buildPill({
    required IconData icon,
    required String label,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      decoration: BoxDecoration(
        color: const Color(0xCC000000), // Solid high-contrast dark badge (80% opacity)
        borderRadius: BorderRadius.circular(999),
        border: Border.all(
          color: Colors.white.withValues(alpha: 0.35),
          width: 1,
        ),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            icon,
            size: 13,
            color: TagaTowersDesignTokens.softMint,
          ),
          const SizedBox(width: 5),
          Text(
            label,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 11.5,
              fontWeight: FontWeight.w600,
              letterSpacing: -0.1,
            ),
          ),
        ],
      ),
    );
  }
}
