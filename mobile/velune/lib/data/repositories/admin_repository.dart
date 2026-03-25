import 'dart:async';

/// Admin features are intentionally not wired into mobile UI routes
/// for security/policy reasons.
abstract class AdminRepository {
  Future<void> health();
}

class MobileAdminRepository implements AdminRepository {
  @override
  Future<void> health() {
    throw UnimplementedError(
      'Admin APIs are not exposed in the mobile app build.',
    );
  }
}

