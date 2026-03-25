import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../application/providers/notifications_provider.dart';
import '../widgets/empty_state.dart';
import '../widgets/error_banner.dart';
import '../widgets/loading_skeleton.dart';

class NotificationsScreen extends ConsumerWidget {
  const NotificationsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final pingAsync = ref.watch(notificationsPingProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Notifications'),
      ),
      body: pingAsync.when(
        loading: () => const Center(
          child: LoadingSkeleton(width: 220, height: 60),
        ),
        error: (err, st) => Center(
          child: ErrorBanner(
            title: 'Notification service offline',
            message: err.toString(),
            onRetry: () {
              ref.refresh(notificationsPingProvider);
            },
          ),
        ),
        data: (ping) {
          if (ping.status.isEmpty) {
            return const EmptyState(
              title: 'No notifications',
              message: 'Your notification service did not return a status.',
            );
          }

          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              ListTile(
                leading: const Icon(Icons.notifications_active),
                title: const Text('Connection status'),
                subtitle: Text('Service: ${ping.status}'),
              ),
              const SizedBox(height: 12),
              const EmptyState(
                title: 'No notification feed yet',
                message:
                    'This app currently only pings the notification service. A full inbox will be added next.',
              ),
            ],
          );
        },
      ),
    );
  }
}

