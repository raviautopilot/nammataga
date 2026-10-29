import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:go_router/go_router.dart';
import '../../../core/constants/app_assets.dart';
import '../../../core/constants/app_colors.dart';

/// Premium cinematic splash screen celebrating the architectural landmark
/// of TAGA Towers seamlessly integrated into an emerald atmosphere.
class SplashScreen extends StatefulWidget {
  const SplashScreen({super.key});

  @override
  State<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends State<SplashScreen>
    with TickerProviderStateMixin {
  late final AnimationController _entranceController;
  late final Animation<double> _fadeIn;
  late final Animation<Offset> _slideUp;

  late final AnimationController _breathingController;
  late final Animation<double> _subtleScale;

  late final AnimationController _progressController;
  late final Animation<double> _progressAnimation;

  String _statusText = 'Official Member Portal';
  Timer? _statusTimer;
  Timer? _navigationTimer;

  @override
  void initState() {
    super.initState();

    // 1. Entrance animation (smooth fade & architectural reveal)
    _entranceController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1000),
    );

    _fadeIn = CurvedAnimation(
      parent: _entranceController,
      curve: Curves.easeOutCubic,
    );

    _slideUp = Tween<Offset>(begin: const Offset(0, 0.05), end: Offset.zero)
        .animate(
          CurvedAnimation(
            parent: _entranceController,
            curve: Curves.easeOutCubic,
          ),
        );

