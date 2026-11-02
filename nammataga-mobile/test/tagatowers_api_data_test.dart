import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:dio/dio.dart';
import 'package:nammataga_mobile/core/constants/api_endpoints.dart';
import 'package:nammataga_mobile/features/tagatowers/domain/room_model.dart';
import 'package:nammataga_mobile/features/tagatowers/domain/booking_model.dart';
import 'package:nammataga_mobile/features/tagatowers/domain/payment_model.dart';
import 'package:nammataga_mobile/features/tagatowers/data/tagatowers_repository.dart';
import 'package:nammataga_mobile/features/tagatowers/providers/tagatowers_provider.dart';

// Simple mock Dio adapter for testing repository network flows
class MockDioInterceptor extends Interceptor {
  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    final path = options.path;

    if (path.contains('/towers/rooms')) {
      handler.resolve(Response(
        requestOptions: options,
        statusCode: 200,
        data: [
          {'id': 'apex-1', 'name': 'Apex Suite A/C', 'type': 'apex-suite', 'capacity': 3, 'allowSingleBed': true},
          {'id': 'kurinchi', 'name': 'Kurinchi', 'type': 'ac-room', 'capacity': 2, 'allowSingleBed': true, 'hide': true},
          {'id': 'pavalam', 'name': 'Pavalam', 'type': 'ac-room', 'capacity': 2, 'allowSingleBed': true},
          {'id': 'pasumai', 'name': 'Pasumai', 'type': 'ac-room', 'capacity': 2, 'allowSingleBed': true, 'hide': true},
          {'id': 'hidden-room', 'name': 'Hidden Room', 'type': 'ac-room', 'capacity': 2, 'allowSingleBed': true, 'hide': true},
          {'id': 'gents-dorm', 'name': 'Gents Dormitory', 'type': 'gents-dorm', 'capacity': 12, 'allowSingleBed': true},
        ],
      ));
      return;
    }

    if (path.contains('/towers/availability-range')) {
      handler.resolve(Response(
        requestOptions: options,
        statusCode: 200,
        data: {
          'apex-1': {
            'room': {'id': 'apex-1', 'name': 'Apex Suite', 'type': 'apex-suite', 'capacity': 3, 'allowSingleBed': true},
            'available': true,
            'availableBeds': 3,
            'genderRestriction': '',
          },
          'gents-dorm': {
            'room': {'id': 'gents-dorm', 'name': 'Gents Dorm', 'type': 'gents-dorm', 'capacity': 12, 'allowSingleBed': true},
            'available': true,
            'availableBeds': 8,
            'genderRestriction': 'male',
          },
        },
      ));
      return;
    }

    if (path.contains('/towers/availability')) {
      handler.resolve(Response(
        requestOptions: options,
        statusCode: 200,
        data: {
          'room': {'id': 'apex-1', 'name': 'Apex Suite', 'type': 'apex-suite', 'capacity': 3, 'allowSingleBed': true},
          'available': true,
          'availableBeds': 2,
          'genderRestriction': '',
        },
      ));
      return;
    }

    if (path.contains('/towers/bookings/past')) {
      handler.resolve(Response(
        requestOptions: options,
        statusCode: 200,
        data: [
          {
            'id': 'BK1002',
            'roomId': 'pavalam',
            'roomName': 'Pavalam',
            'checkInDate': '2025-10-01',
            'checkOutDate': '2025-10-03',
            'bookerName': 'Jane Doe',
            'bookerId': 'TAGA001',
            'bookingFor': 'self',
            'bedCount': 1,
            'paymentStatus': 'confirmed',
            'advanceAmount': 200,
            'bookingStatus': 'completed',
          }
        ],
      ));
      return;
    }

