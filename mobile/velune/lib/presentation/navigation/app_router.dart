import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../application/providers/auth_session_provider.dart';
import '../../application/providers/core_providers.dart';
import '../screens/boot_screen.dart';
import '../screens/budgets_screen.dart';
import '../screens/dashboard_screen.dart';
import '../screens/login_screen.dart';
import '../screens/notifications_screen.dart';
import '../screens/reports_screen.dart';
import '../screens/register_screen.dart';
import '../screens/settings_screen.dart';
import '../screens/transactions_screen.dart';
import '../screens/transaction_detail_screen.dart';

final goRouterProvider = Provider<GoRouter>((ref) {
  final authAsync = ref.watch(authSessionProvider);

  final authState = authAsync.valueOrNull;
  final isAuthed =
      authState != null && authState.status == AuthSessionStatus.authenticated && authState.user != null;

  bool isPublicLocation(String location) =>
      location == '/login' || location == '/register';

  return GoRouter(
    initialLocation: '/',
    redirect: (context, state) {
      // While we don't know the auth state yet, show the boot screen.
      if (authAsync.isLoading) return null;

      final location = state.matchedLocation;

      if (isAuthed) {
        if (location == '/') return '/dashboard';
        if (isPublicLocation(location)) return '/dashboard';
        return null;
      }

      // Not authenticated: force login for all protected routes.
      if (location == '/') return '/login';
      if (!isPublicLocation(location)) return '/login';
      return null;
    },
    routes: [
      GoRoute(
        path: '/',
        builder: (context, state) => const BootScreen(),
      ),
      GoRoute(
        path: '/login',
        builder: (context, state) => const LoginScreen(),
      ),
      GoRoute(
        path: '/register',
        builder: (context, state) => const RegisterScreen(),
      ),
      GoRoute(
        path: '/dashboard',
        builder: (context, state) => const DashboardScreen(),
      ),
      GoRoute(
        path: '/transactions',
        builder: (context, state) => const TransactionsScreen(),
      ),
      GoRoute(
        path: '/transactions/:id',
        builder: (context, state) {
          final id = state.pathParameters['id'] ?? '';
          return TransactionDetailScreen(transactionId: id);
        },
      ),
      GoRoute(
        path: '/budgets',
        builder: (context, state) => const BudgetsScreen(),
      ),
      GoRoute(
        path: '/reports',
        builder: (context, state) => const ReportsScreen(),
      ),
      GoRoute(
        path: '/notifications',
        builder: (context, state) =>
            const NotificationsScreen(),
      ),
      GoRoute(
        path: '/settings',
        builder: (context, state) => const SettingsScreen(),
      ),
    ],
  );
});

