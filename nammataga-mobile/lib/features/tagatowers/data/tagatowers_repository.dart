import 'package:dio/dio.dart';
import '../../../core/constants/api_endpoints.dart';
import '../../../core/network/api_client.dart';
import '../domain/booking_model.dart';
import '../domain/payment_model.dart';
import '../domain/room_model.dart';

/// Data repository handling all network communication for TAGA Towers
class TagaTowersRepository {
  final Dio _dio;

  TagaTowersRepository({Dio? dio}) : _dio = dio ?? ApiClient.instance;

  /// Fetches all rooms and filters out hidden rooms (`kurinchi`, `pasumai` or `hide == true`)
  /// Endpoint: GET /api/towers/rooms
  Future<List<Room>> getRooms() async {
    try {
      final response = await _dio.get(ApiEndpoints.towersRooms);
      if (response.statusCode == 200 && response.data is List) {
        final list = (response.data as List)
            .whereType<Map>()
            .map((item) => Room.fromJson(Map<String, dynamic>.from(item)))
            .where((room) {
              final idLower = room.id.toLowerCase().trim();
              return !room.hide && idLower != 'kurinchi' && idLower != 'pasumai';
            })
            .toList();
        return list;
      }
      return [];
    } on DioException catch (e) {
      throw Exception(_extractErrorMessage(e, 'Failed to fetch towers rooms'));
    } catch (e) {
      throw Exception('Failed to fetch towers rooms: $e');
    }
  }

  /// Bulk checks availability of all rooms across a date range
  /// Endpoint: GET /api/towers/availability-range?checkIn=YYYY-MM-DD&checkOut=YYYY-MM-DD
  Future<Map<String, RoomAvailability>> checkAvailabilityRange(
    String checkIn,
    String checkOut,
  ) async {
    try {
      final response = await _dio.get(
        ApiEndpoints.towersAvailabilityRange(checkIn, checkOut),
      );
      if (response.statusCode == 200 && response.data is Map) {
        final rawMap = response.data as Map;
        final Map<String, RoomAvailability> result = {};
        rawMap.forEach((key, value) {
          if (value is Map) {
            result[key.toString()] = RoomAvailability.fromJson(
              Map<String, dynamic>.from(value),
            );
          }
        });
        return result;
      }
      return {};
    } on DioException catch (e) {
      throw Exception(_extractErrorMessage(e, 'Failed to check room availability'));
    } catch (e) {
      throw Exception('Failed to check room availability: $e');
    }
  }

  /// Checks availability for a single room on a specific date
  /// Endpoint: GET /api/towers/availability?roomId=...&date=YYYY-MM-DD
  Future<RoomAvailability> checkAvailability(
    String roomId,
    String date,
  ) async {
    try {
      final response = await _dio.get(
        ApiEndpoints.towersAvailability(roomId, date),
      );
      if (response.statusCode == 200 && response.data is Map) {
        return RoomAvailability.fromJson(
          Map<String, dynamic>.from(response.data as Map),
        );
      }
      throw Exception('Invalid server response for room availability');
    } on DioException catch (e) {
      throw Exception(_extractErrorMessage(e, 'Failed to check room availability'));
    } catch (e) {
      throw Exception('Failed to check room availability: $e');
    }
  }

  /// Creates a provisional room booking (PaymentPending)
  /// Endpoint: POST /api/towers/bookings?bookerId=...
  Future<BookingResponse> createBooking(
    CreateBookingRequest req,
    String bookerId,
  ) async {
    try {
      final response = await _dio.post(
        ApiEndpoints.towersBookings,
        queryParameters: {'bookerId': bookerId},
        data: req.toJson(),
      );

      if ((response.statusCode == 200 || response.statusCode == 201) &&
          response.data is Map) {
        return BookingResponse.fromJson(
          Map<String, dynamic>.from(response.data as Map),
        );
      }
      throw Exception('Failed to create booking: invalid response');
    } on DioException catch (e) {
      throw Exception(_extractErrorMessage(e, 'Failed to create room booking'));
    } catch (e) {
      throw Exception('Failed to create room booking: $e');
    }
  }