    if (path.contains('/towers/bookings') && options.method == 'POST') {
      handler.resolve(Response(
        requestOptions: options,
        statusCode: 201,
        data: {
          'id': 'BK1001',
          'roomId': 'apex-1',
          'roomName': 'Apex Suite',
          'checkInDate': '2026-12-01',
          'checkOutDate': '2026-12-02',
          'bookerName': 'John Doe',
          'bookerId': 'TAGA001',
          'bookingFor': 'self',
          'bedCount': 1,
          'paymentStatus': 'pending',
          'advanceAmount': 200,
          'bookingStatus': 'upcoming',
        },
      ));
      return;
    }

    if (path.contains('/towers/bookings') && options.method == 'GET') {
      handler.resolve(Response(
        requestOptions: options,
        statusCode: 200,
        data: [
          {
            'id': 'BK1001',
            'roomId': 'apex-1',
            'roomName': 'Apex Suite',
            'checkInDate': '2026-12-01',
            'checkOutDate': '2026-12-02',
            'bookerName': 'John Doe',
            'bookerId': 'TAGA001',
            'bookingFor': 'self',
            'bedCount': 1,
            'paymentStatus': 'confirmed',
            'advanceAmount': 200,
            'bookingStatus': 'upcoming',
          }
        ],
      ));
      return;
    }

    if (path.contains('/towers/bookings/') && options.method == 'DELETE') {
      handler.resolve(Response(
        requestOptions: options,
        statusCode: 200,
        data: {'message': 'Booking cancelled'},
      ));
      return;
    }

    if (path.contains('/towers/create-order') && options.method == 'POST') {
      handler.resolve(Response(
        requestOptions: options,
        statusCode: 200,
        data: {
          'key': 'rzp_test_123',
          'order': {
            'id': 'order_998877',
            'amount': 20000,
            'currency': 'INR',
          },
        },
      ));
      return;
    }

    if (path.contains('/towers/verify-payment') && options.method == 'POST') {
      handler.resolve(Response(
        requestOptions: options,
        statusCode: 200,
        data: {'message': 'Payment confirmed'},
      ));
      return;
    }

    handler.next(options);
  }
}

