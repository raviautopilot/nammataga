import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path_provider/path_provider.dart';

import '../../../core/network/api_client.dart';

/// Exception thrown when a document download or storage operation fails.
class DocumentDownloadException implements Exception {
  final String message;
  final int? statusCode;
  final DioException? dioException;
  final FileSystemException? fileSystemException;

  const DocumentDownloadException(
    this.message, {
    this.statusCode,
    this.dioException,
    this.fileSystemException,
  });

  @override
  String toString() => message;
}

/// Service managing document downloads, caching in application storage,
/// and exporting to device Downloads directory.
class DocumentDownloadService {
  final Dio _dio;
  final Set<String> _activeUrls = {};

  static DocumentDownloadService? _instance;

  /// Singleton instance accessor.
  static DocumentDownloadService get instance {
    _instance ??= DocumentDownloadService._internal();
    return _instance!;
  }

  /// Factory constructor returning the singleton or custom Dio instance.
  factory DocumentDownloadService({Dio? dio}) {
    if (dio != null) {
      return DocumentDownloadService._internal(dio: dio);
    }
    return instance;
  }

  DocumentDownloadService._internal({Dio? dio})
      : _dio = dio ?? ApiClient.instance;

  /// Set of URLs currently being downloaded to prevent duplicate simultaneous tasks.
  Set<String> get activeUrls => Set.unmodifiable(_activeUrls);

  /// Checks whether a specific document URL is actively downloading.
  bool isDownloading(String url) => _activeUrls.contains(url);

  /// Retrieves the internal app cache directory for storing documents.
  Future<Directory> _getCacheDirectory() async {
    final appDocDir = await getApplicationDocumentsDirectory();
    final dir = Directory('${appDocDir.path}/documents');
    if (!await dir.exists()) {
      await dir.create(recursive: true);
    }
    return dir;
  }

  /// Checks if a document with [fileName] already exists in app storage cache.
  /// Returns the cached [File] if valid and non-empty, otherwise returns `null`.
  Future<File?> getCachedDocument(String fileName) async {
    try {
      final dir = await _getCacheDirectory();
      final file = File('${dir.path}/$fileName');
      if (await file.exists() && await file.length() > 0) {
        return file;
      }
      return null;
    } catch (_) {
      return null;
    }
  }

  /// Downloads a document from [url] and caches it as [fileName] in app storage.
  ///
  /// - Reports live progress via [onProgress] `(receivedBytes, totalBytes)`.
  /// - Prevents duplicate concurrent downloads for the same [url].
  /// - Safely handles timeouts, network interruptions, and disk write errors.
  /// - If [saveToPublicDownloads] is true, also copies the document to device Downloads directory.
  /// - Returns the saved local [File].
  Future<File> downloadDocument(
    String url,
    String fileName, {
    void Function(int received, int total)? onProgress,
    bool saveToPublicDownloads = false,
  }) async {
    if (_activeUrls.contains(url)) {
      throw const DocumentDownloadException(
        'A download for this document is already in progress.',
      );
    }

    _activeUrls.add(url);
    File? tempPartFile;

    try {
      final cacheDir = await _getCacheDirectory();
      final targetFile = File('${cacheDir.path}/$fileName');
      tempPartFile = File('${cacheDir.path}/$fileName.part');

      // 1. If already cached and valid, reuse cached file
      if (await targetFile.exists() && await targetFile.length() > 0) {
        if (saveToPublicDownloads) {
          await copyToDownloadsDirectory(targetFile, fileName);
        }
        return targetFile;
      }

      // 2. Perform download via Dio with streaming response into temp file
      final response = await _dio.download(
        url,
        tempPartFile.path,
        onReceiveProgress: onProgress,
        options: Options(
          responseType: ResponseType.bytes,
          followRedirects: true,
          receiveTimeout: const Duration(seconds: 60),
          sendTimeout: const Duration(seconds: 30),
        ),
      );

      if (response.statusCode != null &&
          (response.statusCode! < 200 || response.statusCode! >= 300)) {
        throw DocumentDownloadException(
          'Server returned HTTP ${response.statusCode} while downloading document.',
          statusCode: response.statusCode,
        );
      }

      // 3. Verify downloaded file integrity
      if (!await tempPartFile.exists() || await tempPartFile.length() == 0) {
        throw const DocumentDownloadException(
          'Downloaded document is empty or corrupted.',
        );
      }

      // 4. Safely promote part file to target file
      if (await targetFile.exists()) {
        await targetFile.delete();
      }
      final savedFile = await tempPartFile.rename(targetFile.path);

      // 5. Optionally export to public device downloads directory
      if (saveToPublicDownloads) {
        await copyToDownloadsDirectory(savedFile, fileName);
      }

      return savedFile;
    } on DioException catch (e) {
      if (tempPartFile != null && await tempPartFile.exists()) {
        try {
          await tempPartFile.delete();
        } catch (_) {}
      }
      throw DocumentDownloadException(_mapDioError(e), dioException: e);
    } on FileSystemException catch (e) {
      if (tempPartFile != null && await tempPartFile.exists()) {
        try {
          await tempPartFile.delete();
        } catch (_) {}
      }
      throw DocumentDownloadException(
        'Storage write error: ${e.message}',
        fileSystemException: e,
      );
    } catch (e) {
      if (tempPartFile != null && await tempPartFile.exists()) {
        try {
          await tempPartFile.delete();
        } catch (_) {}
      }
      if (e is DocumentDownloadException) rethrow;
      throw DocumentDownloadException('Failed to download document: $e');
    } finally {
      _activeUrls.remove(url);
    }
  }

