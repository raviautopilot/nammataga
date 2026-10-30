import 'dart:ui';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import '../../../../core/widgets/bouncy_button.dart';
import '../../domain/gallery_image_model.dart';

/// Immersive full-screen photo viewer featuring high-performance sliding gestures,
/// an interactive synchronized thumbnail filmstrip, pinch/double-tap zoom,
/// and frosted glass metadata overlays displaying backend descriptions.
class GalleryLightbox extends StatefulWidget {
  final List<GalleryImageModel> images;
  final int initialIndex;

  const GalleryLightbox({
    super.key,
    required this.images,
    required this.initialIndex,
  });

  static void open(
    BuildContext context, {
    required List<GalleryImageModel> images,
    required int initialIndex,
  }) {
    Navigator.of(context).push(
      PageRouteBuilder(
        opaque: false,
        barrierColor: Colors.black.withValues(alpha: 0.96),
        pageBuilder: (context, anim1, anim2) => GalleryLightbox(
          images: images,
          initialIndex: initialIndex,
        ),
        transitionsBuilder: (context, anim, secondaryAnim, child) {
          return FadeTransition(
            opacity: CurvedAnimation(
              parent: anim,
              curve: Curves.easeOutCubic,
            ),
            child: child,
          );
        },
      ),
    );
  }

  @override
  State<GalleryLightbox> createState() => _GalleryLightboxState();
}

class _GalleryLightboxState extends State<GalleryLightbox> {
  late PageController _pageController;
  late ScrollController _thumbnailScrollController;
  late int _currentIndex;
  final TransformationController _transformController =
      TransformationController();
  bool _showControls = true;
  bool _isZoomed = false;

