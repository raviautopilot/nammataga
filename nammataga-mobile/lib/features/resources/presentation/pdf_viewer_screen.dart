import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:open_filex/open_filex.dart';
import 'package:printing/printing.dart';
import 'package:share_plus/share_plus.dart';
import 'package:syncfusion_flutter_pdfviewer/pdfviewer.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../core/constants/app_colors.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../services/document_download_service.dart';

/// Full-featured in-app PDF viewing experience for NammaTAGA documents.
///
/// Features:
/// - Smooth rendering via [SfPdfViewer] (cached local file or streamed network)
/// - Pinch-to-zoom, double-tap zoom, and explicit zoom control buttons
/// - Page navigation with previous/next triggers and direct jump dialog
/// - Live download progress and caching for instant offline reuse
/// - AppBar actions: Download to device, Native Share sheet, Native Print, Open in External App
/// - Robust error recovery with Retry and Open in Browser fallbacks
class PdfViewerScreen extends ConsumerStatefulWidget {
  final String title;
  final String url;
  final String? localPath;

  const PdfViewerScreen({
    super.key,
    required this.title,
    required this.url,
    this.localPath,
  });

  @override
  ConsumerState<PdfViewerScreen> createState() => _PdfViewerScreenState();
}

class _PdfViewerScreenState extends ConsumerState<PdfViewerScreen> {
  late final PdfViewerController _pdfViewerController;

  File? _localFile;
  bool _isLoading = true;
  bool _isDownloading = false;
  bool _useNetworkFallback = false;
  double _downloadProgress = 0.0;
  String? _errorMessage;

  int _currentPage = 1;
  int _pageCount = 0;

  @override
  void initState() {
    super.initState();
    _pdfViewerController = PdfViewerController();
    _initializeDocument();
  }

  @override
  void dispose() {
    _pdfViewerController.dispose();
    super.dispose();
  }

  /// Extracts a safe filename from the document URL or title.
  String _getFileName() {
    final uri = Uri.tryParse(widget.url);
    final pathSegments = uri?.pathSegments ?? [];
    if (pathSegments.isNotEmpty) {
      final lastSegment = pathSegments.last.split('?').first.trim();
      if (lastSegment.toLowerCase().endsWith('.pdf')) {
        return Uri.decodeComponent(lastSegment);
      }
    }
    final sanitized = widget.title.replaceAll(RegExp(r'[\\/:*?"<>|]'), '_').trim();
    return sanitized.toLowerCase().endsWith('.pdf') ? sanitized : '$sanitized.pdf';
  }

