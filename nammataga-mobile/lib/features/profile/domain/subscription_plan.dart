library;

/// Represents a subscription plan or contribution option from the backend
class SubscriptionPlan {
  final String id;
  final String name;
  final String description;
  final int? amount; // Amount in Rupees for fixed plans
  final String frequency; // 'Annual' or 'One-time'
  final String status;
  final String? nextDueDate;
  final String? validUntil;
  final bool allowCustomAmount;
  final bool oneTime;
  final bool needBased;

  const SubscriptionPlan({
    required this.id,
    required this.name,
    required this.description,
    this.amount,
    required this.frequency,
    this.status = 'active',
    this.nextDueDate,
    this.validUntil,
    this.allowCustomAmount = false,
    this.oneTime = false,
    this.needBased = false,
  });

  /// Formatted display price
  String get displayAmount {
    if (allowCustomAmount || amount == null) {
      return 'Flexible Amount';
    }
    return '₹${amount!.toString().replaceAllMapped(RegExp(r'(\d{1,3})(?=(\d{3})+(?!\d))'), (m) => '${m[1]},')}';
  }

  factory SubscriptionPlan.fromJson(Map<String, dynamic> json) {
    int? parsedAmount;
    if (json['amount'] != null) {
      final a = json['amount'];
      if (a is int) {
        parsedAmount = a;
      } else if (a is num) {
        parsedAmount = a.toInt();
      } else {
        parsedAmount = int.tryParse(a.toString());
      }
    }

    return SubscriptionPlan(
      id: (json['id'] ?? '').toString(),
      name: (json['name'] ?? '').toString(),
      description: (json['description'] ?? '').toString(),
      amount: parsedAmount,
      frequency: (json['frequency'] ?? 'Annual').toString(),
      status: (json['status'] ?? 'active').toString(),
      nextDueDate: json['nextDueDate']?.toString(),
      validUntil: json['validUntil']?.toString(),
      allowCustomAmount: json['allowCustomAmount'] == true,
      oneTime: json['oneTime'] == true,
      needBased: json['needBased'] == true,
    );
  }
}

/// Represents the created Razorpay order response from `POST /api/subscriptions/create-order`
class RazorpayOrder {
  final String orderId;
  final String key;
  final int amount; // Amount in paise
  final String currency;

  const RazorpayOrder({
    required this.orderId,
    required this.key,
    required this.amount,
    required this.currency,
  });

  factory RazorpayOrder.fromJson(Map<String, dynamic> json) {
    return RazorpayOrder(
      orderId: (json['orderId'] ?? json['order_id'] ?? '').toString(),
      key: (json['key'] ?? '').toString(),
      amount: (json['amount'] is num) ? (json['amount'] as num).toInt() : 0,
      currency: (json['currency'] ?? 'INR').toString(),
    );
  }
}
