import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../application/providers/auth_session_provider.dart';
import '../../core/http/app_error.dart';
import '../widgets/error_banner.dart';

class RegisterScreen extends ConsumerStatefulWidget {
  const RegisterScreen({super.key});

  @override
  ConsumerState<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends ConsumerState<RegisterScreen> {
  final _formKey = GlobalKey<FormState>();

  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _baseCurrencyController = TextEditingController();

  AppError? _error;

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    _baseCurrencyController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Register'),
      ),
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Form(
              key: _formKey,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  if (_error != null)
                    ErrorBanner(
                      title: 'Registration failed',
                      message: _error!.message,
                      onRetry: () => setState(() => _error = null),
                    ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _emailController,
                    decoration: const InputDecoration(labelText: 'Email'),
                    keyboardType: TextInputType.emailAddress,
                    validator: (v) {
                      final value = v?.trim() ?? '';
                      if (value.isEmpty) return 'Email is required.';
                      if (!value.contains('@')) return 'Invalid email.';
                      return null;
                    },
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _passwordController,
                    decoration: const InputDecoration(labelText: 'Password'),
                    obscureText: true,
                    validator: (v) {
                      final value = v ?? '';
                      if (value.isEmpty) return 'Password is required.';
                      if (value.length < 6) return 'Password too short.';
                      return null;
                    },
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _baseCurrencyController,
                    decoration: const InputDecoration(
                      labelText: 'Base Currency (e.g. USD)',
                    ),
                    validator: (v) {
                      final value = (v ?? '').trim();
                      if (value.isEmpty) return 'Base currency is required.';
                      if (value.length != 3) return 'Use a 3-letter currency code.';
                      return null;
                    },
                  ),
                  const SizedBox(height: 18),
                  FilledButton(
                    onPressed: () async {
                      final ok = _formKey.currentState?.validate() ?? false;
                      if (!ok) return;

                      setState(() => _error = null);

                      try {
                        await ref
                            .read(authSessionProvider.notifier)
                            .register(
                              email: _emailController.text.trim(),
                              password: _passwordController.text,
                              baseCurrency:
                                  _baseCurrencyController.text.trim().toUpperCase(),
                            );
                      } catch (e) {
                        final maybeAppError = e is AppError
                            ? e
                            : const AppError(
                                code: 'auth_error',
                                message: 'Unable to register.',
                                statusCode: 400,
                              );
                        setState(() => _error = maybeAppError);
                      }
                    },
                    child: const Text('Create account'),
                  ),
                  const SizedBox(height: 12),
                  TextButton(
                    onPressed: () => context.go('/login'),
                    child: const Text('Back to login'),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

