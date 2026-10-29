import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import '../constants/api_endpoints.dart';
import '../constants/app_colors.dart';

/// Wraps the mobile app in an authentic 19.5:9 smartphone frame when previewed
/// on web or desktop browsers, providing an accurate mobile UI simulation.
class WebPhoneFrame extends StatelessWidget {
  final Widget child;

  const WebPhoneFrame({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    if (!kIsWeb) {
      return child;
    }

    return Scaffold(
      backgroundColor: const Color(0xFF0F241A),
      body: LayoutBuilder(
        builder: (context, constraints) {
          // If the screen is already mobile-width (e.g. Chrome mobile devtools), don't wrap
          if (constraints.maxWidth <= 500) {
            return child;
          }

          // Comfortable flagship smartphone width (e.g. 430px)
          double frameWidth = 430.0;
          if (frameWidth > constraints.maxWidth * 0.92) {
            frameWidth = constraints.maxWidth * 0.92;
          }

          // 19.5:9 proportion height (iPhone 15/16 Pro Max / S24 Ultra standard)
          const double targetRatio = 19.5 / 9.0;
          double frameHeight = frameWidth * targetRatio;

          // Clamp comfortably within viewport height
          if (frameHeight > constraints.maxHeight * 0.94) {
            frameHeight = constraints.maxHeight * 0.94;
            final proportionalWidth = frameHeight / targetRatio;
            if (proportionalWidth >= 400 || constraints.maxWidth < 430) {
              frameWidth = proportionalWidth;
            }
          }

          return Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.symmetric(vertical: 20),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Top Device Status Bar info on desktop canvas
                  Container(
                    margin: const EdgeInsets.only(bottom: 12),
                    padding: const EdgeInsets.symmetric(
                      horizontal: 16,
                      vertical: 6,
                    ),
                    decoration: BoxDecoration(
                      color: Colors.white.withValues(alpha: 0.08),
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(
                        color: Colors.white.withValues(alpha: 0.15),
                      ),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          width: 8,
                          height: 8,
                          decoration: const BoxDecoration(
                            color: Color(0xFF52B788),
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 8),
                        Text(
                          '19.5:9 Mobile Preview • Env: ${ApiEndpoints.environment.toUpperCase()} • ${ApiEndpoints.baseUrl}',
                          style: TextStyle(
                            color: Colors.white.withValues(alpha: 0.85),
                            fontSize: 12,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                  ),

                  // Phone Chassis Container (19.5:9)
                  Container(
                    width: frameWidth,
                    height: frameHeight,
                    decoration: BoxDecoration(
                      color: Colors.black,
                      borderRadius: BorderRadius.circular(50),
                      border: Border.all(
                        color: const Color(0xFF2C3E35),
                        width: 10,
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withValues(alpha: 0.6),
                          blurRadius: 40,
                          spreadRadius: 10,
                          offset: const Offset(0, 20),
                        ),
                        BoxShadow(
                          color: AppColors.primary.withValues(alpha: 0.25),
                          blurRadius: 30,
                          spreadRadius: 2,
                        ),
                      ],
                    ),
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(40),
                      child: Stack(
                        children: [
                          // App Viewport
                          Positioned.fill(child: child),

                          // Top Dynamic Island / Speaker Notch Simulation
                          Align(
                            alignment: Alignment.topCenter,
                            child: Container(
                              margin: const EdgeInsets.only(top: 10),
                              width: 110,
                              height: 26,
                              decoration: BoxDecoration(
                                color: Colors.black,
                                borderRadius: BorderRadius.circular(20),
                              ),
                              child: Row(
                                mainAxisAlignment: MainAxisAlignment.end,
                                children: [
                                  Container(
                                    margin: const EdgeInsets.only(right: 12),
                                    width: 10,
                                    height: 10,
                                    decoration: BoxDecoration(
                                      color: const Color(0xFF141F1A),
                                      shape: BoxShape.circle,
                                      border: Border.all(
                                        color: const Color(0xFF2A3D33),
                                        width: 1.5,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),

                          // Bottom Home Gesture Bar
                          Align(
                            alignment: Alignment.bottomCenter,
                            child: Container(
                              margin: const EdgeInsets.only(bottom: 8),
                              width: 130,
                              height: 4.5,
                              decoration: BoxDecoration(
                                color: Colors.black.withValues(alpha: 0.4),
                                borderRadius: BorderRadius.circular(10),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }
}
