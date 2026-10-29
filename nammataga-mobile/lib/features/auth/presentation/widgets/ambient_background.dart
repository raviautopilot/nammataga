import 'dart:math' as math;
import 'package:flutter/material.dart';
import '../../../../core/constants/app_assets.dart';
import '../../../../core/constants/app_colors.dart';

/// Atmospheric canvas backdrop with continuously animated emerald radial spotlights
/// and a delicate watermark seal of the official association emblem.
class AmbientBackground extends StatefulWidget {
  final Widget child;

  const AmbientBackground({super.key, required this.child});

  @override
  State<AmbientBackground> createState() => _AmbientBackgroundState();
}

class _AmbientBackgroundState extends State<AmbientBackground>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 12),
    )..repeat(reverse: true);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _controller,
      builder: (context, child) {
        final progress = _controller.value;
        final dxTop = math.sin(progress * 2 * math.pi) * 18.0;
        final dyTop = math.cos(progress * 2 * math.pi) * 14.0;
        final dxBottom = math.cos(progress * 2 * math.pi) * 16.0;
        final dyBottom = math.sin(progress * 2 * math.pi) * 12.0;
        final pulseScale = 1.0 + (progress - 0.5) * 0.12;

        return Container(
          color: AppColors.background,
          child: Stack(
            children: [
              // Top-right soft emerald radiance drifting
              Positioned(
                top: -120 + dyTop,
                right: -100 + dxTop,
                child: Transform.scale(
                  scale: pulseScale,
                  child: Container(
                    width: 340,
                    height: 340,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      gradient: RadialGradient(
                        colors: [
                          AppColors.secondary.withValues(
                            alpha: 0.65 + progress * 0.15,
                          ),
                          AppColors.accent.withValues(alpha: 0.2),
                          Colors.transparent,
                        ],
                      ),
                    ),
                  ),
                ),
              ),

              // Bottom-left soft sage glow drifting
              Positioned(
                bottom: -100 + dyBottom,
                left: -80 + dxBottom,
                child: Transform.scale(
                  scale: 2.0 - pulseScale,
                  child: Container(
                    width: 300,
                    height: 300,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      gradient: RadialGradient(
                        colors: [
                          AppColors.accent.withValues(
                            alpha: 0.5 + (1 - progress) * 0.15,
                          ),
                          AppColors.secondary.withValues(alpha: 0.15),
                          Colors.transparent,
                        ],
                      ),
                    ),
                  ),
                ),
              ),

              // Floating subtle ambient light orb
              Positioned(
                top: 260 + dyBottom * 1.5,
                left: 30 + dxTop * 0.8,
                child: Container(
                  width: 140,
                  height: 140,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    gradient: RadialGradient(
                      colors: [
                        const Color(
                          0xFF52B788,
                        ).withValues(alpha: 0.08 + progress * 0.04),
                        Colors.transparent,
                      ],
                    ),
                  ),
                ),
              ),

              // Atmospheric center watermark seal with subtle breathing
              Center(
                child: Transform.scale(
                  scale: 0.98 + progress * 0.04,
                  child: Opacity(
                    opacity: 0.03 + progress * 0.015,
                    child: Image.asset(
                      AppAssets.logo,
                      width: 340,
                      height: 340,
                      fit: BoxFit.contain,
                      errorBuilder: (context, error, stackTrace) => Image.asset(
                        AppAssets.logo1,
                        width: 340,
                        height: 340,
                        fit: BoxFit.contain,
                        errorBuilder: (ctx, err, stack) =>
                            const SizedBox.shrink(),
                      ),
                    ),
                  ),
                ),
              ),

              // Forefront child
              Positioned.fill(child: widget.child),
            ],
          ),
        );
      },
    );
  }
}
