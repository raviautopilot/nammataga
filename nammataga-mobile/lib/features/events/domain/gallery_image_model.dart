import '../../../core/constants/api_endpoints.dart';

/// Domain model representing a photo memory from the NammaTAGA Gallery.
class GalleryImageModel {
  final String id;
  final String title;
  final String description;
  final String event;
  final String imageUrl;
  final String date;
  final int year;

  const GalleryImageModel({
    required this.id,
    required this.title,
    this.description = '',
    required this.event,
    required this.imageUrl,
    required this.date,
    required this.year,
  });

  factory GalleryImageModel.fromJson(Map<String, dynamic> json) {
    final titleStr = json['title'] as String? ?? 'TAGA Memory';
    final descStr = json['description'] as String? ??
        json['caption'] as String? ??
        titleStr;

    return GalleryImageModel(
      id: json['id']?.toString() ?? '',
      title: titleStr,
      description: descStr,
      event: json['event'] as String? ?? 'TAGA Event',
      imageUrl: json['imageUrl'] as String? ??
          json['ImageURL'] as String? ??
          json['image'] as String? ??
          '',
      date: json['date'] as String? ?? '',
      year: json['year'] as int? ?? DateTime.now().year,
    );
  }

  /// Exact description to display under the photo or in detail lightbox
  String get displayDescription =>
      description.isNotEmpty ? description : title;

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'description': description,
      'event': event,
      'imageUrl': imageUrl,
      'date': date,
      'year': year,
    };
  }

  /// Resolves the full static URL for network fetching with proper URI encoding.
  String get fullImageUrl {
    if (imageUrl.isEmpty) return '';

    if (imageUrl.startsWith('http://') || imageUrl.startsWith('https://')) {
      return imageUrl;
    }

    final base = ApiEndpoints.baseUrl.replaceAll(RegExp(r'/+$'), ''); // e.g. 'http://localhost:1801/api'

    String cleanPath = imageUrl;
    if (cleanPath.startsWith('/api/')) {
      cleanPath = cleanPath.substring(5); // strip leading '/api/'
    } else if (cleanPath.startsWith('/')) {
      cleanPath = cleanPath.substring(1); // strip leading '/'
    }

    // Safely encode each segment in path to preserve forward slashes while
    // safely percent-encoding spaces, ampersands, single quotes, commas, etc.
    // Specifically, single quotes (') must be encoded as %27 so web CSS url('')
    // does not terminate prematurely.
    final encodedSegments = cleanPath.split('/').map((segment) {
      try {
        final decoded = Uri.decodeComponent(segment);
        return Uri.encodeComponent(decoded)
            .replaceAll("'", '%27')
            .replaceAll('(', '%28')
            .replaceAll(')', '%29');
      } catch (_) {
        return Uri.encodeComponent(segment)
            .replaceAll("'", '%27')
            .replaceAll('(', '%28')
            .replaceAll(')', '%29');
      }
    }).join('/');

    return '$base/$encodedSegments';
  }
}
