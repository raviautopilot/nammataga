import 'package:flutter/foundation.dart';
import 'package:razorpay_flutter/razorpay_flutter.dart';
import 'razorpay_checkout_stub.dart'
    if (dart.library.js_interop) 'razorpay_checkout_web.dart' as web_checkout;

/// Unified Razorpay checkout manager supporting both Web and Native Mobile platforms.
class RazorpayCheckoutManager {
  Razorpay? _nativeRazorpay;

  /// Opens the Razorpay payment gateway
  void openCheckout({
    required Map<String, dynamic> options,
    required void Function(String paymentId, String orderId, String signature)
    onSuccess,
    required void Function(String error) onError,
    required void Function() onDismiss,
  }) {
    if (kIsWeb) {
      // On Web platform: opens Razorpay Standard Checkout overlay
      try {
        web_checkout.openWebRazorpay(
          options: options,
          onSuccess: onSuccess,
          onError: onError,
          onDismiss: onDismiss,
        );
      } catch (e) {
        onError('Failed to open web checkout: $e');
      }
    } else {
      // On Native Mobile (Android/iOS): uses Razorpay SDK
      try {
        _nativeRazorpay?.clear();
        _nativeRazorpay = Razorpay();

        _nativeRazorpay!.on(
          Razorpay.EVENT_PAYMENT_SUCCESS,
          (PaymentSuccessResponse response) {
            onSuccess(
              response.paymentId ?? '',
              response.orderId ?? '',
              response.signature ?? '',
            );
          },
        );

        _nativeRazorpay!.on(
          Razorpay.EVENT_PAYMENT_ERROR,
          (PaymentFailureResponse response) {
            if (response.code == Razorpay.PAYMENT_CANCELLED) {
              onDismiss();
            } else {
              onError(response.message ?? 'Payment failed');
            }
          },
        );

        _nativeRazorpay!.on(
          Razorpay.EVENT_EXTERNAL_WALLET,
          (ExternalWalletResponse response) {
            onSuccess(response.walletName ?? '', '', '');
          },
        );

        _nativeRazorpay!.open(options);
      } catch (e) {
        onError('Failed to initialize native payment: $e');
      }
    }
  }

  void dispose() {
    if (!kIsWeb) {
      _nativeRazorpay?.clear();
    }
  }
}