  /// Retrieves active/upcoming bookings for a specific member
  /// Endpoint: GET /api/towers/bookings?bookerId=...
  Future<List<BookingResponse>> getUserBookings(String bookerId) async {
    try {
      final response = await _dio.get(
        ApiEndpoints.towersBookings,
        queryParameters: {'bookerId': bookerId},
      );

      if (response.statusCode == 200 && response.data is List) {
        return (response.data as List)
            .whereType<Map>()
            .map((item) => BookingResponse.fromJson(Map<String, dynamic>.from(item)))
            .toList();
      }
      return [];
    } on DioException catch (e) {
      throw Exception(_extractErrorMessage(e, 'Failed to fetch user bookings'));
    } catch (e) {
      throw Exception('Failed to fetch user bookings: $e');
    }
  }

  /// Retrieves past (archived) bookings for a member, optionally filtered by year/month
  /// Endpoint: GET /api/towers/bookings/past?bookerId=...&year=...&month=...
  Future<List<BookingResponse>> getPastBookings(
    String bookerId, [
    String? year,
    String? month,
  ]) async {
    try {
      final response = await _dio.get(
        ApiEndpoints.towersPastBookings(bookerId, year, month),
      );

      if (response.statusCode == 200 && response.data is List) {
        return (response.data as List)
            .whereType<Map>()
            .map((item) => BookingResponse.fromJson(Map<String, dynamic>.from(item)))
            .toList();
      }
      return [];
    } on DioException catch (e) {
      throw Exception(_extractErrorMessage(e, 'Failed to fetch past bookings'));
    } catch (e) {
      throw Exception('Failed to fetch past bookings: $e');
    }
  }

  /// Cancels an existing booking
  /// Endpoint: DELETE /api/towers/bookings/:id
  Future<bool> cancelBooking(String bookingId) async {
    try {
      final response = await _dio.delete(
        ApiEndpoints.towersCancelBooking(bookingId),
      );
      return response.statusCode == 200;
    } on DioException catch (e) {
      throw Exception(_extractErrorMessage(e, 'Failed to cancel booking'));
    } catch (e) {
      throw Exception('Failed to cancel booking: $e');
    }
  }

  /// Creates a Razorpay order for the booking advance payment
  /// Endpoint: POST /api/towers/create-order
  Future<TowersCreateOrderResponse> createOrder({
    required int amountInPaise,
    required Map<String, dynamic> notes,
  }) async {
    try {
      final response = await _dio.post(
        ApiEndpoints.towersCreateOrder,
        data: {
          'amount': amountInPaise,
          'notes': notes,
        },
      );

      if (response.statusCode == 200 && response.data is Map) {
        return TowersCreateOrderResponse.fromJson(
          Map<String, dynamic>.from(response.data as Map),
        );
      }
      throw Exception('Failed to create payment order: invalid response');
    } on DioException catch (e) {
      throw Exception(_extractErrorMessage(e, 'Failed to initialize payment'));
    } catch (e) {
      throw Exception('Failed to initialize payment: $e');
    }
  }

  /// Verifies a completed Razorpay payment and confirms the room booking
  /// Endpoint: POST /api/towers/verify-payment
  Future<bool> verifyPayment(VerifyTowersPaymentRequest req) async {
    try {
      final response = await _dio.post(
        ApiEndpoints.towersVerifyPayment,
        data: req.toJson(),
      );
      return response.statusCode == 200;
    } on DioException catch (e) {
      throw Exception(_extractErrorMessage(e, 'Payment verification failed'));
    } catch (e) {
      throw Exception('Payment verification failed: $e');
    }
  }

  String _extractErrorMessage(DioException e, String fallback) {
    if (e.response?.data is Map) {
      final data = e.response!.data as Map;
      if (data['error'] != null) {
        return data['error'].toString();
      }
      if (data['message'] != null) {
        return data['message'].toString();
      }
    }
    if (e.message != null && e.message!.isNotEmpty) {
      return e.message!;
    }
    return fallback;
  }
}
