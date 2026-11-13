import 'package:flutter/material.dart';

/// Shimmer placeholder for State Executive cards and District directory
class OfficeBearersShimmer extends StatefulWidget {
  final bool isDistrict;

  const OfficeBearersShimmer({
    super.key,
    this.isDistrict = false,
  });

  @override
  State<OfficeBearersShimmer> createState() => _OfficeBearersShimmerState();
}

class _OfficeBearersShimmerState extends State<OfficeBearersShimmer>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;

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
        final opacity = 0.4 + (_controller.value * 0.4);
        return widget.isDistrict
            ? _buildDistrictDirectoryShimmer(opacity)
            : _buildStateExecutiveShimmer(opacity);
      },
    );
  }

  Widget _buildStateExecutiveShimmer(double opacity) {
    return ListView.separated(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
      itemCount: 4,
      separatorBuilder: (_, _) => const SizedBox(height: 14),
      itemBuilder: (context, index) {
        return Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: const Color(0xFFE2EBE5)),
          ),
          child: Row(
            children: [
              Container(
                width: 92,
                height: 122,
                decoration: BoxDecoration(
                  color: Colors.grey.withValues(alpha: opacity * 0.3),
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Container(
                      width: 140,
                      height: 16,
                      decoration: BoxDecoration(
                        color: Colors.grey.withValues(alpha: opacity * 0.3),
                        borderRadius: BorderRadius.circular(4),
                      ),
                    ),
                    const SizedBox(height: 8),
                    Container(
                      width: 100,
                      height: 12,
                      decoration: BoxDecoration(
                        color: Colors.grey.withValues(alpha: opacity * 0.2),
                        borderRadius: BorderRadius.circular(4),
                      ),
                    ),
                    const SizedBox(height: 14),
                    Container(
                      width: 120,
                      height: 10,
                      decoration: BoxDecoration(
                        color: Colors.grey.withValues(alpha: opacity * 0.2),
                        borderRadius: BorderRadius.circular(4),
                      ),
                    ),
                    const SizedBox(height: 6),
                    Container(
                      width: 80,
                      height: 10,
                      decoration: BoxDecoration(
                        color: Colors.grey.withValues(alpha: opacity * 0.2),
                        borderRadius: BorderRadius.circular(4),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildDistrictDirectoryShimmer(double opacity) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFE2EBE5)),
      ),
      child: Column(
        children: List.generate(
          5,
          (index) => Container(
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
            decoration: BoxDecoration(
              border: index == 4
                  ? null
                  : const Border(
                      bottom: BorderSide(color: Color(0xFFEFF3F0)),
                    ),
            ),
            child: Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Container(
                        width: 130,
                        height: 14,
                        decoration: BoxDecoration(
                          color: Colors.grey.withValues(alpha: opacity * 0.3),
                          borderRadius: BorderRadius.circular(4),
                        ),
                      ),
                      const SizedBox(height: 6),
                      Container(
                        width: 80,
                        height: 11,
                        decoration: BoxDecoration(
                          color: Colors.grey.withValues(alpha: opacity * 0.2),
                          borderRadius: BorderRadius.circular(4),
                        ),
                      ),
                    ],
                  ),
                ),
                Container(
                  width: 80,
                  height: 24,
                  decoration: BoxDecoration(
                    color: Colors.grey.withValues(alpha: opacity * 0.2),
                    borderRadius: BorderRadius.circular(6),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
