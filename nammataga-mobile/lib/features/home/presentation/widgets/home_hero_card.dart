import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/constants/app_assets.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../auth/providers/auth_provider.dart';
import '../../../profile/providers/profile_provider.dart';
import '../../providers/home_provider.dart';

/// Compact, professional Green Member Profile Card for TAGA Member Portal.
/// Displays:
/// 1. TAGA branding & Established 1929
/// 2. Subscriber Member badge
/// 3. Dynamic Vanakkam, [Member Name]
/// 4. Dynamic TAGA ID
/// 5. Tamil Nadu Government Agricultural Graduates Association
/// 6. Four official G.O. and registration details:
///    - Recognised G.O. No. 831, Dated 08.08.1929
///    - G.O. No. 48, Dated 06.01.1973
///    - G.O. No. 94, Dated 02.07.2009
///    - Registration Number: 246/2001
class HomeHeroCard extends ConsumerWidget {
  const HomeHeroCard({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sessionAsync = ref.watch(authProvider);
    final profileAsync = ref.watch(memberProfileProvider);
    final aboutAsync = ref.watch(aboutInfoProvider);

    final memberName = profileAsync.value?.fullNameWithInitial.isNotEmpty == true
        ? profileAsync.value!.fullNameWithInitial
        : (sessionAsync.value?.displayName.isNotEmpty == true
            ? sessionAsync.value!.displayName
            : 'Member');
    final tagaId = profileAsync.value?.tagaId ?? '';
    final isPaid =
        profileAsync.value?.isPaid ?? sessionAsync.value?.isPaid ?? false;

    final establishedYear = aboutAsync.value?.establishedYear ?? 1929;

    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: const Color(0xFF092014),
        borderRadius: BorderRadius.circular(22),
        boxShadow: [
          BoxShadow(
            color: AppColors.primary.withValues(alpha: 0.25),
            blurRadius: 18,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(22),
        child: Stack(
          children: [
            // Cinematic emerald gradient overlay
            Positioned.fill(
              child: DecoratedBox(
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: [
                      const Color(0xFF071B10),
                      const Color(0xFF0D2C1B),
                      const Color(0xFF1B4332),
                      const Color(0xFF2D6A4F),
                    ],
                    stops: const [0.0, 0.35, 0.72, 1.0],
                  ),
                ),
              ),
            ),

            // Ambient background radiance glow
            Positioned(
              right: -30,
              top: -30,
              child: Container(
                width: 150,
                height: 150,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  gradient: RadialGradient(
                    colors: [
                      const Color(0xFF52B788).withValues(alpha: 0.22),
                      Colors.transparent,
                    ],
                  ),
                ),
              ),
            ),

            Padding(
              padding: const EdgeInsets.all(18),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Top row: Organization branding & Subscriber Member badge
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          Container(
                            width: 42,
                            height: 42,
                            decoration: BoxDecoration(
                              shape: BoxShape.circle,
                              color: Colors.white,
                              border: Border.all(
                                color: const Color(0xFFD8F3DC),
                                width: 1.8,
                              ),
                              boxShadow: [
                                BoxShadow(
                                  color: const Color(0xFF52B788)
                                      .withValues(alpha: 0.35),
                                  blurRadius: 8,
                                ),
                              ],
                            ),
                            padding: const EdgeInsets.all(1.5),
                            child: ClipOval(
                              child: Image.asset(
                                AppAssets.logo,
                                fit: BoxFit.contain,
                                errorBuilder: (ctx, err, stack) => Image.asset(
                                  AppAssets.logo1,
                                  fit: BoxFit.contain,
                                ),
                              ),
                            ),
                          ),
                          const SizedBox(width: 10),
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Text(
                                'TAGA',
                                style: TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.bold,
                                  color: Colors.white,
                                  letterSpacing: 0.4,
                                ),
                              ),
                              Container(
                                margin: const EdgeInsets.only(top: 2),
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 6,
                                  vertical: 1.5,
                                ),
                                decoration: BoxDecoration(
                                  color: const Color(0xFF2D6A4F),
                                  borderRadius: BorderRadius.circular(6),
                                  border: Border.all(
                                    color: const Color(0xFF52B788),
                                    width: 0.8,
                                  ),
                                ),
                                child: Text(
                                  'Established ${establishedYear > 0 ? establishedYear : 1929}',
                                  style: const TextStyle(
                                    fontSize: 9.5,
                                    fontWeight: FontWeight.w600,
                                    color: Colors.white,
                                    letterSpacing: 0.2,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),

                      // Subscriber Member Badge
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 10,
                          vertical: 4.5,
                        ),
                        decoration: BoxDecoration(
                          color: Colors.white.withValues(alpha: 0.15),
                          borderRadius: BorderRadius.circular(20),
                          border: Border.all(
                            color: const Color(0xFF52B788),
                            width: 1.1,
                          ),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Container(
                              width: 6.5,
                              height: 6.5,
                              decoration: const BoxDecoration(
                                shape: BoxShape.circle,
                                color: Color(0xFF52B788),
                              ),
                            ),
                            const SizedBox(width: 5),
                            Text(
                              isPaid ? 'Subscriber Member' : 'Verified Member',
                              style: const TextStyle(
                                fontSize: 10.5,
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),

                  // Greeting & Dynamic Member Name
                  Text(
                    'Vanakkam, $memberName',
                    style: const TextStyle(
                      fontSize: 19,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                      letterSpacing: -0.3,
                    ),
                  ),

                  // Dynamic TAGA ID
                  if (tagaId.isNotEmpty) ...[
                    const SizedBox(height: 3),
                    Text(
                      'TAGA ID: $tagaId',
                      style: const TextStyle(
                        fontSize: 12.5,
                        fontWeight: FontWeight.w700,
                        color: Color(0xFFB7E4C7),
                        letterSpacing: 0.3,
                      ),
                    ),
                  ],

                  const SizedBox(height: 8),

                  // Full Association Name
                  const Text(
                    'Tamil Nadu Government Agricultural Graduates Association',
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: Color(0xFFD8F3DC),
                      letterSpacing: -0.1,
                      height: 1.3,
                    ),
                  ),

                  const SizedBox(height: 12),

                  // Official Recognised G.O. & Registration Block
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 9,
                    ),
                    decoration: BoxDecoration(
                      color: Colors.black.withValues(alpha: 0.28),
                      borderRadius: BorderRadius.circular(10),
                      border: Border.all(
                        color: Colors.white.withValues(alpha: 0.14),
                        width: 1.0,
                      ),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: const [
                        Text(
                          'Recognised G.O. No. 831, Dated 08.08.1929',
                          style: TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFFE8F5EE),
                            height: 1.4,
                          ),
                        ),
                        SizedBox(height: 2),
                        Text(
                          'G.O. No. 48, Dated 06.01.1973',
                          style: TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFFE8F5EE),
                            height: 1.4,
                          ),
                        ),
                        SizedBox(height: 2),
                        Text(
                          'G.O. No. 94, Dated 02.07.2009',
                          style: TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                            color: Color(0xFFE8F5EE),
                            height: 1.4,
                          ),
                        ),
                        SizedBox(height: 2),
                        Text(
                          'Registration Number: 246/2001',
                          style: TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w700,
                            color: Color(0xFFB7E4C7),
                            height: 1.4,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