  static const MethodChannel _downloadsChannel =
      MethodChannel('com.ravi.taga/downloads');

  /// Copies an existing app-cached file to the device's public or accessible Downloads directory.
  Future<File> copyToDownloadsDirectory(File sourceFile, String fileName) async {
    // 1. On Android, export directly to public Downloads folder via MediaStore
    if (Platform.isAndroid) {
      try {
        final result = await _downloadsChannel.invokeMethod<String>(
          'saveToPublicDownloads',
          {
            'sourcePath': sourceFile.path,
            'fileName': fileName,
            'mimeType': 'application/pdf',
          },
        );
        if (result != null && result.isNotEmpty) {
          final publicFile = File(result);
          if (await publicFile.exists() && await publicFile.length() > 0) {
            return publicFile;
          }
        }
      } catch (_) {
        // Fallback to POSIX file copying if method channel is unavailable
      }
    }

    // 2. Standard filesystem copying fallback
    try {
      Directory? targetDir;

      if (Platform.isAndroid) {
        final standardPublicDownload =
            Directory('/storage/emulated/0/Download');
        if (await standardPublicDownload.exists()) {
          targetDir = standardPublicDownload;
        }
      }

      targetDir ??= await getDownloadsDirectory();
      targetDir ??= await getApplicationDocumentsDirectory();

      final destinationFile = File('${targetDir.path}/$fileName');
      return await sourceFile.copy(destinationFile.path);
    } catch (_) {
      try {
        final fallbackDir = await getApplicationDocumentsDirectory();
        final fallbackFile = File('${fallbackDir.path}/$fileName');
        if (fallbackFile.path == sourceFile.path) {
          return sourceFile;
        }
        return await sourceFile.copy(fallbackFile.path);
      } catch (_) {
        return sourceFile;
      }
    }
  }

  /// Maps DioException to human-readable error messages.
  String _mapDioError(DioException e) {
    switch (e.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return 'Connection timed out. Please check your internet connection.';
      case DioExceptionType.badResponse:
        final status = e.response?.statusCode;
        if (status == 401) {
          return 'Authentication required. Please sign in to download this document.';
        } else if (status == 403) {
          return 'Access restricted. Active subscription required to view this document.';
        } else if (status == 404) {
          return 'Document not found on the server.';
        }
        return 'Server returned error ($status). Please try again later.';
      case DioExceptionType.connectionError:
        return 'Network connection error. Please verify your connection.';
      case DioExceptionType.cancel:
        return 'Download was cancelled.';
      default:
        return e.message ?? 'An unexpected network error occurred.';
    }
  }
}

/// Riverpod provider for [DocumentDownloadService].
final documentDownloadServiceProvider =
    Provider<DocumentDownloadService>((ref) {
  return DocumentDownloadService.instance;
});
