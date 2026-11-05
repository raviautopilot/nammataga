import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../../core/constants/app_colors.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../../tagatowers/presentation/theme/tagatowers_design_tokens.dart';
import '../../domain/external_link.dart';
import '../../providers/resources_provider.dart';

/// Presentation widget displaying the official external government and departmental
/// links catalog for NammaTAGA members.
///
/// Supports explicit links passing (for tests or custom catalogs) or reactive watching
/// of [externalLinksProvider] with automatic fallback to curated official portals.
class ExternalLinksView extends ConsumerWidget {
  /// Optional explicit list of external links. If null, watches [externalLinksProvider].
  final List<ExternalLink>? links;

  /// Optional loading override flag.
  final bool? isLoading;

  /// Optional error message override.
  final String? errorMessage;

  /// Callback when user taps Retry in error state.
  final VoidCallback? onRetry;

  /// Optional tap interceptor callback (e.g. for testing). If null, launches the URL in browser.
  final void Function(ExternalLink link)? onLinkTap;

  /// Scroll physics for the list view.
  final ScrollPhysics? physics;

  /// Whether the list should shrink-wrap its contents.
  final bool shrinkWrap;

  /// Padding around the links catalog.
  final EdgeInsetsGeometry padding;

  const ExternalLinksView({
    super.key,
    this.links,
    this.isLoading,
    this.errorMessage,
    this.onRetry,
    this.onLinkTap,
    this.physics = const BouncingScrollPhysics(),
    this.shrinkWrap = false,
    this.padding = const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
  });

  /// Curated real Tamil Nadu & Central government agricultural engineering portals.
  static const List<ExternalLink> defaultGovernmentLinks = [
    ExternalLink(
      title: 'Agriculture Department',
      url: 'https://www.tnagrisnet.tn.gov.in/',
    ),
    ExternalLink(
      title: 'Horticulture Department',
      url: 'https://tnhorticulture.tn.gov.in/',
    ),
    ExternalLink(
      title: 'Marketing Department',
      url: 'https://www.agrimark.tn.gov.in/',
    ),
    ExternalLink(
      title: 'Agri Engineering Department',
      url: 'https://tnagriculture.in/common_portal/Agriengineering',
    ),
    ExternalLink(
      title: 'All Agri Scheme Dashboard',
      url: 'https://www.tnagrisnet.tn.gov.in/ARS/scheme/apc_dashboard/achReport#',
    ),
    ExternalLink(
      title: 'Kuruvai Dashboard',
      url: 'https://www.tnagrisnet.tn.gov.in/KSP/report/dashboard',
    ),
    ExternalLink(
      title: 'Uzhavar App Portal',
      url: 'https://www.tnagrisnet.tn.gov.in/people_app/dashboard/main/en',
    ),
    ExternalLink(
      title: 'Officials Contact List',
      url: 'https://www.tnagrisnet.tn.gov.in/home/contact/',
    ),
    ExternalLink(
      title: 'TNAU Agritech Portal',
      url: 'http://www.agritech.tnau.ac.in/',
    ),
    ExternalLink(
      title: 'Seed Sathi Portal',
      url: 'https://seedtrace.gov.in/ms014/english',
    ),
    ExternalLink(
      title: 'Seed Certification',
      url: 'https://seedcertification.tn.gov.in/seedcertification/',
    ),
    ExternalLink(
      title: 'Seednet India',
      url: 'https://seednet.gov.in/',
    ),
    ExternalLink(
      title: 'TN Govt Orders',
      url: 'https://www.tn.gov.in/godept_list.php',
    ),
    ExternalLink(
      title: 'TN Contact Directory',
      url: 'https://www.tn.gov.in/contact_directory.php',
    ),
    ExternalLink(
      title: 'TN Press Release',
      url: 'https://www.tn.gov.in/press_release.php',
    ),
    ExternalLink(
      title: 'ICAR Website',
      url: 'https://icar.org.in/en/node',
    ),
    ExternalLink(
      title: 'PPQS',
      url: 'https://www.ppqs.gov.in/en',
    ),
    ExternalLink(
      title: 'Agri Welfare',
      url: 'https://www.agriwelfare.gov.in/',
    ),
    ExternalLink(
      title: 'Agmarknet',
      url: 'https://agmarknet.gov.in/home',
    ),
    ExternalLink(
      title: 'Income Tax Portal',
      url: 'https://eportal.incometax.gov.in/iec/foservices/#/login',
    ),
    ExternalLink(
      title: 'CPS Portal',
      url: 'http://cps.tn.gov.in/public/index.php',
    ),
    ExternalLink(
      title: 'IFHRMS Portal',
      url: 'https://www.karuvoolam.tn.gov.in/web/tnta/',
    ),
  ];

