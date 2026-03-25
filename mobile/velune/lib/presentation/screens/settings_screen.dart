import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../application/providers/settings_provider.dart';
import '../widgets/error_banner.dart';
import '../widgets/loading_skeleton.dart';

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final settingsAsync = ref.watch(settingsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Settings'),
      ),
      body: settingsAsync.when(
        loading: () => const Center(
          child: LoadingSkeleton(width: 220, height: 120),
        ),
        error: (err, st) => Center(
          child: ErrorBanner(
            title: 'Unable to load settings',
            message: err.toString(),
            onRetry: () {
              ref.refresh(settingsProvider);
            },
          ),
        ),
        data: (settings) {
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text(
                'Signed in as',
                style: Theme.of(context).textTheme.labelLarge,
              ),
              const SizedBox(height: 8),
              Text(
                settings.user.email,
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              const SizedBox(height: 16),
              ListTile(
                title: const Text('Base currency'),
                subtitle: Text(settings.currency),
                leading: const Icon(Icons.currency_exchange),
              ),
              ListTile(
                title: const Text('Theme mode'),
                subtitle: Text(settings.themeMode),
                leading: const Icon(Icons.color_lens),
              ),
              const SizedBox(height: 16),
              FilledButton.tonal(
                onPressed: () async {
                  await ref.read(settingsProvider.notifier).logout();
                },
                child: const Text('Log out'),
              ),
            ],
          );
        },
      ),
    );
  }
}

