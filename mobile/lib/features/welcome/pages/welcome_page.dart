import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/providers/app_providers.dart';

const _pink = Color(0xFFEE6F8E);

class WelcomePage extends ConsumerWidget {
  const WelcomePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) => Scaffold(
    backgroundColor: const Color(0xFFFFF5F6),
    body: SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(24, 28, 24, 24),
        child: Column(
          children: [
            const Spacer(),
            Container(
              width: 220,
              height: 220,
              decoration: const BoxDecoration(
                shape: BoxShape.circle,
                color: Color(0xFFFFE3E8),
              ),
              child: Stack(
                alignment: Alignment.center,
                children: [
                  const Icon(
                    Icons.savings_rounded,
                    size: 142,
                    color: Color(0xFFF28AA0),
                  ),
                  Positioned(
                    top: 18,
                    right: 42,
                    child: Icon(
                      Icons.monetization_on,
                      size: 37,
                      color: Colors.amber.shade600,
                    ),
                  ),
                  const Positioned(
                    left: 23,
                    top: 55,
                    child: Icon(
                      Icons.favorite,
                      size: 24,
                      color: Color(0xFFF4A3B4),
                    ),
                  ),
                  const Positioned(
                    right: 20,
                    bottom: 48,
                    child: Icon(
                      Icons.cloud,
                      size: 32,
                      color: Color(0xFFF7B7C4),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 34),
            Text(
              'Quản lý chi tiêu\nthông minh hơn',
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.headlineSmall
                  ?.copyWith(fontWeight: FontWeight.w800, height: 1.2),
            ),
            const SizedBox(height: 13),
            const Text(
              'Chụp ảnh, AI tự động ghi nhận\nchi tiêu cho bạn',
              textAlign: TextAlign.center,
              style: TextStyle(color: Color(0xFF6F6870), height: 1.5),
            ),
            const SizedBox(height: 24),
            const Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                CircleAvatar(radius: 3, backgroundColor: _pink),
                SizedBox(width: 7),
                CircleAvatar(radius: 3, backgroundColor: Color(0xFFF6C4CE)),
                SizedBox(width: 7),
                CircleAvatar(radius: 3, backgroundColor: Color(0xFFF6C4CE)),
              ],
            ),
            const Spacer(),
            SizedBox(
              width: double.infinity,
              child: FilledButton(
                onPressed: () async {
                  await ref.read(appLaunchStorageProvider).markWelcomeSeen();
                  if (context.mounted) context.go('/login');
                },
                style: FilledButton.styleFrom(
                  backgroundColor: _pink,
                  minimumSize: const Size.fromHeight(50),
                  shape: const StadiumBorder(),
                ),
                child: const Text(
                  'Bắt đầu',
                  style: TextStyle(fontWeight: FontWeight.w700),
                ),
              ),
            ),
            TextButton(
              onPressed: () async {
                await ref.read(appLaunchStorageProvider).markWelcomeSeen();
                if (context.mounted) context.go('/login');
              },
              child: const Text(
                'Tôi đã có tài khoản',
                style: TextStyle(color: _pink, fontSize: 12),
              ),
            ),
          ],
        ),
      ),
    ),
  );
}
