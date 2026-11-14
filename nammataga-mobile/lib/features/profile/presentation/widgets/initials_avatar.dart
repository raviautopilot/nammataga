import 'package:flutter/material.dart';

/// Initials badge/monogram avatar with a modern squircle (rounded rectangular)
/// shape to replace the circular avatar format as requested.
class InitialsAvatar extends StatelessWidget {
  final String initials;
  final double size;
  final double fontSize;
  final double borderRadius;

  const InitialsAvatar({
    super.key,
    required this.initials,
    this.size = 80,
    this.fontSize = 26,
    this.borderRadius = 22,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(borderRadius),
        color: const Color(0xFFFAFAF5),
        border: Border.all(
          color: const Color(0xFFD8F3DC),
          width: 3,
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.16),
            blurRadius: 14,
            offset: const Offset(0, 6),
          ),
          BoxShadow(
            color: const Color(0xFF52B788).withValues(alpha: 0.22),
            blurRadius: 20,
            spreadRadius: 2,
          ),
        ],
      ),
      child: Center(
        child: Text(
          initials.isNotEmpty ? initials.toUpperCase() : 'T',
          style: TextStyle(
            fontSize: fontSize,
            fontWeight: FontWeight.bold,
            color: const Color(0xFF092014),
            letterSpacing: 1.2,
          ),
        ),
      ),
    );
  }
}
