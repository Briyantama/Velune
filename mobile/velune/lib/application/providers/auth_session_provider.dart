import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/http/app_error.dart';
import '../../domain/models/user.dart';
import 'core_providers.dart';

enum AuthSessionStatus {
  loading,
  authenticated,
  unauthenticated,
}

class AuthSessionState {
  final AuthSessionStatus status;
  final User? user;
  final AppError? error;

  const AuthSessionState({
    required this.status,
    this.user,
    this.error,
  });

  const AuthSessionState.loading()
      : status = AuthSessionStatus.loading,
        user = null,
        error = null;

  const AuthSessionState.authenticated(User user)
      : status = AuthSessionStatus.authenticated,
        user = user,
        error = null;

  const AuthSessionState.unauthenticated([AppError? error])
      : status = AuthSessionStatus.unauthenticated,
        user = null,
        error = error;
}

class AuthSessionController extends AsyncNotifier<AuthSessionState> {
  @override
  Future<AuthSessionState> build() async {
    final tokenStore = ref.read(tokenStoreProvider);
    final authRepo = ref.read(authRepositoryProvider);

    final refreshToken = await tokenStore.readRefreshToken();
    if (refreshToken == null || refreshToken.isEmpty) {
      return const AuthSessionState.unauthenticated();
    }

    try {
      final user = await authRepo.me();
      return AuthSessionState.authenticated(user);
    } on Object catch (e) {
      final maybeAppError = e is AppError
          ? e
          : (e is Exception && e.toString().contains('auth_expired')
              ? const AppError(
                  code: 'auth_expired',
                  message: 'Session expired.',
                  statusCode: 401,
                )
              : null);

      if (maybeAppError?.code == 'auth_expired') {
        await tokenStore.clearTokens();
        return AuthSessionState.unauthenticated(maybeAppError);
      }

      return AuthSessionState.unauthenticated(maybeAppError);
    }
  }

  Future<void> login({required String email, required String password}) async {
    final authRepo = ref.read(authRepositoryProvider);
    await authRepo.login(email: email, password: password);
    ref.invalidateSelf();
  }

  Future<void> register({
    required String email,
    required String password,
    required String baseCurrency,
  }) async {
    final authRepo = ref.read(authRepositoryProvider);
    await authRepo.register(
      email: email,
      password: password,
      baseCurrency: baseCurrency,
    );
    ref.invalidateSelf();
  }

  Future<void> refresh() async {
    final authRepo = ref.read(authRepositoryProvider);
    await authRepo.refresh();
    ref.invalidateSelf();
  }

  Future<void> logout() async {
    final authRepo = ref.read(authRepositoryProvider);
    await authRepo.logout();
    state = const AsyncValue.data(AuthSessionState.unauthenticated());
  }
}

final authSessionProvider =
    AsyncNotifierProvider<AuthSessionController, AuthSessionState>(
  AuthSessionController.new,
);

