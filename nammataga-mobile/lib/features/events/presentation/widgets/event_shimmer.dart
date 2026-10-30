import 'package:flutter/material.dart';

/// Shimmer skeleton loading effect matching NammaTAGA aesthetic.
class ShimmerLoadingBox extends StatefulWidget {
  final double width;
  final double height;
  final double borderRadius;

  const ShimmerLoadingBox({
    super.key,
    required this.width,
    required this.height,
    this.borderRadius = 12,
  });

  @override
  State<ShimmerLoadingBox> createState() => _ShimmerLoadingBoxState();
}

class _ShimmerLoadingBoxState extends State<ShimmerLoadingBox>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1400),
    )..repeat();
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
        return Container(
          width: widget.width,
          height: widget.height,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(widget.borderRadius),
            gradient: LinearGradient(
              begin: Alignment(-1.0 + 2.0 * _controller.value, -0.3),
              end: Alignment(1.0 + 2.0 * _controller.value, 0.3),
              colors: const [
                Color(0xFFE8EFEA),
                Color(0xFFF6FAF7),
                Color(0xFFE8EFEA),
              ],
            ),
          ),
        );
      },
    );
  }
}

/// Shimmer placeholder list for upcoming events cards.
class UpcomingEventsShimmer extends StatelessWidget {
  const UpcomingEventsShimmer({super.key});

  @override
  Widget build(BuildContext context) {
    return ListView.separated(
      physics: const NeverScrollableScrollPhysics(),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
      itemCount: 3,
      separatorBuilder: (context, index) => const SizedBox(height: 16),
      itemBuilder: (context, index) {
        return Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const ShimmerLoadingBox(width: 64, height: 68, borderRadius: 14),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: const [
                        ShimmerLoadingBox(width: 100, height: 16, borderRadius: 6),
                        SizedBox(height: 8),
                        ShimmerLoadingBox(width: double.infinity, height: 20, borderRadius: 6),
                        SizedBox(height: 6),
                        ShimmerLoadingBox(width: 180, height: 14, borderRadius: 6),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              const Divider(color: Color(0xFFF1F5F2), height: 1),
              const SizedBox(height: 12),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: const [
                  ShimmerLoadingBox(width: 120, height: 18, borderRadius: 6),
                  ShimmerLoadingBox(width: 80, height: 32, borderRadius: 10),
                ],
              ),
            ],
          ),
        );
      },
    );
  }
}

/// Shimmer placeholder for the single-column photo gallery feed.
class GalleryShimmerFeed extends StatelessWidget {
  const GalleryShimmerFeed({super.key});

  @override
  Widget build(BuildContext context) {
    return ListView.separated(
      physics: const NeverScrollableScrollPhysics(),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      itemCount: 2,
      separatorBuilder: (context, index) => const SizedBox(height: 18),
      itemBuilder: (context, index) {
        return Container(
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(22),
            border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: const [
                    ShimmerLoadingBox(width: 52, height: 22, borderRadius: 8),
                    SizedBox(width: 8),
                    ShimmerLoadingBox(width: 160, height: 22, borderRadius: 8),
                  ],
                ),
              ),
              const Padding(
                padding: EdgeInsets.symmetric(horizontal: 12),
                child: ShimmerLoadingBox(
                  width: double.infinity,
                  height: 220,
                  borderRadius: 16,
                ),
              ),
              Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: const [
                    ShimmerLoadingBox(width: double.infinity, height: 18, borderRadius: 6),
                    SizedBox(height: 8),
                    ShimmerLoadingBox(width: 180, height: 14, borderRadius: 6),
                  ],
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}

/// Alias for backwards compatibility with tests
typedef GalleryShimmerGrid = GalleryShimmerFeed;
