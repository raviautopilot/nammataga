import 'package:flutter/material.dart';
import '../../../../core/constants/app_colors.dart';

/// Circular initials avatar strictly conforming to the reference image design.
/// Never displays photos or fabricated stock imagery.
class InitialsAvatar extends StatelessWidget {
  final String initials;
  final double size;
  final double fontSize;

  const InitialsAvatar({
    super.key,
    required this.initials,
    this.size = 76,
    this.fontSize = 26,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        color: const Color(0xFFFAFAF5),
        border: Border.all(color: const Color(0xFFD8F3DC), width: 2.5),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.12),
            blurRadius: 12,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Center(
        child: Text(
          initials.isNotEmpty ? initials : 'T',
          style: TextStyle(
            fontSize: fontSize,
            fontWeight: FontWeight.bold,
            color: AppColors.primaryDark,
            letterSpacing: 0.8,
          ),
        ),
      ),
    );
  }
}
