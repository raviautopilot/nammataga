/// Response model returned by POST /api/towers/create-order
class TowersCreateOrderResponse {
  final String key;
  final String orderId;
  final int amount; // in paise
  final String currency;
  final Map<String, dynamic> rawOrder;

  const TowersCreateOrderResponse({
    required this.key,
    required this.orderId,
    required this.amount,
    required this.currency,
    this.rawOrder = const {},
  });

  /// Amount in Rupees (converted from paise)
  double get amountInRupees => amount / 100.0;

  factory TowersCreateOrderResponse.fromJson(Map<String, dynamic> json) {
    final key = json['key'] as String? ?? '';
    final orderMap = json['order'] is Map
        ? Map<String, dynamic>.from(json['order'] as Map)
        : <String, dynamic>{};

    final orderId = orderMap['id'] as String? ?? json['orderId'] as String? ?? '';
    final amount = (orderMap['amount'] as num?)?.toInt() ??
        (json['amount'] as num?)?.toInt() ??
        0;
    final currency = orderMap['currency'] as String? ??
        json['currency'] as String? ??
        'INR';

    return TowersCreateOrderResponse(
      key: key,
      orderId: orderId,
      amount: amount,
      currency: currency,
      rawOrder: orderMap,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'key': key,
      'orderId': orderId,
      'amount': amount,
      'currency': currency,
      'order': rawOrder,
    };
  }

  @override
  String toString() =>
      'TowersCreateOrderResponse(orderId: $orderId, amount: $amount $currency)';
}

/// Verification request payload sent to POST /api/towers/verify-payment
class VerifyTowersPaymentRequest {
  final String bookingId;
  final String razorpayOrderId;
  final String razorpayPaymentId;
  final String razorpaySignature;

  const VerifyTowersPaymentRequest({
    required this.bookingId,
    required this.razorpayOrderId,
    required this.razorpayPaymentId,
    required this.razorpaySignature,
  });

  Map<String, dynamic> toJson() {
    return {
      'bookingId': bookingId,
      'razorpay_order_id': razorpayOrderId,
      'razorpay_payment_id': razorpayPaymentId,
      'razorpay_signature': razorpaySignature,
    };
  }

  factory VerifyTowersPaymentRequest.fromJson(Map<String, dynamic> json) {
    return VerifyTowersPaymentRequest(
      bookingId: json['bookingId'] as String? ?? '',
      razorpayOrderId: json['razorpay_order_id'] as String? ?? '',
      razorpayPaymentId: json['razorpay_payment_id'] as String? ?? '',
      razorpaySignature: json['razorpay_signature'] as String? ?? '',
    );
  }

  @override
  String toString() =>
      'VerifyTowersPaymentRequest(bookingId: $bookingId, orderId: $razorpayOrderId, paymentId: $razorpayPaymentId)';
}
