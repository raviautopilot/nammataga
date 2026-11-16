import 'dart:ui';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

/// Improvised Toast / Floating Capsule matching Option A (Dynamic Pill style).
///
/// Features:
/// - Floating capsule layout with high-depth soft blur glassmorphism
/// - Glowing circular accent micro-badge (Success, Error, Warning, Info, Exit)
/// - Crisp typography with optional title and structured description
/// - Built-in haptic feedback and optional interactive trailing actions
class TagaToast {
  TagaToast._();

  /// Displays a customized floating capsule toast.
  static void show(
    BuildContext context, {
    required String message,
    String? title,
    IconData? icon,
    Color? accentColor,
    Color? glowColor,
    Duration duration = const Duration(seconds: 4),
    Widget? action,
  }) {
    final messenger = ScaffoldMessenger.maybeOf(context);
    if (messenger == null) return;

    messenger.hideCurrentSnackBar();

    final effectiveAccent = accentColor ?? const Color(0xFF2D6A4F);
    final effectiveGlow = glowColor ?? const Color(0xFF52B788);

    messenger.showSnackBar(
      SnackBar(
        behavior: SnackBarBehavior.floating,
        backgroundColor: Colors.transparent,
        elevation: 0,
        margin: const EdgeInsets.symmetric(horizontal: 18, vertical: 14),
        padding: EdgeInsets.zero,
        duration: duration,
        content: ClipRRect(
          borderRadius: BorderRadius.circular(50),
          child: BackdropFilter(
            filter: ImageFilter.blur(sigmaX: 12, sigmaY: 12),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 11),
              decoration: BoxDecoration(
                color: const Color(0xFF071B12).withValues(alpha: 0.92),
                borderRadius: BorderRadius.circular(50),
                border: Border.all(
                  color: Colors.white.withValues(alpha: 0.16),
                  width: 1.1,
                ),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.35),
                    blurRadius: 24,
                    offset: const Offset(0, 8),
                  ),
                ],
              ),
              child: Row(
                children: [
                  // Glowing circular micro-badge
                  Container(
                    width: 32,
                    height: 32,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: effectiveAccent.withValues(alpha: 0.28),
                      border: Border.all(
                        color: effectiveGlow.withValues(alpha: 0.65),
                        width: 1.2,
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: effectiveGlow.withValues(alpha: 0.32),
                          blurRadius: 10,
                        ),
                      ],
                    ),
                    child: Center(
                      child: Icon(
                        icon ?? LucideIcons.info,
                        color: Colors.white,
                        size: 16,
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),

                  // Content text
                  Expanded(
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        if (title != null && title.isNotEmpty)
                          Padding(
                            padding: const EdgeInsets.only(bottom: 2),
                            child: Text(
                              title,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(
                                fontFamily: 'Roboto',
                                fontSize: 13,
                                fontWeight: FontWeight.w700,
                                color: Colors.white,
                                letterSpacing: -0.2,
                              ),
                            ),
                          ),
                        Text(
                          message,
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(
                            fontFamily: 'Roboto',
                            fontSize: 12.5,
                            fontWeight: FontWeight.w500,
                            color: Colors.white.withValues(alpha: 0.92),
                            height: 1.25,
                          ),
                        ),
                      ],
                    ),
                  ),

                  // Optional inline action
                  if (action != null) ...[
                    const SizedBox(width: 8),
                    action,
                  ],
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  /// Success floating capsule with deep emerald styling.
  static void success(
    BuildContext context,
    String message, {
    String? title,
    Duration duration = const Duration(seconds: 3),
    Widget? action,
  }) {
    HapticFeedback.lightImpact();
    show(
      context,
      message: message,
      title: title,
      icon: LucideIcons.check,
      accentColor: const Color(0xFF1B4332),
      glowColor: const Color(0xFF48BB78),
      duration: duration,
      action: action,
    );
  }

  /// Error floating capsule with rich crimson styling.
  static void error(
    BuildContext context,
    String message, {
    String? title,
    Duration duration = const Duration(seconds: 4),
    Widget? action,
  }) {
    HapticFeedback.mediumImpact();
    show(
      context,
      message: message,
      title: title ?? 'Error',
      icon: LucideIcons.alertCircle,
      accentColor: const Color(0xFF991B1B),
      glowColor: const Color(0xFFF87171),
      duration: duration,
      action: action,
    );
  }

  /// Warning / Caution floating capsule with honey amber styling.
  static void warning(
    BuildContext context,
    String message, {
    String? title,
    Duration duration = const Duration(seconds: 4),
    Widget? action,
  }) {
    HapticFeedback.lightImpact();
    show(
      context,
      message: message,
      title: title,
      icon: LucideIcons.alertTriangle,
      accentColor: const Color(0xFF92400E),
      glowColor: const Color(0xFFFBBF24),
      duration: duration,
      action: action,
    );
  }

  /// Info / System notification floating capsule.
  static void info(
    BuildContext context,
    String message, {
    String? title,
    Duration duration = const Duration(seconds: 3),
    Widget? action,
  }) {
    HapticFeedback.lightImpact();
    show(
      context,
      message: message,
      title: title,
      icon: LucideIcons.info,
      accentColor: const Color(0xFF1E3A8A),
      glowColor: const Color(0xFF60A5FA),
      duration: duration,
      action: action,
    );
  }

  /// App Exit notice toast with 5-second lifespan and optional inline Exit button.
  static void exitNotice(
    BuildContext context, {
    Duration duration = const Duration(seconds: 5),
    VoidCallback? onExitTapped,
  }) {
    show(
      context,
      title: 'Exit Application?',
      message: 'Press back or swipe again to exit',
      icon: LucideIcons.undo2,
      accentColor: const Color(0xFF0F3D2E),
      glowColor: const Color(0xFF52B788),
      duration: duration,
      action: onExitTapped != null
          ? GestureDetector(
              onTap: onExitTapped,
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    colors: [Color(0xFFE5383B), Color(0xFFBA181B)],
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                  ),
                  borderRadius: BorderRadius.circular(20),
                  boxShadow: [
                    BoxShadow(
                      color: const Color(0xFFBA181B).withValues(alpha: 0.35),
                      blurRadius: 6,
                      offset: const Offset(0, 2),
                    ),
                  ],
                ),
                child: const Text(
                  'EXIT',
                  style: TextStyle(
                    fontFamily: 'Roboto',
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                    letterSpacing: 0.5,
                  ),
                ),
              ),
            )
          : null,
    );
  }
}