  @override
  void initState() {
    super.initState();
    _currentIndex = widget.initialIndex;
    _pageController = PageController(initialPage: widget.initialIndex);
    _thumbnailScrollController = ScrollController();
    _transformController.addListener(_handleZoomChange);

    // Initial thumbnail scroll alignment
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _scrollToThumbnail(_currentIndex, animate: false);
    });
  }

  void _handleZoomChange() {
    final scale = _transformController.value.getMaxScaleOnAxis();
    final isZoomed = scale > 1.05;
    if (isZoomed != _isZoomed) {
      setState(() {
        _isZoomed = isZoomed;
      });
    }
  }

  @override
  void dispose() {
    _transformController.removeListener(_handleZoomChange);
    _pageController.dispose();
    _thumbnailScrollController.dispose();
    _transformController.dispose();
    super.dispose();
  }

  void _toggleControls() {
    setState(() {
      _showControls = !_showControls;
    });
  }

  void _resetZoom() {
    if (_transformController.value != Matrix4.identity()) {
      _transformController.value = Matrix4.identity();
    }
    if (_isZoomed) {
      setState(() {
        _isZoomed = false;
      });
    }
  }

  void _onPageChanged(int index) {
    setState(() {
      _currentIndex = index;
    });
    _resetZoom();
    HapticFeedback.selectionClick();
    _scrollToThumbnail(index);
  }

  void _scrollToThumbnail(int index, {bool animate = true}) {
    if (!_thumbnailScrollController.hasClients) return;
    const itemExtent = 56.0; // thumbnail width (48) + margins (8)
    final screenWidth = MediaQuery.of(context).size.width;
    final targetOffset =
        (index * itemExtent) - (screenWidth / 2) + (itemExtent / 2);
    final clampedOffset = targetOffset.clamp(
      0.0,
      _thumbnailScrollController.position.maxScrollExtent,
    );

    if (animate) {
      _thumbnailScrollController.animateTo(
        clampedOffset,
        duration: const Duration(milliseconds: 320),
        curve: Curves.easeOutCubic,
      );
    } else {
      _thumbnailScrollController.jumpTo(clampedOffset);
    }
  }

  void _jumpToPhoto(int index) {
    if (index == _currentIndex) return;
    HapticFeedback.lightImpact();
    _resetZoom();
    _pageController.animateToPage(
      index,
      duration: const Duration(milliseconds: 320),
      curve: Curves.easeInOutCubic,
    );
  }

  @override
  Widget build(BuildContext context) {
    if (widget.images.isEmpty) {
      return const Scaffold(
        backgroundColor: Colors.black,
        body: Center(
          child: Text(
            'No photos available',
            style: TextStyle(color: Colors.white70),
          ),
        ),
      );
    }

    final currentImage = widget.images[_currentIndex];

    return Scaffold(
      backgroundColor: Colors.black,
      body: Stack(
        fit: StackFit.expand,
        children: [
          // Interactive PageView with Bounded Rubber-Band Swipe Physics
          PageView.builder(
            controller: _pageController,
            physics: _isZoomed
                ? const NeverScrollableScrollPhysics()
                : const BouncingScrollPhysics(),
            itemCount: widget.images.length,
            onPageChanged: _onPageChanged,
            itemBuilder: (context, index) {
              final img = widget.images[index];
              return GestureDetector(
                onTap: _toggleControls,
                onDoubleTap: () {
                  if (_transformController.value != Matrix4.identity()) {
                    _resetZoom();
                  } else {
                    _transformController.value =
                        Matrix4.diagonal3Values(2.5, 2.5, 1.0);
                  }
                },
                child: Container(
                  width: double.infinity,
                  height: double.infinity,
                  color: Colors.black,
                  alignment: Alignment.center,
                  child: InteractiveViewer(
                    transformationController: _transformController,
                    minScale: 0.85,
                    maxScale: 4.0,
                    child: Hero(
                      tag: 'gallery_image_${img.id}',
                      child: img.fullImageUrl.isNotEmpty
                          ? CachedNetworkImage(
                              imageUrl: img.fullImageUrl,
                              fit: BoxFit.contain, // Full uncropped display
                              alignment: Alignment.center,
                              placeholder: (context, url) => const Center(
                                child: SizedBox(
                                  width: 36,
                                  height: 36,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2.5,
                                    color: Color(0xFF52B788),
                                  ),
                                ),
                              ),
                              errorWidget: (context, url, error) => Container(
                                margin: const EdgeInsets.all(32),
                                padding: const EdgeInsets.all(24),
                                decoration: BoxDecoration(
                                  color: const Color(0xFF0D2C1B),
                                  borderRadius: BorderRadius.circular(20),
                                  border: Border.all(
                                    color: const Color(0xFF2D6A4F),
                                  ),
                                ),
                                child: Column(
                                  mainAxisSize: MainAxisSize.min,
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: const [
                                    Icon(
                                      LucideIcons.imageOff,
                                      size: 48,
                                      color: Color(0xFF74C69D),
                                    ),
                                    SizedBox(height: 14),
                                    Text(
                                      'Photo unavailable',
                                      textAlign: TextAlign.center,
                                      style: TextStyle(
                                        color: Colors.white70,
                                        fontSize: 13,
                                        fontWeight: FontWeight.w500,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            )
                          : const Icon(
                              LucideIcons.image,
                              size: 64,
                              color: Colors.white54,
                            ),
                    ),
                  ),
                ),
              );
            },
          ),

          // Top Header Bar with Close Button & Photo Counter Pill
          AnimatedPositioned(
            duration: const Duration(milliseconds: 240),
            curve: Curves.easeOutCubic,
            top: _showControls ? 0 : -110,
            left: 0,
            right: 0,
            child: Container(
              padding: EdgeInsets.fromLTRB(
                16,
                MediaQuery.of(context).padding.top + 8,
                16,
                18,
              ),
              decoration: const BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topCenter,
                  end: Alignment.bottomCenter,
                  colors: [
                    Colors.black87,
                    Colors.transparent,
                  ],
                ),
              ),
              child: Row(
                children: [
                  BouncyButton(
                    onTap: () => Navigator.of(context).pop(),
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(20),
                      child: BackdropFilter(
                        filter: ImageFilter.blur(sigmaX: 10, sigmaY: 10),
                        child: Container(
                          width: 40,
                          height: 40,
                          decoration: BoxDecoration(
                            color: Colors.white.withValues(alpha: 0.15),
                            shape: BoxShape.circle,
                            border: Border.all(
                              color: Colors.white.withValues(alpha: 0.28),
                              width: 1.1,
                            ),
                          ),
                          child: const Center(
                            child: Icon(
                              LucideIcons.x,
                              color: Colors.white,
                              size: 20,
                            ),
                          ),
                        ),
                      ),
                    ),
                  ),
                  const Spacer(),
                  // Counter Pill with Year
                  ClipRRect(
                    borderRadius: BorderRadius.circular(20),
                    child: BackdropFilter(
                      filter: ImageFilter.blur(sigmaX: 10, sigmaY: 10),
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 14,
                          vertical: 6,
                        ),
                        decoration: BoxDecoration(
                          color: const Color(0xFF0D2C1B).withValues(alpha: 0.85),
                          borderRadius: BorderRadius.circular(20),
                          border: Border.all(
                            color: const Color(0xFF52B788).withValues(alpha: 0.5),
                            width: 1.1,
                          ),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Container(
                              width: 6,
                              height: 6,
                              decoration: const BoxDecoration(
                                shape: BoxShape.circle,
                                color: Color(0xFF74C69D),
                              ),
                            ),
                            const SizedBox(width: 7),
                            Text(
                              '${_currentIndex + 1} of ${widget.images.length}',
                              style: const TextStyle(
                                color: Colors.white,
                                fontWeight: FontWeight.bold,
                                fontSize: 13,
                                letterSpacing: 0.3,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),

          // Bottom Controls: Interactive Thumbnail Filmstrip + Backend Description Card
          AnimatedPositioned(
            duration: const Duration(milliseconds: 240),
            curve: Curves.easeOutCubic,
            bottom: _showControls ? 0 : -260,
            left: 0,
            right: 0,
            child: Container(
              padding: EdgeInsets.fromLTRB(
                16,
                16,
                16,
                MediaQuery.of(context).padding.bottom + 12,
              ),
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.bottomCenter,
                  end: Alignment.topCenter,
                  colors: [
                    Colors.black.withValues(alpha: 0.95),
                    Colors.black.withValues(alpha: 0.82),
                    Colors.transparent,
                  ],
                  stops: const [0.0, 0.7, 1.0],
                ),
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Interactive Thumbnail Filmstrip (When multiple images exist)
                  if (widget.images.length > 1) ...[
                    SizedBox(
                      height: 52,
                      child: ListView.builder(
                        controller: _thumbnailScrollController,
                        scrollDirection: Axis.horizontal,
                        physics: const BouncingScrollPhysics(),
                        itemCount: widget.images.length,
                        itemBuilder: (context, idx) {
                          final img = widget.images[idx];
                          final isSelected = idx == _currentIndex;

                          return GestureDetector(
                            onTap: () => _jumpToPhoto(idx),
                            child: AnimatedContainer(
                              duration: const Duration(milliseconds: 220),
                              width: 48,
                              height: 48,
                              margin: const EdgeInsets.symmetric(horizontal: 4),
                              decoration: BoxDecoration(
                                borderRadius: BorderRadius.circular(10),
                                border: Border.all(
                                  color: isSelected
                                      ? const Color(0xFF52B788)
                                      : Colors.white.withValues(alpha: 0.22),
                                  width: isSelected ? 2.4 : 1.0,
                                ),
                                boxShadow: isSelected
                                    ? [
                                        BoxShadow(
                                          color: const Color(0xFF52B788)
                                              .withValues(alpha: 0.5),
                                          blurRadius: 8,
                                        ),
                                      ]
                                    : null,
                              ),
                              child: ClipRRect(
                                borderRadius: BorderRadius.circular(8),
                                child: img.fullImageUrl.isNotEmpty
                                    ? CachedNetworkImage(
                                        imageUrl: img.fullImageUrl,
                                        fit: BoxFit.cover,
                                        placeholder: (context, url) => Container(
                                          color: const Color(0xFF0D2C1B),
                                        ),
                                        errorWidget: (context, url, error) => Container(
                                          color: const Color(0xFF0D2C1B),
                                          child: const Icon(
                                            LucideIcons.image,
                                            size: 16,
                                            color: Colors.white54,
                                          ),
                                        ),
                                      )
                                    : Container(color: Colors.grey[900]),
                              ),
                            ),
                          );
                        },
                      ),
                    ),
                    const SizedBox(height: 14),
                  ],

                  // Frosted Glass Description Card (Showing exact backend description)
                  ClipRRect(
                    borderRadius: BorderRadius.circular(18),
                    child: BackdropFilter(
                      filter: ImageFilter.blur(sigmaX: 14, sigmaY: 14),
                      child: Container(
                        width: double.infinity,
                        padding: const EdgeInsets.fromLTRB(16, 14, 16, 14),
                        decoration: BoxDecoration(
                          color: Colors.white.withValues(alpha: 0.12),
                          borderRadius: BorderRadius.circular(18),
                          border: Border.all(
                            color: Colors.white.withValues(alpha: 0.2),
                            width: 1.0,
                          ),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            // Year pill
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 3,
                              ),
                              decoration: BoxDecoration(
                                color: const Color(0xFF1B4332).withValues(alpha: 0.8),
                                borderRadius: BorderRadius.circular(6),
                                border: Border.all(
                                  color: const Color(0xFF74C69D).withValues(alpha: 0.4),
                                  width: 0.8,
                                ),
                              ),
                              child: Text(
                                '${currentImage.year} Memory Archive',
                                style: const TextStyle(
                                  color: Color(0xFFD8F3DC),
                                  fontSize: 11,
                                  fontWeight: FontWeight.w700,
                                ),
                              ),
                            ),
                            const SizedBox(height: 8),

                            // Real backend description
                            Text(
                              currentImage.displayDescription,
                              style: const TextStyle(
                                color: Colors.white,
                                fontSize: 14.5,
                                fontWeight: FontWeight.w700,
                                height: 1.35,
                                letterSpacing: -0.2,
                              ),
                            ),
                          ],
                        ),
                      ),
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
}