    // 2. Restrained architectural breathing motion
    _breathingController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 3200),
    );

    _subtleScale = Tween<double>(begin: 1.0, end: 1.018).animate(
      CurvedAnimation(
        parent: _breathingController,
        curve: Curves.easeInOutSine,
      ),
    );

    // 3. Linear progress indicator (2.6s smooth duration)
    _progressController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 2600),
    );

    _progressAnimation = CurvedAnimation(
      parent: _progressController,
      curve: Curves.easeOutQuad,
    );

    _startCinematicSequence();
  }

  void _startCinematicSequence() {
    HapticFeedback.lightImpact();

    _entranceController.forward();
    _breathingController.repeat(reverse: true);
    _progressController.forward();

    // Tasteful status text transition halfway
    _statusTimer = Timer(const Duration(milliseconds: 1400), () {
      if (mounted) {
        setState(() {
          _statusText = 'Welcome to Namma TAGA';
        });
      }
    });

    // Controlled navigation strictly to /login
    _navigationTimer = Timer(const Duration(milliseconds: 2800), () {
      if (mounted) {
        context.go('/login');
      }
    });
  }

  @override
  void dispose() {
    _statusTimer?.cancel();
    _navigationTimer?.cancel();
    _entranceController.dispose();
    _breathingController.dispose();
    _progressController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF041209),
      body: Stack(
        alignment: Alignment.center,
        children: [
          // 1. Deep cinematic emerald & midnight forest background
          Positioned.fill(
            child: DecoratedBox(
              decoration: const BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topCenter,
                  end: Alignment.bottomCenter,
                  colors: [
                    Color(0xFF041209), // Midnight forest
                    Color(0xFF092013), // Deep emerald
                    Color(0xFF113521), // Signature TAGA forest
                    Color(0xFF06170E), // Base shadow
                  ],
                  stops: [0.0, 0.30, 0.70, 1.0],
                ),
              ),
            ),
          ),

          // 2. Soft atmospheric overhead spotlight
          Positioned(
            top: -50,
            child: IgnorePointer(
              child: Container(
                width: 420,
                height: 420,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: RadialGradient(
                    colors: [
                      const Color(0xFF52B788).withValues(alpha: 0.16),
                      const Color(0xFF2D6A4F).withValues(alpha: 0.06),
                      Colors.transparent,
                    ],
                  ),
                ),
              ),
            ),
          ),

          // 3. Responsive foreground layout with seamless architectural hero
          SafeArea(
            child: LayoutBuilder(
              builder: (context, constraints) {
                return Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 20,
                    vertical: 12,
                  ),
                  child: Column(
                    children: [
                      // TOP: Association Crest & Prestigious Bilingual Header
                      _buildHeader(),

                      // CENTER HERO: TAGA Towers seamlessly integrated into background
                      Expanded(child: _buildTagaTowersHero()),

                      // BOTTOM: Landmark Identity, Motto & Elegant Progress
                      _buildFooter(),
                    ],
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildHeader() {
    return FadeTransition(
      opacity: _fadeIn,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const SizedBox(height: 6),
          // Subtle, unboxed official seal crest
          Container(
            width: 52,
            height: 52,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: Colors.white,
              border: Border.all(
                color: const Color(0xFFD8F3DC).withValues(alpha: 0.8),
                width: 1.8,
              ),
              boxShadow: [
                BoxShadow(
                  color: const Color(0xFF52B788).withValues(alpha: 0.28),
                  blurRadius: 14,
                  spreadRadius: 1,
                ),
              ],
            ),
            padding: const EdgeInsets.all(2),
            child: ClipOval(
              child: Image.asset(
                AppAssets.logo,
                fit: BoxFit.contain,
                errorBuilder: (ctx, err, stack) => Image.asset(
                  AppAssets.logo1,
                  fit: BoxFit.contain,
                  errorBuilder: (c, e, s) => const Icon(
                    Icons.agriculture,
                    color: AppColors.primary,
                    size: 24,
                  ),
                ),
              ),
            ),
          ),
          const SizedBox(height: 8),

          // Tamil Title
          const Text(
            'தமிழ்நாடு அரசு வேளாண்மை பட்டதாரிகள் சங்கம்',
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 12.5,
              fontWeight: FontWeight.bold,
              color: Colors.white,
              letterSpacing: 0.25,
              fontFamilyFallback: ['NotoSansTamil', 'sans-serif'],
            ),
          ),
          const SizedBox(height: 2),

          // English Title
          const Text(
            'TAMIL NADU GOVT. AGRI. GRADUATES ASSOCIATION',
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 9.5,
              fontWeight: FontWeight.w600,
              color: Color(0xFF95D5B2),
              letterSpacing: 0.9,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTagaTowersHero() {
    return SlideTransition(
      position: _slideUp,
      child: FadeTransition(
        opacity: _fadeIn,
        child: AnimatedBuilder(
          animation: _breathingController,
          builder: (context, child) {
            return Transform.scale(
              scale: _subtleScale.value,
              child: Center(
                child: ConstrainedBox(
                  constraints: const BoxConstraints(
                    maxWidth: 380,
                    maxHeight: 460,
                  ),
                  // Double ShaderMask: Feather-dissolves edges seamlessly
                  // so the architectural photograph merges naturally into the emerald background
                  child: ShaderMask(
                    shaderCallback: (Rect bounds) {
                      return const LinearGradient(
                        begin: Alignment.topCenter,
                        end: Alignment.bottomCenter,
                        colors: [
                          Colors.transparent,
                          Colors.black,
                          Colors.black,
                          Colors.transparent,
                        ],
                        stops: [0.0, 0.09, 0.86, 1.0],
                      ).createShader(bounds);
                    },
                    blendMode: BlendMode.dstIn,
                    child: ShaderMask(
                      shaderCallback: (Rect bounds) {
                        return const LinearGradient(
                          begin: Alignment.centerLeft,
                          end: Alignment.centerRight,
                          colors: [
                            Colors.transparent,
                            Colors.black,
                            Colors.black,
                            Colors.transparent,
                          ],
                          stops: [0.0, 0.06, 0.94, 1.0],
                        ).createShader(bounds);
                      },
                      blendMode: BlendMode.dstIn,
                      child: Image.asset(
                        AppAssets.tagatower,
                        fit: BoxFit.contain,
                        errorBuilder: (ctx, err, stack) => const Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              Icons.apartment_rounded,
                              size: 72,
                              color: Color(0xFF74C69D),
                            ),
                            SizedBox(height: 10),
                            Text(
                              'TAGA TOWERS',
                              style: TextStyle(
                                color: Colors.white,
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                                letterSpacing: 1.5,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            );
          },
        ),
      ),
    );
  }

  Widget _buildFooter() {
    return FadeTransition(
      opacity: _fadeIn,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // Landmark Name
          const Text(
            'TAGA TOWERS',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: Colors.white,
              letterSpacing: 2.0,
            ),
          ),
          const SizedBox(height: 2),

          // Tamil Landmark Name
          const Text(
            'தாகா டவர்ஸ்',
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w500,
              color: Color(0xFFB7E4C7),
              letterSpacing: 0.5,
              fontFamilyFallback: ['NotoSansTamil', 'sans-serif'],
            ),
          ),
          const SizedBox(height: 6),

          // Motto Pill
          const Text(
            'பயிர் வளர்ப்போம் • உயிர் காப்போம்  (Estd. 1929)',
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 10.5,
              fontWeight: FontWeight.w500,
              color: Color(0xFF74C69D),
              letterSpacing: 0.2,
              fontFamilyFallback: ['NotoSansTamil', 'sans-serif'],
            ),
          ),
          const SizedBox(height: 16),

          // Sleek hairline progress bar
          Container(
            width: 180,
            height: 2.5,
            decoration: BoxDecoration(
              color: Colors.white.withValues(alpha: 0.10),
              borderRadius: BorderRadius.circular(10),
            ),
            child: AnimatedBuilder(
              animation: _progressAnimation,
              builder: (context, child) {
                return Align(
                  alignment: Alignment.centerLeft,
                  child: Container(
                    width: 180 * _progressAnimation.value,
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        colors: [
                          Color(0xFF52B788),
                          Color(0xFFB7E4C7),
                          Colors.white,
                        ],
                      ),
                      borderRadius: BorderRadius.circular(10),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0xFF52B788).withValues(alpha: 0.5),
                          blurRadius: 6,
                        ),
                      ],
                    ),
                  ),
                );
              },
            ),
          ),
          const SizedBox(height: 8),

          // Tasteful Status Message
          AnimatedSwitcher(
            duration: const Duration(milliseconds: 300),
            child: Text(
              _statusText,
              key: ValueKey<String>(_statusText),
              style: TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.w500,
                color: Colors.white.withValues(alpha: 0.70),
                letterSpacing: 0.3,
              ),
            ),
          ),
          const SizedBox(height: 6),
        ],
      ),
    );
  }
}
