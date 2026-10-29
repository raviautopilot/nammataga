import 'dart:math' as math;
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../core/constants/app_assets.dart';
import '../../../core/constants/app_colors.dart';
import '../../../core/storage/secure_storage_service.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../providers/auth_provider.dart';
import 'widgets/ambient_background.dart';
import 'widgets/change_password_sheet.dart';
import 'widgets/forgot_password_sheet.dart';

class LoginScreen extends ConsumerStatefulWidget {
  final VoidCallback? onLoginSuccess;
  final VoidCallback? onBrowseAsGuest;

  const LoginScreen({
    super.key,
    this.onLoginSuccess,
    this.onBrowseAsGuest,
  });

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen>
    with TickerProviderStateMixin {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();

  bool _showPassword = false;
  bool _rememberMe = true;
  bool _isLoading = false;
  String? _errorMessage;

  // Always-running ambient logo animation controller
  late final AnimationController _ambientController;

  // Error shake animation controller (triggers on invalid credentials)
  late final AnimationController _shakeController;
  late final Animation<double> _shakeAnimation;

  @override
  void initState() {
    super.initState();
    _loadSavedEmail();

    // 14-second smooth continuous ambient loop
    _ambientController = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 14),
    )..repeat();

    // Damped harmonic shake animation on credentials error
    _shakeController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 550),
    );

    _shakeAnimation = TweenSequence<double>([
      TweenSequenceItem(tween: Tween(begin: 0.0, end: -12.0), weight: 1),
      TweenSequenceItem(tween: Tween(begin: -12.0, end: 12.0), weight: 2),
      TweenSequenceItem(tween: Tween(begin: 12.0, end: -8.0), weight: 2),
      TweenSequenceItem(tween: Tween(begin: -8.0, end: 8.0), weight: 2),
      TweenSequenceItem(tween: Tween(begin: 8.0, end: -4.0), weight: 2),
      TweenSequenceItem(tween: Tween(begin: -4.0, end: 0.0), weight: 1),
    ]).animate(CurvedAnimation(
      parent: _shakeController,
      curve: Curves.easeInOut,
    ));

    // Clear error automatically when member starts typing
    _emailController.addListener(_onFieldChanged);
    _passwordController.addListener(_onFieldChanged);
  }

  void _onFieldChanged() {
    if (_errorMessage != null) {
      setState(() => _errorMessage = null);
    }
  }

  Future<void> _loadSavedEmail() async {
    final savedEmail = await StorageService.getRememberEmail();
    if (savedEmail != null && savedEmail.isNotEmpty) {
      setState(() {
        _emailController.text = savedEmail;
        _rememberMe = true;
      });
    }
  }

  @override
  void dispose() {
    _emailController.removeListener(_onFieldChanged);
    _passwordController.removeListener(_onFieldChanged);
    _emailController.dispose();
    _passwordController.dispose();
    _ambientController.dispose();
    _shakeController.dispose();
    super.dispose();
  }

  Future<void> _handleLogin() async {
    final email = _emailController.text.trim();
    final password = _passwordController.text;

    if (!_formKey.currentState!.validate()) {
      HapticFeedback.lightImpact();
      _shakeController.forward(from: 0.0);
      return;
    }

    FocusScope.of(context).unfocus();
    HapticFeedback.mediumImpact();

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final session = await ref
          .read(authProvider.notifier)
          .login(
            identifier: email,
            password: password,
            rememberMe: _rememberMe,
          );

      if (!mounted) return;

      if (session.forceChangePassword) {
        final changed = await ChangePasswordSheet.show(
          context,
          email: session.tempEmail ?? email,
          oldPassword: password,
          isForced: true,
        );

        if (changed == true && mounted) {
          _passwordController.clear();
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text(
                'Password set successfully. Please sign in with your new password.',
              ),
              backgroundColor: AppColors.primary,
            ),
          );
        }
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Welcome back, ${session.displayName}!'),
            backgroundColor: AppColors.primary,
            duration: const Duration(seconds: 2),
          ),
        );
        widget.onLoginSuccess?.call();
      }
    } catch (e) {
      if (mounted) {
        final rawErr = e.toString().replaceAll('Exception: ', '').trim();
        final isAuthFailure = rawErr.toLowerCase().contains('invalid') ||
            rawErr.toLowerCase().contains('credential') ||
            rawErr.toLowerCase().contains('unauthorized') ||
            rawErr.toLowerCase().contains('password') ||
            rawErr.toLowerCase().contains('401');

        setState(() {
          _errorMessage = isAuthFailure ? 'Invalid credentials' : rawErr;
        });

        // Trigger horizontal shake and alert haptic
        _shakeController.forward(from: 0.0);
        HapticFeedback.heavyImpact();
      }
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      body: AmbientBackground(
        child: SafeArea(
          child: Center(
            child: SingleChildScrollView(
              physics: const BouncingScrollPhysics(),
              padding: const EdgeInsets.fromLTRB(18, 16, 18, 24),
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 440),
                child: Form(
                  key: _formKey,
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      // 1. ALWAYS-ANIMATING BRAND LOGO & EDITORIAL ASSOCIATION HEADER
                      _buildCreativeHeader(),
                      const SizedBox(height: 18),

                      // 2. SLEEK FROSTED LOGIN CARD WITH INVALID CREDENTIALS SHAKE
                      AnimatedBuilder(
                        animation: _shakeAnimation,
                        builder: (context, child) {
                          return Transform.translate(
                            offset: Offset(_shakeAnimation.value, 0),
                            child: child,
                          );
                        },
                        child: _buildLoginCard(),
                      ),
                      const SizedBox(height: 20),

                      // 3. OFFICIAL ASSOCIATION FOOTNOTE
                      const Text(
                        'Tamil Nadu Agricultural Graduates Association (TAGA)\nOfficial Member Portal • Estd. 1929',
                        textAlign: TextAlign.center,
                        style: TextStyle(
                          fontSize: 11,
                          color: AppColors.textMuted,
                          height: 1.35,
                          letterSpacing: 0.2,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildCreativeHeader() {
    return Column(
      children: [
        // Always-Running Multi-Layer Animated Logo Assembly
        _buildAnimatedLogo(),
        const SizedBox(height: 14),

        // Tamil Association Title
        const Text(
          'தமிழ்நாடு அரசு வேளாண்மை பட்டதாரிகள் சங்கம்',
          textAlign: TextAlign.center,
          style: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
            letterSpacing: 0.2,
          ),
        ),
        const SizedBox(height: 3),

        // English Association Title
        const Text(
          'TAMIL NADU GOVT. AGRI. GRADUATES ASSOCIATION',
          textAlign: TextAlign.center,
          style: TextStyle(
            fontSize: 10,
            fontWeight: FontWeight.bold,
            color: AppColors.primary,
            letterSpacing: 0.6,
          ),
        ),
        const SizedBox(height: 8),

        // Motto & Estd Pill
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 4),
          decoration: BoxDecoration(
            color: const Color(0xFFE8F5E9),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: const Color(0xFFC8E6C9)),
          ),
          child: const Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(LucideIcons.leaf, size: 11.5, color: AppColors.primary),
              SizedBox(width: 6),
              Flexible(
                child: Text(
                  'பயிர் வளர்ப்போம் • உயிர் காப்போம்  (Estd. 1929)',
                  textAlign: TextAlign.center,
                  style: TextStyle(
                    fontSize: 10.5,
                    fontWeight: FontWeight.w600,
                    color: AppColors.primaryDark,
                  ),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 14),

        const Text(
          'Member Login',
          style: TextStyle(
            fontSize: 22,
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
            letterSpacing: -0.4,
          ),
        ),
        const SizedBox(height: 3),
        const Text(
          'Sign in with your registered email to continue',
          style: TextStyle(fontSize: 12.5, color: AppColors.textSecondary),
        ),
      ],
    );
  }

  /// Always-running multi-layer animated logo:
  /// - Continuous slow rotation of the orbital ring
  /// - Orbiting luminous golden and mint satellites
  /// - Breathing ambient emerald & gold glow halo
  /// - Gentle vertical floating levitation & breathing scale on the core crest
  /// - Continuous diagonal sheen sweep across the emblem
  Widget _buildAnimatedLogo() {
    return AnimatedBuilder(
      animation: _ambientController,
      builder: (context, child) {
        final t = _ambientController.value;
        // Breathing pulsation (4 full sinusoidal breaths per 14-sec loop)
        final breathProgress = math.sin(t * 4 * math.pi);
        final floatDy = breathProgress * 4.5;
        final breathScale = 1.0 + breathProgress * 0.035;
        final haloScale = 1.0 + breathProgress * 0.07;
        final haloOpacity = 0.22 + breathProgress * 0.08;

        // Continuous orbital rotation angle
        final orbitAngle = t * 2 * math.pi;

        const ringSize = 142.0;
        const ringRadius = ringSize / 2;
        const orbitRadius = 67.0;

        // Golden satellite node coordinates along orbital track
        final sat1X = ringRadius + orbitRadius * math.cos(orbitAngle) - 5.5;
        final sat1Y = ringRadius + orbitRadius * math.sin(orbitAngle) - 5.5;

        // Complementary mint satellite node at 180°
        final sat2X = ringRadius + orbitRadius * math.cos(orbitAngle + math.pi) - 4.0;
        final sat2Y = ringRadius + orbitRadius * math.sin(orbitAngle + math.pi) - 4.0;

        return SizedBox(
          width: 156,
          height: 156,
          child: Stack(
            clipBehavior: Clip.none,
            alignment: Alignment.center,
            children: [
              // 1. Ambient Breathing Halo Aura
              Transform.scale(
                scale: haloScale,
                child: Container(
                  width: 130,
                  height: 130,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    boxShadow: [
                      BoxShadow(
                        color: const Color(0xFF52B788).withValues(alpha: haloOpacity),
                        blurRadius: 34,
                        spreadRadius: 8,
                      ),
                      BoxShadow(
                        color: const Color(0xFFD4AF37).withValues(alpha: haloOpacity * 0.4),
                        blurRadius: 20,
                        spreadRadius: 2,
                      ),
                    ],
                  ),
                ),
              ),

              // 2. Continuous Rotating Orbital Track & Radiant Arc
              CustomPaint(
                size: const Size(ringSize, ringSize),
                painter: _OrbitalRingPainter(
                  ringColor: const Color(0xFF52B788).withValues(alpha: 0.25),
                  accentColor: const Color(0xFFD4AF37).withValues(alpha: 0.8),
                  rotationAngle: orbitAngle,
                ),
              ),

              // 3. Luminous Golden Satellite Node Orbiting Continuously
              Positioned(
                left: sat1X,
                top: sat1Y,
                child: Container(
                  width: 11,
                  height: 11,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: const Color(0xFFFDE68A),
                    border: Border.all(
                      color: const Color(0xFFD97706),
                      width: 1.5,
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: const Color(0xFFF59E0B).withValues(alpha: 0.9),
                        blurRadius: 8,
                        spreadRadius: 2,
                      ),
                    ],
                  ),
                ),
              ),

              // 4. Complementary Mint Satellite Node
              Positioned(
                left: sat2X,
                top: sat2Y,
                child: Container(
                  width: 8,
                  height: 8,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: const Color(0xFF52B788),
                    boxShadow: [
                      BoxShadow(
                        color: const Color(0xFF52B788).withValues(alpha: 0.8),
                        blurRadius: 6,
                        spreadRadius: 1,
                      ),
                    ],
                  ),
                ),
              ),

              // 5. Floating Levitation & Breathing Core Logo Crest Pedestal
              Transform.translate(
                offset: Offset(0, floatDy),
                child: Transform.scale(
                  scale: breathScale,
                  child: Container(
                    width: 100,
                    height: 100,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: Colors.white,
                      border: Border.all(
                        color: const Color(0xFFD8F3DC),
                        width: 3.5,
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: AppColors.primary.withValues(alpha: 0.28),
                          blurRadius: 22 - floatDy * 1.5,
                          offset: Offset(0, 10 - floatDy),
                        ),
                        BoxShadow(
                          color: const Color(0xFF52B788).withValues(alpha: 0.16),
                          blurRadius: 12,
                          spreadRadius: 2,
                        ),
                      ],
                    ),
                    padding: const EdgeInsets.all(4),
                    child: ClipOval(
                      child: Stack(
                        fit: StackFit.expand,
                        children: [
                          Image.asset(
                            AppAssets.logo,
                            fit: BoxFit.contain,
                            errorBuilder: (context, error, stackTrace) {
                              return Image.asset(
                                AppAssets.logo1,
                                fit: BoxFit.contain,
                                errorBuilder: (ctx, err, stack) => const Icon(
                                  LucideIcons.sprout,
                                  color: AppColors.primary,
                                  size: 44,
                                ),
                              );
                            },
                          ),
                          // Continuous Sheen Sweep
                          _ShimmerSheenOverlay(progress: (t * 2.5) % 1.0),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildLoginCard() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 22),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
        boxShadow: const [
          BoxShadow(
            color: Color(0x0A1B4332),
            blurRadius: 10,
            offset: Offset(0, 3),
          ),
          BoxShadow(
            color: Color(0x0F2D6A4F),
            blurRadius: 28,
            spreadRadius: 1,
            offset: Offset(0, 10),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Clear "Invalid credentials" Alert Banner when authentication fails
          if (_errorMessage != null) ...[
            Container(
              margin: const EdgeInsets.only(bottom: 16),
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
              decoration: BoxDecoration(
                color: const Color(0xFFFEF2F2),
                borderRadius: BorderRadius.circular(14),
                border: Border.all(
                  color: const Color(0xFFFCA5A5),
                  width: 1.2,
                ),
                boxShadow: const [
                  BoxShadow(
                    color: Color(0x0AE11D48),
                    blurRadius: 8,
                    offset: Offset(0, 2),
                  ),
                ],
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    padding: const EdgeInsets.all(5),
                    decoration: const BoxDecoration(
                      color: Color(0xFFFEE2E2),
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(
                      LucideIcons.alertCircle,
                      color: Color(0xFFDC2626),
                      size: 18,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          _errorMessage!,
                          style: const TextStyle(
                            fontSize: 13.5,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFF991B1B),
                            letterSpacing: -0.1,
                          ),
                        ),
                        if (_errorMessage == 'Invalid credentials') ...[
                          const SizedBox(height: 2),
                          const Text(
                            'The email address or password you entered is incorrect. Please verify your credentials and try again.',
                            style: TextStyle(
                              fontSize: 11.5,
                              color: Color(0xFFB91C1C),
                              height: 1.35,
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),
                  IconButton(
                    icon: const Icon(
                      LucideIcons.x,
                      size: 15,
                      color: Color(0xFF991B1B),
                    ),
                    padding: EdgeInsets.zero,
                    constraints: const BoxConstraints(),
                    onPressed: () {
                      setState(() => _errorMessage = null);
                    },
                  ),
                ],
              ),
            ),
          ],

          // Email Input
          const Text(
            'Email Address',
            style: TextStyle(
              fontSize: 12.5,
              fontWeight: FontWeight.w600,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 6),
          TextFormField(
            controller: _emailController,
            keyboardType: TextInputType.emailAddress,
            textInputAction: TextInputAction.next,
            autofillHints: const [AutofillHints.email, AutofillHints.username],
            style: const TextStyle(fontSize: 14, color: AppColors.textPrimary),
            decoration: const InputDecoration(
              hintText: 'Enter your registered email address',
              prefixIcon: Icon(
                LucideIcons.mail,
                size: 17,
                color: AppColors.primary,
              ),
            ),
            validator: (value) {
              final email = value?.trim() ?? '';
              if (email.isEmpty) {
                return 'Please enter your registered email address';
              }
              final emailRegex = RegExp(
                r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$',
              );
              if (!emailRegex.hasMatch(email)) {
                return 'Please enter a valid email address (e.g. name@example.com)';
              }
              return null;
            },
          ),
          const SizedBox(height: 14),

          // Password Input
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text(
                'Password',
                style: TextStyle(
                  fontSize: 12.5,
                  fontWeight: FontWeight.w600,
                  color: AppColors.textPrimary,
                ),
              ),
              BouncyButton(
                onTap: () => ForgotPasswordSheet.show(context),
                child: const Text(
                  'Forgot Password?',
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: AppColors.primary,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          TextFormField(
            controller: _passwordController,
            obscureText: !_showPassword,
            textInputAction: TextInputAction.done,
            onFieldSubmitted: (_) => _handleLogin(),
            style: const TextStyle(fontSize: 14, color: AppColors.textPrimary),
            decoration: InputDecoration(
              hintText: 'Enter your password',
              prefixIcon: const Icon(
                LucideIcons.lock,
                size: 17,
                color: AppColors.primary,
              ),
              suffixIcon: IconButton(
                icon: Icon(
                  _showPassword ? LucideIcons.eyeOff : LucideIcons.eye,
                  size: 17,
                  color: AppColors.textMuted,
                ),
                onPressed: () {
                  HapticFeedback.selectionClick();
                  setState(() => _showPassword = !_showPassword);
                },
              ),
            ),
            validator: (value) {
              if (value == null || value.isEmpty) {
                return 'Please enter your password';
              }
              return null;
            },
          ),
          const SizedBox(height: 12),

          // Remember Me Checkbox
          Row(
            children: [
              SizedBox(
                width: 20,
                height: 20,
                child: Checkbox(
                  value: _rememberMe,
                  activeColor: AppColors.primary,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(5),
                  ),
                  onChanged: (val) {
                    HapticFeedback.selectionClick();
                    setState(() => _rememberMe = val ?? false);
                  },
                ),
              ),
              const SizedBox(width: 8),
              GestureDetector(
                onTap: () {
                  HapticFeedback.selectionClick();
                  setState(() => _rememberMe = !_rememberMe);
                },
                child: const Text(
                  'Remember me on this device',
                  style: TextStyle(
                    fontSize: 12.5,
                    color: AppColors.textPrimary,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 18),

          // Sign In CTA Button with Emerald Gradient
          BouncyButton(
            onTap: _isLoading ? null : _handleLogin,
            child: Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(vertical: 14),
              decoration: BoxDecoration(
                gradient: const LinearGradient(
                  colors: AppColors.brandGradient,
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
                borderRadius: BorderRadius.circular(14),
                boxShadow: [
                  BoxShadow(
                    color: AppColors.primary.withValues(alpha: 0.28),
                    blurRadius: 14,
                    offset: const Offset(0, 4),
                  ),
                ],
              ),
              child: Center(
                child: _isLoading
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2.2,
                          color: Colors.white,
                        ),
                      )
                    : const Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            LucideIcons.logIn,
                            color: Colors.white,
                            size: 17,
                          ),
                          SizedBox(width: 8),
                          Text(
                            'Sign In',
                            style: TextStyle(
                              fontSize: 14.5,
                              fontWeight: FontWeight.bold,
                              color: Colors.white,
                              letterSpacing: 0.2,
                            ),
                          ),
                        ],
                      ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// Custom painter that renders a smooth orbital track with a rotating radiant arc
class _OrbitalRingPainter extends CustomPainter {
  final Color ringColor;
  final Color accentColor;
  final double rotationAngle;

  _OrbitalRingPainter({
    required this.ringColor,
    required this.accentColor,
    required this.rotationAngle,
  });

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final radius = size.width / 2;

    // Base circular track
    final trackPaint = Paint()
      ..color = ringColor
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1.6;
    canvas.drawCircle(center, radius, trackPaint);

    // Glowing accent arc rotating continuously
    final arcPaint = Paint()
      ..color = accentColor
      ..style = PaintingStyle.stroke
      ..strokeWidth = 2.4
      ..strokeCap = StrokeCap.round;

    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius),
      rotationAngle,
      0.8, // ~45 degree radiant arc
      false,
      arcPaint,
    );

    // Complementary subtle arc on opposite side
    final arc2Paint = Paint()
      ..color = ringColor.withValues(alpha: 0.5)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 2.0
      ..strokeCap = StrokeCap.round;

    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius),
      rotationAngle + math.pi,
      0.5,
      false,
      arc2Paint,
    );
  }

  @override
  bool shouldRepaint(covariant _OrbitalRingPainter oldDelegate) {
    return oldDelegate.rotationAngle != rotationAngle ||
        oldDelegate.ringColor != ringColor ||
        oldDelegate.accentColor != accentColor;
  }
}

/// Continuous diagonal crystalline sheen overlay passing across the logo
class _ShimmerSheenOverlay extends StatelessWidget {
  final double progress;

  const _ShimmerSheenOverlay({required this.progress});

  @override
  Widget build(BuildContext context) {
    // Visible for the first 65% of the sweep period, then rests
    if (progress > 0.65) return const SizedBox.shrink();
    final p = progress / 0.65;

    return ClipOval(
      child: FractionallySizedBox(
        widthFactor: 2.0,
        heightFactor: 2.0,
        alignment: Alignment(-2.2 + p * 4.4, 0),
        child: Transform.rotate(
          angle: 0.5,
          child: Container(
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  Colors.transparent,
                  Colors.white.withValues(alpha: 0.28),
                  Colors.transparent,
                ],
                stops: const [0.35, 0.5, 0.65],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
