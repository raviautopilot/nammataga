import 'package:flutter/material.dart';

/// App color palette strictly matching Namma TAGA Web (`taga-web/src/styles/globals.css`)
class AppColors {
  AppColors._();

  // Brand Primaries
  static const Color primary = Color(0xFF2D6A4F);
  static const Color primaryDark = Color(0xFF1B4332);
  static const Color primaryLight = Color(0xFF40916C);

  // Backgrounds & Surfaces
  static const Color background = Color(0xFFFAFFFE);
  static const Color surface = Color(0xFFFFFFFF);
  static const Color card = Color(0xFFFFFFFF);

  // Text / Foregrounds
  static const Color textPrimary = Color(0xFF1B4332);
  static const Color textSecondary = Color(0xFF52796F);
  static const Color textMuted = Color(0xFF52796F);
  static const Color textOnPrimary = Color(0xFFFFFFFF);

  // Mint Highlights & Accents
  static const Color secondary = Color(0xFFD8F3DC);
  static const Color secondaryDark = Color(0xFF95D5B2);
  static const Color accent = Color(0xFFB7E4C7);
  static const Color ring = Color(0xFF74C69D);

  // Input & Muted Containers
  static const Color muted = Color(0xFFF1F8E9);
  static const Color inputBackground = Color(0xFFF1F8E9);

  // Borders & Dividers
  static const Color border = Color(0x262D6A4F); // rgba(45, 106, 79, 0.15)
  static const Color borderLight = Color(0x1A2D6A4F);

  // Status Colors
  static const Color destructive = Color(0xFFD4183D);
  static const Color success = Color(0xFF2D6A4F);
  static const Color warning = Color(0xFFE76F51);
  static const Color info = Color(0xFF2A9D8F);

  // Gradient definitions
  static const List<Color> brandGradient = [
    Color(0xFF1B4332),
    Color(0xFF2D6A4F),
  ];

  static const List<Color> cardGradient = [
    Color(0xFFFAFFFE),
    Color(0xFFF1F8E9),
  ];

  static const List<Color> tierGradients = [
    Color(0xFF2D6A4F),
    Color(0xFF40916C),
    Color(0xFF52B788),
    Color(0xFF74C69D),
    Color(0xFF95D5B2),
  ];
}
