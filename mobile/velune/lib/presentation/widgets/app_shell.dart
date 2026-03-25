import 'package:flutter/material.dart';

class AppShellDestination {
  final int index;
  final String label;
  final IconData icon;

  const AppShellDestination({
    required this.index,
    required this.label,
    required this.icon,
  });
}

class AppShell extends StatelessWidget {
  final Widget child;
  final List<AppShellDestination> destinations;
  final int currentIndex;
  final void Function(int index) onDestinationSelected;
  final AppBar? appBar;

  const AppShell({
    super.key,
    required this.child,
    required this.destinations,
    required this.currentIndex,
    required this.onDestinationSelected,
    this.appBar,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: appBar,
      body: child,
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: currentIndex,
        onTap: onDestinationSelected,
        type: BottomNavigationBarType.fixed,
        items: destinations
            .map(
              (d) => BottomNavigationBarItem(
                icon: Icon(d.icon),
                label: d.label,
              ),
            )
            .toList(),
      ),
    );
  }
}