  /// Normalizes and safely launches an external portal URL in an external browser.
  static Future<bool> launchExternalUrl(BuildContext context, String rawUrl) async {
    var finalUrl = rawUrl.trim();
    if (finalUrl.isEmpty) {
      if (context.mounted) {
        _showSnackBar(context, 'Invalid external link URL.', isError: true);
      }
      return false;
    }

    // URL normalization: if !url.startsWith('http://') && !url.startsWith('https://'), prepends https://
    if (!finalUrl.startsWith('http://') && !finalUrl.startsWith('https://')) {
      finalUrl = 'https://$finalUrl';
    }

    try {
      final uri = Uri.parse(finalUrl);
      final launched = await launchUrl(
        uri,
        mode: LaunchMode.externalApplication,
      );

      if (!launched && context.mounted) {
        _showSnackBar(
          context,
          'Could not launch external portal: $finalUrl',
          isError: true,
        );
      }
      return launched;
    } catch (e) {
      if (context.mounted) {
        _showSnackBar(
          context,
          'Failed to open portal. Please check your web browser.',
          isError: true,
        );
      }
      return false;
    }
  }

  static void _showSnackBar(
    BuildContext context,
    String message, {
    bool isError = false,
  }) {
    ScaffoldMessenger.of(context).clearSnackBars();
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        behavior: SnackBarBehavior.floating,
        backgroundColor: isError ? AppColors.destructive : AppColors.primaryDark,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        margin: const EdgeInsets.all(16),
        content: Row(
          children: [
            Icon(
              isError ? LucideIcons.alertCircle : LucideIcons.externalLink,
              color: Colors.white,
              size: 18,
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Text(
                message,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 13,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Extracts a clean host domain for subtle secondary labeling (e.g. "tnagrisnet.tn.gov.in").
  static String _extractDomain(String rawUrl) {
    try {
      var normalized = rawUrl.trim();
      if (!normalized.startsWith('http://') && !normalized.startsWith('https://')) {
        normalized = 'https://$normalized';
      }
      final uri = Uri.parse(normalized);
      final host = uri.host.replaceFirst(RegExp(r'^www\.'), '');
      return host.isNotEmpty ? host : 'Official Portal';
    } catch (_) {
      return 'Official Portal';
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    // 1. If explicit links were provided, render direct presentation
    if (links != null) {
      if (isLoading == true) {
        return _buildLoadingState();
      }
      if (errorMessage != null && errorMessage!.isNotEmpty) {
        return _buildErrorState(context, errorMessage!, onRetry);
      }
      if (links!.isEmpty) {
        return _buildEmptyState(context);
      }
      return _buildLinksList(context, links!);
    }

    // 2. Otherwise watch Riverpod provider
    final linksAsync = ref.watch(externalLinksProvider);

    return linksAsync.when(
      loading: () => _buildLoadingState(),
      error: (error, _) => _buildErrorState(
        context,
        error.toString(),
        () => ref.refresh(externalLinksProvider),
      ),
      data: (fetchedLinks) {
        final effectiveLinks = fetchedLinks.isNotEmpty
            ? fetchedLinks
            : defaultGovernmentLinks;

        if (effectiveLinks.isEmpty) {
          return _buildEmptyState(context);
        }
        return _buildLinksList(context, effectiveLinks);
      },
    );
  }

  /// Builds the full scrollable links catalog.
  Widget _buildLinksList(BuildContext context, List<ExternalLink> items) {
    return ListView.separated(
      physics: physics,
      shrinkWrap: shrinkWrap,
      padding: padding,
      itemCount: items.length,
      separatorBuilder: (_, _) => const SizedBox(height: 12),
      itemBuilder: (context, index) {
        final item = items[index];
        return _buildLinkTile(context, item);
      },
    );
  }

  /// Builds an individual portal link tile following TAGA design system.
  Widget _buildLinkTile(BuildContext context, ExternalLink link) {
    final domain = _extractDomain(link.url);

    return BouncyButton(
      onTap: () {
        HapticFeedback.lightImpact();
        if (onLinkTap != null) {
          onLinkTap!(link);
        } else {
          launchExternalUrl(context, link.url);
        }
      },
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        decoration: BoxDecoration(
          color: AppColors.cardBackground,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: TagaTowersDesignTokens.subtleBorder,
            width: 1.2,
          ),
          boxShadow: TagaTowersDesignTokens.softShadow,
        ),
        child: Row(
          children: [
            // Left portal emblem badge
            Container(
              width: 44,
              height: 44,
              decoration: BoxDecoration(
                color: AppColors.secondary.withValues(alpha: 0.55),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: AppColors.secondaryDark.withValues(alpha: 0.4),
                  width: 1.2,
                ),
              ),
              child: const Center(
                child: Icon(
                  LucideIcons.globe,
                  color: AppColors.primaryDark,
                  size: 20,
                ),
              ),
            ),
            const SizedBox(width: 14),

            // Portal title and domain
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    link.title,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      fontSize: 15,
                      fontWeight: FontWeight.w600,
                      color: AppColors.primaryDark,
                      height: 1.25,
                      letterSpacing: -0.2,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      Container(
                        width: 6,
                        height: 6,
                        decoration: const BoxDecoration(
                          shape: BoxShape.circle,
                          color: AppColors.primaryLight,
                        ),
                      ),
                      const SizedBox(width: 6),
                      Expanded(
                        child: Text(
                          domain,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            fontSize: 12,
                            color: AppColors.textSecondary,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),

            // Trailing external link icon
            Container(
              width: 34,
              height: 34,
              decoration: BoxDecoration(
                color: AppColors.muted,
                borderRadius: BorderRadius.circular(10),
                border: Border.all(
                  color: AppColors.borderLight,
                  width: 1,
                ),
              ),
              child: const Center(
                child: Icon(
                  LucideIcons.externalLink,
                  color: AppColors.primary,
                  size: 16,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Empty catalog state.
  Widget _buildEmptyState(BuildContext context) {
    return Center(
      child: SingleChildScrollView(
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 48),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 72,
              height: 72,
              decoration: BoxDecoration(
                color: AppColors.muted,
                shape: BoxShape.circle,
                border: Border.all(
                  color: AppColors.borderLight,
                  width: 1.5,
                ),
              ),
              child: Center(
                child: Icon(
                  LucideIcons.externalLink,
                  size: 32,
                  color: AppColors.textSecondary.withValues(alpha: 0.7),
                ),
              ),
            ),
            const SizedBox(height: 18),
            const Text(
              'No external links found',
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 17,
                fontWeight: FontWeight.bold,
                color: AppColors.primaryDark,
                letterSpacing: -0.2,
              ),
            ),
            const SizedBox(height: 8),
            const Text(
              'Official government and departmental portals will appear here when published.',
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 13,
                color: AppColors.textSecondary,
                height: 1.45,
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// Helpful error card with Retry action.
  Widget _buildErrorState(
    BuildContext context,
    String message,
    VoidCallback? onRetryCallback,
  ) {
    return Center(
      child: SingleChildScrollView(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 32),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 24),
          decoration: BoxDecoration(
            color: const Color(0xFFFFFBEB),
            borderRadius: BorderRadius.circular(20),
            border: Border.all(
              color: const Color(0xFFFCD34D),
              width: 1.2,
            ),
            boxShadow: const [
              BoxShadow(
                color: Color(0x0A000000),
                blurRadius: 16,
                offset: Offset(0, 4),
              ),
            ],
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 54,
                height: 54,
                decoration: BoxDecoration(
                  color: const Color(0xFFFEF3C7),
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: const Color(0xFFF59E0B).withValues(alpha: 0.5),
                    width: 1.2,
                  ),
                ),
                child: const Center(
                  child: Icon(
                    LucideIcons.alertCircle,
                    color: Color(0xFFD97706),
                    size: 26,
                  ),
                ),
              ),
              const SizedBox(height: 14),
              const Text(
                'Unable to Load External Links',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF78350F),
                ),
              ),
              const SizedBox(height: 8),
              Text(
                message.isNotEmpty
                    ? message
                    : 'A connection issue occurred while fetching the official links directory.',
                textAlign: TextAlign.center,
                style: const TextStyle(
                  fontSize: 13,
                  color: Color(0xFF92400E),
                  height: 1.4,
                ),
              ),
              const SizedBox(height: 18),
              if (onRetryCallback != null)
                BouncyButton(
                  onTap: () {
                    HapticFeedback.lightImpact();
                    onRetryCallback();
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 20,
                      vertical: 10,
                    ),
                    decoration: BoxDecoration(
                      color: AppColors.primary,
                      borderRadius: BorderRadius.circular(12),
                      boxShadow: [
                        BoxShadow(
                          color: AppColors.primary.withValues(alpha: 0.25),
                          blurRadius: 8,
                          offset: const Offset(0, 3),
                        ),
                      ],
                    ),
                    child: const Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          LucideIcons.refreshCw,
                          color: Colors.white,
                          size: 15,
                        ),
                        SizedBox(width: 8),
                        Text(
                          'Retry',
                          style: TextStyle(
                            color: Colors.white,
                            fontSize: 13.5,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }

  /// Shimmering / clean loading state.
  Widget _buildLoadingState() {
    return SingleChildScrollView(
      physics: const NeverScrollableScrollPhysics(),
      padding: padding,
      child: Column(
        children: List.generate(
          5,
          (index) => Padding(
            padding: const EdgeInsets.only(bottom: 12),
            child: Container(
              height: 72,
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(
                  color: TagaTowersDesignTokens.subtleBorder,
                  width: 1.2,
                ),
              ),
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
              child: Row(
                children: [
                  Container(
                    width: 44,
                    height: 44,
                    decoration: BoxDecoration(
                      color: AppColors.muted,
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Container(
                          height: 14,
                          width: double.infinity,
                          decoration: BoxDecoration(
                            color: AppColors.muted,
                            borderRadius: BorderRadius.circular(4),
                          ),
                        ),
                        const SizedBox(height: 8),
                        Container(
                          height: 10,
                          width: 120,
                          decoration: BoxDecoration(
                            color: AppColors.muted,
                            borderRadius: BorderRadius.circular(4),
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 12),
                  Container(
                    width: 32,
                    height: 32,
                    decoration: BoxDecoration(
                      color: AppColors.muted,
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