void main() {
  group('ApiEndpoints TAGA Towers verification', () {
    test('Endpoints URLs are well formed', () {
      expect(ApiEndpoints.towersRooms, contains('/towers/rooms'));
      expect(ApiEndpoints.towersAvailability('apex-1', '2026-11-01'), contains('/towers/availability?roomId=apex-1&date=2026-11-01'));
      expect(ApiEndpoints.towersAvailabilityRange('2026-11-01', '2026-11-03'), contains('/towers/availability-range?checkIn=2026-11-01&checkOut=2026-11-03'));
      expect(ApiEndpoints.towersBookings, contains('/towers/bookings'));
      expect(ApiEndpoints.towersPastBookings('TAGA001'), contains('/towers/bookings/past?bookerId=TAGA001'));
      expect(ApiEndpoints.towersPastBookings('TAGA001', '2025'), contains('year=2025'));
      expect(ApiEndpoints.towersPastBookings('TAGA001', '2025', '10'), contains('month=10'));
      expect(ApiEndpoints.towersCancelBooking('BK1001'), contains('/towers/bookings/BK1001'));
      expect(ApiEndpoints.towersCreateOrder, contains('/towers/create-order'));
      expect(ApiEndpoints.towersVerifyPayment, contains('/towers/verify-payment'));
      expect(ApiEndpoints.towersBannerImage, contains('/images/tagatower_final.png'));
    });
  });

  group('TAGA Towers Domain Models', () {
    test('Room parsing, single bed allowance, and advance calculation', () {
      const room = Room(
        id: 'apex-1',
        name: 'Apex Suite A/C',
        type: 'apex-suite',
        capacity: 3,
        allowSingleBed: true,
      );

      final json = room.toJson();
      final parsed = Room.fromJson(json);

      expect(parsed.id, 'apex-1');
      expect(parsed.name, 'Apex Suite A/C');
      expect(parsed.type, 'apex-suite');
      expect(parsed.capacity, 3);
      expect(parsed.allowSingleBed, isTrue);
    });

    test('RoomAvailability parsing', () {
      final json = {
        'room': {
          'id': 'gents-dorm',
          'name': 'Gents Dormitory',
          'type': 'gents-dorm',
          'capacity': 12,
          'allowSingleBed': true,
        },
        'available': true,
        'availableBeds': 5,
        'genderRestriction': 'male',
      };

      final avail = RoomAvailability.fromJson(json);
      expect(avail.room.id, 'gents-dorm');
      expect(avail.available, isTrue);
      expect(avail.availableBeds, 5);
      expect(avail.genderRestriction, 'male');
    });

    test('BookingResponse parsing and cancellation eligibility', () {
      final now = DateTime.now();
      final futureCheckIn = '${now.year + 1}-06-15';
      final futureCheckOut = '${now.year + 1}-06-16';

      final booking = BookingResponse(
        id: 'BK999',
        roomId: 'pavalam',
        roomName: 'Pavalam',
        checkInDate: futureCheckIn,
        checkOutDate: futureCheckOut,
        bookerName: 'Test Member',
        bookerId: 'TAGA001',
        bookerPhone: '9876543210',
        bookingFor: 'self',
        bedCount: 1,
        paymentStatus: 'confirmed',
        advanceAmount: 200,
        bookingStatus: 'upcoming',
      );

      expect(booking.isCancellable, isTrue);
      expect(booking.advanceAmount, 200);

      final cancelledBooking = BookingResponse(
        id: 'BK998',
        roomId: 'pavalam',
        roomName: 'Pavalam',
        checkInDate: futureCheckIn,
        checkOutDate: futureCheckOut,
        bookerName: 'Test Member',
        bookerId: 'TAGA001',
        bookingFor: 'self',
        bedCount: 1,
        paymentStatus: 'cancelled',
        advanceAmount: 200,
        bookingStatus: 'cancelled',
      );
      expect(cancelledBooking.isCancellable, isFalse);
    });

    test('TowersCreateOrderResponse and VerifyTowersPaymentRequest', () {
      final orderJson = {
        'key': 'rzp_test_123',
        'order': {
          'id': 'order_123456',
          'amount': 20000,
          'currency': 'INR',
        },
      };

      final orderRes = TowersCreateOrderResponse.fromJson(orderJson);
      expect(orderRes.key, 'rzp_test_123');
      expect(orderRes.orderId, 'order_123456');
      expect(orderRes.amount, 20000);
      expect(orderRes.amountInRupees, 200.0);
      expect(orderRes.currency, 'INR');

      const verifyReq = VerifyTowersPaymentRequest(
        bookingId: 'BK123',
        razorpayOrderId: 'order_123456',
        razorpayPaymentId: 'pay_7890',
        razorpaySignature: 'sig_abc',
      );
      final verifyJson = verifyReq.toJson();
      expect(verifyJson['bookingId'], 'BK123');
      expect(verifyJson['razorpay_order_id'], 'order_123456');
      expect(verifyJson['razorpay_payment_id'], 'pay_7890');
      expect(verifyJson['razorpay_signature'], 'sig_abc');
    });

    test('DateRangeParams formatting and night calculation', () {
      final checkIn = DateTime(2026, 11, 10);
      final checkOut = DateTime(2026, 11, 12);
      final params = DateRangeParams(checkIn: checkIn, checkOut: checkOut);

      expect(params.checkInFormatted, '2026-11-10');
      expect(params.checkOutFormatted, '2026-11-12');
      expect(params.totalNights, 2);
    });
  });

  group('TagaTowersRepository API tests', () {
    late Dio mockDio;
    late TagaTowersRepository repository;

    setUp(() {
      mockDio = Dio();
      mockDio.interceptors.add(MockDioInterceptor());
      repository = TagaTowersRepository(dio: mockDio);
    });

    test('getRooms filters out hidden rooms (kurinchi, pasumai, and hide: true)', () async {
      final rooms = await repository.getRooms();
      expect(rooms.length, 3);
      expect(rooms.any((r) => r.id == 'apex-1'), isTrue);
      expect(rooms.any((r) => r.id == 'pavalam'), isTrue);
      expect(rooms.any((r) => r.id == 'gents-dorm'), isTrue);
      expect(rooms.any((r) => r.id == 'kurinchi'), isFalse);
      expect(rooms.any((r) => r.id == 'pasumai'), isFalse);
      expect(rooms.any((r) => r.id == 'hidden-room'), isFalse);
    });

    test('checkAvailabilityRange returns map of room availability', () async {
      final availMap = await repository.checkAvailabilityRange('2026-12-01', '2026-12-02');
      expect(availMap.containsKey('apex-1'), isTrue);
      expect(availMap['apex-1']!.available, isTrue);
      expect(availMap['apex-1']!.availableBeds, 3);
      expect(availMap['gents-dorm']!.genderRestriction, 'male');
    });

    test('createBooking returns BookingResponse', () async {
      const req = CreateBookingRequest(
        roomId: 'apex-1',
        checkInDate: '2026-12-01',
        checkOutDate: '2026-12-02',
        bookerPhone: '9876543210',
        bookingFor: 'self',
        bedCount: 1,
      );

      final booking = await repository.createBooking(req, 'TAGA001');
      expect(booking.id, 'BK1001');
      expect(booking.roomId, 'apex-1');
      expect(booking.paymentStatus, 'pending');
    });

    test('getUserBookings and getPastBookings return bookings list', () async {
      final active = await repository.getUserBookings('TAGA001');
      expect(active.length, 1);
      expect(active.first.id, 'BK1001');

      final past = await repository.getPastBookings('TAGA001', '2025');
      expect(past.length, 1);
      expect(past.first.id, 'BK1002');
      expect(past.first.bookingStatus, 'completed');
    });

    test('cancelBooking sends DELETE request', () async {
      final success = await repository.cancelBooking('BK1001');
      expect(success, isTrue);
    });

    test('createOrder and verifyPayment execute successfully', () async {
      final orderRes = await repository.createOrder(
        amountInPaise: 20000,
        notes: {'booking_id': 'BK1001'},
      );
      expect(orderRes.key, 'rzp_test_123');
      expect(orderRes.orderId, 'order_998877');

      const verifyReq = VerifyTowersPaymentRequest(
        bookingId: 'BK1001',
        razorpayOrderId: 'order_998877',
        razorpayPaymentId: 'pay_123',
        razorpaySignature: 'sig_123',
      );
      final verified = await repository.verifyPayment(verifyReq);
      expect(verified, isTrue);
    });
  });

  group('BookingFlowNotifier Riverpod tests', () {
    test('State initial values and room selection updates', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      final notifier = container.read(bookingFlowNotifierProvider.notifier);
      var state = container.read(bookingFlowNotifierProvider);

      expect(state.bookingFor, 'self');
      expect(state.bedCount, 1);

      const testRoom = Room(
        id: 'apex-1',
        name: 'Apex Suite',
        type: 'apex-suite',
        capacity: 3,
        allowSingleBed: true,
      );

      notifier.selectRoom(testRoom);
      state = container.read(bookingFlowNotifierProvider);

      expect(state.selectedRoom?.id, 'apex-1');
      expect(state.advanceAmount, 200); // 1 bed * 200

      // Change bed count
      notifier.setBedCount(2);
      state = container.read(bookingFlowNotifierProvider);
      expect(state.bedCount, 2);
      expect(state.guestDetails.length, 2);

      // Set booking target to guest
      notifier.setBookingFor('guest');
      state = container.read(bookingFlowNotifierProvider);
      expect(state.bookingFor, 'guest');

      // Reset
      notifier.reset();
      state = container.read(bookingFlowNotifierProvider);
      expect(state.selectedRoom, isNull);
    });
  });
}
