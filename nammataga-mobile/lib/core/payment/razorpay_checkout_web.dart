import 'dart:convert';
import 'dart:js_interop';

@JS('openRazorpayWebCheckout')
external void _openRazorpayWebCheckout(
  JSString optionsJson,
  JSFunction onSuccess,
  JSFunction onError,
  JSFunction onDismiss,
);

void openWebRazorpay({
  required Map<String, dynamic> options,
  required void Function(String paymentId, String orderId, String signature)
  onSuccess,
  required void Function(String error) onError,
  required void Function() onDismiss,
}) {
  final jsonStr = jsonEncode(options).toJS;

  final successCb =
      (JSString paymentId, JSString orderId, JSString signature) {
        onSuccess(
          paymentId.toDart,
          orderId.toDart,
          signature.toDart,
        );
      }.toJS;

  final errorCb =
      (JSString error) {
        onError(error.toDart);
      }.toJS;

  final dismissCb =
      () {
        onDismiss();
      }.toJS;

  _openRazorpayWebCheckout(jsonStr, successCb, errorCb, dismissCb);
}
