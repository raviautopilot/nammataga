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
    this.borderRadius = 8,
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

/// Shimmer placeholder row for category chips.
class CategoryChipsShimmer extends StatelessWidget {
  const CategoryChipsShimmer({super.key});

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 48,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        physics: const NeverScrollableScrollPhysics(),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        itemCount: 4,
        separatorBuilder: (context, index) => const SizedBox(width: 8),
        itemBuilder: (context, index) {
          final widths = [130.0, 150.0, 110.0, 140.0];
          return ShimmerLoadingBox(
            width: widths[index % widths.length],
            height: 32,
            borderRadius: 20,
          );
        },
      ),
    );
  }
}

/// Shimmer placeholder for a single full-width category row.
class CategoryCardShimmer extends StatelessWidget {
  const CategoryCardShimmer({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 15),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: const Color(0xFFE5E7EB),
          width: 1.0,
        ),
      ),
      child: const Row(
        children: [
          ShimmerLoadingBox(width: 44, height: 44, borderRadius: 10),
          SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                ShimmerLoadingBox(width: 160, height: 16, borderRadius: 4),
                SizedBox(height: 6),
                ShimmerLoadingBox(width: 90, height: 12, borderRadius: 4),
              ],
            ),
          ),
          SizedBox(width: 8),
          ShimmerLoadingBox(width: 16, height: 16, borderRadius: 8),
        ],
      ),
    );
  }
}

/// Shimmer placeholder list for Category Directory landing page.
class CategoryDirectoryShimmer extends StatelessWidget {
  final int itemCount;

  const CategoryDirectoryShimmer({
    super.key,
    this.itemCount = 7,
  });

  @override
  Widget build(BuildContext context) {
    return ListView.separated(
      physics: const NeverScrollableScrollPhysics(),
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
      itemCount: itemCount,
      separatorBuilder: (context, index) => const SizedBox(height: 10),
      itemBuilder: (context, index) => const CategoryCardShimmer(),
    );
  }
}

/// Shimmer placeholder for a single document card.
class DocumentCardShimmer extends StatelessWidget {
  const DocumentCardShimmer({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: const Color(0xFFE2EBE5),
          width: 1.1,
        ),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Left PDF badge placeholder
          const ShimmerLoadingBox(
            width: 44,
            height: 48,
            borderRadius: 12,
          ),
          const SizedBox(width: 12),

          // Middle column details
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const ShimmerLoadingBox(
                  width: double.infinity,
                  height: 16,
                  borderRadius: 6,
                ),
                const SizedBox(height: 6),
                const ShimmerLoadingBox(
                  width: 180,
                  height: 14,
                  borderRadius: 6,
                ),
                const SizedBox(height: 10),
                Row(
                  children: const [
                    ShimmerLoadingBox(
                      width: 58,
                      height: 20,
                      borderRadius: 10,
                    ),
                    SizedBox(width: 6),
                    ShimmerLoadingBox(
                      width: 64,
                      height: 20,
                      borderRadius: 10,
                    ),
                    SizedBox(width: 6),
                    ShimmerLoadingBox(
                      width: 76,
                      height: 20,
                      borderRadius: 10,
                    ),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(width: 10),

          // Right action button placeholder
          const ShimmerLoadingBox(
            width: 32,
            height: 32,
            borderRadius: 8,
          ),
        ],
      ),
    );
  }
}

/// Shimmer placeholder list for documents view.
class ResourcesShimmer extends StatelessWidget {
  final int itemCount;
  final bool includeChips;

  const ResourcesShimmer({
    super.key,
    this.itemCount = 5,
    this.includeChips = false,
  });

  @override
  Widget build(BuildContext context) {
    final list = ListView.separated(
      physics: const NeverScrollableScrollPhysics(),
      shrinkWrap: true,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      itemCount: itemCount,
      separatorBuilder: (context, index) => const SizedBox(height: 12),
      itemBuilder: (context, index) => const DocumentCardShimmer(),
    );

    if (!includeChips) {
      return list;
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const CategoryChipsShimmer(),
        const SizedBox(height: 8),
        Expanded(child: list),
      ],
    );
  }
}
