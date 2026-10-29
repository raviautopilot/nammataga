import 'package:flutter/material.dart';

/// Circular initials avatar strictly conforming to the reference image design.
/// Features a cream surface, dual border effect with gold/mint accents,
/// and bold dark forest green initials. Never displays stock photos or camera uploads.
class InitialsAvatar extends StatelessWidget {
  final String initials;
  final double size;
  final double fontSize;

  const InitialsAvatar({
    super.key,
    required this.initials,
    this.size = 84,
    this.fontSize = 28,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        color: const Color(0xFFFAFAF5),
        border: Border.all(
          color: const Color(0xFFD8F3DC),
          width: 3.5,
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.16),
            blurRadius: 14,
            offset: const Offset(0, 6),
          ),
          BoxShadow(
            color: const Color(0xFF52B788).withValues(alpha: 0.2),
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
