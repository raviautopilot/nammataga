import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../auth/providers/auth_provider.dart';
import '../data/tagatowers_repository.dart';
import '../domain/booking_model.dart';
import '../domain/room_model.dart';
export 'booking_flow_provider.dart';

/// Repository Provider
final tagaTowersRepositoryProvider = Provider<TagaTowersRepository>((ref) {
  return TagaTowersRepository();
});

/// FutureProvider for the active rooms list
final towersRoomsProvider = FutureProvider.autoDispose<List<Room>>((ref) async {
  final repo = ref.watch(tagaTowersRepositoryProvider);
  return repo.getRooms();
});

/// Alias for towersRoomsProvider
final roomsListProvider = towersRoomsProvider;

/// FutureProvider.family for bulk availability over a date range
final towersAvailabilityRangeProvider = FutureProvider.autoDispose
    .family<Map<String, RoomAvailability>, DateRangeParams>((ref, params) async {
  final repo = ref.watch(tagaTowersRepositoryProvider);
  return repo.checkAvailabilityRange(
    params.checkInFormatted,
    params.checkOutFormatted,
  );
});

/// Alias for towersAvailabilityRangeProvider
final availabilityRangeProvider = towersAvailabilityRangeProvider;

/// FutureProvider for user's active/upcoming bookings
final userBookingsProvider = FutureProvider.autoDispose
    .family<List<BookingResponse>, String>((ref, bookerId) async {
  if (bookerId.isEmpty) return [];
  final repo = ref.watch(tagaTowersRepositoryProvider);
  return repo.getUserBookings(bookerId);
});

/// FutureProvider for the current logged-in member's active/upcoming bookings
final currentUserBookingsProvider =
    FutureProvider.autoDispose<List<BookingResponse>>((ref) async {
  final session = ref.watch(authProvider).value;
  final userMap = session?.user;
  final bookerId = (userMap?['taga_id'] ??
          userMap?['tagaId'] ??
          userMap?['sno']?.toString() ??
          userMap?['id']?.toString() ??
          '')
      .toString();
  if (bookerId.isEmpty) return [];
  final repo = ref.watch(tagaTowersRepositoryProvider);
  return repo.getUserBookings(bookerId);
});

/// FutureProvider for user's past archived bookings
final pastBookingsProvider = FutureProvider.autoDispose
    .family<List<BookingResponse>, String?>((ref, year) async {
  final session = ref.watch(authProvider).value;
  final userMap = session?.user;
  final bookerId = (userMap?['taga_id'] ??
          userMap?['tagaId'] ??
          userMap?['sno']?.toString() ??
          userMap?['id']?.toString() ??
          '')
      .toString();
  if (bookerId.isEmpty) return [];
  final repo = ref.watch(tagaTowersRepositoryProvider);
  return repo.getPastBookings(bookerId, year);
});

/// FutureProvider to fetch a single booking by ID (searches both active and past bookings)
final bookingByIdProvider = FutureProvider.autoDispose
    .family<BookingResponse?, String>((ref, bookingId) async {
  if (bookingId.isEmpty) return null;
  final session = ref.watch(authProvider).value;
  final userMap = session?.user;
  final bookerId = (userMap?['taga_id'] ??
          userMap?['tagaId'] ??
          userMap?['sno']?.toString() ??
          userMap?['id']?.toString() ??
          '')
      .toString();
  if (bookerId.isEmpty) return null;

  final repo = ref.watch(tagaTowersRepositoryProvider);

  try {
    final active = await repo.getUserBookings(bookerId);
    final found = active.where((b) => b.id == bookingId).firstOrNull;
    if (found != null) return found;
  } catch (_) {}

  try {
    final past = await repo.getPastBookings(bookerId);
    final found = past.where((b) => b.id == bookingId).firstOrNull;
    if (found != null) return found;
  } catch (_) {}

  return null;
});

