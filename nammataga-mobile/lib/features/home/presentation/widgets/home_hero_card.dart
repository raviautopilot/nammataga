import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/constants/app_assets.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../auth/providers/auth_provider.dart';
import '../../providers/home_provider.dart';

class HomeHeroCard extends ConsumerWidget {
  const HomeHeroCard({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sessionAsync = ref.watch(authProvider);
    final aboutAsync = ref.watch(aboutInfoProvider);

    final isGuest = sessionAsync.value?.isGuest ?? true;
    final displayName = sessionAsync.value?.displayName ?? 'Guest';

    return aboutAsync.when(
      data: (about) => _buildHero(
        context: context,
        isGuest: isGuest,
        displayName: displayName,
        establishedYear: about.establishedYear,
        associationName: about.name,
        tagline: about.tagline,
        heroImageUrl: about.heroImageUrl,
      ),
      loading: () =>
          _buildHeroSkeleton(isGuest: isGuest, displayName: displayName),
      error: (error, stack) => _buildHero(
        context: context,
        isGuest: isGuest,
        displayName: displayName,
        establishedYear: 0,
        associationName: '',
        tagline: '',
        heroImageUrl: '',
      ),
    );
  }

  Widget _buildHero({
    required BuildContext context,
    required bool isGuest,
    required String displayName,
    required int establishedYear,
    required String associationName,
    required String tagline,
    required String heroImageUrl,
  }) {
    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: const Color(0xFF092014),
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: AppColors.primary.withValues(alpha: 0.28),
            blurRadius: 22,
            offset: const Offset(0, 8),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(24),
        child: Stack(
          children: [
            // Background Image if provided by backend
            if (heroImageUrl.isNotEmpty)
              Positioned.fill(
                child: Image.network(
                  heroImageUrl,
                  fit: BoxFit.cover,
                  errorBuilder: (context, error, stackTrace) =>
                      const SizedBox.shrink(),
                ),
              ),

            // Cinematic emerald gradient overlay
            Positioned.fill(
              child: DecoratedBox(
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: [
                      const Color(0xFF071B10).withValues(alpha: 0.94),
                      const Color(0xFF0D2C1B).withValues(alpha: 0.92),
                      const Color(0xFF1B4332).withValues(alpha: 0.88),
                      const Color(0xFF2D6A4F).withValues(alpha: 0.82),
                    ],
                  ),
                ),
              ),
            ),

            // Ambient background radiance ring
            Positioned(
              right: -30,
              top: -30,
              child: Container(
                width: 160,
                height: 160,
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
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Top row: Emblem, Established Badge (from backend), Status Pill
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          Container(
                            width: 44,
                            height: 44,
                            decoration: BoxDecoration(
                              shape: BoxShape.circle,
                              color: Colors.white,
                              border: Border.all(
                                color: const Color(0xFFD8F3DC),
                                width: 2,
                              ),
                              boxShadow: [
                                BoxShadow(
                                  color: const Color(
                                    0xFF52B788,
                                  ).withValues(alpha: 0.35),
                                  blurRadius: 10,
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
                                ),
                              ),
                            ),
                          ),
                          const SizedBox(width: 10),
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Text(
                                'Namma TAGA',
                                style: TextStyle(
                                  fontSize: 15,
                                  fontWeight: FontWeight.bold,
                                  color: Colors.white,
                                  letterSpacing: 0.2,
                                ),
                              ),
                              if (establishedYear > 0)
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
                                    'Established $establishedYear',
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

                      // Member/Guest Access Pill
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 10,
                          vertical: 4.5,
                        ),
                        decoration: BoxDecoration(
                          color: Colors.white.withValues(alpha: 0.15),
                          borderRadius: BorderRadius.circular(20),
                          border: Border.all(
                            color: isGuest
                                ? Colors.white.withValues(alpha: 0.25)
                                : const Color(0xFF52B788),
                            width: 1.1,
                          ),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Container(
                              width: 6.5,
                              height: 6.5,
                              decoration: BoxDecoration(
                                shape: BoxShape.circle,
                                color: isGuest
                                    ? const Color(0xFFE2EBE5)
                                    : const Color(0xFF52B788),
                              ),
                            ),
                            const SizedBox(width: 5),
                            Text(
                              isGuest ? 'Guest Access' : 'Verified Member',
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
                  const SizedBox(height: 18),

                  // Greeting & Name
                  Text(
                    'Vanakkam, $displayName',
                    style: const TextStyle(
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                      letterSpacing: -0.3,
                    ),
                  ),

                  // Official Association Name from Backend API
                  if (associationName.isNotEmpty) ...[
                    const SizedBox(height: 6),
                    Text(
                      associationName,
                      style: const TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                        color: Color(0xFFD8F3DC),
                        letterSpacing: -0.1,
                      ),
                    ),
                  ],

                  // Dynamic Tagline from Backend API
                  if (tagline.isNotEmpty) ...[
                    const SizedBox(height: 8),
                    Text(
                      tagline,
                      style: TextStyle(
                        fontSize: 12,
                        color: Colors.white.withValues(alpha: 0.88),
                        height: 1.45,
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildHeroSkeleton({
    required bool isGuest,
    required String displayName,
  }) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: const Color(0xFF092014),
        borderRadius: BorderRadius.circular(24),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Container(
                    width: 44,
                    height: 44,
                    decoration: const BoxDecoration(
                      shape: BoxShape.circle,
                      color: Color(0xFF133924),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Container(
                        width: 90,
                        height: 14,
                        decoration: BoxDecoration(
                          color: const Color(0xFF133924),
                          borderRadius: BorderRadius.circular(4),
                        ),
                      ),
                      const SizedBox(height: 4),
                      Container(
                        width: 70,
                        height: 10,
                        decoration: BoxDecoration(
                          color: const Color(0xFF133924),
                          borderRadius: BorderRadius.circular(4),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              Container(
                width: 80,
                height: 22,
                decoration: BoxDecoration(
                  color: const Color(0xFF133924),
                  borderRadius: BorderRadius.circular(20),
                ),
              ),
            ],
          ),
          const SizedBox(height: 20),
          Text(
            'Vanakkam, $displayName',
            style: const TextStyle(
              fontSize: 20,
              fontWeight: FontWeight.bold,
              color: Colors.white,
            ),
          ),
          const SizedBox(height: 10),
          Container(
            width: double.infinity * 0.7,
            height: 14,
            decoration: BoxDecoration(
              color: const Color(0xFF133924),
              borderRadius: BorderRadius.circular(4),
            ),
          ),
          const SizedBox(height: 6),
          Container(
            width: double.infinity * 0.9,
            height: 12,
            decoration: BoxDecoration(
              color: const Color(0xFF133924),
              borderRadius: BorderRadius.circular(4),
            ),
          ),
        ],
      ),
    );
  }
}
