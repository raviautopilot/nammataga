import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import 'gallery_screen.dart';
import 'upcoming_events_screen.dart';

/// Main creative hub for NammaTAGA Events & Gallery, featuring a bespoke
/// mobile-native emerald gradient header, fluid micro-interactions, and segmented tabs.
class EventsScreen extends StatefulWidget {
  const EventsScreen({super.key});

  @override
  State<EventsScreen> createState() => _EventsScreenState();
}

class _EventsScreenState extends State<EventsScreen> {
  int _selectedTab = 0; // 0 = Upcoming, 1 = Gallery
  late final PageController _pageController;

  @override
  void initState() {
    super.initState();
    _pageController = PageController(initialPage: _selectedTab);
  }

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  void _onTabSelected(int index) {
    if (_selectedTab == index) return;
    HapticFeedback.selectionClick();
    setState(() {
      _selectedTab = index;
    });
    _pageController.animateToPage(
      index,
      duration: const Duration(milliseconds: 320),
      curve: Curves.easeInOutCubic,
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAF7),
      body: Column(
        children: [
          // Bespoke Native Emerald Artistic Header (No backend banner)
          Container(
            decoration: BoxDecoration(
              borderRadius: const BorderRadius.vertical(
                bottom: Radius.circular(30),
              ),
              boxShadow: [
                BoxShadow(
                  color: const Color(0xFF04190E).withValues(alpha: 0.38),
                  blurRadius: 22,
                  offset: const Offset(0, 8),
                ),
              ],
            ),
            clipBehavior: Clip.antiAlias,
            child: Stack(
              children: [
                // Rich multi-stop emerald gradient background
                Positioned.fill(
                  child: Container(
                    decoration: const BoxDecoration(
                      gradient: LinearGradient(
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                        colors: [
                          Color(0xFF04170D), // Deep rich forest obsidian
                          Color(0xFF0A2B1B), // Signature TAGA deep emerald
                          Color(0xFF14452F), // Forest green
                          Color(0xFF1D5A3D), // Vibrant highlight green
                        ],
                        stops: [0.0, 0.35, 0.72, 1.0],
                      ),
                    ),
                  ),
                ),

                // Ambient Radial Light Glow (Top-Right)
                Positioned(
                  top: -50,
                  right: -40,
                  width: 220,
                  height: 220,
                  child: Container(
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      gradient: RadialGradient(
                        colors: [
                          const Color(0xFF52B788).withValues(alpha: 0.28),
                          Colors.transparent,
                        ],
                      ),
                    ),
                  ),
                ),

                // Ambient Soft Glow (Bottom-Left)
                Positioned(
                  bottom: -40,
                  left: -30,
                  width: 180,
                  height: 180,
                  child: Container(
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      gradient: RadialGradient(
                        colors: [
                          const Color(0xFF2D6A4F).withValues(alpha: 0.32),
                          Colors.transparent,
                        ],
                      ),
                    ),
                  ),
                ),

                // Subtle Watermark Leaf Motif
                Positioned(
                  right: -10,
                  bottom: -15,
                  child: IgnorePointer(
                    child: Icon(
                      LucideIcons.leaf,
                      size: 135,
                      color: Colors.white.withValues(alpha: 0.04),
                    ),
                  ),
                ),

                // Header Foreground Content
                SafeArea(
                  bottom: false,
                  child: Padding(
                    padding: const EdgeInsets.fromLTRB(16, 8, 16, 20),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Nav row
                        Row(
                          children: [
                            BouncyButton(
                              onTap: () {
                                HapticFeedback.lightImpact();
                                if (context.canPop()) {
                                  context.pop();
                                } else {
                                  context.go('/home');
                                }
                              },
                              child: Container(
                                width: 38,
                                height: 38,
                                decoration: BoxDecoration(
                                  color: Colors.white.withValues(alpha: 0.14),
                                  shape: BoxShape.circle,
                                  border: Border.all(
                                    color: Colors.white.withValues(alpha: 0.25),
                                    width: 1.1,
                                  ),
                                ),
                                child: const Center(
                                  child: Icon(
                                    LucideIcons.arrowLeft,
                                    size: 18,
                                    color: Colors.white,
                                  ),
                                ),
                              ),
                            ),
                            const SizedBox(width: 12),
                            // Web-styled Pill Badge
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 10,
                                vertical: 4.5,
                              ),
                              decoration: BoxDecoration(
                                color: const Color(0xFF2D6A4F).withValues(alpha: 0.85),
                                borderRadius: BorderRadius.circular(20),
                                border: Border.all(
                                  color: const Color(0xFF74C69D).withValues(alpha: 0.45),
                                  width: 1.0,
                                ),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: const [
                                  Icon(
                                    LucideIcons.sparkles,
                                    size: 11,
                                    color: Color(0xFFB7E4C7),
                                  ),
                                  SizedBox(width: 5),
                                  Text(
                                    'Events & Media',
                                    style: TextStyle(
                                      fontSize: 11,
                                      fontWeight: FontWeight.bold,
                                      color: Colors.white,
                                      letterSpacing: 0.3,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),

                        const SizedBox(height: 14),

                        // Main Title & Subtitle
                        const Text(
                          'TAGA Events',
                          style: TextStyle(
                            fontSize: 27,
                            fontWeight: FontWeight.w900,
                            color: Colors.white,
                            letterSpacing: -0.6,
                            height: 1.1,
                          ),
                        ),
                        const SizedBox(height: 4),
                        const Text(
                          'Discover upcoming programs and explore association memories',
                          style: TextStyle(
                            fontSize: 12.5,
                            color: Color(0xFFD8F3DC),
                            fontWeight: FontWeight.w500,
                            height: 1.35,
                          ),
                        ),

                        const SizedBox(height: 18),

                        // Creative Segment Tab Switcher
                        Container(
                          padding: const EdgeInsets.all(4),
                          decoration: BoxDecoration(
                            color: Colors.black.withValues(alpha: 0.35),
                            borderRadius: BorderRadius.circular(16),
                            border: Border.all(
                              color: Colors.white.withValues(alpha: 0.15),
                              width: 1.0,
                            ),
                          ),
                          child: Row(
                            children: [
                              Expanded(
                                child: _TabButton(
                                  title: 'Upcoming Events',
                                  icon: LucideIcons.calendarCheck2,
                                  isSelected: _selectedTab == 0,
                                  onTap: () => _onTabSelected(0),
                                ),
                              ),
                              Expanded(
                                child: _TabButton(
                                  title: 'Photo Gallery',
                                  icon: LucideIcons.images,
                                  isSelected: _selectedTab == 1,
                                  onTap: () => _onTabSelected(1),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),

          // Tab Pages
          Expanded(
            child: PageView(
              controller: _pageController,
              onPageChanged: (index) {
                HapticFeedback.selectionClick();
                setState(() {
                  _selectedTab = index;
                });
              },
              children: const [
                UpcomingEventsScreen(),
                GalleryScreen(),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _TabButton extends StatelessWidget {
  final String title;
  final IconData icon;
  final bool isSelected;
  final VoidCallback onTap;

  const _TabButton({
    required this.title,
    required this.icon,
    required this.isSelected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 220),
        curve: Curves.easeOutCubic,
        padding: const EdgeInsets.symmetric(vertical: 10),
        decoration: BoxDecoration(
          color: isSelected ? Colors.white : Colors.transparent,
          borderRadius: BorderRadius.circular(12),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.18),
                    blurRadius: 8,
                    offset: const Offset(0, 2),
                  ),
                ]
              : null,
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              icon,
              size: 15,
              color: isSelected
                  ? AppColors.primaryDark
                  : Colors.white.withValues(alpha: 0.8),
            ),
            const SizedBox(width: 8),
            Text(
              title,
              style: TextStyle(
                fontSize: 12.5,
                fontWeight: isSelected ? FontWeight.w800 : FontWeight.w600,
                color: isSelected
                    ? AppColors.primaryDark
                    : Colors.white.withValues(alpha: 0.85),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