  /// Initializes the document by checking existing local file, app cache,
  /// or initiating a progress-tracked download.
  Future<void> _initializeDocument() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
      _downloadProgress = 0.0;
      _useNetworkFallback = false;
    });

    // 1. Check if an explicit localPath was provided and exists
    if (widget.localPath != null && widget.localPath!.isNotEmpty) {
      final explicitFile = File(widget.localPath!);
      if (await explicitFile.exists() && await explicitFile.length() > 0) {
        if (mounted) {
          setState(() {
            _localFile = explicitFile;
            _isLoading = false;
          });
        }
        return;
      }
    }

    // 2. Check if the document is already cached in application storage
    final fileName = _getFileName();
    final cached = await DocumentDownloadService.instance.getCachedDocument(fileName);
    if (cached != null) {
      if (mounted) {
        setState(() {
          _localFile = cached;
          _isLoading = false;
        });
      }
      return;
    }

    // 3. Download the document to application cache with progress tracking
    setState(() => _isDownloading = true);
    try {
      final downloaded = await DocumentDownloadService.instance.downloadDocument(
        widget.url,
        fileName,
        onProgress: (received, total) {
          if (total > 0 && mounted) {
            setState(() {
              _downloadProgress = received / total;
            });
          }
        },
      );

      if (mounted) {
        setState(() {
          _localFile = downloaded;
          _isDownloading = false;
          _isLoading = false;
        });
      }
    } catch (_) {
      // If download failed, fall back to streaming directly via SfPdfViewer.network
      if (mounted) {
        setState(() {
          _isDownloading = false;
          _useNetworkFallback = true;
          _isLoading = false;
        });
      }
    }
  }

  /// Ensures a local file is available before triggering actions like Share, Print, or External viewer.
  Future<File?> _ensureLocalFile() async {
    if (_localFile != null && await _localFile!.exists()) {
      return _localFile;
    }

    if (!mounted) return null;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Downloading document for processing...'),
        duration: Duration(seconds: 2),
      ),
    );

    try {
      final fileName = _getFileName();
      final downloaded = await DocumentDownloadService.instance.downloadDocument(
        widget.url,
        fileName,
      );
      if (mounted) {
        setState(() => _localFile = downloaded);
      }
      return downloaded;
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to prepare document: $e'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
      return null;
    }
  }

  /// Exports the document to the device's public Downloads directory.
  Future<void> _handleDownloadToDevice() async {
    HapticFeedback.lightImpact();

    try {
      final fileName = _getFileName();
      if (_localFile != null && await _localFile!.exists()) {
        await DocumentDownloadService.instance.copyToDownloadsDirectory(
          _localFile!,
          fileName,
        );
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Saving document to device...'),
              duration: Duration(seconds: 2),
            ),
          );
        }
        final file = await DocumentDownloadService.instance.downloadDocument(
          widget.url,
          fileName,
          saveToPublicDownloads: true,
        );
        if (mounted) {
          setState(() => _localFile = file);
        }
      }

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Row(
              children: [
                const Icon(LucideIcons.checkCircle2, color: Colors.white, size: 20),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    'Saved "$fileName" to device Downloads',
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
            backgroundColor: AppColors.primary,
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Download failed: $e'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
    }
  }

  /// Triggers the native system share sheet.
  Future<void> _handleShare() async {
    HapticFeedback.lightImpact();
    final file = await _ensureLocalFile();
    if (file == null) return;

    try {
      await SharePlus.instance.share(
        ShareParams(
          files: [XFile(file.path)],
          text: widget.title,
          subject: widget.title,
        ),
      );
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Share failed: $e'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
    }
  }

  /// Sends the document to the system print dialog.
  Future<void> _handlePrint() async {
    HapticFeedback.lightImpact();
    final file = await _ensureLocalFile();
    if (file == null) return;

    try {
      final bytes = await file.readAsBytes();
      await Printing.layoutPdf(
        name: widget.title,
        onLayout: (format) async => bytes,
      );
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Print failed: $e'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
    }
  }

  /// Opens the PDF with the device's installed PDF viewers.
  Future<void> _handleOpenExternally() async {
    HapticFeedback.lightImpact();
    final file = await _ensureLocalFile();
    if (file == null) return;

    try {
      final result = await OpenFilex.open(file.path);
      if (result.type != ResultType.done && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Could not open external app: ${result.message}'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error launching external viewer: $e'),
            backgroundColor: AppColors.destructive,
          ),
        );
      }
    }
  }

  /// Prompts a dialog allowing user to jump directly to a page number.
  void _showPageJumpDialog() {
    if (_pageCount <= 1) return;
    final textController = TextEditingController(text: '$_currentPage');

    showDialog<void>(
      context: context,
      builder: (dialogContext) {
        return AlertDialog(
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
          title: const Text(
            'Jump to Page',
            style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Enter page number (1 – $_pageCount):',
                style: const TextStyle(fontSize: 13, color: AppColors.textSecondary),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: textController,
                keyboardType: TextInputType.number,
                autofocus: true,
                inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                decoration: InputDecoration(
                  filled: true,
                  fillColor: AppColors.muted,
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: const BorderSide(color: AppColors.borderLight),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(12),
                    borderSide: const BorderSide(color: AppColors.primary, width: 1.5),
                  ),
                ),
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(dialogContext).pop(),
              child: const Text('Cancel', style: TextStyle(color: AppColors.textSecondary)),
            ),
            ElevatedButton(
              onPressed: () {
                final target = int.tryParse(textController.text.trim());
                if (target != null && target >= 1 && target <= _pageCount) {
                  _pdfViewerController.jumpToPage(target);
                  Navigator.of(dialogContext).pop();
                }
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.primary,
                foregroundColor: Colors.white,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
              ),
              child: const Text('Go'),
            ),
          ],
        );
      },
    );
  }

  /// Opens the PDF URL in the default browser.
  Future<void> _openInBrowser() async {
    final uri = Uri.tryParse(widget.url);
    if (uri != null) {
      try {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Could not launch browser: $e'),
              backgroundColor: AppColors.destructive,
            ),
          );
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        titleSpacing: 0,
        backgroundColor: AppColors.surface,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(LucideIcons.arrowLeft, color: AppColors.primary),
          onPressed: () => Navigator.of(context).pop(),
          tooltip: 'Back',
        ),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              widget.title,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            if (_pageCount > 0)
              Text(
                'Page $_currentPage of $_pageCount',
                style: const TextStyle(
                  fontSize: 12,
                  color: AppColors.textSecondary,
                  fontWeight: FontWeight.w500,
                ),
              ),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(LucideIcons.download, size: 20),
            tooltip: 'Download to device',
            onPressed: _handleDownloadToDevice,
          ),
          IconButton(
            icon: const Icon(LucideIcons.share2, size: 20),
            tooltip: 'Share',
            onPressed: _handleShare,
          ),
          PopupMenuButton<String>(
            icon: const Icon(LucideIcons.moreVertical, size: 20),
            tooltip: 'More actions',
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
            onSelected: (value) {
              switch (value) {
                case 'print':
                  _handlePrint();
                  break;
                case 'external':
                  _handleOpenExternally();
                  break;
                case 'browser':
                  _openInBrowser();
                  break;
              }
            },
            itemBuilder: (context) => [
              const PopupMenuItem(
                value: 'print',
                child: Row(
                  children: [
                    Icon(LucideIcons.printer, size: 18, color: AppColors.textPrimary),
                    SizedBox(width: 12),
                    Text('Print Document', style: TextStyle(fontSize: 14)),
                  ],
                ),
              ),
              const PopupMenuItem(
                value: 'external',
                child: Row(
                  children: [
                    Icon(LucideIcons.externalLink, size: 18, color: AppColors.textPrimary),
                    SizedBox(width: 12),
                    Text('Open in External App', style: TextStyle(fontSize: 14)),
                  ],
                ),
              ),
              const PopupMenuItem(
                value: 'browser',
                child: Row(
                  children: [
                    Icon(LucideIcons.globe, size: 18, color: AppColors.textPrimary),
                    SizedBox(width: 12),
                    Text('Open in Browser', style: TextStyle(fontSize: 14)),
                  ],
                ),
              ),
            ],
          ),
        ],
      ),
      body: _buildBody(),
      bottomNavigationBar: _errorMessage == null && !_isLoading && !_isDownloading
          ? _buildBottomBar()
          : null,
    );
  }

  Widget _buildBody() {
    if (_isDownloading) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 40),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 72,
                height: 72,
                decoration: const BoxDecoration(
                  color: AppColors.secondary,
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  LucideIcons.fileDown,
                  color: AppColors.primary,
                  size: 34,
                ),
              ),
              const SizedBox(height: 20),
              const Text(
                'Downloading Document...',
                style: TextStyle(
                  fontSize: 17,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'Preparing file for optimal viewing & offline access',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 13,
                  color: AppColors.textSecondary.withValues(alpha: 0.9),
                ),
              ),
              const SizedBox(height: 24),
              ClipRRect(
                borderRadius: BorderRadius.circular(10),
                child: LinearProgressIndicator(
                  value: _downloadProgress > 0 ? _downloadProgress : null,
                  backgroundColor: AppColors.secondary,
                  valueColor: const AlwaysStoppedAnimation<Color>(AppColors.primary),
                  minHeight: 8,
                ),
              ),
              if (_downloadProgress > 0) ...[
                const SizedBox(height: 10),
                Text(
                  '${(_downloadProgress * 100).toInt()}%',
                  style: const TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                    color: AppColors.primary,
                  ),
                ),
              ],
            ],
          ),
        ),
      );
    }

    if (_errorMessage != null) {
      return _buildErrorView();
    }

    if (_isLoading) {
      return const Center(
        child: CircularProgressIndicator(color: AppColors.primary),
      );
    }

    // Build PDF Viewer (either from cached file or network fallback)
    if (_localFile != null) {
      return SfPdfViewer.file(
        _localFile!,
        controller: _pdfViewerController,
        canShowScrollHead: true,
        canShowScrollStatus: true,
        enableDoubleTapZooming: true,
        onDocumentLoaded: (details) {
          if (mounted) {
            setState(() {
              _pageCount = details.document.pages.count;
              _currentPage = _pdfViewerController.pageNumber;
            });
          }
        },
        onPageChanged: (details) {
          if (mounted) {
            setState(() {
              _currentPage = details.newPageNumber;
            });
          }
        },
        onDocumentLoadFailed: (details) {
          if (mounted) {
            setState(() {
              _errorMessage = details.description;
            });
          }
        },
      );
    }

    if (_useNetworkFallback) {
      return SfPdfViewer.network(
        widget.url,
        controller: _pdfViewerController,
        canShowScrollHead: true,
        canShowScrollStatus: true,
        enableDoubleTapZooming: true,
        onDocumentLoaded: (details) {
          if (mounted) {
            setState(() {
              _pageCount = details.document.pages.count;
              _currentPage = _pdfViewerController.pageNumber;
            });
          }
        },
        onPageChanged: (details) {
          if (mounted) {
            setState(() {
              _currentPage = details.newPageNumber;
            });
          }
        },
        onDocumentLoadFailed: (details) {
          if (mounted) {
            setState(() {
              _errorMessage = details.description;
            });
          }
        },
      );
    }

    return _buildErrorView();
  }

  /// Clean, informative error view with Retry and Open in Browser buttons.
  Widget _buildErrorView() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 80,
              height: 80,
              decoration: BoxDecoration(
                color: AppColors.destructive.withValues(alpha: 0.1),
                shape: BoxShape.circle,
              ),
              child: const Icon(
                LucideIcons.fileX,
                color: AppColors.destructive,
                size: 40,
              ),
            ),
            const SizedBox(height: 20),
            const Text(
              'Unable to Open Document',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: 10),
            Text(
              _errorMessage ??
                  'The document could not be loaded or parsed. Please verify your internet connection or open it in your browser.',
              textAlign: TextAlign.center,
              style: const TextStyle(
                fontSize: 13,
                color: AppColors.textSecondary,
                height: 1.4,
              ),
            ),
            const SizedBox(height: 28),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                BouncyButton(
                  onTap: () {
                    HapticFeedback.lightImpact();
                    _initializeDocument();
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
                    decoration: BoxDecoration(
                      color: AppColors.primary,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(LucideIcons.rotateCcw, size: 16, color: Colors.white),
                        SizedBox(width: 8),
                        Text(
                          'Retry',
                          style: TextStyle(
                            color: Colors.white,
                            fontWeight: FontWeight.w600,
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(width: 14),
                BouncyButton(
                  onTap: () {
                    HapticFeedback.lightImpact();
                    _openInBrowser();
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 12),
                    decoration: BoxDecoration(
                      color: AppColors.muted,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: AppColors.borderLight),
                    ),
                    child: const Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(LucideIcons.globe, size: 16, color: AppColors.textPrimary),
                        SizedBox(width: 8),
                        Text(
                          'Open in Browser',
                          style: TextStyle(
                            color: AppColors.textPrimary,
                            fontWeight: FontWeight.w600,
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  /// Bottom navigation and zoom toolbar.
  Widget _buildBottomBar() {
    return Container(
      padding: EdgeInsets.only(
        left: 16,
        right: 16,
        top: 8,
        bottom: 8 + MediaQuery.of(context).padding.bottom,
      ),
      decoration: BoxDecoration(
        color: AppColors.surface,
        border: Border(
          top: BorderSide(color: AppColors.borderLight),
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.04),
            blurRadius: 8,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          // Previous page button
          IconButton(
            icon: const Icon(LucideIcons.chevronLeft),
            color: AppColors.primary,
            disabledColor: AppColors.textMuted.withValues(alpha: 0.3),
            tooltip: 'Previous page',
            onPressed: _currentPage > 1
                ? () {
                    HapticFeedback.lightImpact();
                    _pdfViewerController.previousPage();
                  }
                : null,
          ),

          // Tappable Page indicator
          GestureDetector(
            onTap: _showPageJumpDialog,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
              decoration: BoxDecoration(
                color: AppColors.muted,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: AppColors.borderLight),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(LucideIcons.bookOpen, size: 14, color: AppColors.primary),
                  const SizedBox(width: 6),
                  Text(
                    _pageCount > 0 ? 'Page $_currentPage of $_pageCount' : 'Loading...',
                    style: const TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: AppColors.textPrimary,
                    ),
                  ),
                ],
              ),
            ),
          ),

          // Next page button
          IconButton(
            icon: const Icon(LucideIcons.chevronRight),
            color: AppColors.primary,
            disabledColor: AppColors.textMuted.withValues(alpha: 0.3),
            tooltip: 'Next page',
            onPressed: _currentPage < _pageCount
                ? () {
                    HapticFeedback.lightImpact();
                    _pdfViewerController.nextPage();
                  }
                : null,
          ),

          // Zoom controls divider
          Container(
            height: 24,
            width: 1,
            color: AppColors.borderLight,
          ),

          // Zoom Out
          IconButton(
            icon: const Icon(LucideIcons.zoomOut, size: 20),
            color: AppColors.textPrimary,
            tooltip: 'Zoom out',
            onPressed: () {
              HapticFeedback.lightImpact();
              final current = _pdfViewerController.zoomLevel;
              _pdfViewerController.zoomLevel = (current - 0.25).clamp(1.0, 3.5);
            },
          ),

          // Zoom In
          IconButton(
            icon: const Icon(LucideIcons.zoomIn, size: 20),
            color: AppColors.textPrimary,
            tooltip: 'Zoom in',
            onPressed: () {
              HapticFeedback.lightImpact();
              final current = _pdfViewerController.zoomLevel;
              _pdfViewerController.zoomLevel = (current + 0.25).clamp(1.0, 3.5);
            },
          ),

          // Reset Zoom
          IconButton(
            icon: const Icon(LucideIcons.rotateCcw, size: 18),
            color: AppColors.textSecondary,
            tooltip: 'Reset zoom',
            onPressed: () {
              HapticFeedback.lightImpact();
              _pdfViewerController.zoomLevel = 1.0;
            },
          ),
        ],
      ),
    );
  }
}
