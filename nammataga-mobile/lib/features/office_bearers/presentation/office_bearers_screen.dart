import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../domain/district_office_bearer.dart';
import '../domain/state_executive_member.dart';
import '../providers/office_bearers_provider.dart';
import 'widgets/district_bearer_profile_screen.dart';
import 'widgets/district_bearer_row.dart';
import 'widgets/office_bearers_shimmer.dart';
import 'widgets/state_executive_card.dart';
import 'widgets/state_executive_profile_screen.dart';

/// Office Bearers Screen matching NammaTAGA's emerald design system:
/// - Bespoke Emerald Gradient Executive Header with back button, term badge, and title
/// - Premium segmented pill tab switcher with haptic feedback
/// - State Executive: rectangular portrait profile cards with micro actions
/// - District Bearers: clean dropdown + photograph-free mobile table directory
class OfficeBearersScreen extends ConsumerStatefulWidget {
  const OfficeBearersScreen({super.key});

  @override
  ConsumerState<OfficeBearersScreen> createState() =>
      _OfficeBearersScreenState();
}

class _OfficeBearersScreenState extends ConsumerState<OfficeBearersScreen> {
  int _selectedTab = 0; // 0 = State Executive, 1 = District Bearers
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
      duration: const Duration(milliseconds: 300),
      curve: Curves.easeInOutCubic,
    );
  }

  void _handleBack() {
    HapticFeedback.lightImpact();
    if (context.canPop()) {
      context.pop();
    } else {
      context.go('/home');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAF7),
      body: Column(
        children: [
          // 1. Bespoke Native Emerald Artistic Header Card
          _buildExecutiveHeader(),

          // 2. Tab Pages (State Executive & District Bearers)
          Expanded(
            child: PageView(
              controller: _pageController,
              onPageChanged: (index) {
                HapticFeedback.selectionClick();
                setState(() {
                  _selectedTab = index;
                });
              },
              children: [
                _buildStateExecutiveSection(),
                _buildDistrictBearersSection(),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// Builds the signature emerald gradient header matching Events & Resources
  Widget _buildExecutiveHeader() {
    return Container(
      decoration: BoxDecoration(
        borderRadius: const BorderRadius.vertical(
          bottom: Radius.circular(28),
        ),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF04190E).withValues(alpha: 0.35),
            blurRadius: 20,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      clipBehavior: Clip.antiAlias,
      child: Stack(
        children: [
          // Multi-stop Emerald Brand Gradient
          Positioned.fill(
            child: Container(
              decoration: const BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                  colors: [
                    Color(0xFF04170D),
                    Color(0xFF0A2B1B),
                    Color(0xFF14452F),
                    Color(0xFF1D5A3D),
                  ],
                  stops: [0.0, 0.35, 0.72, 1.0],
                ),
              ),
            ),
          ),

          // Ambient Top-Right Radial Glow
          Positioned(
            top: -45,
            right: -35,
            width: 200,
            height: 200,
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

          // Ambient Bottom-Left Glow
          Positioned(
            bottom: -40,
            left: -30,
            width: 170,
            height: 170,
            child: Container(
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                gradient: RadialGradient(
                  colors: [
                    const Color(0xFF2D6A4F).withValues(alpha: 0.30),
                    Colors.transparent,
                  ],
                ),
              ),
            ),
          ),

          // Subtle Watermark Motif
          Positioned(
            right: -10,
            bottom: -15,
            child: IgnorePointer(
              child: Icon(
                LucideIcons.users,
                size: 130,
                color: Colors.white.withValues(alpha: 0.04),
              ),
            ),
          ),

          // Foreground Content
          SafeArea(
            bottom: false,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 18),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Navigation Bar Row (Back Button, Term Pill, Refresh Button)
                  Row(
                    children: [
                      BouncyButton(
                        onTap: _handleBack,
                        scaleDownFactor: 0.90,
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
                              color: Colors.white,
                              size: 18,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),

                      // Pill Badge
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 10,
                          vertical: 4.5,
                        ),
                        decoration: BoxDecoration(
                          color:
                              const Color(0xFF2D6A4F).withValues(alpha: 0.85),
                          borderRadius: BorderRadius.circular(20),
                          border: Border.all(
                            color: const Color(0xFF74C69D)
                                .withValues(alpha: 0.45),
                            width: 1.0,
                          ),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: const [
                            Icon(
                              LucideIcons.shieldCheck,
                              size: 12,
                              color: Color(0xFFB7E4C7),
                            ),
                            SizedBox(width: 5),
                            Text(
                              'Term: 2025–2027',
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

                      const Spacer(),

                      // Refresh button
                      BouncyButton(
                        onTap: () {
                          HapticFeedback.lightImpact();
                          ref.invalidate(stateExecutiveProvider);
                          ref.invalidate(bearerDistrictsProvider);
                          final selected = ref.read(selectedDistrictProvider);
                          if (selected != null) {
                            ref.invalidate(districtBearersFamily(selected));
                          }
                        },
                        scaleDownFactor: 0.90,
                        child: Container(
                          width: 38,
                          height: 38,
                          decoration: BoxDecoration(
                            color: Colors.white.withValues(alpha: 0.12),
                            shape: BoxShape.circle,
                            border: Border.all(
                              color: Colors.white.withValues(alpha: 0.2),
                              width: 1.0,
                            ),
                          ),
                          child: const Center(
                            child: Icon(
                              LucideIcons.refreshCw,
                              color: Colors.white,
                              size: 16,
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),

                  const SizedBox(height: 14),

                  // Main Heading & Tagline
                  const Text(
                    'TAGA Leadership',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 26,
                      fontWeight: FontWeight.w900,
                      letterSpacing: -0.6,
                      height: 1.1,
                    ),
                  ),
                  const SizedBox(height: 4),
                  const Text(
                    'Dedicated team of agriculture graduates serving across Tamil Nadu',
                    style: TextStyle(
                      color: Color(0xFFD8F3DC),
                      fontSize: 12.5,
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
                            title: 'State Executive',
                            icon: LucideIcons.crown,
                            isSelected: _selectedTab == 0,
                            onTap: () => _onTabSelected(0),
                          ),
                        ),
                        Expanded(
                          child: _TabButton(
                            title: 'District Bearers',
                            icon: LucideIcons.mapPin,
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
    );
  }

  // ---------------------------------------------------------------------------
  // 1. STATE EXECUTIVE SECTION
  // ---------------------------------------------------------------------------
  Widget _buildStateExecutiveSection() {
    final stateExecutiveAsync = ref.watch(stateExecutiveProvider);

    return stateExecutiveAsync.when(
      loading: () => const OfficeBearersShimmer(isDistrict: false),
      error: (error, _) => _buildErrorState(
        message: error.toString(),
        onRetry: () => ref.invalidate(stateExecutiveProvider),
      ),
      data: (members) {
        if (members.isEmpty) {
          return _buildEmptyState(
            title: 'No State Executive Members',
            message: 'State executive committee details will be updated soon.',
            onRefresh: () => ref.invalidate(stateExecutiveProvider),
          );
        }

        return RefreshIndicator(
          color: AppColors.primary,
          backgroundColor: Colors.white,
          onRefresh: () async => ref.invalidate(stateExecutiveProvider),
          child: ListView.builder(
            padding: const EdgeInsets.fromLTRB(16, 14, 16, 28),
            physics: const AlwaysScrollableScrollPhysics(
              parent: BouncingScrollPhysics(),
            ),
            itemCount: members.length + 1,
            itemBuilder: (context, index) {
              if (index == 0) {
                return _buildSectionTitle(
                  title: 'State Office Bearers',
                  subtitle:
                      'The core leadership team elected by association members',
                  count: members.length,
                );
              }
              final member = members[index - 1];
              return Padding(
                padding: const EdgeInsets.only(bottom: 12),
                child: StateExecutiveCard(
                  member: member,
                  onTap: () => _openStateExecutiveProfile(member),
                ),
              );
            },
          ),
        );
      },
    );
  }

  void _openStateExecutiveProfile(StateExecutiveMember member) {
    HapticFeedback.lightImpact();
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => StateExecutiveProfileScreen(member: member),
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // 2. DISTRICT BEARERS SECTION
  // ---------------------------------------------------------------------------
  Widget _buildDistrictBearersSection() {
    final districtsAsync = ref.watch(bearerDistrictsProvider);
    final selectedDistrict = ref.watch(selectedDistrictProvider);

    return districtsAsync.when(
      loading: () => const Padding(
        padding: EdgeInsets.all(16),
        child: OfficeBearersShimmer(isDistrict: true),
      ),
      error: (error, _) => _buildErrorState(
        message: error.toString(),
        onRetry: () => ref.invalidate(bearerDistrictsProvider),
      ),
      data: (districts) {
        if (districts.isEmpty) {
          return _buildEmptyState(
            title: 'No Districts Available',
            message: 'District records could not be loaded at this time.',
            onRefresh: () => ref.invalidate(bearerDistrictsProvider),
          );
        }

        return RefreshIndicator(
          color: AppColors.primary,
          backgroundColor: Colors.white,
          onRefresh: () async {
            ref.invalidate(bearerDistrictsProvider);
            if (selectedDistrict != null) {
              ref.invalidate(districtBearersFamily(selectedDistrict));
            }
          },
          child: SingleChildScrollView(
            physics: const AlwaysScrollableScrollPhysics(
              parent: BouncingScrollPhysics(),
            ),
            padding: const EdgeInsets.fromLTRB(16, 14, 16, 32),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                _buildSectionTitle(
                  title: 'District Office Bearers',
                  subtitle:
                      'Select a district to view office bearers serving that region',
                  count: districts.length,
                  countLabel: 'districts',
                ),

                // District Selector Card
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(18),
                    border: Border.all(
                      color: const Color(0xFFE2EBE5),
                      width: 1.1,
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: const Color(0xFF04190E).withValues(alpha: 0.05),
                        blurRadius: 12,
                        offset: const Offset(0, 3),
                      ),
                    ],
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.all(6),
                            decoration: BoxDecoration(
                              color: const Color(0xFFE8F5EE),
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: const Icon(
                              LucideIcons.mapPin,
                              size: 15,
                              color: AppColors.primary,
                            ),
                          ),
                          const SizedBox(width: 8),
                          const Text(
                            'Select District',
                            style: TextStyle(
                              fontSize: 15,
                              fontWeight: FontWeight.w800,
                              color: AppColors.textPrimary,
                              letterSpacing: -0.2,
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 6),
                      Text(
                        'Choose from any of the ${districts.length} districts in Tamil Nadu',
                        style: const TextStyle(
                          fontSize: 12,
                          color: AppColors.textSecondary,
                        ),
                      ),
                      const SizedBox(height: 12),

                      // Dropdown Menu
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 14),
                        decoration: BoxDecoration(
                          color: const Color(0xFFF8FAF9),
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(
                            color: selectedDistrict != null
                                ? AppColors.primary
                                : const Color(0xFFC8E6C9),
                            width: selectedDistrict != null ? 1.5 : 1.0,
                          ),
                        ),
                        child: DropdownButtonHideUnderline(
                          child: DropdownButton<String>(
                            value: selectedDistrict,
                            hint: const Text(
                              'Choose a district...',
                              style: TextStyle(
                                fontSize: 13.5,
                                color: Color(0xFF94A3B8),
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                            isExpanded: true,
                            icon: const Icon(
                              LucideIcons.chevronDown,
                              color: AppColors.primary,
                              size: 18,
                            ),
                            dropdownColor: Colors.white,
                            borderRadius: BorderRadius.circular(12),
                            items: districts.map((district) {
                              return DropdownMenuItem<String>(
                                value: district,
                                child: Text(
                                  district,
                                  style: const TextStyle(
                                    fontSize: 14,
                                    fontWeight: FontWeight.w600,
                                    color: AppColors.textPrimary,
                                  ),
                                ),
                              );
                            }).toList(),
                            onChanged: (value) {
                              if (value != null) {
                                HapticFeedback.selectionClick();
                                ref
                                    .read(selectedDistrictProvider.notifier)
                                    .selectDistrict(value);
                              }
                            },
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 18),

                // Directory Display Area
                if (selectedDistrict == null)
                  _buildNoDistrictSelectedHint()
                else
                  _buildDistrictDirectory(selectedDistrict),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildSectionTitle({
    required String title,
    required String subtitle,
    required int count,
    String countLabel = 'members',
  }) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(2, 4, 2, 12),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          Container(
            width: 3.5,
            height: 18,
            decoration: BoxDecoration(
              color: AppColors.primary,
              borderRadius: BorderRadius.circular(2),
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(
                    fontSize: 17,
                    fontWeight: FontWeight.w800,
                    color: AppColors.textPrimary,
                    letterSpacing: -0.3,
                  ),
                ),
                Text(
                  subtitle,
                  style: const TextStyle(
                    fontSize: 11.5,
                    color: AppColors.textSecondary,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
            decoration: BoxDecoration(
              color: const Color(0xFFE8F5EE),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: const Color(0xFF74C69D).withValues(alpha: 0.35),
              ),
            ),
            child: Text(
              '$count $countLabel',
              style: const TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.w700,
                color: AppColors.primaryDark,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildNoDistrictSelectedHint() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 38),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF04190E).withValues(alpha: 0.04),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: const Color(0xFFE8F5EE),
              shape: BoxShape.circle,
              border: Border.all(color: const Color(0xFFC8E6C9)),
            ),
            child: const Icon(
              LucideIcons.map,
              color: AppColors.primary,
              size: 32,
            ),
          ),
          const SizedBox(height: 14),
          const Text(
            'Select a District Above',
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w800,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 6),
          const Text(
            'Choose your district from the dropdown above to browse the official office bearers serving your region.',
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 12.5,
              color: AppColors.textSecondary,
              height: 1.45,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDistrictDirectory(String district) {
    final bearersAsync = ref.watch(districtBearersFamily(district));

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        // Directory Header
        Padding(
          padding: const EdgeInsets.fromLTRB(2, 4, 2, 10),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  const Icon(
                    LucideIcons.award,
                    size: 16,
                    color: AppColors.primary,
                  ),
                  const SizedBox(width: 7),
                  Text(
                    '$district District Directory',
                    style: const TextStyle(
                      fontSize: 15,
                      fontWeight: FontWeight.w800,
                      color: AppColors.primaryDark,
                    ),
                  ),
                ],
              ),
              bearersAsync.maybeWhen(
                data: (list) => Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                  decoration: BoxDecoration(
                    color: const Color(0xFFE8F5EE),
                    borderRadius: BorderRadius.circular(10),
                    border: Border.all(
                      color: const Color(0xFF74C69D).withValues(alpha: 0.35),
                    ),
                  ),
                  child: Text(
                    '${list.length} Bearers',
                    style: const TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w700,
                      color: AppColors.primaryDark,
                    ),
                  ),
                ),
                orElse: () => const SizedBox.shrink(),
              ),
            ],
          ),
        ),

        bearersAsync.when(
          loading: () => const OfficeBearersShimmer(isDistrict: true),
          error: (error, _) => _buildErrorState(
            message: error.toString(),
            onRetry: () => ref.invalidate(districtBearersFamily(district)),
          ),
          data: (bearers) {
            if (bearers.isEmpty) {
              return Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 20, vertical: 36),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(18),
                  border: Border.all(color: const Color(0xFFE2EBE5)),
                ),
                child: Column(
                  children: [
                    const Icon(
                      LucideIcons.userX,
                      color: Color(0xFF94A3B8),
                      size: 30,
                    ),
                    const SizedBox(height: 12),
                    const Text(
                      'Data Coming Soon',
                      style: TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.bold,
                        color: AppColors.textPrimary,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      'Office bearers data for $district district will be updated soon.',
                      textAlign: TextAlign.center,
                      style: const TextStyle(
                        fontSize: 12.5,
                        color: AppColors.textSecondary,
                      ),
                    ),
                  ],
                ),
              );
            }

            // Structured Table-like Container (NO PHOTOS, NO AVATARS)
            return Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(18),
                border: Border.all(color: const Color(0xFFE2EBE5), width: 1.1),
                boxShadow: [
                  BoxShadow(
                    color: const Color(0xFF04190E).withValues(alpha: 0.05),
                    blurRadius: 14,
                    offset: const Offset(0, 3),
                  ),
                ],
              ),
              child: ClipRRect(
                borderRadius: BorderRadius.circular(17),
                child: Column(
                  children: [
                    // Table Header Row
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 11,
                      ),
                      color: const Color(0xFFE8F5EE),
                      child: const Row(
                        children: [
                          Expanded(
                            flex: 6,
                            child: Text(
                              'NAME & DESIGNATION',
                              style: TextStyle(
                                fontSize: 11,
                                fontWeight: FontWeight.w800,
                                color: AppColors.primaryDark,
                                letterSpacing: 0.5,
                              ),
                            ),
                          ),
                          SizedBox(width: 8),
                          Text(
                            'ACTIONS',
                            style: TextStyle(
                              fontSize: 11,
                              fontWeight: FontWeight.w800,
                              color: AppColors.primaryDark,
                              letterSpacing: 0.5,
                            ),
                          ),
                          SizedBox(width: 14),
                        ],
                      ),
                    ),

                    // Table Body Rows
                    ...bearers.asMap().entries.map((entry) {
                      final index = entry.key;
                      final bearer = entry.value;
                      return DistrictBearerRow(
                        bearer: bearer,
                        isEven: index.isEven,
                        isLast: index == bearers.length - 1,
                        onTap: () =>
                            _openDistrictBearerProfile(bearer, district),
                      );
                    }),
                  ],
                ),
              ),
            );
          },
        ),
      ],
    );
  }

  void _openDistrictBearerProfile(
    DistrictOfficeBearer bearer,
    String district,
  ) {
    HapticFeedback.lightImpact();
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => DistrictBearerProfileScreen(
          bearer: bearer,
          district: district,
        ),
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // ERROR & EMPTY STATES
  // ---------------------------------------------------------------------------
  Widget _buildErrorState({
    required String message,
    required VoidCallback onRetry,
  }) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: const Color(0xFFFEE2E2),
                borderRadius: BorderRadius.circular(16),
              ),
              child: const Icon(
                LucideIcons.alertTriangle,
                color: Color(0xFFDC2626),
                size: 28,
              ),
            ),
            const SizedBox(height: 14),
            const Text(
              'Unable to Load Office Bearers',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              message,
              textAlign: TextAlign.center,
              style: const TextStyle(
                fontSize: 12.5,
                color: AppColors.textSecondary,
                height: 1.4,
              ),
            ),
            const SizedBox(height: 18),
            ElevatedButton.icon(
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.primary,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(
                  horizontal: 20,
                  vertical: 10,
                ),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(10),
                ),
                elevation: 0,
              ),
              icon: const Icon(LucideIcons.refreshCw, size: 14),
              label: const Text('Retry'),
              onPressed: onRetry,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyState({
    required String title,
    required String message,
    required VoidCallback onRefresh,
  }) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: const Color(0xFFE8F5EE),
                borderRadius: BorderRadius.circular(16),
              ),
              child: const Icon(
                LucideIcons.users,
                color: AppColors.primary,
                size: 28,
              ),
            ),
            const SizedBox(height: 14),
            Text(
              title,
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              message,
              textAlign: TextAlign.center,
              style: const TextStyle(
                fontSize: 12.5,
                color: AppColors.textSecondary,
                height: 1.4,
              ),
            ),
            const SizedBox(height: 18),
            OutlinedButton.icon(
              style: OutlinedButton.styleFrom(
                foregroundColor: AppColors.primary,
                side: const BorderSide(color: AppColors.primary),
                padding: const EdgeInsets.symmetric(
                  horizontal: 18,
                  vertical: 9,
                ),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(10),
                ),
              ),
              icon: const Icon(LucideIcons.refreshCw, size: 14),
              label: const Text('Refresh'),
              onPressed: onRefresh,
            ),
          ],
        ),
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
                    color: Colors.black.withValues(alpha: 0.12),
                    blurRadius: 6,
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
              color: isSelected ? AppColors.primaryDark : const Color(0xFFD8F3DC),
            ),
            const SizedBox(width: 7),
            Text(
              title,
              style: TextStyle(
                fontSize: 13,
                fontWeight: isSelected ? FontWeight.w800 : FontWeight.w600,
                color: isSelected ? AppColors.primaryDark : const Color(0xFFD8F3DC),
                letterSpacing: -0.1,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
