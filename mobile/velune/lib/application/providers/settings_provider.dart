import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/http/app_error.dart';
import '../../domain/models/user.dart';
import 'auth_session_provider.dart';

class SettingsState {
  final User user;
  final String themeMode;
  final String currency;

  const SettingsState({
    required this.user,
    required this.themeMode,
    required this.currency,
  });
}

/// Theme preference is kept client-side for now.
final themeModePreferenceProvider = StateProvider<String>((ref) => 'system');

/// Currency preference is kept client-side for now.
final currencyPreferenceProvider = StateProvider<String>((ref) => '');

class SettingsController extends AsyncNotifier<SettingsState> {
  @override
  Future<SettingsState> build() async {
    final auth = await ref.watch(authSessionProvider.future);
    if (auth.status != AuthSessionStatus.authenticated || auth.user == null) {
      throw const AppError(
        code: 'unauthenticated',
        message: 'Please log in.',
        statusCode: 401,
      );
    }

    final user = auth.user!;
    final themeMode = ref.watch(themeModePreferenceProvider);
    final currencyPref = ref.read(currencyPreferenceProvider);
    final currency = currencyPref.isEmpty ? user.baseCurrency : currencyPref;

    if (currencyPref.isEmpty) {
      ref.read(currencyPreferenceProvider.notifier).state = currency;
    }

    return SettingsState(
      user: user,
      themeMode: themeMode,
      currency: currency,
    );
  }

  Future<void> logout() async {
    await ref.read(authSessionProvider.notifier).logout();
  }
}

final settingsProvider =
    AsyncNotifierProvider<SettingsController, SettingsState>(
  SettingsController.new,
);

